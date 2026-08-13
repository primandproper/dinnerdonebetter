/*
Package syncsource adapts this application's repositories to the two read seams platform-go's
search sync defines: searchsync.Fetcher, which the change feed reads one document per event
through, and searchsync.Scanner, which a reindex walks the whole source with.

There is one type here rather than nine because the nine differ only in three functions: how to
read one row, how to page over IDs, and how to turn a row into the subset that gets indexed.
Everything else — omitting rows that have since been deleted, holding the byte ordering a
reindex depends on, wrapping a row in a Document — is the same work every time, and the version
of this that wrote it out per entity is the version where one of the nine quietly disagrees with
the other eight.

Scan is implemented in terms of Fetch on purpose. Both have to produce the same document for the
same row or a reindex will overwrite what the change feed wrote with a differently-shaped copy,
and the cheapest way to guarantee agreement is to have one call the other: the scan query names
the next page of IDs and the fetch turns them into documents. It costs a second round trip per
page, on the background path where that is affordable, and it removes the possibility of two
row-to-document transforms drifting apart.
*/
package syncsource

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
	textsearch "github.com/primandproper/platform-go/v10/search/text"
)

// o11yName names this package's loggers, spans and metrics.
const o11yName = "search_sync_source"

type (
	// FetchFunc reads one row by ID. It returns sql.ErrNoRows when the row is gone, which is
	// an expected outcome here rather than a failure — see Fetch.
	FetchFunc[E any] func(ctx context.Context, id string) (*E, error)

	// ScanFunc returns up to limit IDs sorting strictly after `after`, in ascending byte
	// order. It is the repository's ScanXIDsForReindex method.
	ScanFunc func(ctx context.Context, after string, limit int) ([]string, error)

	// ConvertFunc turns a domain object into the subset that is actually indexed.
	ConvertFunc[E, T any] func(*E) *T
)

// Source is a searchsync.Fetcher and searchsync.Scanner over one entity type.
//
// E is the domain object the repository returns; T is the search subset the index holds.
type Source[E, T any] struct {
	fetch   FetchFunc[E]
	scan    ScanFunc
	convert ConvertFunc[E, T]
	name    string
}

// New builds a Source. name appears in the error messages, so that a failure says which index
// it came from rather than only that a fetch failed.
func New[E, T any](name string, fetch FetchFunc[E], scan ScanFunc, convert ConvertFunc[E, T]) *Source[E, T] {
	return &Source[E, T]{name: name, fetch: fetch, scan: scan, convert: convert}
}

var (
	_ searchsync.Fetcher[struct{}] = (*Source[struct{}, struct{}])(nil)
	_ searchsync.Scanner[struct{}] = (*Source[struct{}, struct{}])(nil)
)

// Name is the index this Source feeds, and the name its Syncer and Reindexer carry.
func (s *Source[E, T]) Name() string { return s.name }

// NewSyncer builds the Syncer that applies one index event for this Source's entity.
//
// It owns no goroutine and reads from no queue: Handle is a jobs.Handler, and the Pool calling
// it supplies concurrency, retry with backoff, dead-lettering and a draining shutdown. Nothing
// here reimplements any of that.
func NewSyncer[E, T any](
	source *Source[E, T],
	index textsearch.IndexManager,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*searchsync.Syncer[T], error) {
	target, err := searchsync.TextTarget[T](index)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "building %s search target", source.Name())
	}

	syncer, err := searchsync.NewSyncer(
		source.Name(),
		source,
		target,
		searchsync.WithSyncerLogger(logging.NewNamedLogger(logger, o11yName)),
		searchsync.WithSyncerTracerProvider(tracerProvider),
		searchsync.WithSyncerMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "building %s syncer", source.Name())
	}

	return syncer, nil
}

// NewReindexer builds the reindex backstop for this Source's index.
//
// The Syncer keeps an index current; a Reindexer rebuilds one. They are separate because they
// answer different failures — a Syncer cannot repair an index that was already wrong before the
// first event was written, and a full walk is far too expensive to be the steady-state path.
//
// No pruner is configured, so a reindex converges the documents the source has and leaves any
// orphan it does not name. Pruning needs an Enumerator over what the index currently holds, and
// none of the text backends behind textsearch.Index can enumerate. Deletions reach the index
// through the change feed instead, which is where they are timely anyway.
func NewReindexer[E, T any](
	source *Source[E, T],
	index textsearch.IndexManager,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	opts ...searchsync.ReindexOption,
) (*searchsync.Reindexer[T], error) {
	target, err := searchsync.TextTarget[T](index)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "building %s search target", source.Name())
	}

	reindexer, err := searchsync.NewReindexer(
		source.Name(),
		source,
		target,
		append([]searchsync.ReindexOption{
			searchsync.WithReindexLogger(logging.NewNamedLogger(logger, o11yName)),
			searchsync.WithReindexTracerProvider(tracerProvider),
			searchsync.WithReindexMetricsProvider(metricsProvider),
		}, opts...)...,
	)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "building %s reindexer", source.Name())
	}

	return reindexer, nil
}

// Fetch returns the current document for each of ids, omitting any whose row no longer exists.
//
// The omission is the interesting half of the contract. A missing row is not an error and must
// not be reported as one: it is how the Syncer learns that a row was deleted between the event
// being written and the event being applied, and it responds by removing the document rather
// than leaving a tombstone in the index. Returning an error here instead would retry the event
// until it dead-lettered, and leave the deleted document in the index the whole time.
func (s *Source[E, T]) Fetch(ctx context.Context, ids ...string) ([]searchsync.Document[T], error) {
	documents := make([]searchsync.Document[T], 0, len(ids))

	for _, id := range ids {
		entity, err := s.fetch(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, platformerrors.Wrapf(err, "fetching %s document %q", s.name, id)
		}

		if entity == nil {
			continue
		}

		documents = append(documents, searchsync.Document[T]{ID: id, Body: s.convert(entity)})
	}

	return documents, nil
}

// Scan returns up to limit documents whose IDs sort strictly after `after`, in ascending byte
// order.
//
// The re-sort at the end is not redundant with the query's ORDER BY. Fetch is documented to
// return documents in any order, and it also drops the ones whose rows have gone, so what comes
// back is neither ordered nor the same length as what went in. A reindex merges this stream
// against the index's own ordered stream to decide what to prune, and hands the caller's
// ordering straight to that merge — so an unordered page here is not a cosmetic problem, it is
// live documents being deleted.
func (s *Source[E, T]) Scan(ctx context.Context, after string, limit int) ([]searchsync.Document[T], error) {
	ids, err := s.scan(ctx, after, limit)
	if err != nil {
		return nil, platformerrors.Wrapf(err, "scanning %s IDs for reindex", s.name)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	documents, err := s.Fetch(ctx, ids...)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(documents, func(a, b searchsync.Document[T]) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})

	return documents, nil
}
