package scheduler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditSweeper registers the audit log's retention sweeper with the injector.
//
// It runs here, beside the outbox relay, for the same two reasons: it is a polling loop that
// must not be tied to a request, and it needs only the database. Unlike the relay it is not
// safe to scale casually — each tick deletes and rewrites a prune watermark per scope, and two
// replicas sweeping the same scope would contend on those transactions rather than divide the
// work. One replica is the intended shape.
//
// It deletes from the one table this application treats as immutable, which is a strange thing
// to schedule, so it is worth being precise about why it is safe. It only ever removes a prefix
// of a scope's chain, never a row from the middle, so the survivors stay contiguous and
// verifiable against each other; and it records the hash of the last entry it removed as that
// scope's watermark, in the same transaction as the delete, so the oldest surviving entry still
// links to something and Verify can tell retention's gap from a deletion.
func RegisterAuditSweeper(i do.Injector) {
	do.Provide[*audit.Sweeper](i, func(i do.Injector) (*audit.Sweeper, error) {
		return audit.NewSweeper(
			do.MustInvoke[context.Context](i),
			&do.MustInvoke[*config.SchedulerConfig](i).Audit,
			do.MustInvoke[database.Client](i),
			audit.WithSweeperLogger(do.MustInvoke[logging.Logger](i)),
			audit.WithSweeperTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
			audit.WithSweeperMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
