package scheduler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	auditcfg "github.com/primandproper/platform-go/v10/audit/config"
	"github.com/primandproper/platform-go/v10/database"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	"github.com/primandproper/platform-go/v10/retention"

	"github.com/samber/do/v2"
)

// RegisterRetentionSweeper registers the sweeper that enforces this application's retention
// policies, of which the audit log is currently the only one.
//
// platform-go v10 took the sweep loop out of the audit package: what audit owns now is a
// retention.Policy describing its window and its bounds, and a generic sweeper drives it. That
// is why this is a retention.Sweeper holding an audit policy rather than an audit.Sweeper —
// adding a second table to prune is a second Policy in the slice below, not a second loop.
//
// It lives in this process because it is background work over the database, which is what this
// process is for. Its Run method is deliberately unused: the scheduler drives Sweep as a
// registered job instead, so exactly one replica prunes per tick because the distributed lock
// says so, rather than because the deployment happens to run one replica. The sweep is safe to
// run concurrently — the audit policy prunes a prefix of a chain inside a transaction — but this
// is work that deletes, and doing it several times over for one result is not a thing to leave
// to convention.
//
// Pruning the one table this application treats as immutable is a strange thing to schedule, so
// it is worth being precise about why it is safe. The audit target only ever removes a prefix of
// a scope's chain, never a row from the middle, so the survivors stay contiguous and verifiable
// against each other; and it records the hash of the last entry it removed as that scope's
// watermark, in the same transaction as the delete, so the oldest surviving entry still links to
// something and Verify can tell retention's gap from a deletion.
func RegisterRetentionSweeper(i do.Injector) {
	do.Provide[*retention.Sweeper](i, func(i do.Injector) (*retention.Sweeper, error) {
		// Copied rather than passed by reference, because the two fields below are
		// overwritten and the config struct is shared with whatever else reads it.
		cfg := do.MustInvoke[*config.SchedulerConfig](i).Audit

		// Pinned, not validated. Neither field has a second legal value: the prefix has to
		// equal the one the migration rendered the tables under, and the migrations are
		// Postgres. A deployment that set either differently would not be configuring
		// retention, it would be pointing the sweep at tables that do not exist — and a
		// sweep of a table that isn't there reports success forever.
		cfg.TablePrefix = ddbaudit.TablePrefix
		cfg.Dialect = do.MustInvoke[database.Client](i).Dialect()

		ctx := do.MustInvoke[context.Context](i)

		auditPolicy, err := auditcfg.NewRetentionPolicy(ctx, &cfg)
		if err != nil {
			return nil, err
		}

		return retention.NewSweeper(
			ctx,
			&do.MustInvoke[*config.SchedulerConfig](i).Retention,
			do.MustInvoke[database.Client](i),
			[]retention.Policy{auditPolicy},
			retention.WithSweeperLogger(do.MustInvoke[logging.Logger](i)),
			retention.WithSweeperTracerProvider(do.MustInvoke[tracing.Provider](i)),
			retention.WithSweeperMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
