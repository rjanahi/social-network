package likerepo

import (
	"context"
	"database/sql"
	"fmt"

	"socialnetwork/pkg/apis/like"
)

type LikesRepository struct {
	db *sql.DB
}

func NewLikesRepository(db *sql.DB) *LikesRepository {
	return &LikesRepository{db: db}
}

func (r *LikesRepository) StoreInteraction(ctx context.Context, interaction like.Like) error {
	var postID sql.NullInt32
	var commentID sql.NullInt32

	if interaction.PostID != nil {
		postID = sql.NullInt32{Int32: int32(*interaction.PostID), Valid: true}
	} else {
		postID = sql.NullInt32{Valid: false}
	}

	if interaction.CommentID != nil {
		commentID = sql.NullInt32{Int32: int32(*interaction.CommentID), Valid: true}
	} else {
		commentID = sql.NullInt32{Valid: false}
	}

	query := `
        INSERT INTO likes (user_id, post_id, comment_id, is_like, created_at)
        VALUES ($1, $2, $3, $4, datetime('now'))
    `
	_, err := r.db.ExecContext(ctx, query, interaction.UserID, postID, commentID, interaction.IsLike)
	if err != nil {
		return fmt.Errorf("store interaction failed: %w", err)
	}

	return nil
}

func (r *LikesRepository) RemovePostInteraction(ctx context.Context, userID, postID int) error {
	query := `
        DELETE FROM likes WHERE user_id = $1 AND post_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, userID, postID)
	if err != nil {
		return fmt.Errorf("remove post interaction failed: %w", err)
	}
	return nil
}

func (r *LikesRepository) RemoveCommentInteraction(ctx context.Context, userID, commentID int) error {
	query := `
        DELETE FROM likes WHERE user_id = $1 AND comment_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, userID, commentID)
	if err != nil {
		return fmt.Errorf("remove comment interaction failed: %w", err)
	}
	return nil
}

func (r *LikesRepository) CheckPostInteractions(ctx context.Context, userID, postID int) (like like.Like, err error) {
	query := `
		SELECT user_id, post_id, comment_id, is_like FROM likes WHERE user_id = $1 AND post_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, userID, postID)
	if err = row.Scan(&like.UserID, &like.PostID, &like.CommentID, &like.IsLike); err != nil {
		return like, fmt.Errorf("get post interaction failed: %w", err)
	}
	return like, nil
}

func (r *LikesRepository) CheckCommentInteractions(ctx context.Context, userID, commentID int) (like like.Like, err error) {
	query := `
		SELECT user_id, post_id, comment_id, is_like FROM likes WHERE user_id = $1 AND comment_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, userID, commentID)
	if err := row.Scan(&like.UserID, &like.PostID, &like.CommentID, &like.IsLike); err != nil {
		return like, fmt.Errorf("get comment interaction failed: %w", err)
	}
	return like, nil
}

func (r *LikesRepository) GetUserIDFromSession(ctx context.Context, token string) (int, error) {
	query := `
		SELECT user_id FROM sessions WHERE token = $1
	`
	row := r.db.QueryRowContext(ctx, query, token)
	var userID int
	if err := row.Scan(&userID); err != nil {
		return 0, fmt.Errorf("get user id from session failed: %w", err)
	}
	return userID, nil
}

func (r *LikesRepository) GetPostsInteractions(ctx context.Context, postID int) (resp like.GetInteractionsResponse, err error) {
	likesQuery := `
		SELECT post_id, COUNT(*) FROM likes WHERE post_id = $1 AND is_like = TRUE
	`
	dislikesQuery := `
		SELECT post_id, COUNT(*) FROM likes WHERE post_id = $1 AND is_like = FALSE
	`
	row := r.db.QueryRowContext(ctx, likesQuery, postID)
	if err = row.Scan(&resp.PostID, &resp.Likes); err != nil {
		return resp, fmt.Errorf("get post likes failed: %w", err)
	}
	row = r.db.QueryRowContext(ctx, dislikesQuery, postID)
	if err = row.Scan(&resp.PostID, &resp.Dislikes); err != nil {
		return resp, fmt.Errorf("get post dislikes failed: %w", err)
	}

	return resp, nil
}

func (r *LikesRepository) GetCommentsInteractions(ctx context.Context, commentID int) (resp like.GetInteractionsResponse, err error) {
	query := `
		SELECT comment_id, 
		       COUNT(*) FILTER (WHERE is_like = TRUE) AS likes, 
		       COUNT(*) FILTER (WHERE is_like = FALSE) AS dislikes
		FROM likes 
		WHERE comment_id = $1
		GROUP BY comment_id
	`

	row := r.db.QueryRowContext(ctx, query, commentID)
	if err = row.Scan(&resp.CommentID, &resp.Likes, &resp.Dislikes); err != nil {
		if err == sql.ErrNoRows {
			return resp, nil
		}
		return resp, fmt.Errorf("failed to get interactions for comment %d: %w", commentID, err)
	}

	return resp, nil
}
