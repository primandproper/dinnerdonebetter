package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/manager"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ manager.OAuth2Manager = (*OAuth2Manager)(nil)

type OAuth2Manager struct {
	mock.Mock
}

// CreateOAuth2Client is a mock function.
func (m *OAuth2Manager) CreateOAuth2Client(ctx context.Context, input *oauth.OAuth2ClientCreationRequestInput) (*oauth.OAuth2Client, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*oauth.OAuth2Client), args.Error(1)
}

// GetOAuth2Client is a mock function.
func (m *OAuth2Manager) GetOAuth2Client(ctx context.Context, oauth2ClientID string) (*oauth.OAuth2Client, error) {
	args := m.Called(ctx, oauth2ClientID)
	return args.Get(0).(*oauth.OAuth2Client), args.Error(1)
}

// GetOAuth2Clients is a mock function.
func (m *OAuth2Manager) GetOAuth2Clients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[oauth.OAuth2Client], error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[oauth.OAuth2Client]), args.Error(1)
}

// ArchiveOAuth2Client is a mock function.
func (m *OAuth2Manager) ArchiveOAuth2Client(ctx context.Context, oauth2ClientID string) error {
	args := m.Called(ctx, oauth2ClientID)
	return args.Error(0)
}
