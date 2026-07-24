package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	commentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/mock"
)

var _ commentsmanager.CommentsDataManager = (*MockCommentsDataManager)(nil)

type MockCommentsDataManager struct {
	mock.Mock
}

func (m *MockCommentsDataManager) CreateComment(ctx context.Context, input *comments.CommentCreationRequestInput) (*comments.Comment, error) {
	returnValues := m.Called(ctx, input)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*comments.Comment), returnValues.Error(1)
}

func (m *MockCommentsDataManager) GetComment(ctx context.Context, id string) (*comments.Comment, error) {
	returnValues := m.Called(ctx, id)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*comments.Comment), returnValues.Error(1)
}

func (m *MockCommentsDataManager) GetCommentsForReference(ctx context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
	returnValues := m.Called(ctx, targetType, referencedID, filter)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[comments.Comment]), returnValues.Error(1)
}

func (m *MockCommentsDataManager) UpdateComment(ctx context.Context, id, belongsToUser string, input *comments.CommentUpdateRequestInput) error {
	returnValues := m.Called(ctx, id, belongsToUser, input)
	return returnValues.Error(0)
}

func (m *MockCommentsDataManager) ArchiveComment(ctx context.Context, id string) error {
	returnValues := m.Called(ctx, id)
	return returnValues.Error(0)
}

func (m *MockCommentsDataManager) ArchiveCommentsForReference(ctx context.Context, targetType, referencedID string) error {
	returnValues := m.Called(ctx, targetType, referencedID)
	return returnValues.Error(0)
}
