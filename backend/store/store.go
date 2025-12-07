package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	ID             int64     `json:"id"`
	SenderID       string    `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	Content        string    `json:"content"`
	ChannelID      string    `json:"channel_id"`
	Timestamp      time.Time `json:"timestamp"`
	IsOutgoing     bool      `json:"is_outgoing"`
}

type Store struct {
	db  *sql.DB
	key []byte
}

func NewStore(username string) (*Store, error) {
	dbName := fmt.Sprintf("redware_%s.db", username)
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id TEXT,
		sender_username TEXT,
		content TEXT,
		channel_id TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_outgoing BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS joined_nodes (
		node_id TEXT PRIMARY KEY,
		address TEXT,
		name TEXT,
		description TEXT,
		icon TEXT,
		owner_id TEXT,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_server BOOLEAN DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS channels (
		channel_id TEXT PRIMARY KEY,
		owner_id TEXT,
		name TEXT,
		topic TEXT,
		type TEXT,
		is_private BOOLEAN,
		icon TEXT,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	_, _ = s.db.Exec("ALTER TABLE joined_nodes ADD COLUMN is_server BOOLEAN DEFAULT 0")

	_, err = s.db.Exec(`
	CREATE TABLE IF NOT EXISTS peers (
		user_id TEXT PRIMARY KEY,
		username TEXT,
		avatar TEXT,
		status TEXT,
		bio TEXT,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)
	return err
}

func (s *Store) SavePeer(userID, username, avatar, status, bio string) error {
	_, err := s.db.Exec(`
		INSERT INTO peers (user_id, username, avatar, status, bio, last_seen)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
			username=excluded.username,
			avatar=excluded.avatar,
			status=excluded.status,
			bio=excluded.bio,
			last_seen=CURRENT_TIMESTAMP
	`, userID, username, avatar, status, bio)
	return err
}

func (s *Store) GetPeer(userID string) (*PeerMetadata, error) {
	var p PeerMetadata
	err := s.db.QueryRow("SELECT user_id, username, avatar FROM peers WHERE user_id = ?", userID).Scan(&p.UserID, &p.Username, &p.Avatar)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type JoinedNode struct {
	NodeID      string    `json:"node_id"`
	Address     string    `json:"address"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	OwnerID     string    `json:"owner_id"`
	JoinedAt    time.Time `json:"joined_at"`
	IsServer    bool      `json:"is_server"`
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveMessage(senderID, senderUsername, content, channelID string, isOutgoing bool) error {
	_, err := s.db.Exec("INSERT INTO messages (sender_id, sender_username, content, channel_id, is_outgoing) VALUES (?, ?, ?, ?, ?)",
		senderID, senderUsername, content, channelID, isOutgoing)
	return err
}

func (s *Store) GetMessages(peerID string) ([]Message, error) {
	var rows *sql.Rows
	var err error

	if len(peerID) > 0 && peerID[0] == '#' {
		rows, err = s.db.Query(`
			SELECT id, sender_id, sender_username, content, COALESCE(channel_id, ''), timestamp, is_outgoing 
			FROM messages 
			WHERE channel_id = ? 
			ORDER BY timestamp ASC`, peerID)
	} else {
		rows, err = s.db.Query(`
			SELECT id, sender_id, sender_username, content, COALESCE(channel_id, ''), timestamp, is_outgoing 
			FROM messages 
			WHERE (sender_id = ? OR (is_outgoing = 1 AND sender_id = ?)) 
			AND (channel_id IS NULL OR channel_id = '')
			ORDER BY timestamp ASC`, peerID, peerID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.SenderUsername, &m.Content, &m.ChannelID, &m.Timestamp, &m.IsOutgoing); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *Store) GetMessagesAfter(timestamp int64, channelID string, peerID string) ([]Message, error) {
	startTime := time.Unix(timestamp, 0)

	var rows *sql.Rows
	var err error

	if channelID != "" {
		rows, err = s.db.Query(`
			SELECT id, sender_id, sender_username, content, COALESCE(channel_id, ''), timestamp, is_outgoing 
			FROM messages 
			WHERE channel_id = ? AND timestamp > ?
			ORDER BY timestamp ASC`, channelID, startTime)
	} else {
		rows, err = s.db.Query(`
			SELECT id, sender_id, sender_username, content, COALESCE(channel_id, ''), timestamp, is_outgoing 
			FROM messages 
			WHERE sender_id = ? AND (channel_id IS NULL OR channel_id = '') AND timestamp > ?
			ORDER BY timestamp ASC`, peerID, startTime)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.SenderUsername, &m.Content, &m.ChannelID, &m.Timestamp, &m.IsOutgoing); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *Store) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Store) decrypt(ciphertextBase64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

type PeerMetadata struct {
	UserID   string
	Username string
	Avatar   string
}

func (s *Store) GetRecentPeers() ([]PeerMetadata, error) {
	rows, err := s.db.Query(`
		SELECT sender_id, 
		       MAX(CASE WHEN is_outgoing = 0 THEN sender_username ELSE '' END) as username
		FROM messages 
		WHERE (channel_id IS NULL OR channel_id = '')
		GROUP BY sender_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []PeerMetadata
	for rows.Next() {
		var p PeerMetadata
		if err := rows.Scan(&p.UserID, &p.Username); err != nil {
			return nil, err
		}
		if p.Username == "" {
			p.Username = p.UserID
		}
		peers = append(peers, p)
	}
	return peers, nil
}

func (s *Store) SaveJoinedNode(nodeID, address, name, description, icon, ownerID string, isServer bool) error {
	_, err := s.db.Exec(`
		INSERT INTO joined_nodes (node_id, address, name, description, icon, owner_id, is_server, joined_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(node_id) DO UPDATE SET
			address=excluded.address,
			name=excluded.name,
			description=excluded.description,
			icon=excluded.icon,
			owner_id=excluded.owner_id,
			is_server=excluded.is_server,
			joined_at=CURRENT_TIMESTAMP
	`, nodeID, address, name, description, icon, ownerID, isServer)
	return err
}

func (s *Store) GetJoinedNodes() ([]JoinedNode, error) {
	rows, err := s.db.Query("SELECT node_id, address, name, description, icon, owner_id, joined_at, COALESCE(is_server, 0) FROM joined_nodes ORDER BY joined_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []JoinedNode
	for rows.Next() {
		var n JoinedNode
		if err := rows.Scan(&n.NodeID, &n.Address, &n.Name, &n.Description, &n.Icon, &n.OwnerID, &n.JoinedAt, &n.IsServer); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (s *Store) RemoveJoinedNode(nodeID string) error {
	_, err := s.db.Exec("DELETE FROM joined_nodes WHERE node_id = ?", nodeID)
	return err
}

func (s *Store) DeleteChannel(channelID string) error {
	_, err := s.db.Exec("DELETE FROM channels WHERE channel_id = ?", channelID)
	return err
}
func (s *Store) SaveChannel(channelID, ownerID, name, topic, channelType string, isPrivate bool, icon, description string) error {
	_, err := s.db.Exec(`
		INSERT INTO channels (channel_id, owner_id, name, topic, type, is_private, icon, description) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(channel_id) DO UPDATE SET 
			name=excluded.name, 
			topic=excluded.topic, 
			type=excluded.type, 
			is_private=excluded.is_private,
			icon=excluded.icon,
			description=excluded.description
	`, channelID, ownerID, name, topic, channelType, isPrivate, icon, description)
	return err
}

func (s *Store) GetChannels() ([]JoinedNode, error) {
	return nil, errors.New("not implemented")
}

type StoreChannel struct {
	ChannelID   string
	OwnerID     string
	Name        string
	Topic       string
	Type        string
	IsPrivate   bool
	Icon        string
	Description string
}

func (s *Store) GetAllChannels() ([]StoreChannel, error) {
	rows, err := s.db.Query("SELECT channel_id, owner_id, name, topic, type, is_private, icon, description FROM channels")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []StoreChannel
	for rows.Next() {
		var c StoreChannel
		if err := rows.Scan(&c.ChannelID, &c.OwnerID, &c.Name, &c.Topic, &c.Type, &c.IsPrivate, &c.Icon, &c.Description); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, nil
}

func (s *Store) GetChannelsByOwner(ownerID string) ([]StoreChannel, error) {
	rows, err := s.db.Query("SELECT channel_id, owner_id, name, topic, type, is_private, icon, description FROM channels WHERE owner_id = ?", ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []StoreChannel
	for rows.Next() {
		var c StoreChannel
		if err := rows.Scan(&c.ChannelID, &c.OwnerID, &c.Name, &c.Topic, &c.Type, &c.IsPrivate, &c.Icon, &c.Description); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, nil
}
