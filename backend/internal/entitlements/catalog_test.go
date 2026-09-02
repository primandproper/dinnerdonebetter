package entitlements

import (
	"testing"

	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"
	"github.com/primandproper/platform-go/v13/identifiers"
	platformmetering "github.com/primandproper/platform-go/v13/metering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shippedCatalog builds the catalog this service ships with: the code-declared features and the
// plans DefaultPlans supplies, registered in the order the platform requires them.
//
// It is what internal/config.DefaultEntitlementsConfig produces by the longer route, through the
// rendered config file and the platform's config constructor. The two are held together by
// TestDefaultEntitlementsConfig over in that package, which builds this catalog from the
// configured plans instead of from these.
func shippedCatalog(t *testing.T) *platformentitlements.Catalog {
	t.Helper()

	catalog := platformentitlements.NewCatalog()

	features := Features()
	for idx := range features {
		require.NoError(t, catalog.RegisterFeature(features[idx]), "feature %q", features[idx].Key)
	}

	plans := DefaultPlans()
	for idx := range plans {
		require.NoError(t, catalog.RegisterPlan(plans[idx]), "plan %q", plans[idx].Name)
	}

	return catalog
}

func TestQuotaSource(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		registry, err := appmetering.NewRegistry()
		require.NoError(t, err)

		source, err := platformentitlements.NewQuotaSource(
			shippedCatalog(t),
			platformentitlements.NewStaticPlanSource(FreePlan),
			registry,
		)
		require.NoError(t, err)
		assert.NotNil(t, source)
	})

	T.Run("the shipped catalog limits nothing", func(t *testing.T) {
		t.Parallel()

		// The parallel to the empty limits table this replaced: nothing is limited yet, and
		// this is here so that the first limit is a deliberate act with a test to change
		// rather than something that lands unnoticed.
		//
		// It asks through the quota source rather than reading the grants, because the quota
		// source is what metering enforces — a grant that said unlimited and a source that
		// reported something else is exactly the disagreement this arrangement exists to
		// make impossible.
		ctx := t.Context()

		registry, err := appmetering.NewRegistry()
		require.NoError(t, err)

		catalog := shippedCatalog(t)

		for _, plan := range catalog.PlanNames() {
			source, sourceErr := platformentitlements.NewQuotaSource(
				catalog,
				platformentitlements.NewStaticPlanSource(plan),
				registry,
			)
			require.NoError(t, sourceErr)

			for _, meter := range registry.MeterNames() {
				quota, quotaErr := source.QuotaFor(ctx, identifiers.New(), meter)
				require.NoError(t, quotaErr, "plan %q meter %q", plan, meter)

				assert.Equal(t, platformmetering.BehaviorAllowOverage, quota.Behavior, "plan %q meter %q", plan, meter)
				assert.Equal(t, platformentitlements.UnlimitedLimit, quota.Limit, "plan %q meter %q", plan, meter)
			}
		}
	})

	T.Run("every metered feature is granted by every plan", func(t *testing.T) {
		t.Parallel()

		// A meter no plan includes is answered with ErrNoQuota, which metering reports
		// rather than treating as unlimited — so a feature added to the catalog and left out
		// of a plan refuses that plan's accounts rather than waving them through.
		catalog := shippedCatalog(t)

		for _, plan := range catalog.PlanNames() {
			for _, key := range catalog.FeatureKeys() {
				_, included := catalog.GrantFor(plan, key)
				assert.True(t, included, "plan %q excludes feature %q", plan, key)
			}
		}
	})

	T.Run("a quota is bucketed by the meter's own period", func(t *testing.T) {
		t.Parallel()

		// The period is derived from the registry rather than declared beside the grant,
		// which is what makes a quota over a window its meter does not bucket by
		// unrepresentable rather than an error on the read path.
		ctx := t.Context()

		registry, err := appmetering.NewRegistry()
		require.NoError(t, err)

		source, err := platformentitlements.NewQuotaSource(
			shippedCatalog(t),
			platformentitlements.NewStaticPlanSource(SubscriberPlan),
			registry,
		)
		require.NoError(t, err)

		for _, name := range registry.MeterNames() {
			meter, ok := registry.Meter(name)
			require.True(t, ok)

			quota, quotaErr := source.QuotaFor(ctx, identifiers.New(), name)
			require.NoError(t, quotaErr)

			assert.Equal(t, meter.Period, quota.Period, "meter %q", name)
		}
	})
}
