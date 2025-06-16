package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	database "socialnetwork/pkg/db"

	"github.com/google/uuid"
)

// Session structure
type Session struct {
	UserID    int
	Token     string
	ExpiresAt time.Time
}

func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		var credentials struct {
			UserOremail string `json:"userOremail"`
			Password    string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&credentials); err != nil {
			fmt.Println(" JSON Decoding Error:", err)
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}

		// Validate login
		valid, errorNum := ValidateLog(credentials.UserOremail, credentials.Password, db)
		if !valid {
			var errorMessage string
			switch errorNum {
			case 1:
				errorMessage = "Username/Email and Password cannot be empty."
			case 2:
				errorMessage = "User not found."
			case 3:
				errorMessage = "Incorrect password."
			case 4:
				errorMessage = "Internal server error."
			default:
				errorMessage = "An unknown error occurred."
			}

			response := map[string]string{"message": errorMessage}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		//  Get user ID
		userID, err := database.GetUserID(db, credentials.UserOremail)
		if err != nil {
			fmt.Println(" Error getting user ID:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//  Check if there is an active session
		activeSessionID, err := database.GetActiveSessionbyUserID(db, userID)
		if err != nil {
			fmt.Println(" Error checking active session:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log active session details
		fmt.Println(" Active Session ID:", activeSessionID)

		// If an active session exists, invalidate the old session
		if activeSessionID != -1 {
			err = database.DeleteSession(db, activeSessionID)
			if err != nil {
				fmt.Println(" Error deleting old session:", err)
				http.Error(w, "Failed to invalidate previous session", http.StatusInternalServerError)
				return
			}
			fmt.Println(" Old session invalidated.")
		}

		//  Generate new session token and expiration
		sessionToken := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		//  Store new session in the database
		err = database.InsertSession(db, userID, sessionToken, expiresAt)
		if err != nil {
			fmt.Println(" Error inserting new session:", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		//  Set session token as a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
			Path:     "/",
		})

		fmt.Println(" Login successful for User ID:", userID)

		//  Send success response
		response := map[string]string{"message": "Login successful."}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

