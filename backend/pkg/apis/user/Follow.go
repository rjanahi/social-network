package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"socialnetwork/pkg/apis/chat"
)

// FollowUser handles follow requests
func FollowUser(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	followerID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get target user ID and privacy status
	var targetID int
	var isPrivate bool
	err := db.QueryRow("SELECT id, isPrivate FROM users WHERE username = ?", req.Username).Scan(&targetID, &isPrivate)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if already following
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM userFollow WHERE follower_id = ? AND following_id = ?", followerID, targetID).Scan(&count)
	if err == nil && count > 0 {
		http.Error(w, "Already following this user", http.StatusBadRequest)
		return
	}

	if isPrivate {
		// For private profiles, check for existing follow request
		var existingReqStatus string
		err = db.QueryRow("SELECT status FROM follow_requests WHERE requester_id = ? AND target_id = ?", followerID, targetID).Scan(&existingReqStatus)
		if err == nil {
			// row exists
			if existingReqStatus == "pending" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Follow request already pending",
					"status":  "pending",
				})
				return
			}
			// If previously declined or accepted, allow re-opening by setting to pending
			_, err = db.Exec("UPDATE follow_requests SET status = 'pending', updated_at = CURRENT_TIMESTAMP WHERE requester_id = ? AND target_id = ?", followerID, targetID)
			if err != nil {
				http.Error(w, "Failed to update follow request", http.StatusInternalServerError)
				return
			}
		} else if err == sql.ErrNoRows {
			// Not found — insert a new pending request
			_, err = db.Exec(`
				INSERT INTO follow_requests (requester_id, target_id, status) 
				VALUES (?, ?, 'pending')`, followerID, targetID)
			if err != nil {
				http.Error(w, "Failed to send follow request", http.StatusInternalServerError)
				return
			}
		} else {
			// Some DB error
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if err != nil {
			http.Error(w, "Failed to send follow request", http.StatusInternalServerError)
			return
		}

		// Broadcast follow request notification to the target user via WebSocket
		if hub != nil {
			var requesterUsername string
			_ = db.QueryRow("SELECT username FROM users WHERE id = ?", followerID).Scan(&requesterUsername)
			notif := chat.Frontend{
				Type:      "follow_request",
				From:      followerID,
				To:        targetID,
				Username:  requesterUsername,
				Content:   fmt.Sprintf("%s wants to follow you", requesterUsername),
				Timestamp: time.Now(),
			}
			hub.Mutex.RLock()
			if client, ok := hub.Clients[targetID]; ok {
				select {
				case client.Send <- notif:
				default:
				}
			}
			hub.Mutex.RUnlock()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Follow request sent",
			"status":  "pending",
		})
	} else {
		// For public profiles, follow immediately
		_, err = db.Exec("INSERT INTO userFollow (follower_id, following_id) VALUES (?, ?)", followerID, targetID)
		if err != nil {
			http.Error(w, "Failed to follow user", http.StatusInternalServerError)
			return
		}

		// Broadcast immediate follow notification to the target user
		if hub != nil {
			var followerUsername string
			_ = db.QueryRow("SELECT username FROM users WHERE id = ?", followerID).Scan(&followerUsername)
			notif := chat.Frontend{
				Type:      "new_follower",
				From:      followerID,
				To:        targetID,
				Username:  followerUsername,
				Content:   fmt.Sprintf("%s started following you", followerUsername),
				Timestamp: time.Now(),
			}
			hub.Mutex.RLock()
			if client, ok := hub.Clients[targetID]; ok {
				select {
				case client.Send <- notif:
				default:
				}
			}
			hub.Mutex.RUnlock()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Now following user",
			"status":  "following",
		})
	}
}

// UnfollowUser handles unfollowing
func UnfollowUser(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	followerID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get target user ID
	var targetID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&targetID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Remove from userFollow table
	_, err = db.Exec("DELETE FROM userFollow WHERE follower_id = ? AND following_id = ?", followerID, targetID)
	if err != nil {
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	// Broadcast unfollow/update notifications via WebSocket so UI updates in real-time
	if hub != nil {
		var followerUsername string
		_ = db.QueryRow("SELECT username FROM users WHERE id = ?", followerID).Scan(&followerUsername)

		// Notify the target that someone unfollowed them
		unfollowNotif := chat.Frontend{
			Type:      "unfollow_update",
			From:      followerID,
			To:        targetID,
			Username:  followerUsername,
			Content:   fmt.Sprintf("%s unfollowed you", followerUsername),
			Timestamp: time.Now(),
		}
		hub.Mutex.RLock()
		if client, ok := hub.Clients[targetID]; ok {
			select {
			case client.Send <- unfollowNotif:
			default:
			}
		}
		hub.Mutex.RUnlock()

		// Notify the actor (follower) that their follow state changed (useful for updating lists)
		followUpdate := chat.Frontend{
			Type:      "follow_update",
			From:      followerID,
			To:        followerID,
			Username:  followerUsername,
			Content:   fmt.Sprintf("You unfollowed user %d", targetID),
			Timestamp: time.Now(),
		}
		hub.Mutex.RLock()
		if client, ok := hub.Clients[followerID]; ok {
			select {
			case client.Send <- followUpdate:
			default:
			}
		}
		hub.Mutex.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Unfollowed user",
	})
}

// HandleFollowRequest handles accepting or declining follow requests
func HandleFollowRequest(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	targetID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		RequesterID int    `json:"requester_id"`
		Action      string `json:"action"` // "accept" or "decline"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate requester id and action
	if req.RequesterID <= 0 {
		http.Error(w, "Invalid requester_id", http.StatusBadRequest)
		return
	}
	if req.Action != "accept" && req.Action != "decline" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Ensure there is a pending follow request before acting
	var existingStatus string
	err := db.QueryRow("SELECT status FROM follow_requests WHERE requester_id = ? AND target_id = ?", req.RequesterID, targetID).Scan(&existingStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No pending follow request found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingStatus != "pending" {
		http.Error(w, "Follow request is not pending", http.StatusBadRequest)
		return
	}

	if req.Action == "accept" {
		// Add to userFollow table. Use INSERT OR IGNORE to avoid UNIQUE constraint
		// failures if the follow relation already exists.
		_, err := db.Exec("INSERT OR IGNORE INTO userFollow (follower_id, following_id) VALUES (?, ?)", req.RequesterID, targetID)
		if err != nil {
			http.Error(w, "Failed to accept follow request", http.StatusInternalServerError)
			return
		}

		// Update request status and check affected rows
		res, err := db.Exec("UPDATE follow_requests SET status = 'accepted', updated_at = CURRENT_TIMESTAMP WHERE requester_id = ? AND target_id = ? AND status = 'pending'", req.RequesterID, targetID)
		if err != nil {
			http.Error(w, "Failed to update follow request", http.StatusInternalServerError)
			return
		}
		if rr, _ := res.RowsAffected(); rr == 0 {
			http.Error(w, "No pending follow request to accept", http.StatusBadRequest)
			return
		}

		// Notify requester via WebSocket (response) and notify target that they have a new follower
		if hub != nil {
			var targetUsername string
			_ = db.QueryRow("SELECT username FROM users WHERE id = ?", targetID).Scan(&targetUsername)
			// response to requester
			respNotif := chat.Frontend{
				Type:      "follow_request_response",
				From:      targetID,
				To:        req.RequesterID,
				Username:  targetUsername,
				Content:   fmt.Sprintf("%s accepted your follow request", targetUsername),
				Timestamp: time.Now(),
			}
			hub.Mutex.RLock()
			if client, ok := hub.Clients[req.RequesterID]; ok {
				select {
				case client.Send <- respNotif:
				default:
				}
			}
			hub.Mutex.RUnlock()
			// Also notify the target (owner) that they have a new follower
			var followerUsername string
			_ = db.QueryRow("SELECT username FROM users WHERE id = ?", req.RequesterID).Scan(&followerUsername)
			followerNotif := chat.Frontend{
				Type:      "new_follower",
				From:      req.RequesterID,
				To:        targetID,
				Username:  followerUsername,
				Content:   fmt.Sprintf("%s started following you", followerUsername),
				Timestamp: time.Now(),
			}
			hub.Mutex.RLock()
			if client, ok := hub.Clients[targetID]; ok {
				select {
				case client.Send <- followerNotif:
				default:
				}
			}
			hub.Mutex.RUnlock()
		}

	} else if req.Action == "decline" {
		// Decline: update status
		res, err := db.Exec("UPDATE follow_requests SET status = 'declined', updated_at = CURRENT_TIMESTAMP WHERE requester_id = ? AND target_id = ? AND status = 'pending'", req.RequesterID, targetID)
		if err != nil {
			http.Error(w, "Failed to decline follow request", http.StatusInternalServerError)
			return
		}
		if rr, _ := res.RowsAffected(); rr == 0 {
			http.Error(w, "No pending follow request to decline", http.StatusBadRequest)
			return
		}
		// Notify requester
		if hub != nil {
			var targetUsername string
			_ = db.QueryRow("SELECT username FROM users WHERE id = ?", targetID).Scan(&targetUsername)
			notif := chat.Frontend{
				Type:      "follow_request_response",
				From:      targetID,
				To:        req.RequesterID,
				Username:  targetUsername,
				Content:   fmt.Sprintf("%s declined your follow request", targetUsername),
				Timestamp: time.Now(),
			}
			hub.Mutex.RLock()
			if client, ok := hub.Clients[req.RequesterID]; ok {
				select {
				case client.Send <- notif:
				default:
				}
			}
			hub.Mutex.RUnlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Follow request %sed", req.Action),
	})
}

// GetAllUsers returns a list of all users for following/browsing
func GetAllUsers(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	currentUserID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get search parameter
	searchQuery := r.URL.Query().Get("search")

	var query string
	var args []interface{}

	if searchQuery != "" {
		query = `
			SELECT u.id, u.username, u.firstname, u.lastname, u.email, u.date_of_birth, 
			       u.nickname, u.bio, u.avatar_url, u.isPrivate,
			       CASE WHEN uf.following_id IS NOT NULL THEN 1 ELSE 0 END as is_following,
			       CASE WHEN fr.id IS NOT NULL THEN fr.status ELSE NULL END as request_status
			FROM users u
			LEFT JOIN userFollow uf ON u.id = uf.following_id AND uf.follower_id = ?
			LEFT JOIN follow_requests fr ON u.id = fr.target_id AND fr.requester_id = ? AND fr.status = 'pending'
			WHERE u.id != ? AND (u.username LIKE ? OR u.firstname LIKE ? OR u.lastname LIKE ?)
			ORDER BY u.firstname, u.lastname`

		searchPattern := "%" + searchQuery + "%"
		args = []interface{}{currentUserID, currentUserID, currentUserID, searchPattern, searchPattern, searchPattern}
	} else {
		query = `
			SELECT u.id, u.username, u.firstname, u.lastname, u.email, u.date_of_birth,
			       u.nickname, u.bio, u.avatar_url, u.isPrivate,
			       CASE WHEN uf.following_id IS NOT NULL THEN 1 ELSE 0 END as is_following,
			       CASE WHEN fr.id IS NOT NULL THEN fr.status ELSE NULL END as request_status
			FROM users u
			LEFT JOIN userFollow uf ON u.id = uf.following_id AND uf.follower_id = ?
			LEFT JOIN follow_requests fr ON u.id = fr.target_id AND fr.requester_id = ? AND fr.status = 'pending'
			WHERE u.id != ?
			ORDER BY u.firstname, u.lastname`

		args = []interface{}{currentUserID, currentUserID, currentUserID}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var user struct {
			ID            int            `json:"id"`
			Username      string         `json:"username"`
			Fname         string         `json:"firstname"`
			Lname         string         `json:"lastname"`
			Email         sql.NullString `json:"email"`
			DateOfBirth   sql.NullString `json:"date_of_birth"`
			Nickname      sql.NullString `json:"nickname"`
			Bio           sql.NullString `json:"bio"`
			Avatar        string         `json:"avatar"`
			IsPrivate     bool           `json:"isPrivate"`
			IsFollowing   bool           `json:"isFollowing"`
			RequestStatus sql.NullString `json:"requestStatus"`
		}

		err := rows.Scan(&user.ID, &user.Username, &user.Fname, &user.Lname,
			&user.Email, &user.DateOfBirth, &user.Nickname, &user.Bio,
			&user.Avatar, &user.IsPrivate, &user.IsFollowing, &user.RequestStatus)
		if err != nil {
			continue
		}

		userMap := map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"firstname":   user.Fname,
			"lastname":    user.Lname,
			"avatar":      user.Avatar,
			"isPrivate":   user.IsPrivate,
			"isFollowing": user.IsFollowing,
		}

		// Only include detailed info if user has public profile OR current user is following them
		if !user.IsPrivate || user.IsFollowing {
			if user.Email.Valid {
				userMap["email"] = user.Email.String
			}
			if user.DateOfBirth.Valid {
				userMap["dateOfBirth"] = user.DateOfBirth.String
			}
			if user.Nickname.Valid && user.Nickname.String != "" {
				userMap["nickname"] = user.Nickname.String
			}
			if user.Bio.Valid && user.Bio.String != "" {
				userMap["bio"] = user.Bio.String
			}
		}

		if user.RequestStatus.Valid {
			userMap["requestStatus"] = user.RequestStatus.String
		}

		users = append(users, userMap)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}

// GetFollowStatus returns the follow status between two users
func GetFollowStatus(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	currentUserID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username parameter required", http.StatusBadRequest)
		return
	}

	// Get target user ID
	var targetID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&targetID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if following
	var isFollowing bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM userFollow WHERE follower_id = ? AND following_id = ?", currentUserID, targetID).Scan(&isFollowing)
	if err != nil {
		isFollowing = false
	}

	// Check for pending request
	var requestStatus string
	err = db.QueryRow("SELECT status FROM follow_requests WHERE requester_id = ? AND target_id = ?", currentUserID, targetID).Scan(&requestStatus)
	if err != nil {
		requestStatus = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"isFollowing":   isFollowing,
		"requestStatus": requestStatus,
	})
}
