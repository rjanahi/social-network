package database

import (
	"database/sql"
)

// HasDeclinedInvitation checks if a user has previously declined an invitation to a group
func HasDeclinedInvitation(db *sql.DB, groupID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND status = 'declined'
	`
	var count int
	err := db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdatePreviousInvitationStatus updates a previously declined invitation to pending
func UpdatePreviousInvitationStatus(tx *sql.Tx, groupID, userID, inviterID int) error {
	// First check if there's a declined invitation
	query := `
		SELECT id FROM group_invitations 
		WHERE group_id = ? AND invitee_id = ? AND status = 'declined'
		ORDER BY created_at DESC LIMIT 1
	`
	var invitationID int
	err := tx.QueryRow(query, groupID, userID).Scan(&invitationID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No declined invitation found, insert a new one
			return InsertGroupInvitationTx(tx, groupID, inviterID, userID)
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
