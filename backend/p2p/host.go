package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"projekt-redware/backend/p2p/proto"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	google_proto "google.golang.org/protobuf/proto"
)

type HistoryProvider interface {
	GetMessagesForSync(channelID string, afterTimestamp int64) ([]*proto.ChatMessage, error)
}

type EventHandler interface {
	OnChatMessage(msg *proto.ChatMessage)
}

type MessageHandler interface {
	HandleMessage(msg *proto.ChatMessage)
}

type Peer struct {
	UserID   string
	Username string
}

type P2PHost struct {
	Host            host.Host
	Ctx             context.Context
	DHT             *dht.IpfsDHT
	PubSub          *pubsub.PubSub
	Discovery       mdns.Service
	HistoryProvider HistoryProvider
	EventHandler    EventHandler
	MessageHandler  MessageHandler
}

func NewP2PHost(ctx context.Context, port int, eventHandler EventHandler) (*P2PHost, error) {
	priv, _, err := getIdentity()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.Identity(priv),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create GossipSub: %w", err)
	}

	notifee := &discoveryNotifee{h: h}
	ser := mdns.NewMdnsService(h, "redware-discovery", notifee)
	if err := ser.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mDNS service: %w", err)
	}

	log.Printf("Libp2p Host created with ID: %s", h.ID())
	log.Printf("Listening on: %v", h.Addrs())

	p2pHost := &P2PHost{
		Host:         h,
		Ctx:          ctx,
		DHT:          kademliaDHT,
		PubSub:       ps,
		Discovery:    ser,
		EventHandler: eventHandler,
	}

	h.SetStreamHandler("/redware/history/1.0.0", p2pHost.HandleHistoryStream)
	h.SetStreamHandler("/redware/chat/1.0.0", p2pHost.HandleChatStream)

	return p2pHost, nil
}

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Printf("Discovered new peer %s", pi.ID.String())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		log.Printf("Failed to connect to peer %s: %v", pi.ID.String(), err)
	}
}

func getIdentity() (crypto.PrivKey, crypto.PubKey, error) {
	keyPath := "identity.key"
	if _, err := os.Stat(keyPath); err == nil {
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, nil, err
		}
		priv, err := crypto.UnmarshalPrivateKey(data)
		if err != nil {
			return nil, nil, err
		}
		return priv, priv.GetPublic(), nil
	}

	log.Println("Generating new identity...")
	priv, pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	data, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		return nil, nil, err
	}

	return priv, pub, nil
}

func (p *P2PHost) Close() {
	if p.Host != nil {
		p.Host.Close()
	}
}

func (p *P2PHost) FindPeer(ctx context.Context, peerIDStr string) (peer.AddrInfo, error) {
	id, err := peer.Decode(peerIDStr)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("invalid peer ID: %w", err)
	}

	return p.DHT.FindPeer(ctx, id)
}

func (p *P2PHost) JoinTopic(topicName string) (*pubsub.Topic, error) {
	return p.PubSub.Join(topicName)
}

func (p *P2PHost) Publish(topic *pubsub.Topic, data []byte) error {
	return topic.Publish(p.Ctx, data)
}

func (p *P2PHost) Subscribe(topic *pubsub.Topic) (*pubsub.Subscription, error) {
	return topic.Subscribe()
}

func (p *P2PHost) HandleHistoryStream(s network.Stream) {
	defer s.Close()

	buf := make([]byte, 4096)
	n, err := s.Read(buf)
	if err != nil {
		log.Printf("Failed to read history request: %v", err)
		return
	}

	var req proto.HistoryFetchRequest
	if err := google_proto.Unmarshal(buf[:n], &req); err != nil {
		log.Printf("Failed to unmarshal history request: %v", err)
		return
	}

	log.Printf("Received HistoryFetchRequest for channel %s from %s", req.ChannelId, req.RequesterId)

	if p.HistoryProvider == nil {
		log.Printf("HistoryProvider is nil")
		return
	}

	messages, err := p.HistoryProvider.GetMessagesForSync(req.ChannelId, req.LastMessageTimestamp)
	if err != nil {
		log.Printf("Failed to get messages for sync: %v", err)
		return
	}

	resp := &proto.HistoryFetchResponse{
		Messages: messages,
	}

	data, err := google_proto.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal history response: %v", err)
		return
	}

	if _, err := s.Write(data); err != nil {
		log.Printf("Failed to write history response: %v", err)
		return
	}
}

func (p *P2PHost) FetchHistory(ctx context.Context, providerID string, channelID string, lastTimestamp int64) ([]*proto.ChatMessage, error) {
	pid, err := peer.Decode(providerID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider ID: %w", err)
	}

	s, err := p.Host.NewStream(ctx, pid, "/redware/history/1.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to open history stream: %w", err)
	}
	defer s.Close()

	req := &proto.HistoryFetchRequest{
		RequesterId:          p.Host.ID().String(),
		ChannelId:            channelID,
		LastMessageTimestamp: lastTimestamp,
	}
	data, err := google_proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if _, err := s.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	buf := make([]byte, 1024*1024)
	n, err := s.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp proto.HistoryFetchResponse
	if err := google_proto.Unmarshal(buf[:n], &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Messages, nil
}

func (p *P2PHost) HandleChatStream(s network.Stream) {
	defer s.Close()

	buf := make([]byte, 4096)
	n, err := s.Read(buf)
	if err != nil {
		log.Printf("Failed to read chat message: %v", err)
		return
	}

	var msg proto.ChatMessage
	if err := google_proto.Unmarshal(buf[:n], &msg); err != nil {
		log.Printf("Failed to unmarshal chat message: %v", err)
		return
	}

	log.Printf("Received Direct Message from %s: %s", msg.SenderUsername, msg.Content)

	if p.EventHandler != nil {
		p.EventHandler.OnChatMessage(&msg)
	}
}

func (p *P2PHost) SendDirectMessage(ctx context.Context, peerID string, msg *proto.ChatMessage) error {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	s, err := p.Host.NewStream(ctx, pid, "/redware/chat/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open chat stream: %w", err)
	}
	defer s.Close()

	data, err := google_proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := s.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
