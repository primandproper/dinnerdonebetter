package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"

	"github.com/samber/do/v2"
)

// RegisterUserSyncer registers the Syncer that applies users-index events.
//
// It is registered for the process that consumes the index topic. Handle is a jobs.Handler, so
// the Pool that runs it is wired where the other consumers are rather than here.
func RegisterUserSyncer(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*searchsync.Syncer[UserSearchSubset], error) {
		return NewUserSyncer(
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[UserTextSearcher](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}

// RegisterUserReindexer registers the users reindex backstop, which runs as a scheduled job.
func RegisterUserReindexer(i do.Injector) {
	do.Provide(i, func(i do.Injector) (*searchsync.Reindexer[UserSearchSubset], error) {
		return NewUserReindexer(
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[UserTextSearcher](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
