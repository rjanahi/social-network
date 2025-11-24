package database

import (
	"database/sql"
	"time"
)

// GetReceivedInvitations returns invitations received by a user
func GetReceivedInvitations(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT gi.id, gi.group_id, gi.inviter_id, gi.status, gi.created_at, 
		       g.title AS group_name, u.username AS inviter_name
		FROM group_invitations gi
		JOIN groups g ON gi.group_id = g.id
		JOIN users u ON gi.inviter_id = u.id
		WHERE gi.invitee_id = ? AND gi.status = 'pending'
		ORDER BY gi.created_at DESC
	`

	// Execute query
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var invitations []map[string]interface{}
	for rows.Next() {
		var id, groupID, inviterID int
		var status, groupName, inviterName string
		var createdAt time.Time

		if err := rows.Scan(&id, &groupID, &inviterID, &status, &createdAt, &groupName, &inviterName); err != nil {
			return nil, err
		}

		invitation := map[string]interface{}{
			"id":           id,
			"group_id":     groupID,
			"inviter_id":   inviterID,
			"status":       status,
			"group_name":   groupName,
			"inviter_name": inviterName,
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
