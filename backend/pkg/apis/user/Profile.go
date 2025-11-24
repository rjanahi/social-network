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
	"strconv"
	"strings"
	"time"

	chat "socialnetwork/pkg/apis/chat"

	"golang.org/x/crypto/bcrypt"
)

type UserProfile struct {
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Fname       string `json:"fname"`
	Lname       string `json:"lname"`
	Email       string `json:"email"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Password    string `json:"password"`
	Bio         string `json:"bio"`
	Avatar      string `json:"avatar"`
	IsPrivate   bool   `json:"isPrivate"`
	DateOfBirth string `json:"dateOfBirth"`
}

func IsFollowing(db *sql.DB, followerUsername, followingUsername string) bool {
	// First get the user IDs for the usernames
	var followerID, followingID int

	// Get follower ID
	err := db.QueryRow(`SELECT id FROM users WHERE username = ?`, followerUsername).Scan(&followerID)
	if err != nil {
		fmt.Println("Error getting follower ID:", err)
		return false
	}

	// Get following ID
	err = db.QueryRow(`SELECT id FROM users WHERE username = ?`, followingUsername).Scan(&followingID)
	if err != nil {
		fmt.Println("Error getting following ID:", err)
		return false
	}

	// Check if the follow relationship exists in userFollow table
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM userFollow 
		WHERE follower_id = ? AND following_id = ?`,
		followerID, followingID).Scan(&count)

	if err != nil {
		fmt.Println("Error checking follow status:", err)
		return false
	}

	return count > 0
}

// Add this simple response struct
type ProfilePrivacyResponse struct {
	Profile      UserProfile `json:"profile"`
	Posts        interface{} `json:"posts"`
	IsFollowing  bool        `json:"isFollowing"`
	CanViewPosts bool        `json:"canViewPosts"`
}

func GetProfileHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, username string) {
	var bio sql.NullString
	var profile UserProfile

	err := db.QueryRow(`
    SELECT username, nickname, firstname, lastname, email, age, gender, bio , avatar_url, isPrivate
    FROM users WHERE username = ?`,
		username).Scan(
		&profile.Username,
		&profile.Nickname,
		&profile.Fname,
		&profile.Lname,
		&profile.Email,
		&profile.Age,
		&profile.Gender,
		&bio,
		&profile.Avatar,
		&profile.IsPrivate,
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

	profileObj := map[string]interface{}{
		"username":  profile.Username,
		"fname":     profile.Fname,
		"lname":     profile.Lname,
		"email":     profile.Email,
		"age":       profile.Age,
		"gender":    profile.Gender,
		"bio":       profile.Bio,
		"avatar":    profile.Avatar,
		"isPrivate": profile.IsPrivate,
		"nickname":  profile.Nickname,
	}
	fmt.Println("Fetched profile:", profileObj)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"profile": profileObj,
	})
}

func UpdateProfileHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, currentUsername string) {
	// Handle both JSON and multipart form data
	contentType := r.Header.Get("Content-Type")
	fmt.Println("[UpdateProfile] Content-Type:", contentType)

	var formValues map[string]string

	if strings.Contains(contentType, "application/json") {
		// Parse JSON request
		fmt.Println("[UpdateProfile] Parsing as JSON")
		var jsonData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Convert to string map
		formValues = make(map[string]string)
		for key, value := range jsonData {
			if value != nil {
				formValues[key] = fmt.Sprintf("%v", value)
			}
		}
	} else {
		// Try multipart form (for file uploads)
		fmt.Println("[UpdateProfile] Parsing as multipart form")
		const maxMem = 10 << 20 // 10MB
		if err := r.ParseMultipartForm(maxMem); err != nil {
			fmt.Println("[UpdateProfile] Parse error:", err)
			http.Error(w, "could not parse form", http.StatusBadRequest)
			return
		}
		fmt.Println("[UpdateProfile] Successfully parsed multipart form")

		// Debug: log received multipart form keys and files
		if r.MultipartForm != nil {
			for k, v := range r.MultipartForm.Value {
				fmt.Printf("[UpdateProfile] Form value %s: %v\n", k, v)
			}
			for k, files := range r.MultipartForm.File {
				for _, fh := range files {
					fmt.Printf("[UpdateProfile] Form file field %s: filename=%s, size=%d, header=%v\n", k, fh.Filename, fh.Size, fh.Header)
				}
			}
		}

		// Convert form values to map
		formValues = make(map[string]string)
		for key, values := range r.PostForm {
			if len(values) > 0 {
				formValues[key] = values[0]
			}
		}
	}

	// 2) load existing user (include id + password hash)
	var existing UserProfile
	var existingBio sql.NullString
	err := db.QueryRow(`
        SELECT username, nickname, firstname, lastname, email, age, gender, password, bio, avatar_url, isPrivate
        FROM users WHERE username = ?`, currentUsername).
		Scan(
			&existing.Username,
			&existing.Nickname,
			&existing.Fname,
			&existing.Lname,
			&existing.Email,
			&existing.Age,
			&existing.Gender,
			&existing.Password, // this is the bcrypt hash
			&existingBio,
			&existing.Avatar,
			&existing.IsPrivate,
		)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if existingBio.Valid {
		existing.Bio = existingBio.String
	}

	// 3) avatar handling (reset, upload, or keep)
	avatarURL := existing.Avatar
	if formValues["resetAvatar"] == "true" {
		avatarURL = "/img/avatars/images.png"
	} else if strings.Contains(contentType, "multipart/form-data") {
		// Only try to read form file if content-type is multipart
		file, header, err := r.FormFile("avatarInput")
		if err != nil {
			if err == http.ErrMissingFile {
				// no file provided - that's fine
				fmt.Println("[UpdateProfile] No avatar file provided")
			} else {
				fmt.Println("[UpdateProfile] FormFile error:", err)
				http.Error(w, "error reading avatar: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err == nil && header.Filename != "" {
			defer file.Close()

			// Read a small header to detect MIME
			// Prefer the Content-Type provided in the multipart header (browsers set this).
			// Fall back to content sniffing only if the header is absent.
			headerMime := ""
			if header != nil {
				headerMime = header.Header.Get("Content-Type")
			}
			// Read up to 512 bytes for detection if needed
			head := make([]byte, 512)
			n, _ := file.Read(head)
			detected := http.DetectContentType(head[:n])
			// reset
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				http.Error(w, "could not read avatar", http.StatusInternalServerError)
				return
			}
			mime := headerMime
			if mime == "" {
				mime = detected
			}
			fmt.Println("[UpdateProfile] header mime:", headerMime, "detected:", detected, "using:", mime)
			if !strings.HasPrefix(mime, "image/") {
				http.Error(w, "Uploaded file is not an image", http.StatusBadRequest)
				return
			}

			// Save into Next.js public dir — prefer STATIC_DIR when provided so the
			// backend can be started from different working directories.
			staticDir := os.Getenv("STATIC_DIR")
			var uploadDir string
			if staticDir != "" {
				uploadDir = filepath.Join(staticDir, "public", "img", "avatars")
			} else {
				uploadDir = filepath.Join("..", "frontend-next", "public", "img", "avatars")
			}
			if err := os.MkdirAll(uploadDir, 0o755); err != nil {
				log.Println("avatar save: MkdirAll error:", err)
				http.Error(w, "could not save avatar", http.StatusInternalServerError)
				return
			}

			base := filepath.Base(strings.ReplaceAll(header.Filename, " ", "-"))
			ext := strings.ToLower(filepath.Ext(base))
			if ext == "" {
				// derive from MIME if missing (support common image types)
				switch mime {
				case "image/png":
					ext = ".png"
				case "image/jpeg":
					ext = ".jpg"
				case "image/gif":
					ext = ".gif"
				case "image/webp":
					ext = ".webp"
				default:
					// fallback to jpg if unknown image subtype
					ext = ".jpg"
				}
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

			copied, err := io.Copy(dst, file)
			if err != nil {
				log.Println("avatar save: Copy error:", err)
				http.Error(w, "could not save avatar", http.StatusInternalServerError)
				return
			}
			fmt.Println("[UpdateProfile] avatar saved:", dstPath, "bytesCopied:", copied)
			avatarURL = "/img/avatars/" + finalName
		}
	}

	// helpers
	get := func(key, old string) string {
		v := strings.TrimSpace(formValues[key])
		if v == "" {
			return old
		}
		return v
	}
	parseBool := func(val string, fallback bool) bool {
		if val == "" {
			return fallback
		}
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "1", "true", "on", "yes":
			return true
		case "0", "false", "off", "no":
			return false
		default:
			return fallback
		}
	}
	parseInt := func(val string, fallback int) int {
		if strings.TrimSpace(val) == "" {
			return fallback
		}
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
		return fallback
	}

	// 4) read text fields with fallbacks
	age := parseInt(formValues["age"], existing.Age)
	isPrivate := parseBool(formValues["isPrivate"], existing.IsPrivate)

	username := get("username", existing.Username)
	nickname := get("nickname", existing.Nickname) // <- fixed fallback
	fname := get("firstname", existing.Fname)
	lname := get("lastname", existing.Lname)
	email := get("email", existing.Email)
	gender := get("gender", existing.Gender)
	dateOfBirth := get("dateOfBirth", existing.DateOfBirth)

	// If a dateOfBirth is provided, compute the age from it and normalize the date
	if strings.TrimSpace(dateOfBirth) != "" {
		if computedAge, normalized, err := ComputeAgeFromDOB(dateOfBirth); err == nil {
			age = computedAge
			dateOfBirth = normalized
		} else {
			fmt.Println("warning: could not parse dateOfBirth during profile update:", err)
			// keep submitted/parsed age in this case
		}
	}

	bio := ""
	if b := strings.TrimSpace(formValues["bio"]); b != "" {
		bio = b
	}

	oldPassword := formValues["password"]
	newPassword := formValues["newPassword"]
	passwordHash := existing.Password // default: unchanged

	fmt.Println("New password: ", newPassword)
	// 4a) if newPassword provided, verify oldPassword and re-hash
	if strings.TrimSpace(newPassword) != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(existing.Password), []byte(oldPassword)); err != nil {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Wrong password",
			})
			return
		}
		newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Println("New password hashing error:", err)
			http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
			return
		}
		passwordHash = string(newHash)
	} else {
		// If newPassword is empty, keep the old password hash
		passwordHash = existing.Password
	}

	// (Optional) enforce username/email uniqueness if changed
	if username != existing.Username {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ? `, username).Scan(&count); err == nil && count > 0 {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Username already taken",
			})
			return
		}
	}
	if email != existing.Email {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ? `, email).Scan(&count); err == nil && count > 0 {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Email already in use",
			})
			return
		}
	}

	fmt.Println("Updating profile:", map[string]interface{}{
		"username":  username,
		"nickname":  nickname,
		"fname":     fname,
		"lname":     lname,
		"email":     email,
		"age":       age,
		"password":  oldPassword,
		"newPass":   newPassword,
		"gender":    gender,
		"bio":       bio,
		"avatar":    avatarURL,
		"isPrivate": isPrivate,
	})

	// 5) UPDATE by id (stable key)
	_, err = db.Exec(`
        UPDATE users SET username = ?, nickname = ? ,firstname = ?, lastname = ?, email = ?, password = ?, age = ?, gender = ?, bio = ?, avatar_url = ?, isPrivate = ?, date_of_birth = ? WHERE username = ?`,
		username, nickname, fname, lname, email, passwordHash, age, gender, bio, avatarURL, isPrivate, dateOfBirth, currentUsername,
	)
	if err != nil {
		log.Println("update profile error:", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// 6) respond (never include password/hash)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Profile updated",
	})
}

func GetOtherProfile(db *sql.DB, username string) (UserProfile, error) {
	var bio sql.NullString
	var profile UserProfile

	err := db.QueryRow(`
		SELECT username, nickname, firstname, lastname, email, age, gender, password, bio, avatar_url, isPrivate
		FROM users WHERE username = ?`,
		username).Scan(
		&profile.Username,
		&profile.Nickname,
		&profile.Fname,
		&profile.Lname,
		&profile.Email,
		&profile.Age,
		&profile.Gender,
		&profile.Password,
		&bio,
		&profile.Avatar,
		&profile.IsPrivate,
	)
	if err != nil {
		return UserProfile{}, err
	}

	profile.Bio = ""
	if bio.Valid {
		profile.Bio = bio.String
	}

	return profile, nil
}

// GetCompleteProfileHandler returns user profile with posts, followers, and following
func GetCompleteProfileHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Validate session
	userID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user profile
	var profile UserProfile
	var bio sql.NullString
	var dateOfBirth sql.NullString

	err := db.QueryRow(`
		SELECT username, nickname, firstname, lastname, email, age, gender, bio, avatar_url, isPrivate, date_of_birth
		FROM users WHERE id = ?`, userID).Scan(
		&profile.Username,
		&profile.Nickname,
		&profile.Fname,
		&profile.Lname,
		&profile.Email,
		&profile.Age,
		&profile.Gender,
		&bio,
		&profile.Avatar,
		&profile.IsPrivate,
		&dateOfBirth,
	)
	if err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	if bio.Valid {
		profile.Bio = bio.String
	}
	if dateOfBirth.Valid {
		profile.DateOfBirth = dateOfBirth.String
	}

	// Get user's posts
	posts, err := getUserPosts(db, userID)
	if err != nil {
		log.Printf("Error fetching user posts: %v", err)
		posts = []interface{}{}
	}

	// Get followers
	followers, err := getFollowers(db, userID)
	if err != nil {
		log.Printf("Error fetching followers: %v", err)
		followers = []interface{}{}
	}

	// Get following
	following, err := getFollowing(db, userID)
	if err != nil {
		log.Printf("Error fetching following: %v", err)
		following = []interface{}{}
	}

	// Get pending follow requests
	pendingRequests, err := getPendingFollowRequests(db, userID)
	if err != nil {
		log.Printf("Error fetching pending requests: %v", err)
		pendingRequests = []interface{}{}
	}

	response := map[string]interface{}{
		"success":         true,
		"profile":         profile,
		"posts":           posts,
		"followers":       followers,
		"following":       following,
		"pendingRequests": pendingRequests,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUserPosts(db *sql.DB, userID int) ([]interface{}, error) {
	query := `
		SELECT p.id, p.title, p.content, p.imgOrgif, p.created_at, 
		       COALESCE(p.privacy_level, 0) as privacy_level,
		       COALESCE(lc.like_count, 0) as like_count,
		       COALESCE(cc.comment_count, 0) as comment_count
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as like_count 
			FROM likes 
			GROUP BY post_id
		) lc ON p.id = lc.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as comment_count 
			FROM comments 
			GROUP BY post_id
		) cc ON p.id = cc.post_id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []interface{}
	for rows.Next() {
		var post struct {
			ID           int    `json:"id"`
			Title        string `json:"title"`
			Content      string `json:"content"`
			ImgOrgif     string `json:"imgOrgif"`
			CreatedAt    string `json:"created_at"`
			PrivacyLevel int    `json:"privacy_level"`
			LikeCount    int    `json:"like_count"`
			CommentCount int    `json:"comment_count"`
		}

		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgOrgif,
			&post.CreatedAt, &post.PrivacyLevel, &post.LikeCount, &post.CommentCount)
		if err != nil {
			continue
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func getFollowers(db *sql.DB, userID int) ([]interface{}, error) {
	query := `
		SELECT u.id, u.username, u.firstname, u.lastname, u.avatar_url
		FROM users u
		INNER JOIN userFollow uf ON u.id = uf.follower_id
		WHERE uf.following_id = ?`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []interface{}
	for rows.Next() {
		var follower struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Fname    string `json:"firstname"`
			Lname    string `json:"lastname"`
			Avatar   string `json:"avatar"`
		}

		err := rows.Scan(&follower.ID, &follower.Username, &follower.Fname, &follower.Lname, &follower.Avatar)
		if err != nil {
			continue
		}

		followers = append(followers, follower)
	}

	return followers, nil
}

func getFollowing(db *sql.DB, userID int) ([]interface{}, error) {
	query := `
		SELECT u.id, u.username, u.firstname, u.lastname, u.avatar_url
		FROM users u
		INNER JOIN userFollow uf ON u.id = uf.following_id
		WHERE uf.follower_id = ?`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var following []interface{}
	for rows.Next() {
		var user struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Fname    string `json:"firstname"`
			Lname    string `json:"lastname"`
			Avatar   string `json:"avatar"`
		}

		err := rows.Scan(&user.ID, &user.Username, &user.Fname, &user.Lname, &user.Avatar)
		if err != nil {
			continue
		}

		following = append(following, user)
	}

	return following, nil
}

func getPendingFollowRequests(db *sql.DB, userID int) ([]interface{}, error) {
	query := `
		SELECT u.id, u.username, u.firstname, u.lastname, u.avatar_url, fr.created_at
		FROM users u
		INNER JOIN follow_requests fr ON u.id = fr.requester_id
		WHERE fr.target_id = ? AND fr.status = 'pending'
		ORDER BY fr.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []interface{}
	for rows.Next() {
		var request struct {
			ID        int    `json:"id"`
			Username  string `json:"username"`
			Fname     string `json:"firstname"`
			Lname     string `json:"lastname"`
			Avatar    string `json:"avatar"`
			CreatedAt string `json:"created_at"`
		}

		err := rows.Scan(&request.ID, &request.Username, &request.Fname, &request.Lname, &request.Avatar, &request.CreatedAt)
		if err != nil {
			continue
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// UpdateProfileJSONHandler handles profile updates via JSON (simpler than multipart)
func UpdateProfileJSONHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, currentUsername string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var updateData struct {
		Fname       string `json:"fname"`
		Lname       string `json:"lname"`
		Nickname    string `json:"nickname"`
		Email       string `json:"email"`
		Age         int    `json:"age"`
		Gender      string `json:"gender"`
		DateOfBirth string `json:"dateOfBirth"`
		Bio         string `json:"bio"`
		IsPrivate   bool   `json:"isPrivate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// If a dateOfBirth is provided in JSON, compute/override the age
	if strings.TrimSpace(updateData.DateOfBirth) != "" {
		if computedAge, normalized, err := ComputeAgeFromDOB(updateData.DateOfBirth); err == nil {
			updateData.Age = computedAge
			updateData.DateOfBirth = normalized
		} else {
			fmt.Println("warning: could not parse dateOfBirth in JSON profile update:", err)
		}
	}

	// Update the user profile
	query := `
		UPDATE users SET 
			username = ?, nickname = ?, firstname = ?, lastname = ?, email = ?, 
			age = ?, gender = ?, bio = ?, isPrivate = ?, date_of_birth = ?
		WHERE username = ?`

	_, err := db.Exec(query,
		currentUsername, // keep same username
		updateData.Nickname,
		updateData.Fname,
		updateData.Lname,
		updateData.Email,
		updateData.Age,
		updateData.Gender,
		updateData.Bio,
		updateData.IsPrivate,
		updateData.DateOfBirth,
		currentUsername,
	)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
	})
}

// TogglePrivacy toggles the user's profile privacy setting
func TogglePrivacy(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	userID, loggedIn := ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current privacy status
	var currentPrivacy bool
	err := db.QueryRow("SELECT isPrivate FROM users WHERE id = ?", userID).Scan(&currentPrivacy)
	if err != nil {
		http.Error(w, "Failed to get current privacy status", http.StatusInternalServerError)
		return
	}

	// Toggle the privacy
	newPrivacy := !currentPrivacy
	_, err = db.Exec("UPDATE users SET isPrivate = ? WHERE id = ?", newPrivacy, userID)
	if err != nil {
		http.Error(w, "Failed to update privacy setting", http.StatusInternalServerError)
		return
	}

	// Broadcast privacy update to connected clients so UI can refresh (e.g., users list)
	if hub != nil {
		var username string
		_ = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
		notif := chat.Frontend{
			Type:      "privacy_update",
			From:      userID,
			Username:  username,
			IsPrivate: newPrivacy,
			Timestamp: time.Now(),
		}
		hub.Mutex.RLock()
		for _, client := range hub.Clients {
			select {
			case client.Send <- notif:
			default:
			}
		}
		hub.Mutex.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"isPrivate": newPrivacy,
		"message":   fmt.Sprintf("Profile is now %s", map[bool]string{true: "private", false: "public"}[newPrivacy]),
	})
}
