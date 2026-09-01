package issuereports

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	platformissuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIssueReportsRepository registers the issue reports store with the injector.
func RegisterIssueReportsRepository(i do.Injector) {
	do.Provide[platformissuereports.Store](i, func(i do.Injector) (platformissuereports.Store, error) {
		return ProvideIssueReportsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		)
	})
}
