package oauthmock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/mock"
)

var _ oauth.Repository = (*RepositoryMock)(nil)

type RepositoryMock struct {
	mock.Mock
}

// GetOAuth2ClientByClientID is a mock function.
func (m *RepositoryMock) GetOAuth2ClientByClientID(ctx context.Context, clientID string) (*oauth.OAuth2Client, error) {
	returnValues := m.Called(ctx, clientID)
	return returnValues.Get(0).(*oauth.OAuth2Client), returnValues.Error(1)
}

// GetOAuth2ClientByDatabaseID is a mock function.
func (m *RepositoryMock) GetOAuth2ClientByDatabaseID(ctx context.Context, clientID string) (*oauth.OAuth2Client, error) {
	returnValues := m.Called(ctx, clientID)
	return returnValues.Get(0).(*oauth.OAuth2Client), returnValues.Error(1)
}

// GetOAuth2Clients is a mock function.
func (m *RepositoryMock) GetOAuth2Clients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[oauth.OAuth2Client], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[oauth.OAuth2Client]), returnValues.Error(1)
}

// CreateOAuth2Client is a mock function.
func (m *RepositoryMock) CreateOAuth2Client(ctx context.Context, input *oauth.OAuth2ClientDatabaseCreationInput) (*oauth.OAuth2Client, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*oauth.OAuth2Client), returnValues.Error(1)
}

// ArchiveOAuth2Client is a mock function.
func (m *RepositoryMock) ArchiveOAuth2Client(ctx context.Context, clientID string) error {
	return m.Called(ctx, clientID).Error(0)
}

func (m *RepositoryMock) CreateOAuth2ClientToken(ctx context.Context, input *oauth.OAuth2ClientTokenDatabaseCreationInput) (*oauth.OAuth2ClientToken, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*oauth.OAuth2ClientToken), returnValues.Error(1)
}

func (m *RepositoryMock) GetOAuth2ClientTokenByCode(ctx context.Context, code string) (*oauth.OAuth2ClientToken, error) {
	returnValues := m.Called(ctx, code)
	return returnValues.Get(0).(*oauth.OAuth2ClientToken), returnValues.Error(1)
}

func (m *RepositoryMock) GetOAuth2ClientTokenByAccess(ctx context.Context, access string) (*oauth.OAuth2ClientToken, error) {
	returnValues := m.Called(ctx, access)
	return returnValues.Get(0).(*oauth.OAuth2ClientToken), returnValues.Error(1)
}

func (m *RepositoryMock) GetOAuth2ClientTokenByRefresh(ctx context.Context, refresh string) (*oauth.OAuth2ClientToken, error) {
	returnValues := m.Called(ctx, refresh)
	return returnValues.Get(0).(*oauth.OAuth2ClientToken), returnValues.Error(1)
}

func (m *RepositoryMock) DeleteOAuth2ClientTokenByAccess(ctx context.Context, access string) error {
	return m.Called(ctx, access).Error(0)
}

func (m *RepositoryMock) DeleteOAuth2ClientTokenByCode(ctx context.Context, code string) error {
	return m.Called(ctx, code).Error(0)
}

func (m *RepositoryMock) DeleteOAuth2ClientTokenByRefresh(ctx context.Context, refresh string) error {
	return m.Called(ctx, refresh).Error(0)
}

func NewRepositoryMock() *RepositoryMock {
	return &RepositoryMock{}
}
