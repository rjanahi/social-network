package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	u "socialnetwork/pkg/apis/user"
	"socialnetwork/pkg/db"
)

// Post structure
type Post struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
}

// CreatePost handles post submission
func CreatePost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var postData Post
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&postData); err != nil {
		fmt.Println(" JSON Decoding Error:", err)
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Validate post fields
	if postData.Title == "" || postData.Content == "" {
		http.Error(w, "Title and Content cannot be empty.", http.StatusBadRequest)
		return
	}

	// Insert the post into the database
	postID, createdAt, err := database.InsertPost(db, userID, postData.Title, postData.Content)
	if err != nil {
		fmt.Println(" Error inserting post:", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	if len(postData.Categories) < 1 {
		category := "none"
		categoryID, err := database.GetCategoryID(db, category)
		if err != nil {
			fmt.Println(" Error getting category ID:", err)
			http.Error(w, "Failed to retrieve category", http.StatusInternalServerError)
			return
		}

		err = database.InsertPostCategory(db, int(postID), categoryID)
		if err != nil {
			fmt.Println(" Error linking post to category:", err)
			http.Error(w, "Failed to associate post with category", http.StatusInternalServerError)
			return
		}
	} else {
		// Insert categories into the database
		for _, categoryName := range postData.Categories {
			categoryID, err := database.GetCategoryID(db, categoryName)
			if err != nil {
				fmt.Println(" Error getting category ID:", err)
				http.Error(w, "Failed to retrieve category", http.StatusInternalServerError)
				return
			}

			err = database.InsertPostCategory(db, int(postID), categoryID)
			if err != nil {
				fmt.Println(" Error linking post to category:", err)
				http.Error(w, "Failed to associate post with category", http.StatusInternalServerError)
				return
			}
		}
	}

	// Send success response
	response := map[string]interface{}{
		"success":   true,
		"message":   "Post created successfully.",
		"postID":    postID,
		"createdAt": createdAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
