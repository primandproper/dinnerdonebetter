package scheduler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuditSweeper registers the audit log's retention sweeper with the injector.
//
// It lives in this process because it is background work over the database, which is what this
// process is for. Its Run method is deliberately unused: the scheduler drives Sweep as a
// registered job instead, so exactly one replica prunes per tick because the distributed lock
// says so, rather than because the deployment happens to run one replica. The Sweeper is safe
// to run concurrently — it prunes a prefix of a chain inside a transaction — but this is work
// that deletes, and doing it several times over for one result is not a thing to leave to
// convention.
//
// It deletes from the one table this application treats as immutable, which is a strange thing
// to schedule, so it is worth being precise about why it is safe. It only ever removes a prefix
// of a scope's chain, never a row from the middle, so the survivors stay contiguous and
// verifiable against each other; and it records the hash of the last entry it removed as that
// scope's watermark, in the same transaction as the delete, so the oldest surviving entry still
// links to something and Verify can tell retention's gap from a deletion.
func RegisterAuditSweeper(i do.Injector) {
	do.Provide[*audit.Sweeper](i, func(i do.Injector) (*audit.Sweeper, error) {
		// Copied rather than passed by reference, because the two fields below are
		// overwritten and the config struct is shared with whatever else reads it.
		cfg := do.MustInvoke[*config.SchedulerConfig](i).Audit

		// Pinned, not validated. Neither field has a second legal value: the prefix has to
		// equal the one the migration rendered the tables under, and the migrations are
		// Postgres. A deployment that set either differently would not be configuring
		// retention, it would be pointing the sweeper at tables that do not exist — and a
		// sweeper pruning a table that isn't there reports success forever.
		cfg.TablePrefix = ddbaudit.TablePrefix
		cfg.Dialect = do.MustInvoke[database.Client](i).Dialect()

		return audit.NewSweeper(
			do.MustInvoke[context.Context](i),
			&cfg,
			do.MustInvoke[database.Client](i),
			audit.WithSweeperLogger(do.MustInvoke[logging.Logger](i)),
			audit.WithSweeperTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
			audit.WithSweeperMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
