package post

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

	"socialnetwork/pkg/apis/chat"
	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
)

// Post structure with privacy
type Post struct {
	Title             string   `json:"title"`
	Content           string   `json:"content"`
	ImgOrGif          string   `json:"imgOrgif"`
	Categories        []string `json:"categories"`
	PrivacyLevel      int      `json:"privacy_level"` // 0=public, 1=followers, 2=selected
	SelectedFollowers []int    `json:"selected_followers"`
}

// CreatePost handles post submission with privacy
func CreatePost(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session and get user ID
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// 2) parse multipart form (up to 10 MB in memory)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// 3) pull out text fields
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	if title == "" || content == "" {
		http.Error(w, "Title and Content cannot be empty.", http.StatusBadRequest)
		return
	}

	privacyLevel, _ := strconv.Atoi(r.FormValue("privacy_level"))
	// Validate privacy level
	if privacyLevel < 0 || privacyLevel > 2 {
		privacyLevel = 0 // Default to public
	}

	// 4) handle optional image upload
	imgOrGif := ""
	file, header, err := r.FormFile("imgOrgif")
	if err == nil {
		defer file.Close()

		// sanitize extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".gif" && ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			http.Error(w, "Image must be a GIF, PNG, or JPG.", http.StatusBadRequest)
			return
		}

		// determine upload directory from STATIC_DIR if present so backend can be run
		// from repo root or backend/ without breaking paths
		staticDir := os.Getenv("STATIC_DIR")
		uploadDir := ""
		if staticDir != "" {
			uploadDir = filepath.Join(staticDir, "public", "img", "posts")
		} else {
			// fallback to original relative path
			uploadDir = filepath.Join("..", "frontend-next", "public", "img", "posts")
		}
		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			log.Println("mkdir error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// generate a safe, timestamped filename to avoid collisions and missing-dir issues
		base := filepath.Base(strings.ReplaceAll(header.Filename, " ", "-"))
		ext = strings.ToLower(filepath.Ext(base))
		nameNoExt := strings.TrimSuffix(base, ext)
		if ext == "" {
			// fallback to .jpg if extension missing
			ext = ".jpg"
		}
		stamp := time.Now().UTC().Format("20060102T150405")
		finalName := fmt.Sprintf("%s_%s%s", nameNoExt, stamp, ext)

		dstPath := filepath.Join(uploadDir, finalName)
		dstFile, err := os.Create(dstPath)
		if err != nil {
			log.Println("file create error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, file); err != nil {
			log.Println("file copy error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// this is the URL your frontend can use to display the image
		imgOrGif = "/img/posts/" + finalName
	} else if err != http.ErrMissingFile {
		http.Error(w, "Error reading uploaded file", http.StatusBadRequest)
		return
	}

	// 5) parse categories (checkboxes with name="category")
	cats := r.MultipartForm.Value["category"]
	if len(cats) == 0 {
		cats = []string{"none"}
	}

	// 6) insert the post record
	// Insert the post into the database with privacy
	postID, createdAt, err := database.InsertPostWithPrivacy(db, userID, title, content, imgOrGif, privacyLevel)
	if err != nil {
		fmt.Println("Error inserting post:", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// 7) link categories
	for _, catName := range cats {
		catID, err := database.GetCategoryID(db, catName)
		if err != nil {
			log.Println("GetCategoryID error:", err)
			http.Error(w, "Failed to retrieve category", http.StatusInternalServerError)
			return
		}
		if err := database.InsertPostCategory(db, int(postID), catID); err != nil {
			log.Println("InsertPostCategory error:", err)
			http.Error(w, "Failed to link category", http.StatusInternalServerError)
			return
		}
	}

	selectedFollowersStr := r.MultipartForm.Value["selected_followers"]
	var selectedFollowersInt []int

	for _, strID := range selectedFollowersStr {
		id, err := strconv.Atoi(strID)
		if err != nil {
			// handle the error appropriately, maybe skip or log
			continue
		}
		selectedFollowersInt = append(selectedFollowersInt, id)
	}
	// Handle private post permissions (privacy level 2)
	if privacyLevel == 2 && len(selectedFollowersInt) > 0 {
		err = database.AddPostPermissions(db, int(postID), selectedFollowersInt)
		if err != nil {
			fmt.Println("Error adding post permissions:", err)
			http.Error(w, "Failed to set post permissions", http.StatusInternalServerError)
			return
		}
	}

	// Get username for WebSocket broadcast
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		username = "Unknown"
	}

	// Broadcast new post notification via WebSocket to all connected users
	if hub != nil {
		// Fetch the complete post data to send to clients
		var avatar sql.NullString
		postData := map[string]interface{}{
			"id":         postID,
			"title":      title,
			"content":    content,
			"image":      imgOrGif,
			"username":   username,
			"userID":     userID,
			"createdAt":  createdAt,
			"categories": cats,
		}

		// Get user avatar if exists (use the same column other helpers expect)
		db.QueryRow("SELECT avatar_url FROM users WHERE id = ?", userID).Scan(&avatar)
		if avatar.Valid {
			// frontend expects `avatar_url` in post objects
			postData["avatar_url"] = avatar.String
		}

		postNotification := chat.Frontend{
			Type:      "new_post",
			From:      userID,
			Username:  username,
			PostId:    int(postID),
			Content:   content,
			Timestamp: time.Now(),
		}

		// Send to all connected clients with the post data encoded in Content as JSON
		postJSON, _ := json.Marshal(postData)
		postNotification.Content = string(postJSON)

		hub.Mutex.RLock()
		for _, client := range hub.Clients {
			select {
			case client.Send <- postNotification:
			default:
				// Skip if client buffer is full
			}
		}
		hub.Mutex.RUnlock()
	}

	// Send success response
	response := map[string]interface{}{
		"success":       true,
		"message":       "Post created successfully.",
		"postID":        postID,
		"createdAt":     createdAt,
		"privacy_level": privacyLevel,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserFollowers returns followers for post privacy selection
func GetUserFollowers(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Get followers
	followers, err := database.GetUserFollowers(db, userID)
	if err != nil {
		fmt.Println("Error getting followers:", err)
		http.Error(w, "Failed to get followers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"followers": followers,
	})
}
