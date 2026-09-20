package indexing

import (
	"context"
	"math"

	"github.com/primandproper/dinnerdonebetter/backend/internal/indexstamp"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	searchsync "github.com/primandproper/platform-go/v14/searchsync"
	syncsource "github.com/primandproper/platform-go/v14/searchsync/source"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// o11yName names the loggers, spans and metrics of the search sync sources built here. It
// keeps the name the deleted internal/search/syncsource used, so nothing downstream of a log
// query has to change.
const o11yName = "search_sync_source"

// UserSource reads users as search documents, for both the change feed and a reindex.
//
// The two reads are platform's SearchIndexWriter, curried into the shapes syncsource wants:
// neither of its methods takes a scope, because the sync services itself rather than acting
// for a caller, and both take an executor this closes over.
//
// The scan walks every live user and does not consult last_indexed_at. That is deliberate
// upstream and it is what this application's own scan did: a backstop that trusted the stamp
// would be trusting bookkeeping written by the thing it exists to check, and an event dropped
// before it was applied leaves a current stamp on a document that was never sent.
func UserSource(
	client database.Client,
	store platformidentity.Store,
) (*syncsource.Source[platformidentity.User, UserSearchSubset], error) {
	fetch := func(ctx context.Context, id string) (*platformidentity.User, error) {
		return store.GetUser(ctx, client.Reader(), tenancy.Global(), id)
	}

	scan := func(ctx context.Context, after string, limit int) ([]string, error) {
		// The page size is a uint8 upstream, which is a ceiling rather than an
		// inconvenience: a reindex page is a batch of documents held in memory and
		// then written to an index, and there is no size above 255 anybody wants.
		return store.ScanUsersForReindex(ctx, client.Reader(), after, clampPage(limit))
	}

	return syncsource.New(IndexTypeUsers, fetch, scan, ConvertUserToUserSearchSubset)
}

// clampPage fits a page size into the byte the scan takes.
//
// A caller asking for more than a byte holds gets a byte's worth rather than a wrapped
// number, and one asking for nothing gets the scan's own default by asking for zero.
func clampPage(limit int) uint8 {
	switch {
	case limit < 0:
		return 0
	case limit > math.MaxUint8:
		return math.MaxUint8
	default:
		return uint8(limit)
	}
}

// UserStamps is the write a Syncer stamps through.
//
// platform's MarkUsersAsIndexed is handed the whole flushed set, which is what the buffer
// exists to produce: one UPDATE per flush rather than one per document. The row count it
// answers with is dropped, because a set naming ids that have since been erased stamps fewer
// rows than it named, and that is a directory being written to while it is indexed rather
// than an error.
//
// It returns the write rather than the buffer so the buffer is still built by indexstamp,
// which is what gives the container a Shutdown to call: a batching.Buffer owns a goroutine,
// and one acquired anywhere else is one nothing retires.
func UserStamps(client database.Client, store platformidentity.Store) func(context.Context, []string) error {
	return func(ctx context.Context, ids []string) error {
		return client.WithTransaction(ctx, func(tx database.Tx) error {
			_, err := store.MarkUsersAsIndexed(ctx, tx, ids)

			return err
		})
	}
}

// NewUserSyncer builds the Syncer that applies one users-index event.
//
// It replaces the scheduler that used to publish an index request for every user a sampler
// thought looked stale. The events now come from the transactions that changed the rows.
func NewUserSyncer(
	client database.Client,
	store platformidentity.Store,
	index UserTextSearcher,
	stamps *indexstamp.Buffer,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*searchsync.Syncer[UserSearchSubset], error) {
	src, err := UserSource(client, store)
	if err != nil {
		return nil, err
	}

	opts := append(
		o11yOptions(logger, tracerProvider, metricsProvider),
		syncsource.WithSyncerOptions(searchsync.WithSyncerStamper(stamps)),
	)

	return syncsource.NewSyncer(src, index, opts...)
}

// NewUserReindexer builds the reindex backstop for the users index.
//
// It is given no stamper on purpose. A reindex writes every document there is, so stamping it
// would make last_indexed_at read as when the last rebuild ran rather than how current each
// document is — which is the question the column exists to answer.
func NewUserReindexer(
	client database.Client,
	store platformidentity.Store,
	index UserTextSearcher,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*searchsync.Reindexer[UserSearchSubset], error) {
	src, err := UserSource(client, store)
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
