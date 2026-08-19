package indexstamp

import (
	"context"
	"sync"
	"testing"

	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v11/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuffer_Shutdown(T *testing.T) {
	T.Parallel()

	T.Run("flushes what it is holding", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		var (
			mu      sync.Mutex
			written []string
		)

		buffer, err := New(
			func(_ context.Context, ids []string) error {
				mu.Lock()
				defer mu.Unlock()
				written = append(written, ids...)

				return nil
			},
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
		)
		require.NoError(t, err)

		buffer.Add("first", "second")

		// Nothing has flushed yet: the buffer is holding these until its interval elapses or
		// it is closed. Shutdown is the half that matters — a process that exits without it
		// loses every stamp still in memory, which is what this asserts against.
		require.NoError(t, buffer.Shutdown(ctx))

		mu.Lock()
		defer mu.Unlock()
		assert.ElementsMatch(t, []string{"first", "second"}, written)
	})
}
