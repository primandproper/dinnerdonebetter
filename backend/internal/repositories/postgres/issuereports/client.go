/*
Package issuereports is platform-go's issue report store, recording through platform's hooks.

The reports and their triage lifecycle are platform's, and so is the recording:
issuereports.RecordingHooks writes an audit entry and emits an event for every write, on the
write's transaction, through the recording.Recorder this application registers. An entry is filed
under the reporter, because the Recorder files by subject, and names the requester as the one who
did it — a triager resolving somebody's report is recorded as the triager. The report's free text
is hashed in the diff, so what somebody typed is never copied into the one table their erasure
cannot reach.

What this package still decides is the table prefix, and nothing else.
*/
package issuereports

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// ProvideIssueReportsRepository provides platform's issue report store, with platform's recording
// hooks on its writes.
func ProvideIssueReportsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
	recorder *platformrecording.Recorder,
) (platformissuereports.Store, error) {
	hooks, err := platformissuereports.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the issue reports recording hooks")
	}

	store, err := platformissuereports.NewSQLStore(
		client,
		platformissuereports.WithTablePrefix(branding.TablePrefix),
		platformissuereports.WithHooks(hooks),
		platformissuereports.WithStoreLogger(logger),
		platformissuereports.WithStoreTracerProvider(tracerProvider),
		platformissuereports.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the issue reports store")
	}

	return store, nil
}
