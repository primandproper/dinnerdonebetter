package scheduler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	auditcfg "github.com/primandproper/platform-go/v15/audit/config"
	"github.com/primandproper/platform-go/v15/retention"
	"github.com/primandproper/primitives-go/v2/database"

	"github.com/samber/do/v2"
)

// RegisterRetentionPolicies registers the retention policies this application's sweep enforces,
// of which the audit log is currently the only one.
//
// platform-go v10 took the sweep loop out of the audit package: what audit owns now is a
// retention.Policy describing its window and its bounds, and a generic sweeper drives it. The
// sweeper itself is platform's, registered from the service.Config Retention block, and it
// resolves the policies from here — adding a second table to prune is a second Policy in the
// slice below, not a second loop.
//
// The sweeper's Run method is deliberately unused: the scheduler drives Sweep as a registered
// job instead, so exactly one replica prunes per tick because the distributed lock says so,
// rather than because the deployment happens to run one replica. The sweep is safe to run
// concurrently — the audit policy prunes a prefix of a chain inside a transaction — but this is
// work that deletes, and doing it several times over for one result is not a thing to leave to
// convention.
//
// Pruning the one table this application treats as immutable is a strange thing to schedule, so
// it is worth being precise about why it is safe. The audit target only ever removes a prefix of
// a scope's chain, never a row from the middle, so the survivors stay contiguous and verifiable
// against each other; and it records the hash of the last entry it removed as that scope's
// watermark, in the same transaction as the delete, so the oldest surviving entry still links to
// something and Verify can tell retention's gap from a deletion.
func RegisterRetentionPolicies(i do.Injector) {
	do.Provide[[]retention.Policy](i, func(i do.Injector) ([]retention.Policy, error) {
		// Copied rather than passed by reference, because the two fields below are
		// overwritten and the config struct is shared with whatever else reads it.
		cfg := do.MustInvoke[*config.SchedulerConfig](i).AuditLog

		// Pinned, not validated. Neither field has a second legal value: the prefix has to
		// equal the one the migration rendered the tables under, and the migrations are
		// Postgres. A deployment that set either differently would not be configuring
		// retention, it would be pointing the sweep at tables that do not exist — and a
		// sweep of a table that isn't there reports success forever.
		cfg.TablePrefix = branding.TablePrefix
		cfg.Dialect = do.MustInvoke[database.Client](i).Dialect()

		auditPolicy, err := auditcfg.NewRetentionPolicy(do.MustInvoke[context.Context](i), &cfg)
		if err != nil {
			return nil, err
		}

		return []retention.Policy{auditPolicy}, nil
	})
}
