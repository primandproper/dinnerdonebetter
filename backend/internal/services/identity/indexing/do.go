package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexstamp"

	"github.com/primandproper/platform-go/v11/batching"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v11/search/sync"

	"github.com/samber/do/v2"
)

// RegisterUserSyncer registers the Syncer that applies users-index events, and the Stamper it
// writes through.
//
// It is registered for the process that consumes the index topic. Handle is a jobs.Handler, so
// the Pool that runs it is wired where the other consumers are rather than here.
func RegisterUserSyncer(i do.Injector) {
	// Named for its index, so that the consumer can resolve the same Stamper the Syncer got in
	// order to close it. do provides one instance per name, which is what makes those the same
	// object rather than two buffers over one index.
	do.ProvideNamed(i, IndexTypeUsers, func(i do.Injector) (*indexstamp.Stamper, error) {
		return indexstamp.New(
			do.MustInvoke[UserTextSearcher](i),
			do.MustInvoke[identity.Repository](i).MarkUserAsIndexed,
			batching.WithLogger(logging.NewNamedLogger(do.MustInvoke[logging.Logger](i), o11yName)),
			batching.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			batching.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide(i, func(i do.Injector) (*searchsync.Syncer[UserSearchSubset], error) {
		return NewUserSyncer(
			do.MustInvoke[identity.Repository](i),
			do.MustInvokeNamed[*indexstamp.Stamper](i, IndexTypeUsers),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}

// RegisterUserReindexer registers the users reindex backstop, which runs as a scheduled job.
//
// It writes to the searcher directly rather than through a Stamper. A reindex writes every
// document there is, so stamping it would turn last_indexed_at into a record of when the last
// rebuild ran — the same value on every row — instead of how current each document is.
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
