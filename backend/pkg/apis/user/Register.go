package user

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"socialnetwork/pkg/db"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type RegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var userData struct {
	Username string `json:"username"`
	Fname    string `json:"fname"`
	Lname    string `json:"lname"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Password string `json:"password"`
}

// Register handles user registration
func Register(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	fmt.Println("Received Request Body:", string(body))
	r.Body = io.NopCloser(bytes.NewReader(body)) // Reassign body so it can be used again

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userData); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	username := userData.Username
	email := strings.ToLower(userData.Email)
	password := userData.Password
	fname := userData.Fname
	lname := userData.Lname
	age := strconv.Itoa(userData.Age)
	gender := userData.Gender

	// Validate user data
	validity, errorNum := ValidateUser(username, email, password, fname, lname, gender, age, db)
	if !validity {
		var errorMessage string
		switch errorNum {
		case 1:
			errorMessage = "You must fill all fields"
		case 2:
			errorMessage = "Username already exists"
		case 3:
			errorMessage = "Email already exists"
		case 5:
			errorMessage = "Invalid email format"
		case 7:
			errorMessage = "Invalid password format. Your password must be at least 10 characters long and contain uppercase, lowercase, numbers, and special characters."
		default:
			errorMessage = "Sorry, an unknown error occurred"
		}

		response := RegistrationResponse{
			Success: false,
			Message: errorMessage,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user record in the database
	if _, err := database.InsertUser(db, username, email, fname, lname, age, gender, string(hashedPassword)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send success response
	response := RegistrationResponse{
		Success: true,
		Message: "User registered successfully.",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func checkIfUsernameExists(db *sql.DB, username string) bool {
	usernames, _ := database.GetAllUserNames(db)
	for _, user := range usernames {
		if username == user {
			return true
		}
	}
	return false
}

func checkIfEmaileExists(db *sql.DB, email string) bool {
	emails, _ := database.GetAllUserEmails(db)
	for _, usermail := range emails {
		if strings.EqualFold(usermail, email) {
			return true
		}
	}
	return false
}

func ValidateUser(username, email, password, fname, lname, age, gender string, db *sql.DB) (bool, int) {
	if username == "" || email == "" || password == "" {
		return false, 1
	}

	if checkIfUsernameExists(db, username) {
		return false, 2
	}

	if !ValidateEmailFormat(email) {
		return false, 5
	}

	if checkIfEmaileExists(db, email) {
		return false, 3
	}

	if !CheckIfPassValid(password) {
		return false, 7
	}

	return true, 0
}

func ValidateLog(userOremail, password string, db *sql.DB) (bool, int) {
	if userOremail == "" || password == "" {
		return false, 1
	}

	if !checkIfUsernameExists(db, userOremail) && !checkIfEmaileExists(db, userOremail) {
		return false, 2
	}

	pass, err := GetUserPassword(db, userOremail)
	if err != nil {
		return false, 4
	}

	err = bcrypt.CompareHashAndPassword([]byte(pass), []byte(password)) //check if both are similar
	if err != nil {
		return false, 3
	}

	return true, 0
}

func GetUserPassword(db *sql.DB, userOremail string) (string, error) {
	isUsername := checkIfUsernameExists(db, userOremail)
	isEmail := checkIfEmaileExists(db, userOremail)

	if isUsername {
		query := `SELECT password FROM users WHERE username = ?`
		var pass string
		err := db.QueryRow(query, userOremail).Scan(&pass)
		if err != nil {
			if err == sql.ErrNoRows {
				// No rows found for the given username
				return "", nil
			}
			return "", err
		}
		return pass, nil
	} else if isEmail {
		email := strings.ToLower(userOremail)
		query := `SELECT password FROM users WHERE email = ?`
		var pass string
		err := db.QueryRow(query, email).Scan(&pass)
		if err != nil {
			if err == sql.ErrNoRows {
				// No rows found for the given username
				return "", nil
			}
			return "", err
		}
		return pass, nil
	}
	return "", nil
}

func ValidateEmailFormat(email string) bool {
	emailRegex := regexp.MustCompile(`^(?i)([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$`)

	// Check if email matches the regexp
	return emailRegex.MatchString(email)
}

func CheckIfPassValid(pass string) bool {
	if len(pass) < 10 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range pass {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasNumber = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}

func ValidateSession(db *sql.DB, r *http.Request) (int, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false // No session found
	}

	var userID int
	var expiresAt time.Time
	query := `SELECT user_id, expires_at FROM sessions WHERE token = ?`
	err = db.QueryRow(query, cookie.Value).Scan(&userID, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return 0, false // Session expired or not found
	}

	return userID, true
}
