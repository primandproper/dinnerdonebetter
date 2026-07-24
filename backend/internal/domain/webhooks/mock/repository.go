package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/mock"
)

var _ webhooks.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// WebhookExists is a mock function.
func (m *Repository) WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error) {
	args := m.Called(ctx, webhookID, accountID)
	return args.Bool(0), args.Error(1)
}

// GetWebhook is a mock function.
func (m *Repository) GetWebhook(ctx context.Context, webhookID, accountID string) (*webhooks.Webhook, error) {
	args := m.Called(ctx, webhookID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.Webhook), args.Error(1)
}

// GetWebhooks is a mock function.
func (m *Repository) GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[webhooks.Webhook]), args.Error(1)
}

// GetWebhooksForAccountAndEvent is a mock function.
func (m *Repository) GetWebhooksForAccountAndEvent(ctx context.Context, accountID, eventType string) ([]*webhooks.Webhook, error) {
	args := m.Called(ctx, accountID, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*webhooks.Webhook), args.Error(1)
}

// CreateWebhook is a mock function.
func (m *Repository) CreateWebhook(ctx context.Context, input *webhooks.WebhookDatabaseCreationInput) (*webhooks.Webhook, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.Webhook), args.Error(1)
}

// ArchiveWebhook is a mock function.
func (m *Repository) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	args := m.Called(ctx, webhookID, accountID)
	return args.Error(0)
}

// AddWebhookTriggerConfig is a mock function.
func (m *Repository) AddWebhookTriggerConfig(ctx context.Context, accountID string, input *webhooks.WebhookTriggerConfigDatabaseCreationInput) (*webhooks.WebhookTriggerConfig, error) {
	args := m.Called(ctx, accountID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerConfig), args.Error(1)
}

// ArchiveWebhookTriggerConfig is a mock function.
func (m *Repository) ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, configID string) error {
	args := m.Called(ctx, webhookID, configID)
	return args.Error(0)
}

// CreateWebhookTriggerEvent is a mock function (catalog).
func (m *Repository) CreateWebhookTriggerEvent(ctx context.Context, input *webhooks.WebhookTriggerEventDatabaseCreationInput) (*webhooks.WebhookTriggerEvent, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerEvent), args.Error(1)
}

// GetWebhookTriggerEvent is a mock function (catalog).
func (m *Repository) GetWebhookTriggerEvent(ctx context.Context, id string) (*webhooks.WebhookTriggerEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerEvent), args.Error(1)
}

// GetWebhookTriggerEvents is a mock function (catalog).
func (m *Repository) GetWebhookTriggerEvents(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.WebhookTriggerEvent], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[webhooks.WebhookTriggerEvent]), args.Error(1)
}

// UpdateWebhookTriggerEvent is a mock function (catalog).
func (m *Repository) UpdateWebhookTriggerEvent(ctx context.Context, id string, input *webhooks.WebhookTriggerEventUpdateRequestInput) error {
	args := m.Called(ctx, id, input)
	return args.Error(0)
}

// ArchiveWebhookTriggerEvent is a mock function (catalog).
func (m *Repository) ArchiveWebhookTriggerEvent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
