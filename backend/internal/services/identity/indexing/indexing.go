package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/search/syncsource"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
)

// UserSource reads users as search documents, for both the change feed and a reindex.
func UserSource(repo identity.Repository) *syncsource.Source[identity.User, UserSearchSubset] {
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
	return syncsource.NewSyncer(UserSource(repo), index, logger, tracerProvider, metricsProvider)
}

// NewUserReindexer builds the reindex backstop for the users index.
func NewUserReindexer(
	repo identity.Repository,
	index UserTextSearcher,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*searchsync.Reindexer[UserSearchSubset], error) {
	return syncsource.NewReindexer(UserSource(repo), index, logger, tracerProvider, metricsProvider)
}
