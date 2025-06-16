package like

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	u "socialnetwork/pkg/apis/user"
	"log/slog"
	"net/http"
)

type LikesController struct {
	s LikesService
}

func NewLikesController(s LikesService) *LikesController {
	return &LikesController{s: s}
}

func (c *LikesController) LikeDislikePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		return
	}

	var req InteractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "err", err)
		return
	}

	if req.PostID == nil {
		return
	}

	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	like, err := c.s.CheckPostInteractions(r.Context(), userID, *req.PostID)
	if errors.Is(err, sql.ErrNoRows) {
		if err = c.s.InteractWithPost(r.Context(), userID, *req.PostID, req.IsLike); err != nil {
			slog.ErrorContext(r.Context(), "Error in interacting with post", "err", err)

			return
		}
	} else if like.IsLike == req.IsLike {
		if err = c.s.RemovePostInteraction(r.Context(), userID, *req.PostID); err != nil {
			slog.ErrorContext(r.Context(), "Error removing post interaction", "err", err)

			return
		}
	} else {
		if err = c.s.RemovePostInteraction(r.Context(), userID, *req.PostID); err != nil {
			slog.ErrorContext(r.Context(), "Error removing post interaction", "err", err)

			return
		}
		if err = c.s.InteractWithPost(r.Context(), userID, *req.PostID, req.IsLike); err != nil {
			slog.ErrorContext(r.Context(), "Error creating a new interaction with post", "err", err)

			return
		}
	}

	//  Fetch updated like/dislike count to send back to frontend
	updatedCounts, err := c.s.GetPostsInteractions(r.Context(), *req.PostID)
	if err != nil {
		http.Error(w, "Failed to fetch updated interactions", http.StatusInternalServerError)
		return
	}

	//  Send JSON response with updated counts
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Interaction updated successfully",
		"likes":    updatedCounts.Likes,
		"dislikes": updatedCounts.Dislikes,
	})
}

func (c *LikesController) InteractWithComment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		return
	}

	var req InteractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(" Error decoding request body:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.CommentID == nil {
		fmt.Println(" Missing Comment ID")
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	userID, loggedIn := u.ValidateSession(db, r)
	if !loggedIn {
		fmt.Println(" Unauthorized User")
		http.Error(w, "Unauthorized. Please log in.", http.StatusUnauthorized)
		return
	}

	fmt.Println(" User is logged in:", userID)

	// Check if interaction exists
	like, err := c.s.CheckCommentInteractions(r.Context(), userID, *req.CommentID)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println(" Storing new interaction for comment", *req.CommentID)
		if err = c.s.InteractWithComment(r.Context(), userID, *req.CommentID, req.IsLike); err != nil {
			fmt.Println(" Error storing interaction:", err)
			http.Error(w, "Failed to like/dislike comment", http.StatusInternalServerError)
			return
		}
	} else if like.IsLike == req.IsLike {
		fmt.Println(" Removing interaction for comment", *req.CommentID)
		if err = c.s.RemoveCommentInteraction(r.Context(), userID, *req.CommentID); err != nil {
			fmt.Println(" Error removing interaction:", err)
			http.Error(w, "Failed to remove interaction", http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println(" Switching Like to Dislike (or vice versa)")
		if err = c.s.RemoveCommentInteraction(r.Context(), userID, *req.CommentID); err != nil {
			fmt.Println(" Error removing previous interaction:", err)
			http.Error(w, "Failed to remove previous interaction", http.StatusInternalServerError)
			return
		}
		if err = c.s.InteractWithComment(r.Context(), userID, *req.CommentID, req.IsLike); err != nil {
			fmt.Println(" Error updating interaction:", err)
			http.Error(w, "Failed to update interaction", http.StatusInternalServerError)
			return
		}
	}

	fmt.Println(" Interaction updated successfully")

	// Get updated counts
	updatedCounts, err := c.s.GetCommentsInteractions(r.Context(), *req.CommentID)
	if err != nil {
		fmt.Println(" Error fetching updated interactions:", err)
		http.Error(w, "Failed to fetch updated interactions", http.StatusInternalServerError)
		return
	}

	fmt.Printf(" Updated Counts -> Likes: %d, Dislikes: %d\n", updatedCounts.Likes, updatedCounts.Dislikes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Interaction updated successfully",
		"likes":    updatedCounts.Likes,
		"dislikes": updatedCounts.Dislikes,
	})
}

func (c *LikesController) GetInteractions(w http.ResponseWriter, r *http.Request) {
	var req GetInteractionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "err", err)
		return
	}

	var resp GetInteractionsResponse
	var err error
	if req.PostID != nil {
		resp, err = c.s.GetPostsInteractions(r.Context(), *req.PostID)
		if err != nil {

			return
		}
	} else if req.CommentID != nil {
		resp, err = c.s.GetCommentsInteractions(r.Context(), *req.CommentID)
		if err != nil {

			return
		}
	} else {
		return
	}

	//  Ensure zero values are returned if no interactions exist
	if resp.Likes == 0 && resp.Dislikes == 0 {
		resp.Likes = 0
		resp.Dislikes = 0
	}

	response := map[string]int{"likes": resp.Likes, "dislikes": resp.Dislikes}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
