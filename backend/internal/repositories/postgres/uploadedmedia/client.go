/*
Package uploadedmedia is platform-go's upload registry, assembled for this
application. The registry itself is platform's: the schema, the paging, the
tenancy column, the key uniqueness and the ownership the reads answer from all
live there, and so does the recording — mediaregistry.RecordingHooks writes the
audit entry and the event every registry write owes, on the write's transaction,
through the recording.Recorder this process built.

What this package still decides is the table prefix, which has to match the
prefix the migration was rendered with (see internal/repositories/postgres/migrations),
and the observability names.

# There is no update

The registry has no statement that assigns a column after the insert, and the
absence is deliberate rather than missing: every column is a fact about bytes
that are already in a bucket, so an "update" that moved a row's key or content
type would be a row that had stopped describing its object. Changing what an
uploaded object is means storing new bytes and registering them.
*/
package uploadedmedia

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "uploaded_media_db_client"
)

// ProvideUploadedMediaRepository provides platform's upload registry, recording
// every write through recorder.
func ProvideUploadedMediaRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	recorder *platformrecording.Recorder,
	client database.Client,
) (mediaregistry.Store, error) {
	hooks, err := mediaregistry.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the upload registry's recording hooks")
	}

	store, err := mediaregistry.NewSQLStore(
		client,
		mediaregistry.WithTablePrefix(branding.TablePrefix),
		mediaregistry.WithHooks(hooks),
		mediaregistry.WithStoreLogger(logging.NewNamedLogger(logger, o11yName)),
		mediaregistry.WithStoreTracerProvider(tracerProvider),
		mediaregistry.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the upload registry store")
	}

	return store, nil
}
