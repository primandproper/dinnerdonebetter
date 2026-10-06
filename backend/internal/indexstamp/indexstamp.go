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
	"errors"
	"strings"

	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/primitives-go/v2/batching"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
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

// NamePrefix begins the container name of every index's stamp buffer. Each index registers its
// own, named NamePrefix plus the index, because they are all one type and do resolves by name.
const NamePrefix = "index_stamp."

// ShutdownAll flushes and stops every stamp buffer i has built.
//
// It exists for a process composed by platform's service.New, whose shutdown releases the
// database client before the container is retired. Retiring the container is how these buffers
// have always been flushed, so in such a process their last flush would be written through a
// client that is already closed — and a lost flush is a document that was indexed and is stamped
// as though it never was. Calling this after the consumers that fill the buffers have stopped, and
// before the service releases its clients, puts the flush back where it belongs.
//
// A buffer shut down here is gone from the container, so the container's own shutdown afterwards
// does not flush it a second time. One nobody built is not built to be shut down.
//
// platform's searchsync.Registry owns the stamp buffers of the indexes registered on it, and
// service.New flushes it in exactly this slot; this is what stands in for it until this
// application's syncers are registered there.
func ShutdownAll(ctx context.Context, i do.Injector) error {
	var errs []error

	invoked := i.ListInvokedServices()
	for idx := range invoked {
		name := invoked[idx].Service
		if !strings.HasPrefix(name, NamePrefix) {
			continue
		}

		if err := do.ShutdownNamedWithContext(ctx, i, name); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
