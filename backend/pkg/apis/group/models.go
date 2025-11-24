package group

import (
	"time"
)

type Group struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorID   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	Creator     string    `json:"creator,omitempty"`
	MemberCount int       `json:"member_count,omitempty"`
}

type GroupMember struct {
	ID       int       `json:"id"`
	GroupID  int       `json:"group_id"`
	UserID   int       `json:"user_id"`
	Status   string    `json:"status"`
	IsAdmin  bool      `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
	Username string    `json:"username,omitempty"`
}

type GroupInvitation struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	InviterID int       `json:"inviter_id"`
	InviteeID int       `json:"invitee_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	GroupName string    `json:"group_name,omitempty"`
	Inviter   string    `json:"inviter,omitempty"`
}

type CreateGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type InviteUserRequest struct {
	GroupID int `json:"group_id"`
	UserID  int `json:"user_id"`
}

type ResponseInvitationRequest struct {
	InvitationID int    `json:"invitation_id"`
	Status       string `json:"status"` // "accepted" or "declined"
}

type JoinGroupRequest struct {
	GroupID int `json:"group_id"`
}

type BulkInviteRequest struct {
	GroupID int   `json:"group_id"`
	UserIDs []int `json:"user_ids"`
}

type CancelInvitationRequest struct {
	InvitationID int `json:"invitation_id"`
}

type GroupPost struct {
	ID           int      `json:"id"`
	GroupID      int      `json:"group_id"`
	UserID       int      `json:"user_id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	CreatedAt    string   `json:"created_at"`
	Username     string   `json:"username"`
	Categories   []string `json:"categories"`
	LikeCount    int      `json:"like_count"`
	DislikeCount int      `json:"dislike_count"` // keep only once
	CommentCount int      `json:"comment_count"`
	IsLiked      bool     `json:"is_liked"`
	IsDisliked   bool     `json:"is_disliked"` // keep only once
}
type CreateGroupPostRequest struct {
	GroupID    int      `json:"group_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"` // names
}

type CreateGroupCommentRequest struct {
	GroupID     int    `json:"group_id"`
	GroupPostID int    `json:"group_post_id"`
	Content     string `json:"content"`
}
