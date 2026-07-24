package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/mock"
)

var _ dataprivacy.Repository = (*Repository)(nil)

// Repository is a mock repository for dataprivacy.
type Repository struct {
	mock.Mock
}

// FetchUserDataCollection is a mock function.
func (m *Repository) FetchUserDataCollection(ctx context.Context, userID string) (*dataprivacy.UserDataCollection, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*dataprivacy.UserDataCollection), args.Error(1)
}

// DeleteUser is a mock function.
func (m *Repository) DeleteUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// CreateUserDataDisclosure is a mock function.
func (m *Repository) CreateUserDataDisclosure(ctx context.Context, input *dataprivacy.UserDataDisclosureCreationInput) (*dataprivacy.UserDataDisclosure, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dataprivacy.UserDataDisclosure), args.Error(1)
}

// GetUserDataDisclosure is a mock function.
func (m *Repository) GetUserDataDisclosure(ctx context.Context, disclosureID string) (*dataprivacy.UserDataDisclosure, error) {
	args := m.Called(ctx, disclosureID)
	return args.Get(0).(*dataprivacy.UserDataDisclosure), args.Error(1)
}

// GetUserDataDisclosuresForUser is a mock function.
func (m *Repository) GetUserDataDisclosuresForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[dataprivacy.UserDataDisclosure]), args.Error(1)
}

// MarkUserDataDisclosureCompleted is a mock function.
func (m *Repository) MarkUserDataDisclosureCompleted(ctx context.Context, disclosureID, reportID string) error {
	return m.Called(ctx, disclosureID, reportID).Error(0)
}

// MarkUserDataDisclosureFailed is a mock function.
func (m *Repository) MarkUserDataDisclosureFailed(ctx context.Context, disclosureID string) error {
	return m.Called(ctx, disclosureID).Error(0)
}

// ArchiveUserDataDisclosure is a mock function.
func (m *Repository) ArchiveUserDataDisclosure(ctx context.Context, disclosureID string) error {
	return m.Called(ctx, disclosureID).Error(0)
}
