package database

import (
	"database/sql"
	"time"
)

// GetPendingGroupJoinRequests retrieves all pending join requests for groups where the user is an admin
func GetPendingGroupJoinRequests(db *sql.DB, adminUserID int) ([]map[string]interface{}, error) {
	query := `
		SELECT gm.id, gm.group_id, gm.user_id, gm.status, gm.joined_at,
		       g.title as group_name, u.username
		FROM group_members gm
		JOIN groups g ON gm.group_id = g.id
		JOIN users u ON gm.user_id = u.id
		JOIN group_members admin_gm ON gm.group_id = admin_gm.group_id AND admin_gm.user_id = ? AND admin_gm.is_admin = 1
		WHERE gm.status = 'pending'
		ORDER BY gm.joined_at DESC
	`

	rows, err := db.Query(query, adminUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var id, groupID, userID int
		var status, groupName, username string
		var joinedAt time.Time

		err := rows.Scan(&id, &groupID, &userID, &status, &joinedAt, &groupName, &username)
		if err != nil {
			return nil, err
		}

		request := map[string]interface{}{
			"id":         id,
			"group_id":   groupID,
			"user_id":    userID,
			"status":     status,
			"group_name": groupName,
			"username":   username,
			"joined_at":  joinedAt.Format("2006-01-02 15:04:05"),
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// UpdateGroupMemberStatus updates the status of a group member (for approval/rejection)
func UpdateGroupMemberStatus(db *sql.DB, requestID int, adminUserID int, newStatus string) error {
	// First, verify that the admin is actually an admin of the group
	verifyQuery := `
		SELECT COUNT(*) 
		FROM group_members gm
		JOIN group_members request_gm ON gm.group_id = request_gm.group_id
		WHERE request_gm.id = ? 
		  AND gm.user_id = ? 
		  AND gm.is_admin = 1
	`
	var count int
	err := db.QueryRow(verifyQuery, requestID, adminUserID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows // Admin not found or not authorized
	}

	// Update the status
	updateQuery := `UPDATE group_members SET status = ? WHERE id = ?`
	_, err = db.Exec(updateQuery, newStatus, requestID)
	return err
}

// UpdateGroupMemberStatus updates the status of a group member by group ID and user ID
// This is used when a declined user wants to reapply
func UpdateGroupMemberStatusByGroupAndUser(db *sql.DB, groupID int, userID int, newStatus string) error {
	query := `UPDATE group_members SET status = ?, joined_at = CURRENT_TIMESTAMP WHERE group_id = ? AND user_id = ?`
	_, err := db.Exec(query, newStatus, groupID, userID)
	return err
}
