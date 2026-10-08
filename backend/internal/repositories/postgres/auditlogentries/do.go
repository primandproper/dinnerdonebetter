package auditlogentries

import (
	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditLog registers the audit log with the injector.
func RegisterAuditLog(i do.Injector) {
	do.Provide[*Log](i, func(i do.Injector) (*Log, error) {
		return ProvideAuditLog(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
	})
}

// RegisterPlatformReader registers the reader platform's audit surface takes.
//
// It resolves the log rather than building a second reader, so the surface
// reads the table the recorder writes. See Log.Reader.
func RegisterPlatformReader(i do.Injector) {
	do.Provide[platformaudit.Reader](i, func(i do.Injector) (platformaudit.Reader, error) {
		return do.MustInvoke[*Log](i).Reader(), nil
	})
}

// RegisterPlatformRecorder registers the recorder platform's surfaces record
// into. Like RegisterPlatformReader, it resolves the log's own, so the
// redactions and the table prefix come with it. See Log.Recorder.
func RegisterPlatformRecorder(i do.Injector) {
	do.Provide[platformaudit.Recorder](i, func(i do.Injector) (platformaudit.Recorder, error) {
		return do.MustInvoke[*Log](i).Recorder(), nil
	})
}
