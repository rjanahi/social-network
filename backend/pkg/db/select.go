package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func GetAllUserNames(db *sql.DB) ([]string, error) {
	query := `SELECT username FROM users`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold the results
	var usernames []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}

func GetAllUserEmails(db *sql.DB) ([]string, error) {
	query := `SELECT email FROM users`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func GetUserID(db *sql.DB, username string) (int, error) {
	query := `SELECT id FROM users WHERE username = ?`
	var id int
	err := db.QueryRow(query, username).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil
		}
		return -1, err
	}
	return id, nil
}

func GetUsernameUsingID(db *sql.DB, id int) (string, error) {
	query := `SELECT username FROM users WHERE id = ?`
	var username string
	err := db.QueryRow(query, id).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return username, nil
}

func GetPostIDbyUserID(db *sql.DB, userID int) (int, error) {
	query := `SELECT id FROM posts WHERE user_id = ?`
	var id int
	err := db.QueryRow(query, userID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil
		}
		return -1, err
	}
	return id, nil
}

func GetActiveSessionbyUserID(db *sql.DB, userID int) (int, error) {
	query := `SELECT id FROM sessions WHERE user_id = ? AND expires_at > ?`
	var id int
	currentTime := time.Now().UTC()
	err := db.QueryRow(query, userID, currentTime).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil // No active session
		}
		return -1, err
	}
	return id, nil // Active session exists
}

func GetCategoriesByPostID(db *sql.DB, postID int) ([]string, error) {
	query := `SELECT categories.name FROM categories 
              JOIN post_categories ON categories.id = post_categories.category_id 
              WHERE post_categories.post_id = ?`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var categoryName string
		if err := rows.Scan(&categoryName); err != nil {
			return nil, err
		}
		categories = append(categories, categoryName)
	}

	return categories, nil
}

func GetPostByPostID(db *sql.DB, postID int) ([]map[string]interface{}, error) {
	query := `
       SELECT p.id, u.username, u.firstname, u.lastname, u.avatar_url, p.title, p.content, p.imgOrgif, p.created_at 
       FROM posts p
       JOIN users u ON p.user_id = u.id 
       WHERE p.id = ?`

	rows, err := db.Query(query, postID)
	if err != nil {
		fmt.Println(" Error retrieving posts:", err)
		return nil, err
	}
	defer rows.Close()
	var posts []map[string]interface{}
	for rows.Next() {
		var postID int
		var username, firstname, lastname, avatarURL, title, content, imgOrgif string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &firstname, &lastname, &avatarURL, &title, &content, &imgOrgif, &createdAt)
		if err != nil {
			fmt.Println(" Error scanning post:", err)
			return nil, err
		}

		// Fetch categories for this post
		categories, err := GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println(" Error retrieving categories for post:", err)
			return nil, err
		}

		// Store post in slice
		post := map[string]interface{}{
			"id":         postID,
			"username":   username,
			"firstname":  firstname,
			"lastname":   lastname,
			"avatar_url": avatarURL,
			"title":      title,
			"content":    content,
			"imgOrgif":   imgOrgif,
			"categories": categories, //  Include categories
			"createdAt":  createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPostsByUserID(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
	SELECT p.id, u.username, u.firstname, u.lastname, u.avatar_url, p.title, p.content, p.imgOrgif, p.created_at 
	FROM posts p
	JOIN users u ON p.user_id = u.id 
	WHERE u.id = ? 
	ORDER BY p.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		fmt.Println(" Error retrieving user posts:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var postID int
		var username, firstname, lastname, avatarURL, title, content, imgOrgif string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &firstname, &lastname, &avatarURL, &title, &content, &imgOrgif, &createdAt)
		if err != nil {
			fmt.Println(" Error scanning user post:", err)
			return nil, err
		}

		categories, err := GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println(" Error retrieving categories:", err)
			return nil, err
		}

		post := map[string]interface{}{
			"id":         postID,
			"username":   username,
			"firstname":  firstname,
			"lastname":   lastname,
			"avatar_url": avatarURL,
			"title":      title,
			"content":    content,
			"imgOrgif":   imgOrgif,
			"categories": categories,
			"createdAt":  createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPostByCategoryID(db *sql.DB, catID int) ([]map[string]interface{}, error) {
	query := `
	SELECT p.id, u.username, p.title, p.content,p.imgOrgif, p.created_at 
	FROM posts p
	JOIN post_categories pc ON pc.post_id = p.id
	JOIN categories c ON c.id = pc.category_id
	JOIN users u ON u.id = p.user_id
	WHERE c.id = ?`

	rows, err := db.Query(query, catID)
	if err != nil {
		return nil, fmt.Errorf("error querying posts: %w", err)
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var postID int
		var username, firstname, lastname, avatarURL, title, content, imgOrgif string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &firstname, &lastname, &avatarURL, &title, &content, &imgOrgif, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err)
		}
		categories, err := GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println(" Error retrieving categories:", err)
			return nil, err
		}

		post := map[string]interface{}{
			"id":         postID,
			"username":   username,
			"firstname":  firstname,
			"lastname":   lastname,
			"avatar_url": avatarURL,
			"title":      title,
			"content":    content,
			"imgOrgif":   imgOrgif,
			"categories": categories,
			"createdAt":  createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func GetCategoryIDByName(db *sql.DB, category string) (int, error) {
	var categoryID int
	query := `SELECT id FROM categories WHERE name = ?`

	err := db.QueryRow(query, category).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("category '%s' not found", category)
		}
		return 0, fmt.Errorf("error retrieving category '%s': %v", category, err)
	}

	return categoryID, nil
}

func GetPostIfLiked(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
	SELECT p.id, u.username, p.title, p.content,p.imgOrgif, p.created_at 
	FROM posts p
	JOIN users u ON u.id = p.user_id
	JOIN likes l ON l.post_id = p.id AND l.is_like = 1
	WHERE l.user_id = ?;`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying posts: %w", err)
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var postID int
		var username, firstname, lastname, avatarURL, title, content, imgOrgif string
		var createdAt time.Time

		err := rows.Scan(&postID, &username, &firstname, &lastname, &avatarURL, &title, &content, &imgOrgif, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err)
		}
		categories, err := GetCategoriesByPostID(db, postID)
		if err != nil {
			fmt.Println("Error retrieving categories:", err)
			return nil, err
		}

		post := map[string]interface{}{
			"id":         postID,
			"username":   username,
			"firstname":  firstname,
			"lastname":   lastname,
			"avatar_url": avatarURL,
			"title":      title,
			"content":    content,
			"imgOrgif":   imgOrgif,
			"categories": categories,
			"createdAt":  createdAt.Format("2006-01-02 15:04:05"),
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func GetUserFollowers(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT u.id, u.username, u.email
        FROM users u
        JOIN userFollow f ON u.id = f.follower_id
        WHERE f.following_id = ?
        ORDER BY u.username
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var followers []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email string

		err := rows.Scan(&id, &username, &email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", err)
		}

		follower := map[string]interface{}{
			"id":       id,
			"username": username,
			"email":    email,
		}

		followers = append(followers, follower)
	}

	return followers, nil
}

// GetUserFollowing returns a list of users that the specified user is following
func GetUserFollowing(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT u.id, u.username, u.email
        FROM users u
        JOIN userFollow f ON u.id = f.following_id
        WHERE f.follower_id = ?
        ORDER BY u.username
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()

	var following []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email string

		err := rows.Scan(&id, &username, &email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan following: %w", err)
		}

		user := map[string]interface{}{
			"id":       id,
			"username": username,
			"email":    email,
		}

		following = append(following, user)
	}

	return following, nil
}

// CheckFollowRelationship checks if at least one user follows the other
// Returns true if userA follows userB OR userB follows userA
func CheckFollowRelationship(db *sql.DB, userA, userB int) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM userFollow
		WHERE (follower_id = ? AND following_id = ?)
		   OR (follower_id = ? AND following_id = ?)
	`
	var hasRelationship bool
	err := db.QueryRow(query, userA, userB, userB, userA).Scan(&hasRelationship)
	if err != nil {
		return false, fmt.Errorf("failed to check follow relationship: %w", err)
	}
	return hasRelationship, nil
}

func CheckUserCanSeePost(db *sql.DB, postID, viewerID int) (bool, error) {
	query := `
        SELECT p.privacy_level, p.user_id
        FROM posts p
        WHERE p.id = ?
    `

	var privacyLevel, authorID int
	err := db.QueryRow(query, postID).Scan(&privacyLevel, &authorID)
	if err != nil {
		return false, fmt.Errorf("failed to get post privacy: %w", err)
	}

	// If it's the author's own post
	if authorID == viewerID {
		return true, nil
	}

	// If it's public
	if privacyLevel == 0 {
		return true, nil
	}

	if privacyLevel == 1 {
		var count int
		err := db.QueryRow(`
            SELECT COUNT(*) FROM userFollow 
            WHERE follower_id = ? AND following_id = ?
        `, viewerID, authorID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}

	if privacyLevel == 2 {
		var count int
		err := db.QueryRow(`
            SELECT COUNT(*) FROM post_permissions 
            WHERE post_id = ? AND user_id = ?
        `, postID, viewerID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}

	return false, nil
}

// GetChatHistory retrieves messages between two users
func GetChatHistory(db *sql.DB, userID1, userID2 int) ([]map[string]interface{}, error) {
	query := `
		SELECT m.id, m.sender_id, m.receiver_id, m.content, m.created_at,
		       u.username as sender_username
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) 
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at ASC
	`

	rows, err := db.Query(query, userID1, userID2, userID2, userID1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, senderID, receiverID int
		var content, senderUsername string
		var createdAt time.Time

		err := rows.Scan(&id, &senderID, &receiverID, &content, &createdAt, &senderUsername)
		if err != nil {
			return nil, err
		}

		messages = append(messages, map[string]interface{}{
			"id":        id,
			"from":      senderID,
			"to":        receiverID,
			"content":   content,
			"username":  senderUsername,
			"timestamp": createdAt.Format(time.RFC3339),
		})
	}

	return messages, nil
}
