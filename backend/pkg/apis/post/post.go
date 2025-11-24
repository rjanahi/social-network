package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	u "socialnetwork/pkg/apis/user"
	database "socialnetwork/pkg/db"
)

// GetPosts returns posts visible to the current user based on privacy settings
func GetPosts(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate session and get user ID
	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	// Complex query that respects privacy settings
	query := `
		SELECT 
			p.id, u.username, u.firstname, u.lastname, u.avatar_url, p.user_id, p.title, p.content, 
			COALESCE(p.privacy_level, 0) as privacy_level, 
			p.imgOrgif, 
			p.created_at,
			COALESCE(likes.count, 0) as likes_count,
			COALESCE(dislikes.count, 0) as dislikes_count,
			COALESCE(comments.count, 0) as comments_count
		FROM posts p
		JOIN users u ON p.user_id = u.id
        LEFT JOIN (
            SELECT post_id, COUNT(*) as count 
            FROM likes 
            WHERE is_like = 1 AND comment_id IS NULL 
            GROUP BY post_id
        ) likes ON p.id = likes.post_id
        LEFT JOIN (
            SELECT post_id, COUNT(*) as count 
            FROM likes 
            WHERE is_like = 0 AND comment_id IS NULL 
            GROUP BY post_id
        ) dislikes ON p.id = dislikes.post_id
        LEFT JOIN (
            SELECT post_id, COUNT(*) as count 
            FROM comments 
            GROUP BY post_id
        ) comments ON p.id = comments.post_id
        WHERE 
            -- Public posts (everyone can see)
            COALESCE(p.privacy_level, 0) = 0
            OR
            -- User's own posts (can always see own posts)
            p.user_id = ?
            OR
            -- Almost private posts (only mutual followers can see - both must follow each other)
            (COALESCE(p.privacy_level, 0) = 1 AND EXISTS (
                SELECT 1 FROM userFollow 
                WHERE follower_id = ? AND following_id = p.user_id
            ) AND EXISTS (
                SELECT 1 FROM userFollow 
                WHERE follower_id = p.user_id AND following_id = ?
            ))
            OR
            -- Private posts (only selected followers with mutual follow can see)
            (COALESCE(p.privacy_level, 0) = 2 AND EXISTS (
                SELECT 1 FROM post_permissions 
                WHERE post_id = p.id AND user_id = ?
            ) AND EXISTS (
                SELECT 1 FROM userFollow 
                WHERE follower_id = ? AND following_id = p.user_id
            ) AND EXISTS (
                SELECT 1 FROM userFollow 
                WHERE follower_id = p.user_id AND following_id = ?
            ))
        ORDER BY p.created_at DESC
    `

	rows, err := db.Query(query, userID, userID, userID, userID, userID, userID)
	if err != nil {
		fmt.Println("Error retrieving posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold posts
	var posts []map[string]interface{}
	for rows.Next() {
		var postID, privacyLevel, likesCount, dislikesCount, commentsCount, postUserID int
		var username, firstname, lastname, avatarURL, title, content, imgOrgif string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &firstname, &lastname, &avatarURL, &postUserID, &title, &content, &privacyLevel, &imgOrgif, &createdAt, &likesCount, &dislikesCount, &commentsCount)
		if err != nil {
			fmt.Println("Error scanning post:", err)
			http.Error(w, "Failed to process posts", http.StatusInternalServerError)
			return
		}

		// Fetch categories for this post
		categories, err := database.GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println("Error retrieving categories for post:", err)
			categories = []string{} // Default to empty if error
		}

		// Get privacy level text
		privacyText := "Public"
		switch privacyLevel {
		case 1:
			privacyText = "Almost Private"
		case 2:
			privacyText = "Private"
		}

		// Store post in slice
		post := map[string]interface{}{
			"id":             postID,
			"username":       username,
			"firstname":      firstname,
			"lastname":       lastname,
			"avatar_url":     avatarURL,
			"userID":         postUserID,
			"title":          title,
			"content":        content,
			"privacy_level":  privacyLevel,
			"privacy_text":   privacyText,
			"imgOrgif":       imgOrgif,
			"image":          imgOrgif, // Alias for frontend compatibility
			"categories":     categories,
			"likes_count":    likesCount,
			"dislikes_count": dislikesCount,
			"comments":       commentsCount,
			"createdAt":      createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}

	// Return posts as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func GetPublicPosts(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	query := `
        SELECT DISTINCT 
            p.id, u.username, p.title, p.content, 
            p.imgOrgif,
            COALESCE(p.privacy_level, 0) as privacy_level, 
            p.created_at,
            COALESCE(likes.count, 0) as likes_count,
            COALESCE(dislikes.count, 0) as dislikes_count,
            COALESCE(comments.count, 0) as comments_count
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN (
            SELECT post_id, COUNT(*) as count 
            FROM likes 
            WHERE is_like = 1 AND comment_id IS NULL 
            GROUP BY post_id
        ) likes ON p.id = likes.post_id
        LEFT JOIN (
            SELECT post_id, COUNT(*) as count 
            FROM likes 
            WHERE is_like = 0 AND comment_id IS NULL 
            GROUP BY post_id
        ) dislikes ON p.id = dislikes.post_id
        LEFT JOIN (
            SELECT post_id, COUNT(*) as count 
            FROM comments 
            GROUP BY post_id
        ) comments ON p.id = comments.post_id
        WHERE 
            COALESCE(p.privacy_level, 0) = 0  -- ONLY PUBLIC posts
        ORDER BY p.created_at DESC
    `

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error retrieving public posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var postID, privacyLevel, likesCount, dislikesCount, commentsCount int
		var username, title, content, imgOrgif string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &title, &content, &imgOrgif, &privacyLevel, &createdAt, &likesCount, &dislikesCount, &commentsCount)
		if err != nil {
			fmt.Println("Error scanning post:", err)
			http.Error(w, "Failed to process posts", http.StatusInternalServerError)
			return
		}

		// Fetch categories for this post
		categories, err := database.GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println("Error retrieving categories for post:", err)
			categories = []string{} // Default to empty if error
		}

		// Store post in slice
		post := map[string]interface{}{
			"id":             postID,
			"username":       username,
			"title":          title,
			"content":        content,
			"imgOrgif":       imgOrgif,
			"image":          imgOrgif, // Alias for frontend compatibility
			"categories":     categories,
			"likes_count":    likesCount,
			"dislikes_count": dislikesCount,
			"comments":       commentsCount,
			"createdAt":      createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}

	// Return posts as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func GetProfilePosts(db *sql.DB, w http.ResponseWriter, r *http.Request, username string) {
	// who is being viewed
	var profileUserID int
	if err := db.QueryRow(`SELECT id FROM users WHERE username=?`, username).Scan(&profileUserID); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// who is viewing (0 if not logged in)
	viewerID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		viewerID = 0
	}

	rows, err := db.Query(`
        SELECT DISTINCT
            p.id, u.username, p.title, p.content,
            COALESCE(p.privacy_level, 0) AS privacy_level,
            p.imgOrgif,
            p.created_at,
            COALESCE(likes.count, 0)    AS likes_count,
            COALESCE(dislikes.count, 0) AS dislikes_count,
            COALESCE(comments.count, 0) AS comments_count
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN (
          SELECT post_id, COUNT(*) AS count FROM likes
          WHERE is_like = 1 AND comment_id IS NULL GROUP BY post_id
        ) likes    ON p.id = likes.post_id
        LEFT JOIN (
          SELECT post_id, COUNT(*) AS count FROM likes
          WHERE is_like = 0 AND comment_id IS NULL GROUP BY post_id
        ) dislikes ON p.id = dislikes.post_id
        LEFT JOIN (
          SELECT post_id, COUNT(*) AS count FROM comments
          GROUP BY post_id
        ) comments ON p.id = comments.post_id
        WHERE p.user_id = ?
          AND (
               COALESCE(p.privacy_level,0) = 0                         -- public
            OR (? > 0 AND ? = p.user_id)                               -- viewer is owner
            OR (COALESCE(p.privacy_level,0) = 1 AND EXISTS (           -- followers-only (mutual follow required)
                 SELECT 1 FROM userFollow f
                 WHERE f.follower_id = ? AND f.following_id = p.user_id
            ) AND EXISTS (
                 SELECT 1 FROM userFollow f
                 WHERE f.follower_id = p.user_id AND f.following_id = ?
            ))
            OR (COALESCE(p.privacy_level,0) = 2 AND EXISTS (           -- explicit permission with mutual follow
                 SELECT 1 FROM post_permissions pp
                 WHERE pp.post_id = p.id AND pp.user_id = ?
            ) AND EXISTS (
                 SELECT 1 FROM userFollow f
                 WHERE f.follower_id = ? AND f.following_id = p.user_id
            ) AND EXISTS (
                 SELECT 1 FROM userFollow f
                 WHERE f.follower_id = p.user_id AND f.following_id = ?
            ))
          )
        ORDER BY p.created_at DESC
    `, profileUserID, viewerID, viewerID, viewerID, viewerID, viewerID, viewerID, viewerID)
	if err != nil {
		fmt.Println("Error retrieving profile posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []map[string]any
	for rows.Next() {
		var (
			postID, privacyLevel, likesCount, dislikesCount, commentsCount int
			username, title, content, imgOrgif                             string
			createdAt                                                      time.Time
		)
		if err := rows.Scan(&postID, &username, &title, &content, &privacyLevel, &imgOrgif, &createdAt, &likesCount, &dislikesCount, &commentsCount); err != nil {
			http.Error(w, "Failed to process posts", http.StatusInternalServerError)
			return
		}
		categories, _ := database.GetCategoriesByPostID(db, postID)

		privacyText := map[int]string{0: "Public", 1: "Followers", 2: "Private"}[privacyLevel]

		posts = append(posts, map[string]any{
			"id": postID, "username": username, "title": title, "content": content,
			"privacy_level": privacyLevel, "privacy_text": privacyText,
			"imgOrgif": imgOrgif, "image": imgOrgif, "categories": categories,
			"likes_count": likesCount, "dislikes_count": dislikesCount,
			"comments":  commentsCount,
			"createdAt": createdAt.Format("2006-01-02 15:04:05"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(posts)
}
