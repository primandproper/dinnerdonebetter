package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditLogRepository registers the audit log repository with the injector.
func RegisterAuditLogRepository(i do.Injector) {
	do.Provide[audit.Repository](i, func(i do.Injector) (audit.Repository, error) {
		return ProvideAuditLogRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
