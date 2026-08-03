package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v9/filtering"
)

type (
	// WebhookDataManager describes the interface for webhook business logic: creation,
	// retrieval, archival, subscription management, and secret rotation.
	WebhookDataManager interface {
		CreateWebhook(ctx context.Context, userID, accountID string, input *webhooks.WebhookCreationRequestInput) (*webhooks.WebhookCreationResponse, error)
		GetWebhook(ctx context.Context, webhookID, accountID string) (*webhooks.Webhook, error)
		GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error)
		ArchiveWebhook(ctx context.Context, webhookID, accountID string) error
		AddWebhookTriggerConfig(ctx context.Context, accountID string, input *webhooks.WebhookTriggerConfigCreationRequestInput) (*webhooks.WebhookTriggerConfig, error)
		ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, accountID, configID string) error
		RotateWebhookSecret(ctx context.Context, webhookID, accountID string) (string, error)
		WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error)

		// GetWebhookEventTypes lists the events a webhook may subscribe to.
		//
		// It takes no filter and returns no pagination: the catalog is generated Go with a
		// couple of hundred entries that cannot change between deployments, so paging it
		// would be ceremony over a constant.
		GetWebhookEventTypes(ctx context.Context) []*webhooks.WebhookEventType
	}
)
