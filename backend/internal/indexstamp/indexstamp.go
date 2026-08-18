/*
Package indexstamp records, on the row a search document was built from, when that document
was last written to the search index.

# Why this exists

last_indexed_at is a column nine tables carry. Before platform-go v10 it drove indexing: a
sampler asked "which rows look stale" and published an index request for each, so the column
was both written by the indexer and read by the query that chose its next batch. The reindexer
that replaced the sampler walks every ID in byte order instead — see ScanXIDsForReindex — and
asks the column nothing. That left it written by nobody and read by nobody, and a column in
that state is a lie about the schema rather than a spare field.

Deleting it was the alternative, and it is not free: querygen.StandardCRUD derives whether to
emit the reindex scan from this column's presence, so removing it here would silently stop
nine tables getting the query the reindexer walks. This package takes the other answer —
write it — which needs no agreement with the platform and turns the column into the one thing
nothing else can currently say: how current a document in the index is.

# Why the write is buffered

A Stamper wraps the index the Syncer writes through, so a stamp happens exactly when a
document is accepted by the index and not merely when an event is consumed. That puts a write
on the sync path, and the sync path is as concurrent as its jobs.Pool.

One UPDATE per applied document, issued from every worker at once, is the shape
batching.Buffer's documentation is about: statements against the same handful of popular rows,
taking row locks in whatever order each caller built them in, deadlocking on 40P01 and holding
a pool connection while they do. A Buffer collapses repeats of a key, flushes on an interval
from one goroutine, and emits in key order, so however busy the consumer is there is one
stamping write in flight and one lock order.

Nothing waits on the result. A stamp that is lost to a failing flush, or to a shutdown that
outran the interval, costs an observation and nothing else — no read path consults this column,
and the reindexer's correctness does not depend on it. That is what makes Buffer the right half
of the package rather than GroupCommit.
*/
package indexstamp

import (
	"context"
	"errors"
	"strings"

	"github.com/primandproper/platform-go/v11/batching"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	textsearch "github.com/primandproper/platform-go/v11/search/text"
)

// maxPending is how many distinct IDs accumulate before a flush is forced.
//
// It is lower than batching's own default because a flush here is a loop of single-row
// UPDATEs rather than one statement — the repositories' MarkXAsIndexed is the write, and
// giving it a bulk sibling would mean a bespoke query on all nine tables that
// querygen.StandardCRUD does not emit. A few hundred round trips fit inside a flush timeout
// comfortably; a few thousand would not.
const maxPending = 256

var (
	// ErrNilIndex is returned when New is given no index to wrap.
	ErrNilIndex = errors.New("nil index provided")

	// ErrNilMarkFunc is returned when New is given no way to stamp a row.
	ErrNilMarkFunc = errors.New("nil mark function provided")
)

// MarkFunc stamps one row as indexed. It is a repository's MarkXAsIndexed method.
type MarkFunc func(ctx context.Context, id string) error

// Stamper is a textsearch.IndexManager that also records, on the row each document came from,
// that the document was written.
//
// It is an IndexManager rather than a wrapper around the Syncer because that is the level at
// which "the document was indexed" is true: an event for a row that has since been deleted
// reaches the Syncer and turns into a delete, and a document the index refused was not
// indexed. Neither should stamp anything, and neither does.
//
// A Stamper owns the goroutine its Buffer owns, and must be Closed.
type Stamper struct {
	index  textsearch.IndexManager
	buffer *batching.Buffer[string]
}

var _ textsearch.IndexManager = (*Stamper)(nil)

// New wraps index so that every document it accepts stamps its row through mark.
//
// opts are passed to the underlying batching.Buffer, after the defaults this package sets, so
// a caller supplies the observability pillars — and, in a test, a shorter flush interval —
// with batching's own options.
func New(index textsearch.IndexManager, mark MarkFunc, opts ...batching.Option) (*Stamper, error) {
	if index == nil {
		return nil, ErrNilIndex
	}

	if mark == nil {
		return nil, ErrNilMarkFunc
	}

	buffer, err := batching.NewBuffer(
		stampAll(mark),
		append([]batching.Option{
			// Emitting in ID order is what turns contention between a flush and any other
			// writer of these rows into a queue rather than a deadlock cycle.
			batching.WithOrder(strings.Compare),
			batching.WithMaxPending(maxPending),
		}, opts...)...,
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building index stamp buffer")
	}

	return &Stamper{index: index, buffer: buffer}, nil
}

// Index writes the document, and buffers a stamp for its row once the index has taken it.
//
// The stamp is buffered after the write rather than alongside it, so a failed index leaves
// last_indexed_at saying what it said before: that this document has not been written since
// whenever it last was.
func (s *Stamper) Index(ctx context.Context, id string, value any) error {
	if err := s.index.Index(ctx, id, value); err != nil {
		return err
	}

	s.buffer.Add(id)

	return nil
}

// Delete removes the document and stamps nothing. There is no row left to stamp — a delete
// reaches the index because the row was archived or vanished — and the column is about
// documents the index holds.
func (s *Stamper) Delete(ctx context.Context, id string) error {
	return s.index.Delete(ctx, id)
}

// Wipe empties the index and stamps nothing, for the same reason Delete does not: it is the
// removal half of a rebuild, and the writes that follow it are what stamp.
func (s *Stamper) Wipe(ctx context.Context) error {
	return s.index.Wipe(ctx)
}

// Close stops the flusher and writes whatever stamps it still holds, bounded by ctx.
//
// Call it after whatever is writing through this Stamper has drained, so a document indexed
// during shutdown still stamps.
func (s *Stamper) Close(ctx context.Context) error {
	return s.buffer.Close(ctx)
}

// stampAll is the buffer's write: the keys one flush accumulated, stamped one row at a time.
//
// It does not stop at the first failure. The IDs in a batch are unrelated rows, so one that
// cannot be stamped — archived between the index write and the flush, most likely — says
// nothing about the rest, and abandoning them would lose observations for a reason that
// applies to only one.
func stampAll(mark MarkFunc) func(ctx context.Context, ids []string) error {
	return func(ctx context.Context, ids []string) error {
		var errs []error

		for _, id := range ids {
			if err := mark(ctx, id); err != nil {
				errs = append(errs, platformerrors.Wrapf(err, "stamping %q as indexed", id))
			}
		}

		return errors.Join(errs...)
	}
}
