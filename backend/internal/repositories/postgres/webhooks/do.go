package webhooks

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	domainwebhooks "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWebhooksRepository registers the webhooks repository with the injector.
func RegisterWebhooksRepository(i do.Injector) {
	do.Provide[domainwebhooks.Repository](i, func(i do.Injector) (domainwebhooks.Repository, error) {
		return ProvideWebhooksRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		), nil
	})
}
