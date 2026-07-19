package manager

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditDataManager registers the audit data manager with the injector.
func RegisterAuditDataManager(i do.Injector) {
	do.Provide[AuditDataManager](i, func(i do.Injector) (AuditDataManager, error) {
		return NewAuditDataManager(
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[audit.Repository](i),
		), nil
	})
}
