package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWebhookDataManager registers the webhook data manager with the injector.
func RegisterWebhookDataManager(i do.Injector) {
	do.Provide[WebhookDataManager](i, func(i do.Injector) (WebhookDataManager, error) {
		return NewWebhookDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[webhooks.Repository](i),
		)
	})
}
