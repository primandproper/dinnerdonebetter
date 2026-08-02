package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v9/filtering"
)

type (
	// WebhookDataManager describes the interface for webhook business logic (creation, retrieval, archival, trigger config and catalog management).
	WebhookDataManager interface {
		CreateWebhook(ctx context.Context, userID, accountID string, input *webhooks.WebhookCreationRequestInput) (*webhooks.Webhook, error)
		GetWebhook(ctx context.Context, webhookID, accountID string) (*webhooks.Webhook, error)
		GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error)
		ArchiveWebhook(ctx context.Context, webhookID, accountID string) error
		AddWebhookTriggerConfig(ctx context.Context, accountID string, input *webhooks.WebhookTriggerConfigCreationRequestInput) (*webhooks.WebhookTriggerConfig, error)
		ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, configID string) error
		CreateWebhookTriggerEvent(ctx context.Context, input *webhooks.WebhookTriggerEventCreationRequestInput) (*webhooks.WebhookTriggerEvent, error)
		GetWebhookTriggerEvent(ctx context.Context, id string) (*webhooks.WebhookTriggerEvent, error)
		GetWebhookTriggerEvents(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.WebhookTriggerEvent], error)
		UpdateWebhookTriggerEvent(ctx context.Context, id string, input *webhooks.WebhookTriggerEventUpdateRequestInput) error
		ArchiveWebhookTriggerEvent(ctx context.Context, id string) error
		WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error)
	}
)
