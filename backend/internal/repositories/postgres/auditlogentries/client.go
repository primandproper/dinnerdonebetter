package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v13/audit"
	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "audit_log_entries_db_client"
)

// repository is the audit log entry repository implementation.
//
// It holds no database handle. Writes take the caller's executor, so an entry
// commits with the change it describes or not at all; reads go through the
// platform Reader, which owns its own. The schema belongs to the platform as
// well — the uniqueness constraint that makes a forked chain unrepresentable is
// the guarantee rather than an incidental storage detail — so there is no
// generated querier here and no SQL in this package.
type repository struct {
	tracer   tracing.Tracer
	logger   logging.Logger
	recorder platformaudit.Recorder
	reader   platformaudit.Reader
}

// ProvideAuditLogRepository provides a new repository.
//
// The Recorder and Reader are built here rather than injected, so that the
// redaction policy and the table prefix are applied to every audit log in the
// process by construction. A Recorder assembled somewhere else could be assembled
// without them, and the failure would be silent in the direction that matters:
// entries written to a table nothing reads, or secrets written to a table nothing
// can edit.
//
// Pass the metrics provider. audit_chain_breaks is the instrument to alert on —
// everything else this package emits describes throughput, but a non-zero break
// count means the log has stopped being evidence.
func ProvideAuditLogRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
) (audit.Repository, error) {
	if client == nil {
		return nil, platformaudit.ErrNilDatabaseClient
	}

	recorderOptions := []platformaudit.RecorderOption{
		platformaudit.WithRecorderTablePrefix(audit.TablePrefix),
		platformaudit.WithRecorderLogger(logging.EnsureLogger(logger)),
		platformaudit.WithRecorderTracerProvider(tracerProvider),
		platformaudit.WithRecorderMetricsProvider(metricsProvider),
	}
	for resourceType := range audit.Redactions {
		recorderOptions = append(recorderOptions, platformaudit.WithRedaction(resourceType, audit.Redactions[resourceType]))
	}

	recorder, err := platformaudit.NewRecorder(client.Dialect(), recorderOptions...)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building audit recorder")
	}

	reader, err := platformaudit.NewReader(
		client,
		platformaudit.WithReaderTablePrefix(audit.TablePrefix),
		platformaudit.WithReaderLogger(logging.EnsureLogger(logger)),
		platformaudit.WithReaderTracerProvider(tracerProvider),
		platformaudit.WithReaderMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building audit reader")
	}

	return &repository{
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recorder,
		reader:   reader,
	}, nil
}
