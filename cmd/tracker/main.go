package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Tracker struct {
	db *pgxpool.Pool
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://redware:redware@localhost:5433/redware?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	schemaData, err := os.ReadFile("cmd/tracker/schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}
	_, err = db.Exec(ctx, string(schemaData))
	if err != nil {
		log.Fatalf("Failed to apply schema: %v", err)
	}
	log.Println("Database schema applied successfully.")

	tracker := &Tracker{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", tracker.handleRegister)
	mux.HandleFunc("/auth/login", tracker.handleLogin)
	mux.HandleFunc("/heartbeat", tracker.handleHeartbeat)
	mux.HandleFunc("/peers", tracker.handleGetPeers)
	mux.HandleFunc("/channels/create", tracker.handleCreateChannel)
	mux.HandleFunc("/channels/join", tracker.handleJoinChannel)
	mux.HandleFunc("/channels/leave", tracker.handleLeaveChannel)
	mux.HandleFunc("/channels/peers", tracker.handleGetChannelPeers)

	log.Printf("Bonfire Tracker listening on :%s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (t *Tracker) handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Bonfire Tracker Online 🔥")
}

func (t *Tracker) handleHealth(w http.ResponseWriter, r *http.Request) {
	err := t.db.Ping(r.Context())
	if err != nil {
		http.Error(w, "Database offline", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (t *Tracker) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Avatar   string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := t.db.Exec(r.Context(), "INSERT INTO users (username, password_hash, avatar_data) VALUES ($1, $2, $3)", req.Username, req.Password, req.Avatar)
	if err != nil {
		log.Printf("Registration error: %v", err)
		http.Error(w, "Username likely taken", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "User created")
}

func (t *Tracker) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var id string
	var avatar *string
	err := t.db.QueryRow(r.Context(), "SELECT id, avatar_data FROM users WHERE username=$1 AND password_hash=$2", req.Username, req.Password).Scan(&id, &avatar)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	resp := map[string]string{
		"token":  id,
		"avatar": "",
	}
	if avatar != nil {
		resp["avatar"] = *avatar
	}
	json.NewEncoder(w).Encode(resp)
}

func (t *Tracker) handleUpdateAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Avatar string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := t.db.Exec(r.Context(), "UPDATE users SET avatar_data=$1 WHERE id=$2", req.Avatar, userID)
	if err != nil {
		log.Printf("Update Avatar error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Tracker) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		log.Println("Heartbeat unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Port          int    `json:"port"`
		Status        string `json:"status"`
		StatusMessage string `json:"status_message"`
		ServerName    string `json:"server_name"`
		ServerIcon    string `json:"server_icon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if host == "::1" {
		host = "127.0.0.1"
	}

	_, err = t.db.Exec(r.Context(), `
		INSERT INTO users (id, username, password_hash, avatar_data)
		VALUES ($1, $2, 'p2p_auto_generated', '')
		ON CONFLICT (id) DO NOTHING
	`, userID, "User-"+userID[:8])
	if err != nil {
		log.Printf("Heartbeat User Upsert error: %v", err)
	}

	query := `
		INSERT INTO peers (user_id, ip_address, port, last_seen, status, status_message, server_name, server_icon)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6, $7)
		ON CONFLICT (user_id) 
		DO UPDATE SET ip_address = EXCLUDED.ip_address, port = EXCLUDED.port, last_seen = NOW(), status = EXCLUDED.status, status_message = EXCLUDED.status_message, server_name = EXCLUDED.server_name, server_icon = EXCLUDED.server_icon;
	`
	_, err = t.db.Exec(r.Context(), query, userID, host, req.Port, req.Status, req.StatusMessage, req.ServerName, req.ServerIcon)
	if err != nil {
		log.Printf("Heartbeat DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Tracker) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		log.Println("GetPeers unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := t.db.Query(r.Context(), `
		SELECT p.user_id, u.username, p.ip_address, p.port, u.avatar_data, p.status, p.status_message, p.server_name, p.server_icon
		FROM peers p
		JOIN users u ON p.user_id = u.id
		WHERE p.last_seen > NOW() - INTERVAL '45 seconds' AND p.user_id != $1
	`, userID)
	if err != nil {
		log.Printf("GetPeers DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Peer struct {
		UserID        string  `json:"user_id"`
		Username      string  `json:"username"`
		IP            string  `json:"ip"`
		Port          int     `json:"port"`
		Avatar        *string `json:"avatar"`
		Status        string  `json:"status"`
		StatusMessage string  `json:"status_message"`
		ServerName    string  `json:"server_name"`
		ServerIcon    string  `json:"server_icon"`
	}
	var peers []Peer
	for rows.Next() {
		var p Peer
		if err := rows.Scan(&p.UserID, &p.Username, &p.IP, &p.Port, &p.Avatar, &p.Status, &p.StatusMessage, &p.ServerName, &p.ServerIcon); err != nil {
			continue
		}
		peers = append(peers, p)
	}

	json.NewEncoder(w).Encode(peers)
}

func (t *Tracker) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ChannelName string `json:"channel_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var channelID int
	err := t.db.QueryRow(r.Context(), `
		INSERT INTO channels (name, owner_id) VALUES ($1, $2)
		RETURNING id
	`, req.ChannelName, userID).Scan(&channelID)

	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "channels_name_key" (SQLSTATE 23505)` {
			http.Error(w, "Channel name taken", http.StatusConflict)
			return
		}
		log.Printf("CreateChannel DB error: %v", err)
		http.Error(w, "Channel name likely taken", http.StatusConflict)
		return
	}

	_, err = t.db.Exec(r.Context(), `
		INSERT INTO channel_members (channel_id, user_id, last_seen)
		VALUES ($1, $2, NOW())
	`, channelID, userID)
	if err != nil {
		log.Printf("CreateChannel DB error (join): %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       channelID,
		"name":     req.ChannelName,
		"owner_id": userID,
	})
}

func (t *Tracker) handleJoinChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ChannelName string `json:"channel_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var channelID int
	var ownerID string
	err := t.db.QueryRow(r.Context(), `
		SELECT id, owner_id FROM channels WHERE name = $1
	`, req.ChannelName).Scan(&channelID, &ownerID)

	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	_, err = t.db.Exec(r.Context(), `
		INSERT INTO channel_members (channel_id, user_id, last_seen)
		VALUES ($1, $2, NOW())
		ON CONFLICT (channel_id, user_id) DO UPDATE SET last_seen = NOW()
	`, channelID, userID)
	if err != nil {
		log.Printf("JoinChannel DB error (join): %v", err)
	}

	t.returnChannelPeers(w, r, channelID, userID, ownerID)
}

func (t *Tracker) handleLeaveChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ChannelName string `json:"channel_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := t.db.Exec(r.Context(), `
		DELETE FROM channel_members 
		WHERE user_id = $1 AND channel_id = (SELECT id FROM channels WHERE name = $2)
	`, userID, req.ChannelName)
	if err != nil {
		log.Printf("LeaveChannel DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Tracker) handleGetChannelPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channelName := r.URL.Query().Get("name")
	if channelName == "" {
		http.Error(w, "Missing channel name", http.StatusBadRequest)
		return
	}

	var channelID int
	var ownerID string
	err := t.db.QueryRow(r.Context(), "SELECT id, owner_id FROM channels WHERE name = $1", channelName).Scan(&channelID, &ownerID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	t.returnChannelPeers(w, r, channelID, userID, ownerID)
}

func (t *Tracker) returnChannelPeers(w http.ResponseWriter, r *http.Request, channelID int, excludeUserID string, ownerID string) {
	rows, err := t.db.Query(r.Context(), `
		SELECT p.user_id, u.username, p.ip_address, p.port, u.avatar_data, p.status, p.status_message, p.server_name, p.server_icon
		FROM channel_members cm
		JOIN peers p ON cm.user_id = p.user_id
		JOIN users u ON p.user_id = u.id
		WHERE cm.channel_id = $1 AND p.last_seen > NOW() - INTERVAL '45 seconds' AND p.user_id != $2
	`, channelID, excludeUserID)
	if err != nil {
		log.Printf("returnChannelPeers DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Peer struct {
		UserID        string  `json:"user_id"`
		Username      string  `json:"username"`
		IP            string  `json:"ip"`
		Port          int     `json:"port"`
		Avatar        *string `json:"avatar"`
		Status        string  `json:"status"`
		StatusMessage string  `json:"status_message"`
		ServerName    string  `json:"server_name"`
		ServerIcon    string  `json:"server_icon"`
	}
	var peers []Peer
	for rows.Next() {
		var p Peer
		if err := rows.Scan(&p.UserID, &p.Username, &p.IP, &p.Port, &p.Avatar, &p.Status, &p.StatusMessage, &p.ServerName, &p.ServerIcon); err != nil {
			continue
		}
		peers = append(peers, p)
	}

	response := map[string]interface{}{
		"peers":    peers,
		"owner_id": ownerID,
	}

	json.NewEncoder(w).Encode(response)
}

func (t *Tracker) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := t.authenticate(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Status        string `json:"status"`
		StatusMessage string `json:"status_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := t.db.Exec(r.Context(), `
		UPDATE peers SET status = $1, status_message = $2, last_seen = NOW()
		WHERE user_id = $3
	`, req.Status, req.StatusMessage, userID)
	if err != nil {
		log.Printf("UpdateStatus DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Tracker) authenticate(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return ""
	}
	token := authHeader[7:]
	return token
}
