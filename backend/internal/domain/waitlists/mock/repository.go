package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ waitlists.Repository = (*Repository)(nil)

// Repository mocks the waitlists repository.
type Repository struct {
	mock.Mock
}

// WaitlistIsNotExpired is a mock function.
func (m *Repository) WaitlistIsNotExpired(ctx context.Context, waitlistID string) (bool, error) {
	args := m.Called(ctx, waitlistID)
	return args.Bool(0), args.Error(1)
}

// GetWaitlist is a mock function.
func (m *Repository) GetWaitlist(ctx context.Context, waitlistID string) (*waitlists.Waitlist, error) {
	args := m.Called(ctx, waitlistID)
	return args.Get(0).(*waitlists.Waitlist), args.Error(1)
}

// GetWaitlists is a mock function.
func (m *Repository) GetWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[waitlists.Waitlist]), args.Error(1)
}

// GetActiveWaitlists is a mock function.
func (m *Repository) GetActiveWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[waitlists.Waitlist]), args.Error(1)
}

// CreateWaitlist is a mock function.
func (m *Repository) CreateWaitlist(ctx context.Context, input *waitlists.WaitlistDatabaseCreationInput) (*waitlists.Waitlist, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*waitlists.Waitlist), args.Error(1)
}

// UpdateWaitlist is a mock function.
func (m *Repository) UpdateWaitlist(ctx context.Context, waitlist *waitlists.Waitlist) error {
	args := m.Called(ctx, waitlist)
	return args.Error(0)
}

// ArchiveWaitlist is a mock function.
func (m *Repository) ArchiveWaitlist(ctx context.Context, waitlistID string) error {
	args := m.Called(ctx, waitlistID)
	return args.Error(0)
}

// GetWaitlistSignup is a mock function.
func (m *Repository) GetWaitlistSignup(ctx context.Context, waitlistSignupID, waitlistID string) (*waitlists.WaitlistSignup, error) {
	args := m.Called(ctx, waitlistSignupID, waitlistID)
	return args.Get(0).(*waitlists.WaitlistSignup), args.Error(1)
}

// GetWaitlistSignupsForWaitlist is a mock function.
func (m *Repository) GetWaitlistSignupsForWaitlist(ctx context.Context, waitlistID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
	args := m.Called(ctx, waitlistID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[waitlists.WaitlistSignup]), args.Error(1)
}

// GetWaitlistSignupsForUser is a mock function.
func (m *Repository) GetWaitlistSignupsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[waitlists.WaitlistSignup]), args.Error(1)
}

// CreateWaitlistSignup is a mock function.
func (m *Repository) CreateWaitlistSignup(ctx context.Context, input *waitlists.WaitlistSignupDatabaseCreationInput) (*waitlists.WaitlistSignup, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*waitlists.WaitlistSignup), args.Error(1)
}

// UpdateWaitlistSignup is a mock function.
func (m *Repository) UpdateWaitlistSignup(ctx context.Context, waitlistSignup *waitlists.WaitlistSignup) error {
	args := m.Called(ctx, waitlistSignup)
	return args.Error(0)
}

// ArchiveWaitlistSignup is a mock function.
func (m *Repository) ArchiveWaitlistSignup(ctx context.Context, waitlistSignupID string) error {
	args := m.Called(ctx, waitlistSignupID)
	return args.Error(0)
}
