package issuereports

import (
	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIssueReportsRepository registers the issue reports store with the injector.
func RegisterIssueReportsRepository(i do.Injector) {
	do.Provide[platformissuereports.Store](i, func(i do.Injector) (platformissuereports.Store, error) {
		return ProvideIssueReportsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*platformrecording.Recorder](i),
		)
	})
}
