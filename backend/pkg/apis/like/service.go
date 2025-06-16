package like

import "context"

type LikesService struct {
	repo LikesRepository
}

type LikesRepository interface {
	StoreInteraction(ctx context.Context, interaction Like) error
	RemovePostInteraction(ctx context.Context, userID, postID int) error
	RemoveCommentInteraction(ctx context.Context, userID, commentID int) error
	CheckPostInteractions(ctx context.Context, userID, postID int) (Like, error)
	CheckCommentInteractions(ctx context.Context, userID, commentID int) (Like, error)
	GetUserIDFromSession(ctx context.Context, token string) (int, error)
	GetPostsInteractions(ctx context.Context, postID int) (GetInteractionsResponse, error)
	GetCommentsInteractions(ctx context.Context, commentID int) (GetInteractionsResponse, error)
}

func NewLikesService(repo LikesRepository) *LikesService {
	return &LikesService{repo}
}

func (s *LikesService) InteractWithPost(ctx context.Context, userID, postID int, isLike bool) error {
	interaction := Like{
		UserID: userID,
		PostID: &postID,
		IsLike: isLike,
	}
	return s.repo.StoreInteraction(ctx, interaction)
}

func (s *LikesService) RemovePostInteraction(ctx context.Context, userID, postID int) error {
	return s.repo.RemovePostInteraction(ctx, userID, postID)
}

func (s *LikesService) InteractWithComment(ctx context.Context, userID, commentID int, isLike bool) error {
	
	interaction := Like{
		UserID:    userID,
		CommentID: &commentID,
		IsLike:    isLike,
	}
	return s.repo.StoreInteraction(ctx, interaction)
}

func (s *LikesService) RemoveCommentInteraction(ctx context.Context, userID, commentID int) error {
	return s.repo.RemoveCommentInteraction(ctx, userID, commentID)
}

func (s *LikesService) CheckPostInteractions(ctx context.Context, userID, postID int) (Like, error) {
	return s.repo.CheckPostInteractions(ctx, userID, postID)
}

func (s *LikesService) CheckCommentInteractions(ctx context.Context, userID, commentID int) (Like, error) {
	return s.repo.CheckCommentInteractions(ctx, userID, commentID)
}

func (s *LikesService) GetUserIDFromSession(ctx context.Context, token string) (int, error) {
	return s.repo.GetUserIDFromSession(ctx, token)
}

func (s *LikesService) GetPostsInteractions(ctx context.Context, postID int) (GetInteractionsResponse, error) {
	return s.repo.GetPostsInteractions(ctx, postID)
}

func (s *LikesService) GetCommentsInteractions(ctx context.Context, commentID int) (GetInteractionsResponse, error) {
	return s.repo.GetCommentsInteractions(ctx, commentID)
}
