package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "audit_log_entries_db_client"
)

// repository is the audit log entry repository implementation.
//
// Writes take the caller's executor, so an entry commits with the change it
// describes or not at all. Reads take this repository's own read executor: as
// of v14 the platform Reader holds no handle either, so the client is held
// here to supply one. The schema belongs to the platform as
// well — the uniqueness constraint that makes a forked chain unrepresentable is
// the guarantee rather than an incidental storage detail — so there is no
// generated querier here and no SQL in this package.
type repository struct {
	tracer   tracing.Tracer
	logger   logging.Logger
	db       database.Client
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
		client.Dialect(),
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
		db:       client,
		recorder: recorder,
		reader:   reader,
	}, nil
}

// PlatformReader is the reader this repository built, for the surfaces that read
// the log directly.
//
// It is exposed rather than rebuilt because rebuilding is the failure this
// package's construction exists to prevent: a Reader assembled elsewhere could
// be assembled with a different table prefix, and would then answer "no entries"
// about a log that is full. platform's audit/grpc takes an audit.Reader, so this
// is how it gets the one whose prefix matches the Recorder's.
func (q *repository) PlatformReader() platformaudit.Reader { return q.reader }

// ReaderFrom answers with the platform reader behind an audit.Repository.
//
// The assertion cannot fail for a repository this package built, and a
// repository it did not build is a caller who has substituted the audit log —
// in which case there is no platform reader and the surfaces that need one
// should not be mounted.
func ReaderFrom(repo audit.Repository) (platformaudit.Reader, bool) {
	exposer, ok := repo.(interface {
		PlatformReader() platformaudit.Reader
	})
	if !ok {
		return nil, false
	}

	return exposer.PlatformReader(), true
}
