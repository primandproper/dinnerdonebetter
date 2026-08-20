package internalops

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops/generated"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const (
	o11yName = "internalops_db_client"
)

type repository struct {
	database.Client
	tracer           tracing.Tracer
	generatedQuerier generated.Querier
	readDB           database.SQLQueryExecutor
	writeDB          database.SQLQueryExecutor
}

// ProvideInternalOpsRepository provides a new repository.
func ProvideInternalOpsRepository(logger logging.Logger, tracerProvider tracing.Provider, client database.Client) internalops.InternalOpsDataManager {
	c := &repository{
		Client:           client,
		readDB:           client.Reader(),
		writeDB:          client.Writer(),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier: generated.New(),
	}

	return c
}
