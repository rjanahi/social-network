package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
	"strconv"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func GetComments(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		fmt.Println(" Invalid post ID:", postIDStr)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	comments, err := GetCommentsByPostID(db, postID)
	if err != nil {
		fmt.Println(" Error retrieving comments:", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Failed to retrieve comments"}`, http.StatusInternalServerError)
		return
	}

	//  Always return a JSON array (never `nil`)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func CreateComment(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Log received Content-Type
	contentType := r.Header.Get("Content-Type")
	fmt.Println(" Received Content-Type:", contentType)

	// Ensure the request content type is JSON
	if contentType != "application/json" {
		fmt.Println(" Error: Expected JSON but received:", contentType)
		http.Error(w, "Invalid content type, expected application/json", http.StatusUnsupportedMediaType)
		return
	}

	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(" Error reading request body:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse JSON request
	var requestData struct {
		PostID  int    `json:"post_id"`
		Content string `json:"content"`
	}

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		fmt.Println(" JSON Decoding Error:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	fmt.Printf(" Parsed Comment Request: PostID: %d, Content: '%s'\n", requestData.PostID, requestData.Content)

	// Validate inputs
	if requestData.Content == "" {
		http.Error(w, "Comment content cannot be empty.", http.StatusBadRequest)
		return
	}
	if requestData.PostID <= 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Insert comment into the database
	_, _, err = database.InsertComment(db, requestData.PostID, userID, requestData.Content)
	if err != nil {
		fmt.Println(" Error inserting comment:", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Send success response
	response := map[string]interface{}{
		"success": true,
		"message": "Comment added successfully.",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetCommentsByPostID(db *sql.DB, postID int) ([]Comment, error) {
	query := `SELECT c.id, c.user_id, u.username, c.content, c.created_at 
              FROM comments c
              JOIN users u ON c.user_id = u.id
              WHERE c.post_id = ?
              ORDER BY c.created_at ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		fmt.Println(" Database Query Error:", err)
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.Username, &comment.Content, &comment.CreatedAt)
		if err != nil {
			fmt.Println(" Row Scanning Error:", err)
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		fmt.Println(" Iteration Error:", err)
		return nil, err
	}

	if comments == nil {
		comments = []Comment{}
	}

	return comments, nil
}
