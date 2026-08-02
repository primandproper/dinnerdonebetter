package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries/generated"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "audit_log_entries_db_client"
)

// repository is the audit log entry repository implementation.
type repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	readDB           database.SQLQueryExecutor
	writeDB          database.SQLQueryExecutor
}

// ProvideAuditLogRepository provides a new repository.
func ProvideAuditLogRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	client database.Client,
) audit.Repository {
	c := &repository{
		Client:           client,
		readDB:           client.Reader(),
		writeDB:          client.Writer(),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier: generated.New(),
		logger:           logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
