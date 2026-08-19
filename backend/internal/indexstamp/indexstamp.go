/*
Package indexstamp holds the search syncers' last_indexed_at writers.

search/sync stamps through a searchsync.Stamper, and *batching.Buffer[string] already is one.
What it is not is a service the container knows how to retire: a Buffer owns a goroutine and
must be Closed, and do calls Shutdown. This package is that adapter, plus the one decision that
does not belong in nine call sites — which observability the buffers get, and that ids are
handed to the database in bulk.

Registering a Buffer here rather than building one inside a syncer is deliberate, and it is the
platform's own reasoning: a Syncer owns no goroutine and has no lifecycle, so acquiring one
through an option would be a shutdown obligation that nothing in its signature mentions. Here
the obligation is the container's, which is the one thing in the process that already knows how
to discharge it.
*/
package indexstamp

import (
	"context"

	"github.com/primandproper/platform-go/v11/batching"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v11/search/sync"
)

// o11yName names the loggers, spans and metrics of the stamp buffers built here.
const o11yName = "index_stamp"

// Buffer is the stamp writer for one index, with the container's shutdown method on it.
//
// It embeds the platform's Buffer rather than wrapping it method by method, so it satisfies
// searchsync.Stamper by being one.
type Buffer struct {
	_ struct{} `json:"-"`

	*batching.Buffer[string]
}

// Stamper is what a Buffer is for, restated so a call site can name the narrow thing.
var _ searchsync.Stamper = (*Buffer)(nil)

// New builds the buffered stamp writer for one index.
//
// write is handed every id in a flush at once, because one statement per flush is the entire
// reason the write is buffered — the nine MarkXAsIndexed repository methods are exactly that
// shape. Ordering is not set here: NewStampBuffer pins it, since a stamping write that takes
// row locks in whatever order each caller happened to build them in is how a bookkeeping
// column deadlocks endpoints that have nothing to do with it.
func New(
	write func(ctx context.Context, ids []string) error,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*Buffer, error) {
	buffer, err := searchsync.NewStampBuffer(
		write,
		batching.WithLogger(logging.NewNamedLogger(logger, o11yName)),
		batching.WithTracerProvider(tracerProvider),
		batching.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, err
	}

	return &Buffer{Buffer: buffer}, nil
}

// Shutdown flushes what the buffer is holding and stops its goroutine.
//
// It is do's ShutdownerWithContextAndError, so the container retires the buffers as part of its
// own shutdown. Whoever shuts the container down must do it after the things that Add have
// stopped, or the last flush races the work that fills it.
func (b *Buffer) Shutdown(ctx context.Context) error {
	return b.Close(ctx)
}
