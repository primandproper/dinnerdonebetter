/*
Package comments is platform-go's comment store, recording through platform's hooks.

The comments are platform's, and so is the recording: comments.RecordingHooks writes an audit
entry and emits an event for every write, on the write's transaction, through the
recording.Recorder this application registers. An entry is filed under the comment's author,
because the Recorder files by subject, and names the requester as the one who did it — an
administrator archiving somebody's comment is recorded as the administrator. An edit's diff hashes
the body, so what somebody wrote is never copied into the one table their erasure cannot reach.

What this package still decides is the table prefix, and the target catalog it is handed.
*/
package comments

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformcomments "github.com/primandproper/platform-go/v15/comments"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// ProvideCommentsRepository provides platform's comment store, with platform's recording hooks
// on its writes.
//
// The target catalog is a parameter because it is the one thing platform refuses to guess at,
// and a store built without one accepts no writes at all. It is assembled in
// internal/build/comments, where the domains that own the things being commented on are already
// in hand.
func ProvideCommentsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
	recorder *platformrecording.Recorder,
	targets platformcomments.Targets,
) (platformcomments.Store, error) {
	hooks, err := platformcomments.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the comments recording hooks")
	}

	store, err := platformcomments.NewSQLStore(
		client,
		platformcomments.WithTablePrefix(branding.TablePrefix),
		platformcomments.WithTargets(targets),
		platformcomments.WithHooks(hooks),
		platformcomments.WithStoreLogger(logger),
		platformcomments.WithStoreTracerProvider(tracerProvider),
		platformcomments.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the comments store")
	}

	return store, nil
}
