/*
Package payments is platform-go's billing store, assembled for this application.
The catalog, the subscriptions, the purchases and the ledger are platform's: the
schema, the paging, the tenancy column, the uniqueness that turns a redelivered
webhook into a collision instead of a second row, and the guarded status writes
all live there. So does the recording: billing.RecordingHooks writes the audit
entry and the event every write owes, on the write's transaction, through the
recording.Recorder this process built. A redelivery the store refuses —
billing.ErrStatusUnchanged, billing.ErrAlreadyCompleted, an exists sentinel —
calls no hook, so a replay records nothing.

What this package still decides is the table prefix, which has to match the one
the migration was rendered with (see internal/repositories/postgres/migrations),
and the observability names. Which of billing's events a subscriber may receive
is the webhook catalog's decision, not this package's.
*/
package payments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	"github.com/primandproper/platform-go/v15/billing"
	billingcfg "github.com/primandproper/platform-go/v15/billing/config"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "payments_db_client"
)

// ProvidePaymentsRepository provides platform's billing store, recording every
// write through recorder.
//
// The store is assembled through platform's own billing/config rather than by
// naming billing.NewSQLStore's options here, so the knobs are stated once
// upstream.
func ProvidePaymentsRepository(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	recorder *platformrecording.Recorder,
	client database.Client,
) (billing.Store, error) {
	hooks, err := billing.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the billing store's recording hooks")
	}

	store, err := billingcfg.NewStore(
		ctx,
		&billingcfg.Config{TablePrefix: branding.TablePrefix},
		client,
		billingcfg.WithLogger(logging.NewNamedLogger(logger, o11yName)),
		billingcfg.WithTracerProvider(tracerProvider),
		billingcfg.WithMetricsProvider(metricsProvider),
		billingcfg.WithStoreOptions(billing.WithHooks(hooks)),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the billing store")
	}

	return store, nil
}
