package notifications

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v8/database"
	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsRepository registers the notifications repository with the injector.
func RegisterNotificationsRepository(i do.Injector) {
	do.Provide[*Repository](i, func(i do.Injector) (*Repository, error) {
		return ProvideNotificationsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*databasecfg.Config](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		), nil
	})
}
