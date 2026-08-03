package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "audit_log_entries_db_client"
)

// repository is the audit log entry repository implementation.
//
// It owns no SQL of its own. The audit tables belong to the platform — the
// unique index that makes a forked chain unrepresentable and the chain row that
// serializes concurrent writers are the guarantee rather than incidental storage
// details — so this type is the adapter between the platform's Entry and the
// AuditLogEntry shape our API has always spoken, and nothing more.
type repository struct {
	database.Client
	tracer   tracing.Tracer
	logger   logging.Logger
	reader   platformaudit.Reader
	recorder platformaudit.Recorder
}

// ProvideAuditLogRepository provides a new repository.
//
// metricsProvider is worth passing rather than leaving nil: audit_chain_breaks
// is the one instrument here whose non-zero value is an incident on its own,
// because it means the log has stopped being evidence.
func ProvideAuditLogRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	client database.Client,
) (audit.Repository, error) {
	logger = logging.NewNamedLogger(logger, o11yName)

	// The dialect comes from the client rather than from configuration, so the
	// Reader and the Recorder cannot be built for a dialect the database is not.
	reader, err := platformaudit.NewReader(
		client,
		platformaudit.WithReaderTablePrefix(audit.TablePrefix),
		platformaudit.WithReaderLogger(logger),
		platformaudit.WithReaderTracerProvider(tracerProvider),
		platformaudit.WithReaderMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, err
	}

	recorderOptions := []platformaudit.RecorderOption{
		platformaudit.WithRecorderTablePrefix(audit.TablePrefix),
		platformaudit.WithRecorderLogger(logger),
		platformaudit.WithRecorderTracerProvider(tracerProvider),
		platformaudit.WithRecorderMetricsProvider(metricsProvider),
	}
	for resourceType := range audit.Redactions {
		recorderOptions = append(recorderOptions, platformaudit.WithRedaction(resourceType, audit.Redactions[resourceType]))
	}

	recorder, err := platformaudit.NewRecorder(client.Dialect(), recorderOptions...)
	if err != nil {
		return nil, err
	}

	return &repository{
		Client:   client,
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:   logger,
		reader:   reader,
		recorder: recorder,
	}, nil
}
