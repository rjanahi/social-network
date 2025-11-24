package group

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
)

// GetReceivedInvitations returns invitations received by the current user
func GetReceivedInvitations(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	// Get received invitations
	invitations, err := database.GetReceivedInvitations(db, userID)
	if err != nil {
		fmt.Println("Error getting received invitations:", err)
		http.Error(w, "Failed to get invitations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitations)
}
