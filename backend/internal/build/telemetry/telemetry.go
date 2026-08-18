package telemetry

import (
	"context"
	"log"
	"time"

	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

const flushTimeout = 5 * time.Second

// Flush shuts down the injector's metrics provider and force-flushes its tracer provider
// so short-lived worker pods export their spans/metrics before the process exits. Defer it
// immediately after building the injector.
func Flush(ctx context.Context, i do.Injector) {
	shutdownCtx, cancel := context.WithTimeout(ctx, flushTimeout)
	defer cancel()

	if shutdownErr := do.MustInvoke[metrics.Provider](i).Shutdown(shutdownCtx); shutdownErr != nil {
		log.Printf("error shutting down metrics: %v", shutdownErr)
	}
	if flushErr := do.MustInvoke[tracing.Provider](i).ForceFlush(shutdownCtx); flushErr != nil {
		log.Printf("error flushing traces: %v", flushErr)
	}
}
