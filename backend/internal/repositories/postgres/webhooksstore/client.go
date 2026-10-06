/*
Package webhooksstore is platform-go's webhook store, with platform's own
recording installed on it.

It replaces internal/repositories/postgres/webhooks, which owned three tables of
its own — webhooks, webhook_trigger_configs and webhook_trigger_events — and ran
them in parallel with platform's endpoints, joined by a shared identifier. The
endpoints were the delivery machinery and the local rows were the model; now
there is one model, and it is platform's.

One field did not survive, and it was already inert. The local webhooks row
carried a method column, validated on every write and echoed back on every read,
while delivery went through platform's dispatcher — which posts. A client that
asked for PUT was told PUT and sent POST. Dropping the column removes the lie
rather than a capability.

The audit entry and the event every endpoint and subscription write owes are
platform's too: recordinghooks.RecordingHooks records them through the
recording.Recorder this application registers, on the caller's transaction, so
the row, the entry and the event commit together or not at all. What this
package decides is only that the hooks are installed — there is no store without
them to register by mistake.

The cycle an earlier version of this package resolved leniently — the store's
hooks record through an emitter whose dispatcher reads this store — is gone:
platform's emitter builds its own hookless fan-out store over the same tables
(see platform's webhooks/recordinghooks, "Wiring it without a cycle"), and the
dispatcher registered beside this store is for the writes that go through it.
*/
package webhooksstore

import (
	"context"

	platformrecording "github.com/primandproper/platform-go/v15/recording"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	"github.com/primandproper/platform-go/v15/webhooks/recordinghooks"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// ProvideStore builds platform's webhook store, with platform's recording hooks
// over recorder installed on it.
func ProvideStore(
	ctx context.Context,
	cfg *webhookscfg.Config,
	client database.Client,
	recorder *platformrecording.Recorder,
) (platformwebhooks.Store, error) {
	hooks, err := recordinghooks.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the webhooks recording hooks")
	}

	return webhookscfg.NewStore(ctx, cfg, client,
		webhookscfg.WithStoreOptions(platformwebhooks.WithHooks(hooks)),
	)
}
