package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"projekt-redware/backend/p2p"
	"projekt-redware/backend/p2p/proto"
	p2p_proto "projekt-redware/backend/p2p/proto"
	"projekt-redware/backend/store"
	"strings"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	p2pHost       *p2p.P2PHost
	P2PPort       int
	UserID        string
	Username      string
	Status        string
	StatusMessage string
	store         *store.Store

	topics map[string]*pubsub.Topic
	subs   map[string]*pubsub.Subscription

	syncChans map[string]chan *p2p_proto.HistoryOffer
	syncLock  sync.Mutex

	peerPorts     map[string]int
	peerPortsLock sync.Mutex
}

func NewApp(port int) *App {
	return &App{
		P2PPort:       port,
		Status:        "online",
		StatusMessage: "",
		topics:        make(map[string]*pubsub.Topic),
		subs:          make(map[string]*pubsub.Subscription),
		syncChans:     make(map[string]chan *p2p_proto.HistoryOffer),
		peerPorts:     make(map[string]int),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if a.p2pHost == nil {
		host, err := p2p.NewP2PHost(ctx, a.P2PPort, a)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Failed to start Libp2p Host: %v", err))
		} else {
			a.p2pHost = host
			a.p2pHost.HistoryProvider = a
			a.UserID = host.Host.ID().String()
			a.EmitLog(fmt.Sprintf("Libp2p Host Started! ID: %s", a.UserID))
		}
	}
}

func (a *App) Login(username, password string) map[string]string {
	a.Username = username
	dbPath := fmt.Sprintf("redware_%s.db", username)
	appStore, err := store.NewStore(dbPath)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to initialize store: %v", err))
	} else {
		a.store = appStore
		a.EmitLog("Message Store initialized successfully")
	}

	return map[string]string{
		"token":    a.UserID,
		"avatar":   "",
		"user_id":  a.UserID,
		"username": username,
	}
}

func (a *App) Register(username, password string) bool {
	return true
}

func (a *App) UpdateAvatar(avatarURL string) error {
	return nil
}

func (a *App) SetStatus(status, message string) error {
	a.Status = status
	a.StatusMessage = message
	return nil
}

func (a *App) Log(message string) {
	fmt.Println("[Frontend Log]:", message)
}

func (a *App) EmitLog(msg string) {
	fmt.Println("[Backend Log]:", msg)
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "DEBUG_LOG", msg)
	}
}

func (a *App) StartP2P() bool {
	a.EmitLog(fmt.Sprintf("Starting Libp2p Host on port %d...", a.P2PPort))

	host, err := p2p.NewP2PHost(a.ctx, a.P2PPort, a)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to start Libp2p Host: %v", err))
		return false
	}
	a.p2pHost = host
	a.UserID = host.Host.ID().String()

	a.EmitLog(fmt.Sprintf("Libp2p Host Started! ID: %s", a.UserID))

	go func() {
		if a.store != nil {
			nodes, err := a.store.GetJoinedNodes()
			if err != nil {
				a.EmitLog(fmt.Sprintf("Failed to load joined nodes: %v", err))
			} else {
				a.EmitLog(fmt.Sprintf("Found %d joined nodes. Reconnecting...", len(nodes)))
				for _, node := range nodes {
					a.EmitLog(fmt.Sprintf("Reconnecting to %s...", node.Name))
				}
			}
		}
	}()

	return true
}

func (a *App) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
			case <-a.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (a *App) ConnectPeer(userID, address string) bool {
	if a.p2pHost == nil {
		return false
	}

	a.EmitLog(fmt.Sprintf("ConnectPeer called for %s (%s) - Libp2p handles this automatically", userID, address))
	return true
}

func (a *App) GetDiscoveredPeers() []p2p.PeerInfo {
	if a.p2pHost == nil {
		return []p2p.PeerInfo{}
	}

	var peers []p2p.PeerInfo
	allPeers := a.p2pHost.Host.Peerstore().Peers()
	a.EmitLog(fmt.Sprintf("GetDiscoveredPeers: Found %d peers in Peerstore", len(allPeers)))
	a.EmitLog(fmt.Sprintf("GetDiscoveredPeers: Found %d peers in Peerstore", len(allPeers)))
	for _, pid := range allPeers {
		if pid == a.p2pHost.Host.ID() {
			continue
		}

		pInfo := p2p.PeerInfo{
			UserID: pid.String(),
			Status: "online",
		}

		if a.store != nil {
			if profile, err := a.store.GetPeer(pid.String()); err == nil {
				pInfo.Username = profile.Username
				pInfo.Avatar = profile.Avatar
			}
		}

		if pInfo.Username == "" {
			pInfo.Username = pid.String()
		}

		addrs := a.p2pHost.Host.Peerstore().Addrs(pid)
		if len(addrs) > 0 {
			for _, addr := range addrs {
				s := addr.String()
				parts := strings.Split(s, "/")
				if len(parts) >= 3 && (parts[1] == "ip4" || parts[1] == "ip6") {
					pInfo.IP = parts[2]
					break
				}
			}
		}

		a.peerPortsLock.Lock()
		if port, ok := a.peerPorts[pid.String()]; ok {
			pInfo.Port = port
		}
		a.peerPortsLock.Unlock()

		peers = append(peers, pInfo)
	}
	return peers
}

func (a *App) handleHistoryQuery(query *p2p_proto.HistoryQuery, topicName string) {
	if query == nil {
		return
	}
	if query.RequesterId == a.UserID {
		return
	}

	if a.store == nil {
		return
	}

	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	offer := &p2p_proto.HistoryOffer{
		ProviderId:     a.UserID,
		ChannelId:      query.ChannelId,
		AvailableCount: 100,
	}

	packet := &p2p_proto.Packet{
		Payload: &p2p_proto.Packet_HistoryOffer{HistoryOffer: offer},
	}
	data, _ := json.Marshal(packet)
	if topic, ok := a.topics[topicName]; ok {
		a.p2pHost.Publish(topic, data)
	}
}

func (a *App) handleProfileQuery(query *p2p_proto.ProfileQuery, topicName string) {
	if query == nil {
		return
	}
	if query.TargetId == a.UserID {
		resp := &p2p_proto.ProfileResponse{
			UserId:        a.UserID,
			Username:      a.Username,
			AvatarUrl:     "",
			Status:        a.Status,
			StatusMessage: a.StatusMessage,
			P2PPort:       int32(a.P2PPort),
			Bio:           "",
		}

		packet := &p2p_proto.Packet{
			Payload: &p2p_proto.Packet_ProfileResponse{ProfileResponse: resp},
		}
		data, _ := json.Marshal(packet)
		if topic, ok := a.topics[topicName]; ok {
			a.p2pHost.Publish(topic, data)
		}
	}
}

func (a *App) handleProfileResponse(resp *p2p_proto.ProfileResponse) {
	if resp == nil {
		return
	}

	if a.store != nil {
		err := a.store.SavePeer(resp.UserId, resp.Username, resp.AvatarUrl, resp.Status, resp.Bio)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Failed to save peer profile: %v", err))
		}
	}

	a.peerPortsLock.Lock()
	a.peerPorts[resp.UserId] = int(resp.P2PPort)
	a.peerPortsLock.Unlock()

	runtime.EventsEmit(a.ctx, "PEER_UPDATED", resp)
}

func (a *App) JoinNode(nodeID string, joinCode string) bool {
	a.EmitLog(fmt.Sprintf("Joining node %s...", nodeID))

	var name, icon, description string
	peers := a.GetDiscoveredPeers()
	for _, p := range peers {
		if p.UserID == nodeID {
			name = p.Username
			icon = p.Avatar
			break
		}
	}
	if name == "" {
		name = nodeID
	}

	if a.store != nil {
		err := a.store.SaveJoinedNode(
			nodeID,
			nodeID,
			name,
			description,
			icon,
			nodeID,
			true,
		)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Failed to save joined node: %v", err))
			return false
		}
	}

	a.EmitLog(fmt.Sprintf("Successfully joined node %s", name))
	return true
}

func (a *App) GetNodeMembers() []p2p.Peer {
	peers := a.GetDiscoveredPeers()
	var legacyPeers []p2p.Peer
	for _, p := range peers {
		legacyPeers = append(legacyPeers, p2p.Peer{
			UserID:   p.UserID,
			Username: p.Username,
		})
	}
	return legacyPeers
}

func (a *App) GetGuildMembers(guildID string) []*proto.Member {
	return []*proto.Member{}
}

func (a *App) GetPublicNodes() []p2p.PeerInfo {
	return a.GetDiscoveredPeers()
}

func (a *App) GetJoinedNodes() []store.JoinedNode {
	if a.store == nil {
		return []store.JoinedNode{}
	}
	nodes, err := a.store.GetJoinedNodes()
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to get joined nodes: %v", err))
		return []store.JoinedNode{}
	}
	return nodes
}

func (a *App) UpdateServerMetadata(name, description, icon string) {
	runtime.EventsEmit(a.ctx, "localNodeUpdated", map[string]string{
		"name":        name,
		"description": description,
		"icon":        icon,
	})

	if a.store != nil {
		err := a.store.SaveJoinedNode(
			a.UserID,
			"",
			name,
			description,
			icon,
			a.UserID,
			true,
		)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Failed to save local node: %v", err))
		}
	}
}

func (a *App) GetLocalNodeMetadata() map[string]string {
	if a.store == nil {
		return map[string]string{}
	}
	nodes, err := a.store.GetJoinedNodes()
	if err == nil {
		for _, n := range nodes {
			if n.NodeID == a.UserID {
				return map[string]string{
					"name":        n.Name,
					"description": n.Description,
					"icon":        n.Icon,
				}
			}
		}
	}
	return map[string]string{}
}

func (a *App) GetLibp2pID() string {
	if a.p2pHost == nil {
		return ""
	}
	return a.p2pHost.Host.ID().String()
}

func (a *App) JoinChannel(channelName string) map[string]interface{} {
	if a.p2pHost != nil {
		go a.subscribeAndListen(channelName)

		go func() {
			offerCh := make(chan *p2p_proto.HistoryOffer, 10)
			a.syncLock.Lock()
			a.syncChans[channelName] = offerCh
			a.syncLock.Unlock()

			defer func() {
				a.syncLock.Lock()
				delete(a.syncChans, channelName)
				a.syncLock.Unlock()
			}()

			query := &p2p_proto.Packet{
				Payload: &p2p_proto.Packet_HistoryQuery{
					HistoryQuery: &p2p_proto.HistoryQuery{
						RequesterId:          a.UserID,
						ChannelId:            channelName,
						LastMessageTimestamp: 0,
					},
				},
			}
			data, _ := json.Marshal(query)
			if topic, ok := a.topics[channelName]; ok {
				a.p2pHost.Publish(topic, data)
			}

			timeout := time.After(2 * time.Second)
			var bestOffer *p2p_proto.HistoryOffer

		Loop:
			for {
				select {
				case offer := <-offerCh:
					if offer.ProviderId == a.UserID {
						continue
					}
					bestOffer = offer
					break Loop
				case <-timeout:
					break Loop
				}
			}

			if bestOffer != nil {
				a.EmitLog(fmt.Sprintf("Syncing history from %s", bestOffer.ProviderId))
				msgs, err := a.p2pHost.FetchHistory(a.ctx, bestOffer.ProviderId, channelName, 0)
				if err != nil {
					a.EmitLog(fmt.Sprintf("Failed to fetch history: %v", err))
				} else {
					a.EmitLog(fmt.Sprintf("Fetched %d messages", len(msgs)))
					for _, m := range msgs {
						if a.store != nil {
							a.store.SaveMessage(m.SenderId, m.SenderUsername, m.Content, m.ChannelId, false)
						}
						runtime.EventsEmit(a.ctx, "CHAT_MESSAGE", m)
					}
				}
			} else {
				a.EmitLog("No history providers found")
			}
		}()
	}

	return map[string]interface{}{
		"id":       0,
		"owner_id": "",
		"peers":    []p2p.PeerInfo{},
	}
}

func (a *App) subscribeAndListen(topicName string) {
	if _, exists := a.topics[topicName]; exists {
		return
	}

	topic, err := a.p2pHost.JoinTopic(topicName)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to join topic %s: %v", topicName, err))
		return
	}
	a.topics[topicName] = topic

	sub, err := a.p2pHost.Subscribe(topic)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to subscribe to topic %s: %v", topicName, err))
		return
	}
	a.subs[topicName] = sub

	a.EmitLog(fmt.Sprintf("Subscribed to PubSub topic: %s", topicName))

	for {
		msg, err := sub.Next(a.ctx)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Subscription error on %s: %v", topicName, err))
			return
		}

		if msg.ReceivedFrom == a.p2pHost.Host.ID() {
			continue
		}

		var packet p2p_proto.Packet
		if err := json.Unmarshal(msg.Data, &packet); err != nil {
			var chatMsg p2p_proto.ChatMessage
			if err := json.Unmarshal(msg.Data, &chatMsg); err == nil {
				a.EmitLog(fmt.Sprintf("Received legacy message in %s", topicName))
				runtime.EventsEmit(a.ctx, "CHAT_MESSAGE", &chatMsg)
				continue
			}
			a.EmitLog(fmt.Sprintf("Failed to unmarshal pubsub message: %v", err))
			continue
		}

		switch p := packet.Payload.(type) {
		case *p2p_proto.Packet_ChatMessage:
			chatMsg := p.ChatMessage
			a.EmitLog(fmt.Sprintf("Received ChatMessage in %s from %s", topicName, chatMsg.SenderUsername))
			runtime.EventsEmit(a.ctx, "CHAT_MESSAGE", chatMsg)
			if chatMsg.SenderId != a.UserID && a.store != nil {
				a.store.SaveMessage(chatMsg.SenderId, chatMsg.SenderUsername, chatMsg.Content, chatMsg.ChannelId, false)
			}

		case *p2p_proto.Packet_HistoryQuery:
			go a.handleHistoryQuery(p.HistoryQuery, topicName)

		case *p2p_proto.Packet_HistoryOffer:
			go a.handleHistoryOffer(p.HistoryOffer, topicName)
		case *p2p_proto.Packet_ProfileQuery:
			go a.handleProfileQuery(p.ProfileQuery, topicName)
		case *p2p_proto.Packet_ProfileResponse:
			go a.handleProfileResponse(p.ProfileResponse)
		case *p2p_proto.Packet_WebrtcSignal:
			go a.handleWebRTCSignal(p.WebrtcSignal)
		}
	}
}

func (a *App) handleHistoryOffer(offer *p2p_proto.HistoryOffer, topicName string) {
	a.syncLock.Lock()
	ch, ok := a.syncChans[topicName]
	a.syncLock.Unlock()

	if ok {
		select {
		case ch <- offer:
		default:
		}
	}
}

func (a *App) handleWebRTCSignal(signal *p2p_proto.WebRTCSignal) {
	if signal == nil {
		return
	}
	if signal.TargetId != a.UserID {
		return
	}

	a.EmitLog(fmt.Sprintf("Received WebRTC Signal: %s from %s", signal.Type, signal.SenderId))
	runtime.EventsEmit(a.ctx, "WEBRTC_SIGNAL", signal)
}

func (a *App) CreateChannel(channelName, topic string, isPrivate bool, icon, description, channelType string) error {
	if a.store != nil {
		channelID := fmt.Sprintf("%d", time.Now().UnixNano())

		err := a.store.SaveChannel(channelID, a.UserID, channelName, topic, channelType, isPrivate, icon, description)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Failed to save channel: %v", err))
			return err
		}
		a.EmitLog(fmt.Sprintf("Created channel %s (%s)", channelName, channelID))

		runtime.EventsEmit(a.ctx, "CHANNEL_METADATA", nil)
	} else {
		a.EmitLog("Store is nil, cannot create channel")
	}

	return nil
}

func (a *App) UpdateChannelInfo(channelID, topic string, isPrivate bool, icon, description, channelType string) bool {
	if a.store == nil {
		return false
	}

	err := a.store.SaveChannel(channelID, a.UserID, channelID, topic, channelType, isPrivate, icon, description)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to update channel: %v", err))
		return false
	}

	runtime.EventsEmit(a.ctx, "CHANNEL_METADATA", nil)
	return true
}

func (a *App) UpdateNodeSettings(name, icon, description string) error {
	if a.store != nil {
		err := a.store.SaveJoinedNode(
			a.UserID,
			"",
			name,
			description,
			icon,
			a.UserID,
			true,
		)
		if err != nil {
			return fmt.Errorf("failed to save local node settings: %v", err)
		}
	}

	return nil
}

func (a *App) LeaveChannel(channelName string) bool {
	if topic, ok := a.topics[channelName]; ok {
		topic.Close()
		delete(a.topics, channelName)
	}
	if sub, ok := a.subs[channelName]; ok {
		sub.Cancel()
		delete(a.subs, channelName)
	}
	return true
}

func (a *App) DeleteChannel(channelID string) bool {
	if a.store == nil {
		return false
	}
	err := a.store.DeleteChannel(channelID)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to delete channel: %v", err))
		return false
	}

	runtime.EventsEmit(a.ctx, "CHANNEL_DELETED", nil)
	return true
}

func (a *App) GetChannelPeers(channelName string) []p2p.PeerInfo {
	if a.p2pHost == nil {
		return []p2p.PeerInfo{}
	}

	if topic, ok := a.topics[channelName]; ok {
		peerIDs := topic.ListPeers()
		var peers []p2p.PeerInfo
		for _, pid := range peerIDs {
			peers = append(peers, p2p.PeerInfo{
				UserID:   pid.String(),
				Username: pid.String(),
				Status:   "online",
			})
		}
		return peers
	}
	return []p2p.PeerInfo{}
}

func (a *App) SendMessage(targetID, content string) bool {
	msg := &p2p_proto.ChatMessage{
		Id:             fmt.Sprintf("%d", time.Now().UnixNano()),
		ChannelId:      targetID,
		SenderId:       a.UserID,
		SenderUsername: a.Username,
		Content:        content,
		Timestamp:      time.Now().Unix(),
	}

	if len(targetID) > 0 && targetID[0] == '#' && a.p2pHost != nil {
		topicName := targetID
		topic, ok := a.topics[topicName]
		if !ok {
			a.EmitLog(fmt.Sprintf("Not joined to topic %s", topicName))
			return false
		}

		packet := &p2p_proto.Packet{
			Payload: &p2p_proto.Packet_ChatMessage{ChatMessage: msg},
		}
		data, _ := json.Marshal(packet)
		if err := a.p2pHost.Publish(topic, data); err != nil {
			a.EmitLog(fmt.Sprintf("Failed to publish message: %v", err))
			return false
		}
		return true
	}

	if a.p2pHost != nil {
		err := a.p2pHost.SendDirectMessage(a.ctx, targetID, msg)
		if err != nil {
			a.EmitLog(fmt.Sprintf("Failed to send direct message: %v", err))
			return false
		}
		if a.store != nil {
			if err := a.store.SaveMessage(a.UserID, a.Username, content, targetID, true); err != nil {
				a.EmitLog(fmt.Sprintf("Failed to save sent message: %v", err))
			}
		}
		return true
	}

	return false
}

func (a *App) GetMessages(peerID string) []store.Message {
	if a.store == nil {
		return []store.Message{}
	}
	msgs, err := a.store.GetMessages(peerID)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to get messages: %v", err))
		return []store.Message{}
	}
	return msgs
}

func (a *App) GetPeerChannels(peerID string) []*proto.ChannelMetadata {
	if a.store == nil {
		return []*proto.ChannelMetadata{}
	}

	channels, err := a.store.GetChannelsByOwner(peerID)
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to get channels for peer %s: %v", peerID, err))
		return []*proto.ChannelMetadata{}
	}

	var protoChannels []*proto.ChannelMetadata
	for _, c := range channels {
		protoChannels = append(protoChannels, &proto.ChannelMetadata{
			ChannelId:   c.ChannelID,
			Name:        c.Name,
			Topic:       c.Topic,
			IsPrivate:   c.IsPrivate,
			Icon:        c.Icon,
			Description: c.Description,
			Type:        c.Type,
		})
	}
	return protoChannels
}

func (a *App) GetDirectMessages() []p2p.PeerInfo {
	if a.store == nil {
		return []p2p.PeerInfo{}
	}
	recentPeers, err := a.store.GetRecentPeers()
	if err != nil {
		a.EmitLog(fmt.Sprintf("Failed to get recent peers: %v", err))
		return []p2p.PeerInfo{}
	}

	var peers []p2p.PeerInfo
	onlinePeers := a.GetDiscoveredPeers()
	onlineMap := make(map[string]p2p.PeerInfo)
	for _, p := range onlinePeers {
		onlineMap[p.UserID] = p
	}

	for _, rp := range recentPeers {
		if p, ok := onlineMap[rp.UserID]; ok {
			peers = append(peers, p)
		} else {
			peers = append(peers, p2p.PeerInfo{
				UserID:   rp.UserID,
				Username: rp.Username,
				Status:   "offline",
			})
		}
	}
	return peers
}

func (a *App) GetMessagesForSync(channelID string, afterTimestamp int64) ([]*p2p_proto.ChatMessage, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store is nil")
	}

	msgs, err := a.store.GetMessagesAfter(afterTimestamp, channelID, "")
	if err != nil {
		return nil, err
	}

	var protoMsgs []*p2p_proto.ChatMessage
	for _, m := range msgs {
		protoMsgs = append(protoMsgs, &p2p_proto.ChatMessage{
			Id:             fmt.Sprintf("%d", m.ID),
			Content:        m.Content,
			Timestamp:      m.Timestamp.Unix(),
			SenderId:       m.SenderID,
			SenderUsername: m.SenderUsername,
			ChannelId:      m.ChannelID,
		})
	}

	return protoMsgs, nil
}

func (a *App) SendWebRTCSignal(targetID, signalType, payload, channelID string) bool {
	if a.p2pHost == nil {
		return false
	}

	signal := &p2p_proto.WebRTCSignal{
		SenderId:  a.UserID,
		TargetId:  targetID,
		Type:      signalType,
		Payload:   payload,
		ChannelId: channelID,
	}

	packet := &p2p_proto.Packet{
		Payload: &p2p_proto.Packet_WebrtcSignal{WebrtcSignal: signal},
	}
	data, _ := json.Marshal(packet)

	topicName := channelID
	if topicName == "" {
		a.EmitLog("SendWebRTCSignal: ChannelID is empty")
		return false
	}

	if topic, ok := a.topics[topicName]; ok {
		if err := a.p2pHost.Publish(topic, data); err != nil {
			a.EmitLog(fmt.Sprintf("Failed to publish WebRTC signal: %v", err))
			return false
		}
		return true
	} else {
		a.EmitLog(fmt.Sprintf("SendWebRTCSignal: Topic %s not found/joined", topicName))
		return false
	}
}

func (a *App) OnChatMessage(msg *p2p_proto.ChatMessage) {
	a.EmitLog(fmt.Sprintf("App received message from %s: %s", msg.SenderUsername, msg.Content))

	if a.store != nil {
		if err := a.store.SaveMessage(msg.SenderId, msg.SenderUsername, msg.Content, msg.ChannelId, false); err != nil {
			a.EmitLog(fmt.Sprintf("Failed to save message: %v", err))
		}
	}

	runtime.EventsEmit(a.ctx, "CHAT_MESSAGE", msg)
}

type ActivityUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Activity struct {
	ImageUrl string       `json:"imageUrl"`
	Game     string       `json:"game"`
	User     ActivityUser `json:"user"`
	Details  string       `json:"details"`
}

func (a *App) GetActiveNow() []Activity {
	return []Activity{}
}
