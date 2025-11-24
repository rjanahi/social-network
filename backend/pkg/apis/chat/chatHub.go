package chat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	database "socialnetwork/pkg/db"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

type Frontend struct {
	From         int       `json:"from"`
	To           int       `json:"to"`
	GroupID      int       `json:"group_id"`
	Content      string    `json:"content"`
	Username     string    `json:"username"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
	PostId       int       `json:"post_id"`
	CommentId    int       `json:"comment_id"`
	IsLike       bool      `json:"is_like"`
	IsPrivate    bool      `json:"isPrivate,omitempty"`
	YesCount     int       `json:"yes_count,omitempty"`
	NoCount      int       `json:"no_count,omitempty"`
	UserVote     string    `json:"user_vote,omitempty"`
	RequestID    int       `json:"request_id,omitempty"`
	InvitationID int       `json:"invitation_id,omitempty"`
	EventDate    string    `json:"event_date,omitempty"`
}

type Client struct {
	UserID int
	Conn   *websocket.Conn
	Send   chan Frontend
}

type Hub struct {
	Clients      map[int]*Client
	Online       chan *Client
	Offline      chan *Client
	Broadcast    chan Frontend
	MessageStore map[string][]Frontend
	Mutex        sync.RWMutex
	DB           *sql.DB
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		Clients:      make(map[int]*Client),
		Online:       make(chan *Client),
		Offline:      make(chan *Client),
		Broadcast:    make(chan Frontend),
		MessageStore: make(map[string][]Frontend),
		DB:           db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Online:
			h.Mutex.Lock()
			h.Clients[client.UserID] = client
			h.Mutex.Unlock()

			// Broadcast updated online users list to all clients
			h.broadcastOnlineUsers()

		case client := <-h.Offline:
			h.Mutex.Lock()
			delete(h.Clients, client.UserID)
			h.Mutex.Unlock()

			// Broadcast updated online users list to all clients
			h.broadcastOnlineUsers()

		case msg := <-h.Broadcast:
			key := chatKey(msg.From, msg.To)
			h.Mutex.Lock()
			h.MessageStore[key] = append(h.MessageStore[key], msg)
			h.Mutex.Unlock()
			_ = h.saveMessageToDB(msg)
			if toClient, ok := h.Clients[msg.To]; ok {
				toClient.Send <- msg
			}
		}
	}
}

func (h *Hub) GetOnlineUserIDs() []int {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	var userIDs []int
	for id := range h.Clients {
		userIDs = append(userIDs, id)
	}
	return userIDs
}

// broadcastOnlineUsers sends the updated list of online users to all connected clients
func (h *Hub) broadcastOnlineUsers() {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	// Get list of online user IDs with usernames
	type OnlineUser struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}

	var onlineUsers []OnlineUser
	for userID := range h.Clients {
		// Fetch username from database
		var username string
		err := h.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		if err == nil {
			onlineUsers = append(onlineUsers, OnlineUser{ID: userID, Username: username})
		}
	}

	// Create a custom message with the users array
	type OnlineUsersMessage struct {
		Type      string       `json:"type"`
		Users     []OnlineUser `json:"users"`
		Timestamp time.Time    `json:"timestamp"`
	}

	msg := OnlineUsersMessage{
		Type:      "online_users",
		Users:     onlineUsers,
		Timestamp: time.Now(),
	}

	// Send to all connected clients
	for _, client := range h.Clients {
		// Convert to JSON and send directly
		data, err := json.Marshal(msg)
		if err == nil {
			client.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func chatKey(a, b int) string {
	if a < b {
		return fmt.Sprintf("%d-%d", a, b)
	}
	return fmt.Sprintf("%d-%d", b, a)
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[WebSocket] ServeWs called from %s\n", r.RemoteAddr)
	fmt.Printf("[WebSocket] Request URL: %s\n", r.URL.String())
	fmt.Printf("[WebSocket] Request Headers: %v\n", r.Header)

	userIDStr := r.URL.Query().Get("user_id")
	fmt.Printf("[WebSocket] userIDStr from query: '%s'\n", userIDStr)

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		fmt.Printf("[WebSocket] Invalid user ID: userIDStr='%s', err=%v, userID=%d\n", userIDStr, err, userID)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	fmt.Printf("[WebSocket] Valid userID: %d, attempting upgrade\n", userID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[WebSocket] Upgrade Error: %v\n", err)
		return
	}

	fmt.Printf("[WebSocket] Successfully upgraded connection for user %d\n", userID)

	client := &Client{UserID: userID, Conn: conn, Send: make(chan Frontend)}
	hub.Online <- client

	go client.writePump()
	go client.readPump(hub)
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Offline <- c
		c.Conn.Close()
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Frontend
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		// Handle request for online users list
		if msg.Type == "get_online_users" {
			hub.broadcastOnlineUsers()
			continue
		}

		// Handle typing signal separately
		if msg.Type == "typing" {
			hub.Mutex.RLock()
			if toClient, ok := hub.Clients[msg.To]; ok {
				toClient.Send <- msg
			}
			hub.Mutex.RUnlock()
			continue
		}

		if msg.Type == "new_post" || msg.Type == "new_comment" || msg.Type == "new_postLike" || msg.Type == "new_commentLike" {
			hub.Mutex.RLock()
			for _, client := range hub.Clients {
				client.Send <- msg
			}
			hub.Mutex.RUnlock()
			continue
		}

		// Group chat fan-out
		if msg.Type == "group_message" || msg.Type == "new_groupPost" || msg.Type == "new_groupEvent" {
			msg.Timestamp = time.Now()

			if msg.Type == "group_message" {
				// 1) Save to DB
				if err := hub.saveGroupMessageToDB(msg); err != nil {
					// optional: log but don't break fan-out
					fmt.Println("saveGroupMessageToDB error:", err)
				}
			}

			// 2) Find members of the group
			memberIDs, err := hub.getGroupMemberIDs(msg.GroupID)
			if err != nil {
				fmt.Println("getGroupMemberIDs error:", err)
				return
			}

			// 3) Send to every online member (including sender for echo)
			hub.Mutex.RLock()
			for _, mid := range memberIDs {
				if cl, ok := hub.Clients[mid]; ok {
					cl.Send <- msg
				}
			}
			hub.Mutex.RUnlock()

			// handled — skip the 1:1 broadcast path
			continue
		}

		// todo: Notifications
		if msg.Type == "notif" {
			hub.Mutex.RLock()
			for _, client := range hub.Clients {
				if client.UserID != msg.From {
					client.Send <- msg
				}
			}
			hub.Mutex.RUnlock()
			continue
		}

		// Handle private messages with follow validation
		if msg.Type == "private_message" || msg.Type == "message" {
			// Validate that at least one user follows the other
			canChat, err := database.CheckFollowRelationship(hub.DB, msg.From, msg.To)
			if err != nil {
				fmt.Println("Error checking follow relationship:", err)
				continue
			}

			if !canChat {
				// Send error message back to sender
				errorMsg := Frontend{
					Type:      "error",
					Content:   "You can only message users you follow or who follow you",
					Timestamp: time.Now(),
				}
				hub.Mutex.RLock()
				if senderClient, ok := hub.Clients[msg.From]; ok {
					select {
					case senderClient.Send <- errorMsg:
					default:
					}
				}
				hub.Mutex.RUnlock()
				continue
			}

			// Save message to database
			msg.Timestamp = time.Now()
			if err := hub.saveMessageToDB(msg); err != nil {
				fmt.Println("Error saving message:", err)
			}

			// Send to recipient if online
			hub.Mutex.RLock()
			if toClient, ok := hub.Clients[msg.To]; ok {
				toClient.Send <- msg
			}
			// Echo back to sender for confirmation
			if fromClient, ok := hub.Clients[msg.From]; ok {
				fromClient.Send <- msg
			}
			hub.Mutex.RUnlock()
			continue
		}

		msg.Timestamp = time.Now()
		hub.Broadcast <- msg
	}
}

func (c *Client) writePump() {
	for msg := range c.Send {
		data, _ := json.Marshal(msg)
		c.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (h *Hub) saveMessageToDB(msg Frontend) error {
	query := `INSERT INTO messages (sender_id, receiver_id, content, created_at) VALUES (?, ?, ?, ?)`
	_, err := h.DB.Exec(query, msg.From, msg.To, msg.Content, msg.Timestamp)
	return err
}

// getGroupMemberIDs returns all accepted member user_ids for a group.
func (h *Hub) getGroupMemberIDs(groupID int) ([]int, error) {
	const q = `
		SELECT user_id
		FROM group_members
		WHERE group_id = ? AND status = 'accepted'
	`
	rows, err := h.DB.Query(q, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, rows.Err()
}

func (h *Hub) saveGroupMessageToDB(msg Frontend) error {
	// 1) Verify sender is an accepted member of the group
	var allowed int
	checkQ := `
        SELECT COUNT(*) 
        FROM group_members 
        WHERE group_id = ? AND user_id = ? AND status = 'accepted'
    `
	if err := h.DB.QueryRow(checkQ, msg.GroupID, msg.From).Scan(&allowed); err != nil {
		return fmt.Errorf("membership check failed: %w", err)
	}
	if allowed == 0 {
		return fmt.Errorf("user %d is not an accepted member of group %d", msg.From, msg.GroupID)
	}

	// 2) Insert the message
	const q = `
        INSERT INTO group_messages (group_id, sender_id, content, created_at)
        VALUES (?, ?, ?, ?)
    `
	_, err := h.DB.Exec(q, msg.GroupID, msg.From, msg.Content, msg.Timestamp)
	return err
}
