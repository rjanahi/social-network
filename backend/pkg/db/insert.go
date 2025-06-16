package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func InsertUser(db *sql.DB, username, email, fname, lname, age, gender, password string) (int64, error) {
	query := `INSERT INTO users (username, firstname, lastname, age, gender, email, password) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, username, fname, lname, age, gender, email, password)
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

func InsertPost(db *sql.DB, user_id int, title, content string) (int64, time.Time, error) {
	query := `INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)`
	result, err := db.Exec(query, user_id, title, content)
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

func InsertComment(db *sql.DB, postID, userID int, content string) (int64, time.Time, error) {
	query := `INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)`

	// Execute the insert query
	result, err := db.Exec(query, postID, userID, content)
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
