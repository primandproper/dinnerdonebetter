package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ uploadedmedia.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// GetUploadedMedia is a mock function.
func (m *Repository) GetUploadedMedia(ctx context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
	args := m.Called(ctx, uploadedMediaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uploadedmedia.UploadedMedia), args.Error(1)
}

// GetUploadedMediaWithIDs is a mock function.
func (m *Repository) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*uploadedmedia.UploadedMedia, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*uploadedmedia.UploadedMedia), args.Error(1)
}

// GetUploadedMediaForUser is a mock function.
func (m *Repository) GetUploadedMediaForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia]), args.Error(1)
}

// CreateUploadedMedia is a mock function.
func (m *Repository) CreateUploadedMedia(ctx context.Context, input *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uploadedmedia.UploadedMedia), args.Error(1)
}

// UpdateUploadedMedia is a mock function.
func (m *Repository) UpdateUploadedMedia(ctx context.Context, uploadedMedia *uploadedmedia.UploadedMedia) error {
	args := m.Called(ctx, uploadedMedia)
	return args.Error(0)
}

// ArchiveUploadedMedia is a mock function.
func (m *Repository) ArchiveUploadedMedia(ctx context.Context, uploadedMediaID string) error {
	args := m.Called(ctx, uploadedMediaID)
	return args.Error(0)
}
