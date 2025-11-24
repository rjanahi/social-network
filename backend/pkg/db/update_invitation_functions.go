package database

import (
	"database/sql"
)

// UpdateCancelledInvitationStatus updates a previously cancelled invitation to pending
func UpdateCancelledInvitationStatus(tx *sql.Tx, groupID, userID, inviterID int) error {
	// First check if there's a cancelled invitation
	query := `
		SELECT id FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND status = 'cancelled'
		ORDER BY created_at DESC LIMIT 1
	`
	var invitationID int
	err := tx.QueryRow(query, groupID, userID).Scan(&invitationID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No cancelled invitation found, we can try to insert a new one
			return nil
		}
		return err
	}

	// Update the status to pending and the inviter
	update := `
		UPDATE group_invitations 
		SET status = 'pending', inviter_id = ?, created_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.Exec(update, inviterID, invitationID)
	return err
}

// HasCancelledInvitation checks if a user has previously cancelled an invitation to a group
func HasCancelledInvitation(db *sql.DB, groupID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND status = 'cancelled'
	`
	var count int
	err := db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
