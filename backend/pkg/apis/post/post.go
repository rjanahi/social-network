package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	database "socialnetwork/pkg/db"
	"time"
)

func GetPosts(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query posts from database
	rows, err := db.Query(`SELECT posts.id, users.username, posts.title, posts.content, posts.created_at 
                           FROM posts 
                           JOIN users ON posts.user_id = users.id 
                           ORDER BY posts.created_at DESC`)
	if err != nil {
		fmt.Println(" Error retrieving posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold posts
	var posts []map[string]interface{}

	for rows.Next() {
		var postID int
		var username, title, content string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &title, &content, &createdAt)
		if err != nil {
			fmt.Println(" Error scanning post:", err)
			http.Error(w, "Failed to process posts", http.StatusInternalServerError)
			return
		}

		// Fetch categories for this post
		categories, err := database.GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println(" Error retrieving categories for post:", err)
			http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
			return
		}

		// Store post in slice
		post := map[string]interface{}{
			"id":         postID,
			"username":   username,
			"title":      title,
			"content":    content,
			"categories": categories,
			"createdAt":  createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}

	// Return posts as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func GetPostsID(db *sql.DB, w http.ResponseWriter, r *http.Request) ([]int, error) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil, nil
	}

	// Query posts from database
	rows, err := db.Query(`SELECT posts.id, users.username, posts.title, posts.content, posts.created_at 
                           FROM posts 
                           JOIN users ON posts.user_id = users.id 
                           ORDER BY posts.created_at DESC`)
	if err != nil {
		fmt.Println(" Error retrieving posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return nil, nil
	}
	defer rows.Close()

	var posts []int

	for rows.Next() {
		var postID int
		var username, title, content string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &title, &content, &createdAt)
		if err != nil {
			fmt.Println(" Error scanning post:", err)
			http.Error(w, "Failed to process posts", http.StatusInternalServerError)
			return nil, nil
		}
		posts = append(posts, postID)

	}

	return posts, nil
}
