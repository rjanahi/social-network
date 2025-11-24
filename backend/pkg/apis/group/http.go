package group

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"socialnetwork/pkg/apis/chat"
	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
)

// CreateGroup handles group creation
func CreateGroup(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
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
	var groupData CreateGroupRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&groupData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Validate group fields
	if groupData.Title == "" || groupData.Description == "" {
		http.Error(w, "Title and Description cannot be empty.", http.StatusBadRequest)
		return
	}

	// Insert the group into the database
	groupID, createdAt, err := database.InsertGroup(db, userID, groupData.Title, groupData.Description)
	if err != nil {
		fmt.Println("Error inserting group:", err)
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Automatically add creator as admin member
	err = database.InsertGroupMember(db, int(groupID), userID, "accepted", true)
	if err != nil {
		fmt.Println("Error adding creator as group admin:", err)
		http.Error(w, "Failed to add creator to group", http.StatusInternalServerError)
		return
	}

	// Get username for WebSocket broadcast
	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

	// Broadcast new group creation notification to ALL online users
	if hub != nil {
		notification := chat.Frontend{
			Type:      "new_group_created",
			From:      userID,
			Username:  username,
			GroupID:   int(groupID),
			Content:   groupData.Title,
			Timestamp: time.Now(),
		}

		// Broadcast to ALL connected users so everyone sees the new group
		hub.Mutex.RLock()
		for _, client := range hub.Clients {
			select {
			case client.Send <- notification:
			default:
				// Skip if buffer full
			}
		}
		hub.Mutex.RUnlock()
	}

	// Send success response
	response := map[string]interface{}{
		"success":   true,
		"message":   "Group created successfully.",
		"groupID":   groupID,
		"createdAt": createdAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetGroups returns all groups
func GetGroups(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	groups, err := database.GetAllGroupsWithMembership(db, userID)
	if err != nil {
		fmt.Println("Error retrieving groups:", err)
		http.Error(w, "Failed to retrieve groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// GetGroupDetails returns detailed information about a specific group
func GetGroupDetails(db *sql.DB, w http.ResponseWriter, r *http.Request, groupIDStr string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	group, err := database.GetGroupByID(db, groupID)
	if err != nil {
		fmt.Println("Error retrieving group:", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check if user is a member
	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil {
		fmt.Println("Error checking group membership:", err)
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}

	// Get group members
	members, err := database.GetGroupMembers(db, groupID)
	if err != nil {
		fmt.Println("Error retrieving group members:", err)
		http.Error(w, "Failed to retrieve group members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"group":    group,
		"isMember": isMember,
		"members":  members,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InviteUserToGroup handles group invitations
func InviteUserToGroup(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
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
	var inviteData InviteUserRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&inviteData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Check if inviter is a member of the group
	isMember, err := database.IsGroupMember(db, inviteData.GroupID, inviterID)
	if err != nil || !isMember {
		http.Error(w, "You must be a member to invite users", http.StatusForbidden)
		return
	}

	// Check if user is already a member or has pending invitation
	existingStatus, err := database.GetGroupMemberStatus(db, inviteData.GroupID, inviteData.UserID)
	if err == nil && existingStatus != "" {
		http.Error(w, "User is already a member or has a pending invitation", http.StatusConflict)
		return
	}

	// Insert invitation
	err = database.InsertGroupInvitation(db, inviteData.GroupID, inviterID, inviteData.UserID)
	if err != nil {
		fmt.Println("Error creating invitation:", err)
		http.Error(w, "Failed to send invitation", http.StatusInternalServerError)
		return
	}

	// Try to fetch the newly created invitation id
	var invitationID int
	// column name is invitee_id in the schema (not 'invited_id')
	if err := db.QueryRow(`SELECT id FROM group_invitations WHERE group_id = ? AND inviter_id = ? AND invitee_id = ? ORDER BY created_at DESC LIMIT 1`, inviteData.GroupID, inviterID, inviteData.UserID).Scan(&invitationID); err != nil {
		// Log but continue; invitation may have been inserted but we couldn't fetch id
		fmt.Printf("Warning: could not fetch invitation id after insert: %v\n", err)
		invitationID = 0
	}

	// Broadcast a group invitation notification to the invitee
	if hub != nil {
		var inviterUsername string
		_ = db.QueryRow("SELECT username FROM users WHERE id = ?", inviterID).Scan(&inviterUsername)

		notif := chat.Frontend{
			Type:      "group_invitation",
			From:      inviterID,
			To:        inviteData.UserID,
			Username:  inviterUsername,
			GroupID:   inviteData.GroupID,
			Content:   fmt.Sprintf("%s invited you to join a group", inviterUsername),
			Timestamp: time.Now(),
		}
		// Only attach InvitationID if we successfully retrieved it
		if invitationID > 0 {
			notif.InvitationID = invitationID
		}

		hub.Mutex.RLock()
		if client, ok := hub.Clients[inviteData.UserID]; ok {
			select {
			case client.Send <- notif:
			default:
			}
		}
		hub.Mutex.RUnlock()
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Invitation sent successfully.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RespondToGroupInvitation handles accepting/declining group invitations
func RespondToGroupInvitation(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
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
	var responseData ResponseInvitationRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&responseData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	if responseData.Status != "accepted" && responseData.Status != "declined" {
		http.Error(w, "Invalid status. Must be 'accepted' or 'declined'", http.StatusBadRequest)
		return
	}

	// Get group ID and inviter ID before updating
	var groupID, inviterID int
	err := db.QueryRow("SELECT group_id, inviter_id FROM group_invitations WHERE id = ?", responseData.InvitationID).Scan(&groupID, &inviterID)
	if err != nil {
		fmt.Println("Error getting invitation details:", err)
		http.Error(w, "Invitation not found", http.StatusNotFound)
		return
	}

	// Update invitation status
	err = database.UpdateGroupInvitationStatus(db, responseData.InvitationID, userID, responseData.Status)
	if err != nil {
		fmt.Println("Error updating invitation:", err)
		http.Error(w, "Failed to respond to invitation", http.StatusInternalServerError)
		return
	}

	// If accepted, add user to group members
	if responseData.Status == "accepted" {
		err = database.InsertGroupMember(db, groupID, userID, "accepted", false)
		if err != nil {
			fmt.Println("Error adding user to group:", err)
			http.Error(w, "Failed to join group", http.StatusInternalServerError)
			return
		}
	}

	// Broadcast WebSocket notification
	if hub != nil {
		var username string
		db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

		// Notify the user who responded
		userNotif := chat.Frontend{
			Type:         "invitation_response",
			From:         userID,
			To:           userID,
			Username:     username,
			GroupID:      groupID,
			Content:      fmt.Sprintf("You %s the group invitation", responseData.Status),
			InvitationID: responseData.InvitationID,
			Timestamp:    time.Now(),
		}

		hub.Mutex.RLock()
		if client, ok := hub.Clients[userID]; ok {
			select {
			case client.Send <- userNotif:
			default:
			}
		}
		hub.Mutex.RUnlock()

		// Notify the inviter
		inviterNotif := chat.Frontend{
			Type:         "invitation_response",
			From:         userID,
			To:           inviterID,
			Username:     username,
			GroupID:      groupID,
			Content:      fmt.Sprintf("%s %s your group invitation", username, responseData.Status),
			InvitationID: responseData.InvitationID,
			Timestamp:    time.Now(),
		}

		hub.Mutex.RLock()
		if client, ok := hub.Clients[inviterID]; ok {
			select {
			case client.Send <- inviterNotif:
			default:
			}
		}
		hub.Mutex.RUnlock()

		// If accepted, broadcast to all group members
		if responseData.Status == "accepted" {
			memberNotif := chat.Frontend{
				Type:      "group_member_update",
				From:      userID,
				Username:  username,
				GroupID:   groupID,
				Content:   fmt.Sprintf("%s joined the group", username),
				Timestamp: time.Now(),
			}

			// Broadcast to ALL connected users
			hub.Mutex.RLock()
			for _, client := range hub.Clients {
				select {
				case client.Send <- memberNotif:
				default:
				}
			}
			hub.Mutex.RUnlock()
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Invitation %s successfully.", responseData.Status),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RequestToJoinGroup handles join requests
func RequestToJoinGroup(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
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
	var joinData JoinGroupRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&joinData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// First, check if we need to handle a special case (decline->pending)
	var statusCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM group_members 
		WHERE group_id = ? AND user_id = ? AND status = 'declined'`,
		joinData.GroupID, userID).Scan(&statusCount)

	if err != nil {
		fmt.Println("Error checking declined status:", err)
	} else if statusCount > 0 {
		// This user was previously declined, update status to pending
		fmt.Println("Found declined status, updating to pending")
		_, err := db.Exec(`UPDATE group_members 
			SET status = 'pending', joined_at = CURRENT_TIMESTAMP 
			WHERE group_id = ? AND user_id = ?`,
			joinData.GroupID, userID)
		if err != nil {
			fmt.Println("Error updating status:", err)
			http.Error(w, "Failed to send join request", http.StatusInternalServerError)
			return
		}

		// Broadcast join request notification
		sendJoinRequestNotification(db, hub, userID, joinData.GroupID)

		// Success response
		response := map[string]interface{}{
			"success": true,
			"message": "Join request sent successfully. Waiting for group admin approval.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Normal flow - check existing status
	var status string
	err = db.QueryRow(`SELECT status FROM group_members 
		WHERE group_id = ? AND user_id = ?`,
		joinData.GroupID, userID).Scan(&status)

	if err == nil {
		// Record exists, handle based on current status
		if status == "accepted" {
			http.Error(w, "You are already a member of this group", http.StatusConflict)
			return
		} else if status == "pending" {
			http.Error(w, "You already have a pending request for this group", http.StatusConflict)
			return
		}
		// Any other status should have been caught by the declined check above
	} else if err != sql.ErrNoRows {
		// Unexpected error
		fmt.Println("Database error:", err)
		http.Error(w, "Error checking membership status", http.StatusInternalServerError)
		return
	}

	// No existing record or error was sql.ErrNoRows, create a new pending request
	_, err = db.Exec(`INSERT INTO group_members (group_id, user_id, status, is_admin) 
		VALUES (?, ?, 'pending', 0)`,
		joinData.GroupID, userID)

	if err != nil {
		fmt.Println("Error creating join request:", err)
		http.Error(w, "Failed to send join request", http.StatusInternalServerError)
		return
	}

	// Broadcast join request notification
	sendJoinRequestNotification(db, hub, userID, joinData.GroupID)

	response := map[string]interface{}{
		"success": true,
		"message": "Join request sent successfully. Waiting for group admin approval.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LeaveGroup handles a user leaving a group
func LeaveGroup(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
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
	var leaveData struct {
		GroupID int `json:"group_id"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&leaveData); err != nil {
		fmt.Println("JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Check if user is the group creator
	var creatorID int
	err := db.QueryRow("SELECT creator_id FROM groups WHERE id = ?", leaveData.GroupID).Scan(&creatorID)
	if err != nil {
		fmt.Println("Error checking group creator:", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	if creatorID == userID {
		http.Error(w, "Group creator cannot leave the group", http.StatusForbidden)
		return
	}

	// Remove user from group_members
	result, err := db.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", leaveData.GroupID, userID)
	if err != nil {
		fmt.Println("Error leaving group:", err)
		http.Error(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "You are not a member of this group", http.StatusBadRequest)
		return
	}

	// Broadcast WebSocket notification to all users
	if hub != nil {
		var username string
		db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

		notification := chat.Frontend{
			Type:      "group_member_left",
			From:      userID,
			Username:  username,
			GroupID:   leaveData.GroupID,
			Content:   fmt.Sprintf("%s left the group", username),
			Timestamp: time.Now(),
		}

		// Broadcast to ALL connected users
		hub.Mutex.RLock()
		for _, client := range hub.Clients {
			select {
			case client.Send <- notification:
			default:
			}
		}
		hub.Mutex.RUnlock()
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Successfully left the group",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendJoinRequestNotification broadcasts WebSocket notification for join requests
func sendJoinRequestNotification(db *sql.DB, hub *chat.Hub, userID int, groupID int) {
	if hub == nil {
		return
	}

	// Get username and group creator
	var username string
	var creatorID int
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	db.QueryRow("SELECT creator_id FROM groups WHERE id = ?", groupID).Scan(&creatorID)

	// Notify the requesting user
	userNotif := chat.Frontend{
		Type:      "join_request_sent",
		From:      userID,
		To:        userID,
		Username:  username,
		GroupID:   groupID,
		Content:   "Join request pending approval",
		Timestamp: time.Now(),
	}

	hub.Mutex.RLock()
	if client, ok := hub.Clients[userID]; ok {
		select {
		case client.Send <- userNotif:
		default:
		}
	}
	hub.Mutex.RUnlock()

	// Notify the group admin/creator
	adminNotif := chat.Frontend{
		Type:      "group_join_request",
		From:      userID,
		To:        creatorID,
		Username:  username,
		GroupID:   groupID,
		Content:   fmt.Sprintf("%s requested to join your group", username),
		Timestamp: time.Now(),
	}

	// Try to include the join request ID so admin can accept/decline
	var requestID int
	_ = db.QueryRow(`SELECT id FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'pending' ORDER BY joined_at DESC LIMIT 1`, groupID, userID).Scan(&requestID)
	if requestID > 0 {
		adminNotif.RequestID = requestID
		userNotif.RequestID = requestID
	}

	hub.Mutex.RLock()
	if client, ok := hub.Clients[creatorID]; ok {
		select {
		case client.Send <- adminNotif:
		default:
		}
	}
	hub.Mutex.RUnlock()
}

// GetUserGroups returns groups where user is a member
func GetUserGroups(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	groups, err := database.GetUserGroups(db, userID)
	if err != nil {
		fmt.Println("Error retrieving user groups:", err)
		http.Error(w, "Failed to retrieve groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// GetGroupInvitations returns pending invitations for the user
func GetGroupInvitations(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	invitations, err := database.GetUserGroupInvitations(db, userID)
	if err != nil {
		fmt.Println("Error retrieving invitations:", err)
		http.Error(w, "Failed to retrieve invitations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitations)
}

func GetGroupPosts(db *sql.DB, w http.ResponseWriter, r *http.Request, groupIDStr string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil || !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	posts, err := database.GetGroupPosts(db, groupID, userID)
	if err != nil {
		fmt.Println("GetGroupPosts error:", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func HandleGetGroupPostComments(db *sql.DB, w http.ResponseWriter, r *http.Request, groupIDStr, postIDStr string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	postID, err := database.ParseID(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil || !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Fetch the post details
	var post map[string]interface{}
	var postUserID int
	var title, content, username, firstname, lastname, avatarURL, imgOrgif string
	var createdAt time.Time

	err = db.QueryRow(`
	       SELECT gp.id, gp.user_id, gp.title, gp.content, gp.imgOrgif, gp.created_at, u.username, u.firstname, u.lastname, u.avatar_url
	       FROM group_posts gp
	       JOIN users u ON gp.user_id = u.id
	       WHERE gp.id = ? AND gp.group_id = ?
       `, postID, groupID).Scan(&postID, &postUserID, &title, &content, &imgOrgif, &createdAt, &username, &firstname, &lastname, &avatarURL)

	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	post = map[string]interface{}{
		"id":         postID,
		"userID":     postUserID,
		"username":   username,
		"firstname":  firstname,
		"lastname":   lastname,
		"avatar_url": avatarURL,
		"title":      title,
		"content":    content,
		"image":      imgOrgif,
		"imgOrgif":   imgOrgif,
		"createdAt":  createdAt.Format("2006-01-02 15:04:05"),
		"type":       "group",
	}

	comments, err := database.GetGroupPostComments(db, postID)
	if err == sql.ErrNoRows {
		comments = []map[string]interface{}{}
	} else if err != nil {
		http.Error(w, "failed to load comments", http.StatusInternalServerError)
		return
	}

	// Return both post and comments
	response := map[string]interface{}{
		"post":     post,
		"comments": comments,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func ToggleGroupPostLike(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := r.URL.Query().Get("groupId")
	postIDStr := r.URL.Query().Get("postId")
	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid groupId", http.StatusBadRequest)
		return
	}
	postID, err := database.ParseID(postIDStr)
	if err != nil {
		http.Error(w, "Invalid postId", http.StatusBadRequest)
		return
	}

	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil || !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if err := database.AddGroupPostLike(db, postID, userID); err != nil {
			http.Error(w, "Failed to like", http.StatusInternalServerError)
			return
		}
	} else {
		if err := database.RemoveGroupPostLike(db, postID, userID); err != nil {
			http.Error(w, "Failed to unlike", http.StatusInternalServerError)
			return
		}
	}

	// Get username for WebSocket broadcast
	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

	// Get updated like/dislike counts
	likesCount, dislikesCount, err := database.GetGroupPostLikeCounts(db, postID)
	if err != nil {
		http.Error(w, "Failed to fetch counts", http.StatusInternalServerError)
		return
	}

	// Broadcast like update to all group members
	if hub != nil {
		// Prepare count data as JSON
		countsData := map[string]interface{}{
			"post_id":        postID,
			"group_id":       groupID,
			"likes_count":    likesCount,
			"dislikes_count": dislikesCount,
		}
		countsJSON, _ := json.Marshal(countsData)

		likeNotification := chat.Frontend{
			Type:      "group_post_like_update",
			From:      userID,
			Username:  username,
			GroupID:   groupID,
			PostId:    postID,
			Content:   string(countsJSON),
			Timestamp: time.Now(),
		}

		// Get group member IDs
		memberIDs, err := getGroupMemberIDs(db, groupID)
		if err == nil {
			hub.Mutex.RLock()
			for _, memberID := range memberIDs {
				if client, ok := hub.Clients[memberID]; ok {
					select {
					case client.Send <- likeNotification:
					default:
						// Skip if buffer full
					}
				}
			}
			hub.Mutex.RUnlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"likes":    likesCount,
		"dislikes": dislikesCount,
	})
}

func AddGroupPostComment(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	groupIDStr := r.FormValue("group_id")
	groupPostIDStr := r.FormValue("group_post_id")
	content := strings.TrimSpace(r.FormValue("content"))
	groupID, err := database.ParseID(groupIDStr)
	groupPostID, err2 := database.ParseID(groupPostIDStr)
	if err != nil || err2 != nil || groupID <= 0 || groupPostID <= 0 {
		http.Error(w, "Invalid group or post ID", http.StatusBadRequest)
		return
	}

	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil || !isMember {
		http.Error(w, "You must be an accepted member to comment", http.StatusForbidden)
		return
	}

	// Handle optional image/GIF upload
	imgOrgif := ""
	file, header, err := r.FormFile("imgOrgif")
	if err == nil && header != nil && header.Filename != "" {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".gif" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			http.Error(w, "Image must be a GIF, PNG, or JPG.", http.StatusBadRequest)
			return
		}
		uploadDir := "../frontend-next/public/img/group_comments"
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		fname := strings.ReplaceAll(header.Filename, " ", "-")
		dstFile, err := os.Create(filepath.Join(uploadDir, fname))
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		defer dstFile.Close()
		if _, err := io.Copy(dstFile, file); err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		imgOrgif = "/img/group_comments/" + fname
	}

	commentID, createdAt, err := database.InsertGroupPostComment(db, groupPostID, userID, content, imgOrgif)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	var username string
	_ = db.QueryRow(`SELECT username FROM users WHERE id=?`, userID).Scan(&username)

	// Broadcast comment notification to all group members
	if hub != nil {
		// Fetch the full comment data for broadcast
		comments, err := database.GetGroupPostComments(db, groupPostID)
		var newComment map[string]interface{}
		if err == nil {
			for _, c := range comments {
				if cid, ok := c["id"].(int); ok && cid == int(commentID) {
					newComment = c
					break
				}
			}
		}
		commentJSON, _ := json.Marshal(newComment)
		commentNotification := chat.Frontend{
			Type:      "group_post_comment",
			From:      userID,
			Username:  username,
			GroupID:   groupID,
			PostId:    groupPostID,
			CommentId: int(commentID),
			Content:   string(commentJSON),
			Timestamp: time.Now(),
		}

		// Get group member IDs
		memberIDs, err := getGroupMemberIDs(db, groupID)
		if err == nil {
			hub.Mutex.RLock()
			for _, memberID := range memberIDs {
				if client, ok := hub.Clients[memberID]; ok {
					select {
					case client.Send <- commentNotification:
					default:
						// Skip if buffer full
					}
				}
			}
			hub.Mutex.RUnlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"comment_id":    commentID,
		"group_post_id": groupPostID,
		"user_id":       userID,
		"username":      username,
		"content":       content,
		"imgOrgif":      imgOrgif,
		"created_at":    createdAt.Format("2006-01-02 15:04:05"),
	})
}

func DeleteGroupPost(db *sql.DB, w http.ResponseWriter, r *http.Request, groupIDStr, postIDStr string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	postID, err := database.ParseID(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Ensure the post belongs to this group & get owner
	postGroupID, ownerID, err := database.GetGroupPostOwnerAndGroup(db, postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	if postGroupID != groupID {
		http.Error(w, "Post does not belong to this group", http.StatusBadRequest)
		return
	}

	// Only author or group admin can delete
	isAdmin, err := database.IsGroupAdmin(db, groupID, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if userID != ownerID && !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := database.DeleteGroupPost(db, postID); err != nil {
		fmt.Println("DeleteGroupPost error:", err)
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func CreateGroupPost(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := u.ValidateSession(db, r)
	if !ok || userID <= 0 {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	groupIDStr := r.FormValue("group_id")
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categories := r.Form["categories"]

	groupID, err := database.ParseID(groupIDStr)
	if err != nil || groupID <= 0 || title == "" || content == "" {
		http.Error(w, "group_id, title and content are required", http.StatusBadRequest)
		return
	}

	// Membership check
	var count int
	if err := db.QueryRow(`
        SELECT COUNT(*) 
        FROM group_members 
        WHERE group_id = ? AND user_id = ? AND status = 'accepted'
    `, groupID, userID).Scan(&count); err != nil {
		http.Error(w, "Membership check failed", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Forbidden: not a member of this group", http.StatusForbidden)
		return
	}

	// Handle optional image/GIF upload
	imgOrgif := ""
	file, header, err := r.FormFile("imgOrgif")
	if err == nil && header != nil && header.Filename != "" {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".gif" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			http.Error(w, "Image must be a GIF, PNG, or JPG.", http.StatusBadRequest)
			return
		}
		uploadDir := "../frontend-next/public/img/group_posts"
		_ = os.MkdirAll(uploadDir, os.ModePerm)
		fname := strings.ReplaceAll(header.Filename, " ", "-")
		dstFile, err := os.Create(filepath.Join(uploadDir, fname))
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		defer dstFile.Close()
		if _, err := io.Copy(dstFile, file); err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		imgOrgif = "/img/group_posts/" + fname
	}

	// Insert post with image path
	postID64, createdAt, err := database.InsertGroupPost(db, groupID, userID, title, content, imgOrgif)
	if err != nil {
		http.Error(w, "Failed to create group post", http.StatusInternalServerError)
		return
	}
	postID := int(postID64)

	// Attach categories (optional, non-fatal per-category errors)
	for _, raw := range categories {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		catID, err := database.GetCategoryID(db, name)
		if err != nil {
			fmt.Println("GetCategoryID error:", err)
			continue
		}
		if err := database.AddGroupPostCategory(db, postID, catID); err != nil {
			fmt.Println("AddGroupPostCategory error:", err)
			continue
		}
	}

	// Get username for WebSocket broadcast
	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

	// Broadcast new group post to all group members via WebSocket
	if hub != nil {
		// Fetch the full post data for broadcast
		posts, err := database.GetGroupPosts(db, groupID, userID)
		var newPost map[string]interface{}
		if err == nil {
			for _, p := range posts {
				if pid, ok := p["id"].(int); ok && pid == postID {
					newPost = p
					break
				}
			}
		}
		postJSON, _ := json.Marshal(newPost)
		postNotification := chat.Frontend{
			Type:      "new_groupPost",
			From:      userID,
			Username:  username,
			GroupID:   groupID,
			PostId:    postID,
			Content:   string(postJSON),
			Timestamp: time.Now(),
		}

		// Get group member IDs
		memberIDs, err := getGroupMemberIDs(db, groupID)
		if err == nil {
			hub.Mutex.RLock()
			for _, memberID := range memberIDs {
				if client, ok := hub.Clients[memberID]; ok {
					select {
					case client.Send <- postNotification:
					default:
						// Skip if buffer full
					}
				}
			}
			hub.Mutex.RUnlock()
		}
	}

	resp := map[string]any{
		"success":    true,
		"id":         postID,
		"group_id":   groupID,
		"user_id":    userID,
		"title":      title,
		"content":    content,
		"categories": categories,
		"imgOrgif":   imgOrgif,
		"created_at": createdAt.Format("2006-01-02 15:04:05"),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Helper function to get group member IDs
func getGroupMemberIDs(db *sql.DB, groupID int) ([]int, error) {
	rows, err := db.Query(`
        SELECT user_id 
        FROM group_members 
        WHERE group_id = ? AND status = 'accepted'
    `, groupID)
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

// ToggleGroupPostDislike mirrors ToggleGroupPostLike for 👎
func ToggleGroupPostDislike(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := r.URL.Query().Get("groupId")
	postIDStr := r.URL.Query().Get("postId")

	groupID, err := database.ParseID(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid groupId", http.StatusBadRequest)
		return
	}
	postID, err := database.ParseID(postIDStr)
	if err != nil {
		http.Error(w, "Invalid postId", http.StatusBadRequest)
		return
	}

	// must be accepted member
	isMember, err := database.IsGroupMember(db, groupID, userID)
	if err != nil || !isMember {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if err := database.AddGroupPostDislike(db, postID, userID); err != nil {
			http.Error(w, "Failed to dislike", http.StatusInternalServerError)
			return
		}
	} else {
		if err := database.RemoveGroupPostDislike(db, postID, userID); err != nil {
			http.Error(w, "Failed to undislike", http.StatusInternalServerError)
			return
		}
	}

	// Get username for WebSocket broadcast
	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

	// Get updated like/dislike counts
	likesCount, dislikesCount, err := database.GetGroupPostLikeCounts(db, postID)
	if err != nil {
		http.Error(w, "Failed to fetch counts", http.StatusInternalServerError)
		return
	}

	// Broadcast dislike update to all group members
	if hub != nil {
		// Prepare count data as JSON
		countsData := map[string]interface{}{
			"post_id":        postID,
			"group_id":       groupID,
			"likes_count":    likesCount,
			"dislikes_count": dislikesCount,
		}
		countsJSON, _ := json.Marshal(countsData)

		dislikeNotification := chat.Frontend{
			Type:      "group_post_like_update",
			From:      userID,
			Username:  username,
			GroupID:   groupID,
			PostId:    postID,
			Content:   string(countsJSON),
			Timestamp: time.Now(),
		}

		// Get group member IDs
		memberIDs, err := getGroupMemberIDs(db, groupID)
		if err == nil {
			hub.Mutex.RLock()
			for _, memberID := range memberIDs {
				if client, ok := hub.Clients[memberID]; ok {
					select {
					case client.Send <- dislikeNotification:
					default:
						// Skip if buffer full
					}
				}
			}
			hub.Mutex.RUnlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"likes":    likesCount,
		"dislikes": dislikesCount,
	})
}
