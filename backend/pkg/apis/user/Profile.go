package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type UserProfile struct {
	Username string `json:"username"`
	Fname    string `json:"fname"`
	Lname    string `json:"lname"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func GetProfileHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, username string) {
	var bio sql.NullString
	var profile UserProfile

	err := db.QueryRow(`
    SELECT username, firstname, lastname, email, age, gender, password, bio 
    FROM users WHERE username = ?`,
		username).Scan(
		&profile.Username,
		&profile.Fname,
		&profile.Lname,
		&profile.Email,
		&profile.Age,
		&profile.Gender,
		&profile.Password,
		&bio,
	)
	if bio.Valid {
		profile.Bio = bio.String
	} else {
		profile.Bio = "" // Default value when NULL
	}
	if err != nil {
		fmt.Println("Error fetching user:", err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func UpdateProfileHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, currentUsername string) {
	var input UserProfile
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use sql.NullString to safely scan possible NULL value
	var existingBio sql.NullString
	var existing UserProfile
	err := db.QueryRow(`
        SELECT username, firstname, lastname, email, age, gender, password, bio
        FROM users WHERE username = ?`, currentUsername).
		Scan(
			&existing.Username,
			&existing.Fname,
			&existing.Lname,
			&existing.Email,
			&existing.Age,
			&existing.Gender,
			&existing.Password,
			&existingBio,
		)

	if err != nil {
		fmt.Println("Error fetching existing user:", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if existingBio.Valid {
		existing.Bio = existingBio.String
	} else {
		existing.Bio = ""
	}

	// Use old value if field is empty
	if strings.TrimSpace(input.Username) == "" {
		input.Username = existing.Username
	}
	if strings.TrimSpace(input.Fname) == "" {
		input.Fname = existing.Fname
	}
	if strings.TrimSpace(input.Lname) == "" {
		input.Lname = existing.Lname
	}
	if strings.TrimSpace(input.Email) == "" {
		input.Email = existing.Email
	}
	if strings.TrimSpace(input.Gender) == "" {
		input.Gender = existing.Gender
	}
	if input.Age == 0 {
		input.Age = existing.Age
	}

	if strings.TrimSpace(input.Bio) == "" {
		input.Bio = existing.Bio
	}

	// Now update all fields
	_, err = db.Exec(`
        UPDATE users 
        SET username = ?, firstname = ?, lastname = ?, email = ?, age = ?, gender = ?, bio = ?
        WHERE username = ?`,
		input.Username, input.Fname, input.Lname, input.Email, input.Age, input.Gender, input.Bio,
		currentUsername)

	if err != nil {
		fmt.Println("Error updating user:", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Profile updated",
	})
}
