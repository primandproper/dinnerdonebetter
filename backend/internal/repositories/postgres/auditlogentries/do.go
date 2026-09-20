package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	platformaudit "github.com/primandproper/platform-go/v14/audit"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditLogRepository registers the audit log repository with the injector.
func RegisterAuditLogRepository(i do.Injector) {
	do.Provide[audit.Repository](i, func(i do.Injector) (audit.Repository, error) {
		return ProvideAuditLogRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
	})
}

// RegisterPlatformReader registers the reader platform's audit surface takes.
//
// It resolves the repository rather than building a second reader, so the
// surface reads the table the recorder writes. See ReaderFrom.
func RegisterPlatformReader(i do.Injector) {
	do.Provide[platformaudit.Reader](i, func(i do.Injector) (platformaudit.Reader, error) {
		reader, ok := ReaderFrom(do.MustInvoke[audit.Repository](i))
		if !ok {
			return nil, platformerrors.New("audit repository exposes no platform reader")
		}

		return reader, nil
	})
}
