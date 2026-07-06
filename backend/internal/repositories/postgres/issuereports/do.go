package issue_reports

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIssueReportsRepository registers the issue reports repository with the injector.
func RegisterIssueReportsRepository(i do.Injector) {
	do.Provide[issuereports.Repository](i, func(i do.Injector) (issuereports.Repository, error) {
		return ProvideIssueReportsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		), nil
	})
}
