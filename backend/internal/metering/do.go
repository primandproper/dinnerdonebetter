package metering

import (
	"context"

	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"

	"github.com/primandproper/platform-go/v11/capitalism"
	"github.com/primandproper/platform-go/v11/database"
	platformmetering "github.com/primandproper/platform-go/v11/metering"
	meteringcfg "github.com/primandproper/platform-go/v11/metering/config"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/logging"

	"github.com/samber/do/v2"
)

// RegisterRegistry registers the meter registry with the injector.
//
// Every process that meters registers it, and they must all register the same one: the recorder
// writes totals keyed by a meter's period and aggregation, and a second process reading them
// under a different declaration would be reading a different number.
func RegisterRegistry(i do.Injector) {
	do.Provide[*platformmetering.Registry](i, func(do.Injector) (*platformmetering.Registry, error) {
		return NewRegistry()
	})
}

// RegisterStore registers the metering store with the injector.
//
// Prerequisites: *meteringcfg.Config and a database.Client. The store's dialect comes from the
// client rather than from configuration, so it cannot disagree with the database the statements
// actually run against.
func RegisterStore(i do.Injector) {
	do.Provide[platformmetering.Store](i, func(i do.Injector) (platformmetering.Store, error) {
		ctx := do.MustInvoke[context.Context](i)

		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		return meteringcfg.NewStore(
			ctx,
			do.MustInvoke[*meteringcfg.Config](i),
			do.MustInvoke[database.Client](i),
			meteringcfg.WithPillars(pillars),
		)
	})
}

// RegisterRecorder registers the ingest path with the injector.
//
// This is all a process that counts usage needs. Enforcement is a separate component on purpose,
// and a write site should hold a Recorder rather than an Enforcer unless it actually intends to
// refuse something.
//
// No analytics reporter is attached. The platform makes that opt-in because it posts one
// analytics event per accepted record, which for a meter on a hot path is a second data pipeline
// the size of the first, billed by the row.
func RegisterRecorder(i do.Injector) {
	do.Provide[platformmetering.Recorder](i, func(i do.Injector) (platformmetering.Recorder, error) {
		ctx := do.MustInvoke[context.Context](i)

		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		return meteringcfg.NewRecorder(
			ctx,
			do.MustInvoke[*meteringcfg.Config](i),
			do.MustInvoke[platformmetering.Store](i),
			do.MustInvoke[*platformmetering.Registry](i),
			// No period resolver: the calendar resolver is the default, and the calendar
			// month is the period every meter here uses. See the package documentation for
			// why it is not the subscription's billing anchor.
			nil,
			// No analytics reporter — see above.
			nil,
			meteringcfg.WithPillars(pillars),
		)
	})
}

// RegisterEnforcer registers the read path with the injector.
//
// Nothing calls it yet. It is registered so that the day a limit goes on planLimits, the change
// is a Check at one call site rather than a wiring exercise — which is the promise the whole
// count-first ordering rests on.
//
// Two things have to change with that first limit. The enforcer is built without a cache, so
// Check reads the durable total on every call, which is exactly the latency the platform warns
// at length about; attach a cache.Cache[metering.CachedTotal] before any Check reaches a request
// path. And EnforcerConfig.FailOpen decides what happens when the store is unreachable, which is
// a product question — refuse everyone, or serve everyone free — that nobody has been asked yet.
func RegisterEnforcer(i do.Injector) {
	do.Provide[platformmetering.Enforcer](i, func(i do.Injector) (platformmetering.Enforcer, error) {
		ctx := do.MustInvoke[context.Context](i)

		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		registry := do.MustInvoke[*platformmetering.Registry](i)

		quotas, err := NewSubscriptionQuotaSource(
			registry,
			do.MustInvoke[paymentsmanager.PaymentsDataManager](i),
			platformmetering.WithPlanLimitLogger(logging.NewNamedLogger(pillars.Logger, quotaSourceO11yName)),
			platformmetering.WithPlanLimitTracerProvider(pillars.TracerProvider),
			platformmetering.WithPlanLimitMetricsProvider(pillars.MetricsProvider),
		)
		if err != nil {
			return nil, err
		}

		return meteringcfg.NewEnforcer(
			ctx,
			do.MustInvoke[*meteringcfg.Config](i),
			do.MustInvoke[platformmetering.Store](i),
			registry,
			nil,
			quotas,
			// No cache — see above.
			nil,
			meteringcfg.WithPillars(pillars),
		)
	})
}

// RegisterFlusher registers the provider push with the injector.
//
// It belongs in a worker rather than in the API server: a flush is a scheduled pass over a
// backlog, and the credentials it needs to post usage are not credentials a request path has any
// business holding.
//
// Prerequisites: *meteringcfg.Config, a metering Store, and a capitalism.UsageReporter. The
// reporter has no default — the platform refuses to invent one, because a flusher posting
// nowhere still marks usage flushed. The noop reporter is a supported deployment and is what a
// capitalism provider of "noop" yields, so choosing it stays a config decision rather than an
// omission here.
func RegisterFlusher(i do.Injector) {
	do.Provide[*platformmetering.Flusher](i, func(i do.Injector) (*platformmetering.Flusher, error) {
		ctx := do.MustInvoke[context.Context](i)

		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		return meteringcfg.NewFlusher(
			ctx,
			do.MustInvoke[*meteringcfg.Config](i),
			do.MustInvoke[platformmetering.Store](i),
			NewProviderMapper(),
			do.MustInvoke[capitalism.UsageReporter](i),
			meteringcfg.WithPillars(pillars),
		)
	})
}
