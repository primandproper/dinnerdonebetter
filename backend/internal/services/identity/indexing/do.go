package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexstamp"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v12/search/sync"

	"github.com/samber/do/v2"
)

// stampBufferName names the users index's stamp buffer in the container.
const stampBufferName = "index_stamp.users"

// RegisterUserSyncer registers the Syncer that applies users-index events.
//
// It is registered for the process that consumes the index topic. Handle is a jobs.Handler, so
// the Pool that runs it is wired where the other consumers are rather than here.
func RegisterUserSyncer(i do.Injector) {
	// The stamp buffer is named because there is one per index and they are all the same
	// type. do retires it on container shutdown, which is where its goroutine is accounted for.
	do.ProvideNamed(i, stampBufferName, func(i do.Injector) (*indexstamp.Buffer, error) {
		return indexstamp.New(
			do.MustInvoke[identity.Repository](i).MarkUsersAsIndexed,
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})

	do.Provide(i, func(i do.Injector) (*searchsync.Syncer[UserSearchSubset], error) {
		return NewUserSyncer(
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[UserTextSearcher](i),
			do.MustInvokeNamed[*indexstamp.Buffer](i, stampBufferName),
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
