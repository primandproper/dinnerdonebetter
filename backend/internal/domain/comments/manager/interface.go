package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"

	"github.com/primandproper/platform-go/v4/filtering"
)

type (
	// CommentsDataManager describes the interface for comment business logic.
	CommentsDataManager interface {
		CreateComment(ctx context.Context, input *comments.CommentCreationRequestInput) (*comments.Comment, error)
		GetComment(ctx context.Context, id string) (*comments.Comment, error)
		GetCommentsForReference(ctx context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error)
		UpdateComment(ctx context.Context, id, belongsToUser string, input *comments.CommentUpdateRequestInput) error
		ArchiveComment(ctx context.Context, id string) error
		ArchiveCommentsForReference(ctx context.Context, targetType, referencedID string) error
	}
)
