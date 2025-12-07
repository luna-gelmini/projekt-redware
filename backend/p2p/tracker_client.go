package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type TrackerClient struct {
	BaseURL string
	Client  *http.Client
	Token   string
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type PeerInfo struct {
	UserID            string `json:"user_id"`
	Username          string `json:"username"`
	IP                string `json:"ip"`
	Port              int    `json:"port"`
	Avatar            string `json:"avatar"`
	Status            string `json:"status"`
	StatusMessage     string `json:"status_message"`
	ServerName        string `json:"server_name"`
	ServerIcon        string `json:"server_icon"`
	ServerDescription string `json:"server_description"`
}

func NewTrackerClient(baseURL string) *TrackerClient {
	return &TrackerClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TrackerClient) Register(username, password string) error {
	payload := map[string]string{"username": username, "password": password}
	return c.post("/auth/register", payload, nil)
}

func (c *TrackerClient) Login(username, password string) (string, string, error) {
	payload := map[string]string{"username": username, "password": password}
	var resp struct {
		Token  string `json:"token"`
		Avatar string `json:"avatar"`
	}
	err := c.post("/auth/login", payload, &resp)
	if err != nil {
		return "", "", err
	}
	c.Token = resp.Token
	return resp.Token, resp.Avatar, nil
}

func (c *TrackerClient) UpdateAvatar(avatarData string) error {
	payload := map[string]string{"avatar": avatarData}
	return c.post("/user/avatar", payload, nil)
}

func (c *TrackerClient) Heartbeat(port int, status, statusMessage, serverName, serverIcon string) error {
	payload := map[string]interface{}{
		"port":           port,
		"status":         status,
		"status_message": statusMessage,
		"server_name":    serverName,
		"server_icon":    serverIcon,
	}
	err := c.post("/heartbeat", payload, nil)
	if err != nil {
		fmt.Printf("Heartbeat Error: %v\n", err)
	}
	return err
}

func (c *TrackerClient) UpdateStatus(status, statusMessage string) error {
	payload := map[string]string{
		"status":         status,
		"status_message": statusMessage,
	}
	return c.post("/status", payload, nil)
}

func (c *TrackerClient) GetPeers() ([]PeerInfo, error) {
	var peers []PeerInfo
	err := c.get("/peers", &peers)
	if err != nil {
		fmt.Printf("GetPeers Error: %v\n", err)
	}
	return peers, err
}

func (c *TrackerClient) JoinChannel(channelName string) ([]PeerInfo, string, error) {
	payload := map[string]string{"channel_name": channelName}
	var resp struct {
		Peers   []PeerInfo `json:"peers"`
		OwnerID string     `json:"owner_id"`
	}
	err := c.post("/channels/join", payload, &resp)
	return resp.Peers, resp.OwnerID, err
}

func (c *TrackerClient) CreateChannel(channelName string) (int, string, error) {
	payload := map[string]string{"channel_name": channelName}
	var resp struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		OwnerID string `json:"owner_id"`
	}
	err := c.post("/channels/create", payload, &resp)
	return resp.ID, resp.OwnerID, err
}

func (c *TrackerClient) LeaveChannel(channelName string) error {
	payload := map[string]string{"channel_name": channelName}
	return c.post("/channels/leave", payload, nil)
}

func (c *TrackerClient) GetChannelPeers(channelName string) ([]PeerInfo, string, error) {
	var resp struct {
		Peers   []PeerInfo `json:"peers"`
		OwnerID string     `json:"owner_id"`
	}
	encodedName := url.QueryEscape(channelName)
	targetURL := "/channels/peers?name=" + encodedName
	fmt.Printf("GetChannelPeers Requesting: %s\n", targetURL)
	err := c.get(targetURL, &resp)
	return resp.Peers, resp.OwnerID, err
}

func (c *TrackerClient) get(endpoint string, result interface{}) error {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *TrackerClient) post(endpoint string, payload interface{}, result interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}
