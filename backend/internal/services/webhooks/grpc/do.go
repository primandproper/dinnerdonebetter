package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/manager"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWebhooksService registers the webhooks gRPC service with the injector.
func RegisterWebhooksService(i do.Injector) {
	do.Provide[WebhooksMethodPermissions](i, func(i do.Injector) (WebhooksMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[webhookssvc.WebhooksServiceServer](i, func(i do.Injector) (webhookssvc.WebhooksServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[manager.WebhookDataManager](i),
		), nil
	})
}
