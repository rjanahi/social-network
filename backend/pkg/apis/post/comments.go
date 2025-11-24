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

type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	AvatarURL string    `json:"avatar_url"`
	Content   string    `json:"content"`
	ImgOrGif  string    `json:"imgOrgif"`
	Image     string    `json:"image"` // Alias for frontend compatibility
	CreatedAt time.Time `json:"created_at"`
}

// GetComments handles GET /comments?post_id=
func GetComments(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		fmt.Println("Invalid post ID:", postIDStr)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	comments, err := GetCommentsByPostID(db, postID)
	if err != nil {
		fmt.Println("Error retrieving comments:", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Failed to retrieve comments"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// CreateComment handles POST /create-comment with multipart/form-data
func CreateComment(db *sql.DB, hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Println("ParseMultipartForm error:", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	postIDStr := r.FormValue("post_id")
	content := r.FormValue("comment")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
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

		// generate a safe filename
		fname := strings.ReplaceAll(header.Filename, " ", "-")
		dstFile, err := os.Create("../frontend-next/public/img/comments/" + fname)
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
		imgOrGif = "/img/comments/" + fname
	} else if err != http.ErrMissingFile {
		imgOrGif = "" // no file uploaded, set to empty string
	}

	if content == "" && imgOrGif == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	fmt.Println("Image or GIF URL:", imgOrGif)
	if _, _, err := database.InsertComment(db, postID, userID, content, imgOrGif); err != nil {
		fmt.Println("DB insert error:", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Get username for WebSocket broadcast
	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

	// Broadcast new comment notification via WebSocket
	if hub != nil {
		commentNotification := chat.Frontend{
			Type:      "new_comment",
			From:      userID,
			Username:  username,
			PostId:    postID,
			Content:   content,
			Timestamp: time.Now(),
		}

		hub.Mutex.RLock()
		for _, client := range hub.Clients {
			select {
			case client.Send <- commentNotification:
			default:
				// Skip if buffer full
			}
		}
		hub.Mutex.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Comment added successfully",
	})
}

// GetCommentsByPostID fetches comments and maps NULL imgOrgif to nil
func GetCommentsByPostID(db *sql.DB, postID int) ([]Comment, error) {
	const query = `
SELECT
	c.id,
	c.user_id,
	u.username,
	u.firstname,
	u.lastname,
	u.avatar_url,
	c.content,
	COALESCE(c.imgOrgif, '') AS imgOrgif,
	c.created_at
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.post_id = ?
ORDER BY c.created_at ASC
	`

	rows, err := db.Query(query, postID)
	if err != nil {
		fmt.Println("Database Query Error:", err)
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var cmt Comment
		if err := rows.Scan(
			&cmt.ID,
			&cmt.UserID,
			&cmt.Username,
			&cmt.Firstname,
			&cmt.Lastname,
			&cmt.AvatarURL,
			&cmt.Content,
			&cmt.ImgOrGif,
			&cmt.CreatedAt,
		); err != nil {
			fmt.Println("Row Scanning Error:", err)
			return nil, err
		}
		cmt.Image = cmt.ImgOrGif // Set alias for frontend compatibility
		comments = append(comments, cmt)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Iteration Error:", err)
		return nil, err
	}
	return comments, nil
}
