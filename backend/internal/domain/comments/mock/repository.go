package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"

	"github.com/primandproper/platform-go/v4/filtering"

	"github.com/stretchr/testify/mock"
)

var _ comments.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

func (m *Repository) CreateComment(ctx context.Context, input *comments.CommentDatabaseCreationInput) (*comments.Comment, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*comments.Comment), args.Error(1)
}

func (m *Repository) GetComment(ctx context.Context, id string) (*comments.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*comments.Comment), args.Error(1)
}

func (m *Repository) GetCommentsForReference(ctx context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
	args := m.Called(ctx, targetType, referencedID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[comments.Comment]), args.Error(1)
}

func (m *Repository) UpdateComment(ctx context.Context, id, belongsToUser, content string) error {
	return m.Called(ctx, id, belongsToUser, content).Error(0)
}

func (m *Repository) ArchiveComment(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *Repository) ArchiveCommentsForReference(ctx context.Context, targetType, referencedID string) error {
	return m.Called(ctx, targetType, referencedID).Error(0)
}
