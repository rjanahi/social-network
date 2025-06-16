package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func DeleteUser(db *sql.DB, username string) error {
	query := `DELETE FROM users WHERE username = ?`
	_, err := db.Exec(query, username)
	return err
}

func DeleteCategory(db *sql.DB, name string) error {
	query := `DELETE FROM categories WHERE name = ?`
	_, err := db.Exec(query, name)
	return err
}

func DeletePost(db *sql.DB, postID int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := db.Exec(query, postID)
	return err
}

func DeletePostCategory(db *sql.DB, postID, categoryID int) error {
	query := `DELETE FROM post_categories WHERE post_id = ? AND category_id = ?`
	_, err := db.Exec(query, postID, categoryID)
	return err
}

func DeleteComment(db *sql.DB, commentID int) error {
	query := `DELETE FROM comments WHERE id = ?`
	_, err := db.Exec(query, commentID)
	return err
}

func DeleteLike(db *sql.DB, likeID int) error {
	query := `DELETE FROM likes WHERE id = ?`
	_, err := db.Exec(query, likeID)
	return err
}

func DeleteSession(db *sql.DB, sessionID int) error {
	fmt.Println(" Deleting session with ID:", sessionID)
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := db.Exec(query, sessionID)
	if err != nil {
		fmt.Println(" Error deleting session:", err)
	}
	return err
}
