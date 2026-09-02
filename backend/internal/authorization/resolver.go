package authorization

import (
	platformauthz "github.com/primandproper/platform-go/v13/authorization"
	authzdatabase "github.com/primandproper/platform-go/v13/authorization/database"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/dialect"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

// NewDatabaseResolver builds the resolver that answers what a role grants, over the
// policy tables the migrator seeds.
//
// It is a constructor rather than a configuration block because nothing here is a
// choice. The prefix has to be the one the migration rendered or the resolver reads
// tables that do not exist; the provider has to be the database or it answers from a
// policy nothing seeded; and the dialect is Postgres because the migrations are. A
// config field that must not be changed is worse than a constant — the same reasoning
// that wires uploads/registry directly.
//
// No cache is wrapped around it. Resolution is one statement per session build, against
// five roles, and platform's cached decorator wants a shared store to be worth having:
// this deployment provisions no Redis, and a per-process cache would make a policy
// change visible to one replica at a time. That is a fine trade for a table only a
// migration writes, and a bad one to make silently, so it waits for somewhere shared to
// put it. See #1386.
func NewDatabaseResolver(
	db database.SQLQueryExecutor,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (platformauthz.PolicyResolver, error) {
	opts := []authzdatabase.Option{}
	if logger != nil {
		opts = append(opts, authzdatabase.WithLogger(logger))
	}
	if tracerProvider != nil {
		opts = append(opts, authzdatabase.WithTracerProvider(tracerProvider))
	}
	if metricsProvider != nil {
		opts = append(opts, authzdatabase.WithMetricsProvider(metricsProvider))
	}

	// Returned through a variable, and only once err is known to be nil: returning the
	// constructor's result straight through would hand back a non-nil interface
	// wrapping a nil pointer whenever construction failed.
	resolver, err := authzdatabase.NewResolver(
		&authzdatabase.Config{Dialect: dialect.Postgres, TablePrefix: TablePrefix},
		db,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}
