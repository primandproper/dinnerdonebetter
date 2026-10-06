/*
Package waitlists is platform-go's waitlist store, recording through platform's hooks.

The lists and the signups are platform's: the schema, the paging, the tenancy column, the
lifecycle and the withdrawal all live there. So is the recording now — waitlists.RecordingHooks
writes an audit entry and emits an event for every write, on the write's transaction, through the
recording.Recorder this application registers. A signup's entries are filed under the person on
the list, because the Recorder files by subject; a list's are filed where the write ran, because
a list is an administrative row that belongs to nobody.

What this package still decides is the table prefix, and nothing else.
*/
package waitlists

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformrecording "github.com/primandproper/platform-go/v15/recording"
	platformwaitlists "github.com/primandproper/platform-go/v15/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// ProvideWaitlistsRepository provides platform's waitlist store, with platform's recording
// hooks on its writes.
func ProvideWaitlistsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
	recorder *platformrecording.Recorder,
) (platformwaitlists.Store, error) {
	hooks, err := platformwaitlists.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the waitlists recording hooks")
	}

	store, err := platformwaitlists.NewSQLStore(
		client,
		platformwaitlists.WithTablePrefix(branding.TablePrefix),
		platformwaitlists.WithHooks(hooks),
		platformwaitlists.WithStoreLogger(logger),
		platformwaitlists.WithStoreTracerProvider(tracerProvider),
		platformwaitlists.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the waitlists store")
	}

	return store, nil
}
