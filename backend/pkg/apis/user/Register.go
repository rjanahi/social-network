package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	database "socialnetwork/pkg/db"
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
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatarInput"`
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

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")
	fname := r.FormValue("fname")
	lname := r.FormValue("lname")
	age := r.FormValue("age") // your DB uses TEXT
	gender := r.FormValue("gender")
	nickname := r.FormValue("nickname")       // optional
	bio := r.FormValue("bio")                 // optional
	dateOfBirth := r.FormValue("dateOfBirth") // optional

	// Optional avatar
	avatarURL := ""
	file, header, err := r.FormFile("avatarInput")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "error reading avatar", http.StatusBadRequest)
		return
	}
	if err == nil && header != nil && header.Filename != "" {
		defer file.Close()
		base := filepath.Base(strings.ReplaceAll(header.Filename, " ", "-"))
		ext := strings.ToLower(filepath.Ext(base))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			http.Error(w, "Image must be a PNG or JPG.", http.StatusBadRequest)
			return
		}

		uploadDir := filepath.Join("..", "frontend-next", "public", "img", "avatars")
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			log.Println("avatar save: MkdirAll error:", err)
			http.Error(w, "could not save avatar", http.StatusInternalServerError)
			return
		}

		stamp := time.Now().UTC().Format("20060102T150405")
		nameNoExt := strings.TrimSuffix(base, ext)
		finalName := fmt.Sprintf("%s_%s%s", nameNoExt, stamp, ext)

		dstPath := filepath.Join(uploadDir, finalName)
		dst, err := os.Create(dstPath)
		if err != nil {
			log.Println("avatar save: Create error:", err)
			http.Error(w, "could not save avatar", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Println("avatar save: Copy error:", err)
			http.Error(w, "could not save avatar", http.StatusInternalServerError)
			return
		}
		avatarURL = "/img/avatars/" + finalName
	}
	if avatarURL == "" {
		// ensure a safe default (DB default only applies if column omitted)
		avatarURL = "/img/avatars/images.png"
	}

	// Validate requireds only; nickname/bio/avatar are optional
	validity, errorNum := ValidateUser(username, email, password, fname, lname, age, gender, db)
	if !validity {
		var msg string
		switch errorNum {
		case 1:
			msg = "You must fill all fields"
		case 2:
			msg = "Username already exists"
		case 3:
			msg = "Email already exists"
		case 5:
			msg = "Invalid email format"
		case 7:
			msg = "Invalid password format. Your password must be at least 10 characters long and contain uppercase, lowercase, numbers, and special characters."
		default:
			msg = "Sorry, an unknown error occurred"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegistrationResponse{Success: false, Message: msg})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If a dateOfBirth was provided try to compute/override age to keep DB consistent
	if dateOfBirth != "" {
		if computedAge, normalized, err := ComputeAgeFromDOB(dateOfBirth); err == nil {
			age = strconv.Itoa(computedAge)
			dateOfBirth = normalized
		} else {
			// If parsing fails, log but continue with submitted age
			// (InsertUser will accept the provided age string)
			fmt.Println("warning: could not parse dateOfBirth during registration:", err)
		}
	}

	// IMPORTANT: ensure this argument order matches your InsertUser signature/columns
	if _, err := database.InsertUser(
		db,
		username, nickname, email, fname, lname, age, gender, string(hashedPassword), bio, avatarURL, dateOfBirth,
	); err != nil {
		log.Println("InsertUser error:", err) // <— watch this in your console
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RegistrationResponse{Success: true, Message: "User registered successfully."})
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
