package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIssueReportsDataManager registers the issue reports data manager with the injector.
func RegisterIssueReportsDataManager(i do.Injector) {
	do.Provide[IssueReportsDataManager](i, func(i do.Injector) (IssueReportsDataManager, error) {
		return NewIssueReportsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[issuereports.Repository](i),
		)
	})
}
