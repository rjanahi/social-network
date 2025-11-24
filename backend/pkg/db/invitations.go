package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// GetInvitableUsers returns users who aren't already members of a group and match the search query
func GetInvitableUsers(db *sql.DB, groupID int, searchQuery string) ([]map[string]interface{}, error) {
	// Base query gets users who aren't members and don't have pending requests
	query := `
		SELECT u.id, u.username, u.email,
		       (SELECT status FROM group_invitations 
		        WHERE group_id = ? AND invitee_id = u.id AND (status = 'declined' OR status = 'cancelled')
		        ORDER BY created_at DESC LIMIT 1) as previous_status
		FROM users u
		WHERE u.id NOT IN (
			-- Users who are already members or have pending requests
			SELECT user_id FROM group_members WHERE group_id = ?
		) AND u.id NOT IN (
			-- Users who have pending invitations (but allow re-inviting those who declined or cancelled)
			SELECT invitee_id FROM group_invitations 
			WHERE group_id = ? AND status = 'pending'
		)`

	// Add search condition if provided
	params := []interface{}{groupID, groupID, groupID}
	if searchQuery != "" {
		query += ` AND (u.username LIKE ? OR u.email LIKE ?)`
		searchPattern := "%" + searchQuery + "%"
		params = append(params, searchPattern, searchPattern)
	}

	// Add limit and ordering
	query += ` ORDER BY u.username LIMIT 50`

	// Execute query
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email string
		var previousStatus sql.NullString

		if err := rows.Scan(&id, &username, &email, &previousStatus); err != nil {
			return nil, err
		}

		user := map[string]interface{}{
			"id":              id,
			"username":        username,
			"email":           email,
			"previous_status": previousStatus.String,
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return empty array if no users found
	if users == nil {
		users = []map[string]interface{}{}
	}

	return users, nil
}

// HasPendingInvitation checks if a user already has a pending invitation to a group
func HasPendingInvitation(tx *sql.Tx, groupID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND status = 'pending'
	`
	var count int
	err := tx.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// InsertGroupInvitationTx creates a group invitation within a transaction
func InsertGroupInvitationTx(tx *sql.Tx, groupID, inviterID, inviteeID int) error {
	// Try to insert a new pending invitation. If a UNIQUE constraint violation
	// occurs for (group_id, invitee_id), update the existing row to pending
	// and set the inviter_id/created_at accordingly. This prevents race
	// conditions and avoids returning a SQLite UNIQUE constraint error.
	query := `INSERT INTO group_invitations (group_id, inviter_id, invitee_id, status, created_at) VALUES (?, ?, ?, 'pending', CURRENT_TIMESTAMP)`
	_, err := tx.Exec(query, groupID, inviterID, inviteeID)
	if err == nil {
		return nil
	}

	// If the error is a UNIQUE constraint on group_id, invitee_id, attempt to
	// update the existing invitation row to pending instead of failing.
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		update := `UPDATE group_invitations SET status = 'pending', inviter_id = ?, created_at = CURRENT_TIMESTAMP WHERE group_id = ? AND invitee_id = ?`
		_, uerr := tx.Exec(update, inviterID, groupID, inviteeID)
		if uerr != nil {
			return fmt.Errorf("insert failed: %v; update also failed: %v", err, uerr)
		}
		return nil
	}

	return err
}

// GetSentInvitations returns invitations sent by a user, optionally filtered by group
func GetSentInvitations(db *sql.DB, userID int, groupID int) ([]map[string]interface{}, error) {
	query := `
		SELECT gi.id, gi.group_id, gi.invitee_id, gi.status, gi.created_at, 
		       g.title AS group_name, u.username AS invitee_name
		FROM group_invitations gi
		JOIN groups g ON gi.group_id = g.id
		JOIN users u ON gi.invitee_id = u.id
		WHERE gi.inviter_id = ?
	`
	params := []interface{}{userID}

	// Add group filter if provided
	if groupID > 0 {
		query += ` AND gi.group_id = ?`
		params = append(params, groupID)
	}

	query += ` ORDER BY gi.created_at DESC`

	// Execute query
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var invitations []map[string]interface{}
	for rows.Next() {
		var id, groupID, inviteeID int
		var status, groupName, inviteeName string
		var createdAt time.Time

		if err := rows.Scan(&id, &groupID, &inviteeID, &status, &createdAt, &groupName, &inviteeName); err != nil {
			return nil, err
		}

		invitation := map[string]interface{}{
			"id":           id,
			"group_id":     groupID,
			"invitee_id":   inviteeID,
			"status":       status,
			"group_name":   groupName,
			"invitee_name": inviteeName,
			"created_at":   createdAt.Format("2006-01-02 15:04:05"),
		}
		invitations = append(invitations, invitation)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return empty array if no invitations found
	if invitations == nil {
		invitations = []map[string]interface{}{}
	}

	return invitations, nil
}

// CancelInvitation cancels a pending invitation (sets status to 'cancelled')
func CancelInvitation(db *sql.DB, invitationID, inviterID int) error {
	// First check if this invitation belongs to the inviter
	query := `
		SELECT COUNT(*) FROM group_invitations 
		WHERE id = ? AND inviter_id = ? AND status = 'pending'
	`
	var count int
	err := db.QueryRow(query, invitationID, inviterID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("invitation not found or you don't have permission to cancel it")
	}

	// Update the invitation status to 'cancelled'
	update := `UPDATE group_invitations SET status = 'cancelled' WHERE id = ?`
	_, err = db.Exec(update, invitationID)
	return err
}
