package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// ParseID converts string ID to integer
func ParseID(idStr string) (int, error) {
	if idStr == "" {
		return 0, fmt.Errorf("empty ID")
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid ID: %d", id)
	}

	return id, nil
}

// GetAllGroups returns all groups with basic info
func GetAllGroups(db *sql.DB) ([]map[string]interface{}, error) {
	query := `
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at, u.username,
		       COUNT(gm.user_id) as member_count
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		GROUP BY g.id, g.title, g.description, g.creator_id, g.created_at, u.username
		ORDER BY g.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []map[string]interface{}
	for rows.Next() {
		var id, creatorID, memberCount int
		var title, description, creator string
		var createdAt time.Time

		err := rows.Scan(&id, &title, &description, &creatorID, &createdAt, &creator, &memberCount)
		if err != nil {
			return nil, err
		}

		group := map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  description,
			"creator_id":   creatorID,
			"creator":      creator,
			"member_count": memberCount,
			"created_at":   createdAt.Format("2006-01-02 15:04:05"),
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// GetAllGroupsWithMembership returns all groups with current user's membership status
func GetAllGroupsWithMembership(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at, u.username,
		       COUNT(gm.user_id) as member_count,
		       COALESCE(user_gm.status, '') as user_membership_status
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		LEFT JOIN group_members user_gm ON g.id = user_gm.group_id AND user_gm.user_id = ?
		GROUP BY g.id, g.title, g.description, g.creator_id, g.created_at, u.username, user_gm.status
		ORDER BY g.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []map[string]interface{}
	for rows.Next() {
		var id, creatorID, memberCount int
		var title, description, creator, userMembershipStatus string
		var createdAt time.Time

		err := rows.Scan(&id, &title, &description, &creatorID, &createdAt, &creator, &memberCount, &userMembershipStatus)
		if err != nil {
			return nil, err
		}

		group := map[string]interface{}{
			"id":                     id,
			"title":                  title,
			"description":            description,
			"creator_id":             creatorID,
			"creator":                creator,
			"member_count":           memberCount,
			"created_at":             createdAt.Format("2006-01-02 15:04:05"),
			"user_membership_status": userMembershipStatus,
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// GetGroupByID returns detailed group information
func GetGroupByID(db *sql.DB, groupID int) (map[string]interface{}, error) {
	query := `
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at, u.username,
		       COUNT(gm.user_id) as member_count
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
		WHERE g.id = ?
		GROUP BY g.id, g.title, g.description, g.creator_id, g.created_at, u.username`

	var id, creatorID, memberCount int
	var title, description, creator string
	var createdAt time.Time

	err := db.QueryRow(query, groupID).Scan(&id, &title, &description, &creatorID, &createdAt, &creator, &memberCount)
	if err != nil {
		return nil, err
	}

	group := map[string]interface{}{
		"id":           id,
		"title":        title,
		"description":  description,
		"creator_id":   creatorID,
		"creator":      creator,
		"member_count": memberCount,
		"created_at":   createdAt.Format("2006-01-02 15:04:05"),
	}

	return group, nil
}

// IsGroupMember checks if a user is a member of a group
func IsGroupMember(db *sql.DB, groupID, userID int) (bool, error) {
	query := `SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'accepted'`
	var count int
	err := db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetGroupMemberStatus returns the membership status of a user in a group
func GetGroupMemberStatus(db *sql.DB, groupID, userID int) (string, error) {
	query := `SELECT status FROM group_members WHERE group_id = ? AND user_id = ?`
	var status string
	err := db.QueryRow(query, groupID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No membership record found
		}
		return "", err
	}
	return status, nil
}

// UpdateGroupInvitationStatus updates the status of a group invitation
func UpdateGroupInvitationStatus(db *sql.DB, invitationID, userID int, status string) error {
	query := `UPDATE group_invitations SET status = ? WHERE id = ? AND invitee_id = ?`
	result, err := db.Exec(query, status, invitationID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no invitation found or you don't have permission to respond")
	}

	return nil
}

// GetGroupIDFromInvitation gets the group ID from an invitation
func GetGroupIDFromInvitation(db *sql.DB, invitationID int) (int, error) {
	query := `SELECT group_id FROM group_invitations WHERE id = ?`
	var groupID int
	err := db.QueryRow(query, invitationID).Scan(&groupID)
	return groupID, err
}

// GetUserGroups returns groups where user is a member
func GetUserGroups(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT g.id, g.title, g.description, g.creator_id, g.created_at, u.username,
		       COUNT(gm2.user_id) as member_count
		FROM groups g
		JOIN users u ON g.creator_id = u.id
		JOIN group_members gm ON g.id = gm.group_id
		LEFT JOIN group_members gm2 ON g.id = gm2.group_id AND gm2.status = 'accepted'
		WHERE gm.user_id = ? AND gm.status = 'accepted'
		GROUP BY g.id, g.title, g.description, g.creator_id, g.created_at, u.username
		ORDER BY g.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []map[string]interface{}
	for rows.Next() {
		var id, creatorID, memberCount int
		var title, description, creator string
		var createdAt time.Time

		err := rows.Scan(&id, &title, &description, &creatorID, &createdAt, &creator, &memberCount)
		if err != nil {
			return nil, err
		}

		group := map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  description,
			"creator_id":   creatorID,
			"creator":      creator,
			"member_count": memberCount,
			"created_at":   createdAt.Format("2006-01-02 15:04:05"),
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// GetUserGroupInvitations returns pending invitations for a user
func GetUserGroupInvitations(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT gi.id, gi.group_id, gi.inviter_id, gi.status, gi.created_at, g.title, u.username
		FROM group_invitations gi
		JOIN groups g ON gi.group_id = g.id
		JOIN users u ON gi.inviter_id = u.id
		WHERE gi.invitee_id = ? AND gi.status = 'pending'
		ORDER BY gi.created_at DESC`

	fmt.Printf("[GetUserGroupInvitations] Querying for userID: %d\n", userID)
	rows, err := db.Query(query, userID)
	if err != nil {
		fmt.Printf("[GetUserGroupInvitations] Query error: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var invitations []map[string]interface{}
	for rows.Next() {
		var id, groupID, inviterID int
		var groupTitle, inviterName, status string
		var createdAt time.Time

		err := rows.Scan(&id, &groupID, &inviterID, &status, &createdAt, &groupTitle, &inviterName)
		if err != nil {
			fmt.Printf("[GetUserGroupInvitations] Scan error: %v\n", err)
			return nil, err
		}

		invitation := map[string]interface{}{
			"id":           id,
			"group_id":     groupID,
			"inviter_id":   inviterID,
			"status":       status,
			"group_name":   groupTitle,
			"inviter_name": inviterName,
			"created_at":   createdAt.Format("2006-01-02 15:04:05"),
		}
		fmt.Printf("[GetUserGroupInvitations] Found invitation: %+v\n", invitation)
		invitations = append(invitations, invitation)
	}

	fmt.Printf("[GetUserGroupInvitations] Returning %d invitations\n", len(invitations))
	return invitations, nil
}

func GetGroupMembers(db *sql.DB, groupID int) ([]map[string]interface{}, error) {
	query := `
		SELECT u.id, u.username, gm.status
		FROM group_members gm	
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'accepted'
		ORDER BY u.username`
	rows, err := db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []map[string]interface{}
	for rows.Next() {
		var id int
		var username, status string
		if err := rows.Scan(&id, &username, &status); err != nil {
			return nil, err
		}
		member := map[string]interface{}{
			"id":       id,
			"username": username,
			"status":   status,
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}
