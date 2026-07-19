package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/manager"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ manager.WebhookDataManager = (*WebhookDataManager)(nil)

// WebhookDataManager is a mock type for the WebhookDataManager interface.
type WebhookDataManager struct {
	mock.Mock
}

func (m *WebhookDataManager) WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error) {
	args := m.Called(ctx, webhookID, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *WebhookDataManager) CreateWebhook(ctx context.Context, userID, accountID string, input *webhooks.WebhookCreationRequestInput) (*webhooks.Webhook, error) {
	args := m.Called(ctx, userID, accountID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.Webhook), args.Error(1)
}

func (m *WebhookDataManager) GetWebhook(ctx context.Context, webhookID, accountID string) (*webhooks.Webhook, error) {
	args := m.Called(ctx, webhookID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.Webhook), args.Error(1)
}

func (m *WebhookDataManager) GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[webhooks.Webhook]), args.Error(1)
}

func (m *WebhookDataManager) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	args := m.Called(ctx, webhookID, accountID)
	return args.Error(0)
}

func (m *WebhookDataManager) AddWebhookTriggerConfig(ctx context.Context, accountID string, input *webhooks.WebhookTriggerConfigCreationRequestInput) (*webhooks.WebhookTriggerConfig, error) {
	args := m.Called(ctx, accountID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerConfig), args.Error(1)
}

func (m *WebhookDataManager) ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, configID string) error {
	args := m.Called(ctx, webhookID, configID)
	return args.Error(0)
}

func (m *WebhookDataManager) CreateWebhookTriggerEvent(ctx context.Context, input *webhooks.WebhookTriggerEventCreationRequestInput) (*webhooks.WebhookTriggerEvent, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerEvent), args.Error(1)
}

func (m *WebhookDataManager) GetWebhookTriggerEvent(ctx context.Context, id string) (*webhooks.WebhookTriggerEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerEvent), args.Error(1)
}

func (m *WebhookDataManager) GetWebhookTriggerEvents(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.WebhookTriggerEvent], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[webhooks.WebhookTriggerEvent]), args.Error(1)
}

func (m *WebhookDataManager) UpdateWebhookTriggerEvent(ctx context.Context, id string, input *webhooks.WebhookTriggerEventUpdateRequestInput) error {
	args := m.Called(ctx, id, input)
	return args.Error(0)
}

func (m *WebhookDataManager) ArchiveWebhookTriggerEvent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
