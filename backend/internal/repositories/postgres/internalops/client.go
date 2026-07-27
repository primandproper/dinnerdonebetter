package internalops

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/internalops/generated"

	"github.com/primandproper/platform-go/v7/database"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

const (
	o11yName = "internalops_db_client"
)

type repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	readDB           database.SQLQueryExecutor
	writeDB          database.SQLQueryExecutor
}

// ProvideInternalOpsRepository provides a new repository.
func ProvideInternalOpsRepository(logger logging.Logger, tracerProvider tracing.TracerProvider, client database.Client) internalops.InternalOpsDataManager {
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
