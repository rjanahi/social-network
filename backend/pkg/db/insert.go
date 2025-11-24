package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func InsertUser(db *sql.DB, username, nickname, email, fname, lname, age, gender, password, bio, avatar, dateOfBirth string) (int64, error) {
	query := `INSERT INTO users (username, nickname, firstname, lastname, age, gender, email, password, bio, avatar_url, date_of_birth) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, username, nickname, fname, lname, age, gender, email, password, bio, avatar, dateOfBirth)
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	return lastID, err
}

func InsertCategory(db *sql.DB, name string) (int, error) {
	query := `INSERT INTO categories (name) VALUES (?)`
	result, err := db.Exec(query, name)
	if err != nil {
		return -1, err
	}
	// Retrieve the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err // Return zero time and the error if ID retrieval fails
	}
	return int(id), err
}

func InsertPost(db *sql.DB, user_id int, title, content, imgOrgif string) (int64, time.Time, error) {
	query := `INSERT INTO posts (user_id, title, content,imgOrgif) VALUES (?, ?, ?,?)`
	result, err := db.Exec(query, user_id, title, content, imgOrgif)
	if err != nil {
		return -1, time.Time{}, err
	}
	// Retrieve the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return -1, time.Time{}, err // Return zero time and the error if ID retrieval fails
	}

	// Query the created_at timestamp for the newly inserted post
	var createdAt time.Time
	query = `SELECT created_at FROM posts WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&createdAt)
	if err != nil {
		return -1, time.Time{}, err // Return zero time and the error if timestamp retrieval fails
	}

	return id, createdAt, err
}

func InsertPostCategory(db *sql.DB, postID int, categoryID int) error {
	query := `INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`
	_, err := db.Exec(query, postID, categoryID)
	return err
}

func InsertComment(db *sql.DB, postID, userID int, content, imgOrgif string) (int64, time.Time, error) {
	query := `INSERT INTO comments (post_id, user_id, content, imgOrgif) VALUES (?, ?, ?, ?)`

	// Execute the insert query
	result, err := db.Exec(query, postID, userID, content, imgOrgif)
	if err != nil {
		return -1, time.Time{}, err
	}

	// Retrieve the auto-generated ID of the new comment
	id, err := result.LastInsertId()
	if err != nil {
		return -1, time.Time{}, err
	}

	// Query the created_at timestamp for the newly inserted comment
	var createdAt time.Time
	query = `SELECT created_at FROM comments WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&createdAt)
	if err != nil {
		return id, time.Time{}, err
	}

	return id, createdAt, nil
}

func InsertLike(db *sql.DB, userID, postID, commentID int, isLike bool) error {
	query := `INSERT INTO likes (user_id, post_id, comment_id, is_like) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, userID, postID, commentID, isLike)

	return err
}

func InsertSession(db *sql.DB, userID int, token string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, userID, token, expiresAt)
	if err != nil {
		fmt.Println(" SQL Error inserting session:", err)
	}
	return err
}

func GetCategoryID(db *sql.DB, categoryName string) (int, error) {
	query := `SELECT id FROM categories WHERE name = ?`
	var categoryID int
	err := db.QueryRow(query, categoryName).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If category doesn't exist, create it
			categoryID, err = InsertCategory(db, categoryName)
			if err != nil {
				return -1, err
			}
		} else {
			return -1, err
		}
	}
	return categoryID, nil
}

// InsertGroup creates a new group and returns its ID and creation time
func InsertGroup(db *sql.DB, creatorID int, title, description string) (int64, time.Time, error) {
	query := `INSERT INTO groups (creator_id, title, description) VALUES (?, ?, ?)`
	result, err := db.Exec(query, creatorID, title, description)
	if err != nil {
		return -1, time.Time{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, time.Time{}, err
	}

	// Get the creation time
	var createdAt time.Time
	err = db.QueryRow(`SELECT created_at FROM groups WHERE id = ?`, id).Scan(&createdAt)
	if err != nil {
		return id, time.Now(), nil // Return current time if query fails
	}

	return id, createdAt, nil
}

// InsertGroupMember adds a user to a group
func InsertGroupMember(db *sql.DB, groupID int, userID int, status string, isAdmin bool) error {
	// Use an upsert so that if a membership record already exists we update it
	// to the new status (e.g., accepted) instead of failing with a UNIQUE
	// constraint. This handles cases where a pending/declined record exists.
	query := `INSERT INTO group_members (group_id, user_id, status, is_admin, joined_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(group_id, user_id) DO UPDATE SET status=excluded.status, is_admin=excluded.is_admin, joined_at=CURRENT_TIMESTAMP`
	_, err := db.Exec(query, groupID, userID, status, isAdmin)
	return err
}

// InsertGroupInvitation creates a group invitation
func InsertGroupInvitation(db *sql.DB, groupID, inviterID, inviteeID int) error {
	query := `INSERT INTO group_invitations (group_id, inviter_id, invitee_id) VALUES (?, ?, ?)`
	_, err := db.Exec(query, groupID, inviterID, inviteeID)
	return err
}

// InsertPostWithPrivacy creates a new post with privacy level
func InsertPostWithPrivacy(db *sql.DB, userID int, title, content, imgOrGif string, privacyLevel int) (int64, string, error) {
	query := `INSERT INTO posts (user_id, title, content, imgOrGif, privacy_level) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, userID, title, content, imgOrGif, privacyLevel)
	if err != nil {
		return -1, "", err
	}

	// Retrieve the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return -1, "", err
	}

	// Query the created_at timestamp for the newly inserted post
	var createdAt time.Time
	query = `SELECT created_at FROM posts WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&createdAt)
	if err != nil {
		return -1, "", err
	}

	return id, createdAt.Format("2006-01-02 15:04:05"), nil
}

// Add this function to insert.go
func InsertPostPermission(db *sql.DB, postID, userID int) error {
	query := `INSERT INTO post_permissions (post_id, user_id) VALUES (?, ?)`
	_, err := db.Exec(query, postID, userID)
	return err
}

// AddPostPermissions adds specific user permissions for private posts
func AddPostPermissions(db *sql.DB, postID int, userIDs []int) error {
	if len(userIDs) == 0 {
		return nil
	}

	query := `INSERT OR IGNORE INTO post_permissions (post_id, user_id) VALUES (?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.Exec(postID, userID)
		if err != nil {
			return fmt.Errorf("failed to add permission for user %d: %w", userID, err)
		}
	}

	return nil
}
