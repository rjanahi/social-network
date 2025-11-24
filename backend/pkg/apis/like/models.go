package like

type Like struct {
	UserID    int
	PostID    *int
	CommentID *int
	IsLike    bool
}

type InteractRequest struct {
	PostID    *int `json:"post_id,omitempty"`
	CommentID *int `json:"comment_id,omitempty"`
	IsLike    bool `json:"is_like"`
}

type GetInteractionsRequest struct {
	PostID    *int `json:"post_id,omitempty"`
	CommentID *int `json:"comment_id,omitempty"`
}

type GetInteractionsResponse struct {
	PostID    *int `json:"post_id,omitempty"`
	CommentID *int `json:"comment_id,omitempty"`
	Likes     int  `json:"likes"`
	Dislikes  int  `json:"dislikes"`
}
