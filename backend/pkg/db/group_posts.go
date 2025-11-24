package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"
)

func getOrCreateCategoryIDTx(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow(`SELECT id FROM categories WHERE LOWER(name)=LOWER(?)`, name).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := tx.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
		if err != nil {
			return 0, err
		}
		lastID, _ := res.LastInsertId()
		return int(lastID), nil
	}
	return id, err
}

func GetGroupPosts(db *sql.DB, groupID, viewerID int) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
				 SELECT gp.id, gp.group_id, gp.user_id, gp.title, gp.content, gp.created_at,
								gp.imgOrgif,
								u.username, u.firstname, u.lastname, u.avatar_url,
								IFNULL(l.cnt, 0)  AS like_count,
								IFNULL(dl.cnt, 0) AS dislike_count,
								IFNULL(c.cnt, 0)  AS comment_count,
								CASE WHEN ul.user_id  IS NULL THEN 0 ELSE 1 END AS is_liked,
								CASE WHEN ud.user_id IS NULL THEN 0 ELSE 1 END AS is_disliked
				 FROM group_posts gp
				 JOIN users u ON u.id = gp.user_id
				 LEFT JOIN (SELECT group_post_id, COUNT(*) AS cnt FROM group_post_likes GROUP BY group_post_id) l  ON l.group_post_id  = gp.id
				 LEFT JOIN (SELECT group_post_id, COUNT(*) AS cnt FROM group_post_dislikes GROUP BY group_post_id) dl ON dl.group_post_id = gp.id
				 LEFT JOIN (SELECT group_post_id, COUNT(*) AS cnt FROM group_post_comments GROUP BY group_post_id) c  ON c.group_post_id  = gp.id
				 LEFT JOIN (SELECT group_post_id, user_id FROM group_post_likes WHERE user_id = ?)  ul ON ul.group_post_id = gp.id
				 LEFT JOIN (SELECT group_post_id, user_id FROM group_post_dislikes WHERE user_id = ?) ud ON ud.group_post_id = gp.id
				 WHERE gp.group_id = ?
					 AND EXISTS (
								 SELECT 1 FROM group_members m
								 WHERE m.group_id = gp.group_id
									 AND m.user_id  = ?
									 AND m.status   = 'accepted'
					 )
				 ORDER BY gp.created_at DESC
			`, viewerID, viewerID, groupID, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var (
			id, gid, uid                                             int
			likeCount, dislikeCount                                  int
			commentCount                                             int
			isLikedInt, isDislikedInt                                int
			title, content, username, firstname, lastname, avatarURL string
			createdAt                                                time.Time
			imgOrgif                                                 sql.NullString
		)
		if err := rows.Scan(
			&id, &gid, &uid, &title, &content, &createdAt,
			&imgOrgif,
			&username, &firstname, &lastname, &avatarURL,
			&likeCount, &dislikeCount, &commentCount, &isLikedInt, &isDislikedInt,
		); err != nil {
			return nil, err
		}

		// categories
		crows, err := db.Query(`
            SELECT c.name
            FROM group_post_categories gpc
            JOIN categories c ON c.id = gpc.category_id
            WHERE gpc.group_post_id = ?`, id)
		if err != nil {
			return nil, err
		}
		var cats []string
		for crows.Next() {
			var n string
			_ = crows.Scan(&n)
			cats = append(cats, n)
		}
		crows.Close()

		out = append(out, map[string]interface{}{
			"id":            id,
			"group_id":      gid,
			"user_id":       uid,
			"title":         title,
			"content":       content,
			"created_at":    createdAt.Format("2006-01-02 15:04:05"),
			"username":      username,
			"firstname":     firstname,
			"lastname":      lastname,
			"avatar_url":    avatarURL,
			"categories":    cats,
			"imgOrgif":      imgOrgif.String,
			"image":         imgOrgif.String,
			"like_count":    likeCount,
			"dislike_count": dislikeCount,
			"comment_count": commentCount,
			"is_liked":      isLikedInt == 1,
			"is_disliked":   isDislikedInt == 1,
		})
	}
	return out, rows.Err()
}

func GetGroupPostComments(db *sql.DB, groupPostID int) ([]map[string]interface{}, error) {
	rows, err := db.Query(`
			SELECT c.id, c.group_post_id, c.user_id, c.content, c.created_at, c.imgOrgif, u.username, u.firstname, u.lastname, u.avatar_url
			FROM group_post_comments c
			JOIN users u ON u.id = c.user_id
			WHERE c.group_post_id = ?
			ORDER BY c.created_at ASC
		`, groupPostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var (
			id, gid, uid                                      int
			content, username, firstname, lastname, avatarURL string
			createdAt                                         time.Time
			imgOrgif                                          sql.NullString
		)
		if err := rows.Scan(&id, &gid, &uid, &content, &createdAt, &imgOrgif, &username, &firstname, &lastname, &avatarURL); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":            id,
			"group_post_id": gid,
			"user_id":       uid,
			"content":       content,
			"imgOrgif":      imgOrgif.String,
			"image":         imgOrgif.String,
			"createdAt":     createdAt.Format("2006-01-02 15:04:05"),
			"created_at":    createdAt.Format("2006-01-02 15:04:05"),
			"username":      username,
			"firstname":     firstname,
			"lastname":      lastname,
			"avatar_url":    avatarURL,
		})
	}
	return out, rows.Err()
}

func InsertGroupPostComment(db *sql.DB, groupPostID, userID int, content, imgOrgif string) (int64, time.Time, error) {
	query := `INSERT INTO group_post_comments (group_post_id, user_id, content, imgOrgif) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, groupPostID, userID, content, imgOrgif)
	if err != nil {
		return -1, time.Time{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, time.Time{}, err
	}
	var createdAt time.Time
	query = `SELECT created_at FROM group_post_comments WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&createdAt)
	if err != nil {
		return -1, time.Time{}, err
	}
	return id, createdAt, nil
}

func GetGroupPostOwnerAndGroup(db *sql.DB, postID int) (groupID int, ownerID int, err error) {
	err = db.QueryRow(`SELECT group_id, user_id FROM group_posts WHERE id=?`, postID).Scan(&groupID, &ownerID)
	return
}

func IsGroupAdmin(db *sql.DB, groupID, userID int) (bool, error) {
	var isAdmin bool
	err := db.QueryRow(`
		SELECT COALESCE(is_admin, 0) = 1
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status='accepted'`, groupID, userID).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return isAdmin, err
}

func DeleteGroupPost(db *sql.DB, postID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1) Get image path (if any) before deleting
	var imgPath sql.NullString
	if err = tx.QueryRow(`SELECT imgOrgif FROM group_posts WHERE id=?`, postID).Scan(&imgPath); err != nil {
		return err
	}

	// 2) Delete children first (if you don’t have ON DELETE CASCADE)
	if _, err = tx.Exec(`DELETE FROM group_post_likes     WHERE group_post_id=?`, postID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM group_post_dislikes  WHERE group_post_id=?`, postID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM group_post_comments  WHERE group_post_id=?`, postID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM group_post_categories WHERE group_post_id=?`, postID); err != nil {
		return err
	}

	// 3) Delete the post
	if _, err = tx.Exec(`DELETE FROM group_posts WHERE id=?`, postID); err != nil {
		return err
	}

	// 4) Commit DB work before touching filesystem
	if err = tx.Commit(); err != nil {
		return err
	}

	if imgPath.Valid && imgPath.String != "" {
		publicRoot := "../frontend-next/public"
		abs := filepath.Join(publicRoot, filepath.Clean(imgPath.String))
		_ = os.Remove(abs) // ignore error: file might not exist
	}

	return nil
}

func InsertGroupPost(db *sql.DB, groupID int, userID int, title, content, imgOrgif string) (int64, time.Time, error) {
	query := `INSERT INTO group_posts (group_id, user_id, title, content, imgOrgif) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, groupID, userID, title, content, imgOrgif)
	if err != nil {
		return -1, time.Time{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, time.Time{}, err
	}
	var createdAt time.Time
	query = `SELECT created_at FROM group_posts WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&createdAt)
	if err != nil {
		return -1, time.Time{}, err
	}
	return id, createdAt, nil
}

func AddGroupPostCategory(db *sql.DB, groupPostID, categoryID int) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO group_post_categories (group_post_id, category_id) VALUES (?, ?)`,
		groupPostID, categoryID)
	return err
}

func AddGroupPostLike(db *sql.DB, groupPostID, userID int) error {
	if _, err := db.Exec(`DELETE FROM group_post_dislikes WHERE group_post_id=? AND user_id=?`, groupPostID, userID); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO group_post_likes (group_post_id, user_id) VALUES (?, ?)`, groupPostID, userID)
	return err
}

func RemoveGroupPostLike(db *sql.DB, groupPostID, userID int) error {
	_, err := db.Exec(`DELETE FROM group_post_likes WHERE group_post_id = ? AND user_id = ?`,
		groupPostID, userID)
	return err
}

func AddGroupPostDislike(db *sql.DB, groupPostID, userID int) error {
	if _, err := db.Exec(`DELETE FROM group_post_likes WHERE group_post_id=? AND user_id=?`, groupPostID, userID); err != nil {
		return err
	}
	_, err := db.Exec(`INSERT OR IGNORE INTO group_post_dislikes (group_post_id, user_id) VALUES (?, ?)`, groupPostID, userID)
	return err
}

func RemoveGroupPostDislike(db *sql.DB, groupPostID, userID int) error {
	_, err := db.Exec(`DELETE FROM group_post_dislikes WHERE group_post_id = ? AND user_id = ?`,
		groupPostID, userID)
	return err
}

// GetGroupPostLikeCounts returns the like and dislike counts for a group post
func GetGroupPostLikeCounts(db *sql.DB, groupPostID int) (likes int, dislikes int, err error) {
	err = db.QueryRow(`
		SELECT 
			IFNULL((SELECT COUNT(*) FROM group_post_likes WHERE group_post_id = ?), 0) AS likes,
			IFNULL((SELECT COUNT(*) FROM group_post_dislikes WHERE group_post_id = ?), 0) AS dislikes
	`, groupPostID, groupPostID).Scan(&likes, &dislikes)
	return likes, dislikes, err
}
