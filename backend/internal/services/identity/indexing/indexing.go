package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
	syncsource "github.com/primandproper/platform-go/v10/search/sync/source"
)

// o11yName names the loggers, spans and metrics of the search sync sources built here. It
// keeps the name the deleted internal/search/syncsource used, so nothing downstream of a log
// query has to change.
const o11yName = "search_sync_source"

// UserSource reads users as search documents, for both the change feed and a reindex.
func UserSource(repo identity.Repository) (*syncsource.Source[identity.User, UserSearchSubset], error) {
	return syncsource.New(
		IndexTypeUsers,
		repo.GetUser,
		repo.ScanUserIDsForReindex,
		ConvertUserToUserSearchSubset,
	)
}

// NewUserSyncer builds the Syncer that applies one users-index event.
//
// It replaces the scheduler that used to publish an index request for every user a sampler
// thought looked stale. The events now come from the transactions that changed the rows.
func NewUserSyncer(
	repo identity.Repository,
	index UserTextSearcher,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*searchsync.Syncer[UserSearchSubset], error) {
	src, err := UserSource(repo)
	if err != nil {
		return nil, err
	}

	return syncsource.NewSyncer(src, index, o11yOptions(logger, tracerProvider, metricsProvider)...)
}

// NewUserReindexer builds the reindex backstop for the users index.
func NewUserReindexer(
	repo identity.Repository,
	index UserTextSearcher,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*searchsync.Reindexer[UserSearchSubset], error) {
	src, err := UserSource(repo)
	if err != nil {
		return nil, err
	}

	return syncsource.NewReindexer(src, index, o11yOptions(logger, tracerProvider, metricsProvider)...)
}

// o11yOptions is the three pillars as syncsource options. They arrive here separately rather
// than as an observability.Pillars because that is how the container holds them.
func o11yOptions(logger logging.Logger, tracerProvider tracing.Provider, metricsProvider metrics.Provider) []syncsource.Option {
	return []syncsource.Option{
		syncsource.WithLogger(logging.NewNamedLogger(logger, o11yName)),
		syncsource.WithTracerProvider(tracerProvider),
		syncsource.WithMetricsProvider(metricsProvider),
	}
}
