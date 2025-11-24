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

// GetInvitableUsers returns a list of users who aren't already members of the group
func GetInvitableUsers(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get group ID from query params
	groupIDStr := r.URL.Query().Get("groupId")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Validate session and get user ID
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Parse group ID
	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if user is a member of the group
	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil {
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}

	if !isMember {
		http.Error(w, "You must be a member to invite users", http.StatusForbidden)
		return
	}

	// Get search query if provided
	searchQuery := r.URL.Query().Get("q")

	// Get invitable users
	users, err := database.GetInvitableUsers(db, groupID, searchQuery)
	if err != nil {
		fmt.Println("Error getting invitable users:", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	// Send response with proper structure
	response := map[string]interface{}{
		"users": users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendBulkInvitations allows sending invitations to multiple users at once
func SendBulkInvitations(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session and get user ID
	inviterID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var inviteData BulkInviteRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&inviteData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if inviteData.GroupID <= 0 || len(inviteData.UserIDs) == 0 {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Check if inviter is a member of the group
	isMember, err := database.IsGroupMember(db, inviteData.GroupID, inviterID)
	if err != nil || !isMember {
		http.Error(w, "You must be a member to invite users", http.StatusForbidden)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Track successful and failed invitations
	successCount := 0
	failedInvitations := make([]map[string]interface{}, 0)

	// Process each user ID
	for _, userID := range inviteData.UserIDs {
		// Check if user is already a member (but not for pending invitations)
		existingStatus, err := database.GetGroupMemberStatus(db, inviteData.GroupID, userID)
		if err == nil && existingStatus != "" {
			// User is already a member or has a pending request
			failedInvitations = append(failedInvitations, map[string]interface{}{
				"user_id": userID,
				"reason":  "User is already a member or has a pending request",
			})
			continue
		}

		// Check for existing pending invitation
		hasInvitation, err := database.HasPendingInvitation(tx, inviteData.GroupID, userID)
		if err != nil {
			fmt.Printf("Error checking invitation for user %d: %v\n", userID, err)
			failedInvitations = append(failedInvitations, map[string]interface{}{
				"user_id": userID,
				"reason":  "Error checking existing invitations",
			})
			continue
		}

		if hasInvitation {
			failedInvitations = append(failedInvitations, map[string]interface{}{
				"user_id": userID,
				"reason":  "User already has a pending invitation",
			})
			continue
		}

		// First, check if there was a previously cancelled invitation
		// Try to update it if it exists
		err = database.UpdateCancelledInvitationStatus(tx, inviteData.GroupID, userID, inviterID)
		if err != nil {
			fmt.Printf("Error updating cancelled invitation for user %d: %v\n", userID, err)
			failedInvitations = append(failedInvitations, map[string]interface{}{
				"user_id": userID,
				"reason":  "Error updating previous invitation",
			})
			continue
		}

		// Check if there was a previous declined invitation
		hasDeclinedInvitation, err := database.HasDeclinedInvitation(db, inviteData.GroupID, userID)
		if err != nil {
			fmt.Printf("Error checking previous invitations for user %d: %v\n", userID, err)
			failedInvitations = append(failedInvitations, map[string]interface{}{
				"user_id": userID,
				"reason":  "Error checking previous invitations",
			})
			continue
		}

		// If there was a declined invitation, update its status
		if hasDeclinedInvitation {
			err = database.UpdatePreviousInvitationStatus(tx, inviteData.GroupID, userID, inviterID)
			if err != nil {
				fmt.Printf("Error updating previous invitation for user %d: %v\n", userID, err)
				failedInvitations = append(failedInvitations, map[string]interface{}{
					"user_id": userID,
					"reason":  "Error updating previous invitation",
				})
				continue
			}
		} else {
			// Insert new invitation
			err = database.InsertGroupInvitationTx(tx, inviteData.GroupID, inviterID, userID)
			if err != nil {
				fmt.Printf("Error creating invitation for user %d: %v\n", userID, err)
				failedInvitations = append(failedInvitations, map[string]interface{}{
					"user_id": userID,
					"reason":  "Error creating invitation",
				})
				continue
			}
		}

		successCount++
	}

	// Commit or rollback transaction
	if successCount > 0 {
		if err := tx.Commit(); err != nil {
			fmt.Println("Error committing transaction:", err)
			http.Error(w, "Failed to save invitations", http.StatusInternalServerError)
			return
		}

		// Send WebSocket notifications to invited users
		if hub != nil {
			var inviterUsername string
			db.QueryRow("SELECT username FROM users WHERE id = ?", inviterID).Scan(&inviterUsername)

			for _, userID := range inviteData.UserIDs {
				// Only notify successfully invited users
				wasSuccessful := true
				for _, failed := range failedInvitations {
					if failed["user_id"] == userID {
						wasSuccessful = false
						break
					}
				}

				if wasSuccessful {
					notification := chat.Frontend{
						Type:      "group_invitation",
						From:      inviterID,
						To:        userID,
						Username:  inviterUsername,
						GroupID:   inviteData.GroupID,
						Content:   fmt.Sprintf("%s invited you to join a group", inviterUsername),
						Timestamp: time.Now(),
					}

					hub.Mutex.RLock()
					if client, ok := hub.Clients[userID]; ok {
						select {
						case client.Send <- notification:
						default:
						}
					}
					hub.Mutex.RUnlock()
				}
			}
		}
	} else {
		if err := tx.Rollback(); err != nil {
			fmt.Println("Error rolling back transaction:", err)
		}
	}

	// Send response
	response := map[string]interface{}{
		"success":            successCount > 0,
		"success_count":      successCount,
		"failed_count":       len(failedInvitations),
		"failed_invitations": failedInvitations,
		"message":            fmt.Sprintf("Successfully sent %d invitations", successCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSentInvitations returns invitations sent by the current user
func GetSentInvitations(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	// Get group ID from query params if provided
	groupIDStr := r.URL.Query().Get("groupId")
	var groupID int
	var err error
	if groupIDStr != "" {
		groupID, err = database.ParseID(groupIDStr)
		if err != nil {
			http.Error(w, "Invalid group ID", http.StatusBadRequest)
			return
		}

		// Check if user is a member of the group
		isMember, err := database.IsGroupMember(db, groupID, userID)
		if err != nil || !isMember {
			http.Error(w, "You must be a member to view invitations", http.StatusForbidden)
			return
		}
	}

	// Get sent invitations
	invitations, err := database.GetSentInvitations(db, userID, groupID)
	if err != nil {
		fmt.Println("Error getting sent invitations:", err)
		http.Error(w, "Failed to get invitations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitations)
}

// CancelInvitation cancels a pending invitation
func CancelInvitation(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session and get user ID
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var cancelData CancelInvitationRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&cancelData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Cancel invitation
	err := database.CancelInvitation(db, cancelData.InvitationID, userID)
	if err != nil {
		fmt.Println("Error canceling invitation:", err)
		http.Error(w, "Failed to cancel invitation", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Invitation canceled successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
