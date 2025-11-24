package group

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"socialnetwork/pkg/apis/chat"
	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
)

// GetPendingJoinRequests returns all pending group join requests for groups where the user is an admin
func GetPendingJoinRequests(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session and get user ID
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Get pending join requests
	requests, err := database.GetPendingGroupJoinRequests(db, userID)
	if err != nil {
		fmt.Println("Error retrieving pending join requests:", err)
		http.Error(w, "Failed to retrieve pending join requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// RespondToJoinRequest handles approving/rejecting group join requests
type RespondJoinRequestBody struct {
	RequestID int    `json:"request_id"`
	Status    string `json:"status"` // "accepted" or "declined"
}

func RespondToJoinRequest(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session and get user ID
	adminUserID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var requestData RespondJoinRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if requestData.Status != "accepted" && requestData.Status != "declined" {
		http.Error(w, "Invalid status. Must be 'accepted' or 'declined'", http.StatusBadRequest)
		return
	}

	// Get the user_id and group_id from the membership record before updating
	var targetUserID, groupID int
	err := db.QueryRow(`
		SELECT gm.user_id, gm.group_id 
		FROM group_members gm
		JOIN groups g ON gm.group_id = g.id
		WHERE gm.id = ? AND g.creator_id = ? AND gm.status = 'pending'
	`, requestData.RequestID, adminUserID).Scan(&targetUserID, &groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Not authorized or request not found", http.StatusForbidden)
		} else {
			fmt.Println("Error fetching request details:", err)
			http.Error(w, "Failed to process request", http.StatusInternalServerError)
		}
		return
	}

	// Update the status
	err = database.UpdateGroupMemberStatus(db, requestData.RequestID, adminUserID, requestData.Status)
	if err != nil {
		fmt.Println("Error updating join request status:", err)
		http.Error(w, "Failed to update join request", http.StatusInternalServerError)
		return
	}

	// Get admin username for WebSocket notification
	var adminUsername string
	db.QueryRow("SELECT username FROM users WHERE id = ?", adminUserID).Scan(&adminUsername)

	// Broadcast notification to the user whose request was approved/declined
	if hub != nil && targetUserID > 0 {
		notification := chat.Frontend{
			Type:      "group_request_response",
			From:      adminUserID,
			To:        targetUserID,
			Username:  adminUsername,
			GroupID:   groupID,
			Content:   fmt.Sprintf("Your group join request has been %s", requestData.Status),
			Timestamp: time.Now(),
		}

		hub.Mutex.RLock()
		if client, ok := hub.Clients[targetUserID]; ok {
			select {
			case client.Send <- notification:
			default:
				// Skip if buffer full
			}
		}
		hub.Mutex.RUnlock()

		// Also notify the admin (refresh their pending list)
		adminNotif := chat.Frontend{
			Type:      "group_member_update",
			From:      targetUserID,
			To:        adminUserID,
			Username:  adminUsername,
			GroupID:   groupID,
			Content:   fmt.Sprintf("Join request %s", requestData.Status),
			Timestamp: time.Now(),
		}

		hub.Mutex.RLock()
		if client, ok := hub.Clients[adminUserID]; ok {
			select {
			case client.Send <- adminNotif:
			default:
			}
		}
		hub.Mutex.RUnlock()
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Join request %s successfully", requestData.Status),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
