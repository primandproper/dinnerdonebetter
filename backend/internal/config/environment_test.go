package config

import (
	"testing"

	appentitlements "github.com/primandproper/dinnerdonebetter/backend/internal/entitlements"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	"github.com/primandproper/platform-go/v13/entitlements"
	entitlementscfg "github.com/primandproper/platform-go/v13/entitlements/config"
	meteringmock "github.com/primandproper/platform-go/v13/metering/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringOrDefault(T *testing.T) {
	T.Parallel()

	T.Run("with empty string", func(t *testing.T) {
		t.Parallel()

		result := stringOrDefault("", "default")
		assert.Equal(t, "default", result)
	})

	T.Run("with non-empty string", func(t *testing.T) {
		t.Parallel()

		result := stringOrDefault("value", "default")
		assert.Equal(t, "value", result)
	})
}

// The shipped catalog is what the API server boots with, and every way it can be wrong is a
// refusal at startup rather than a wrong answer later: a grant naming a feature nobody declared,
// a quota feature naming a meter nobody registered, a fallback naming a plan nobody configured.
// This is those refusals, run as a test.
func TestDefaultEntitlementsConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := DefaultEntitlementsConfig()

		require.NoError(t, cfg.ValidateWithContext(ctx))
		assert.Equal(t, appentitlements.FreePlan, cfg.Checker.FallbackPlan)
	})

	T.Run("the configured plans build a catalog over the declared features", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := DefaultEntitlementsConfig()

		catalog, err := entitlementscfg.NewCatalog(ctx, &cfg, appentitlements.Features())
		require.NoError(t, err)

		assert.ElementsMatch(t,
			[]string{appentitlements.FreePlan, appentitlements.SubscriberPlan},
			catalog.PlanNames(),
		)
	})

	T.Run("the catalog's limits are the limits metering is handed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := DefaultEntitlementsConfig()

		catalog, err := entitlementscfg.NewCatalog(ctx, &cfg, appentitlements.Features())
		require.NoError(t, err)

		registry, err := appmetering.NewRegistry()
		require.NoError(t, err)

		// NewQuotaSource is where a quota feature naming a meter the registry does not have
		// is caught, which is the one drift between the two catalogs that nothing else
		// notices until a Check errors for that one feature in production.
		source, err := entitlementscfg.NewQuotaSource(
			ctx, &cfg, catalog, entitlements.NewStaticPlanSource(appentitlements.FreePlan), registry,
		)
		require.NoError(t, err)
		assert.NotNil(t, source)
	})

	T.Run("the checker builds", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := DefaultEntitlementsConfig()

		catalog, err := entitlementscfg.NewCatalog(ctx, &cfg, appentitlements.Features())
		require.NoError(t, err)

		// The enforcer is mocked because what it does is metering's business; what matters
		// here is that one is required at all. A catalog with quota features and no enforcer
		// is refused at construction rather than at the first Check of one, so a service
		// wired without it fails to boot rather than passing every test written for the
		// features it does answer.
		checker, err := entitlementscfg.NewChecker(
			ctx,
			&cfg,
			catalog,
			entitlements.NewStaticPlanSource(appentitlements.SubscriberPlan),
			&meteringmock.EnforcerMock{},
			nil,
			nil,
		)
		require.NoError(t, err)
		assert.NotNil(t, checker)
	})
}
