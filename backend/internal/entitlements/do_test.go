package entitlements

import (
	"context"
	"testing"

	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"

	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"
	entitlementscfg "github.com/primandproper/platform-go/v13/entitlements/config"
	platformmetering "github.com/primandproper/platform-go/v13/metering"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// injectorWithPrerequisites builds a container holding everything the registrations under test
// read, minus the ones they register themselves.
//
// The plan source is static rather than the subscription-backed one, because what this is about
// is the wiring: the account-to-plan question has its own tests, and answering it here would
// mean standing up the whole payments manager to learn nothing about the container.
func injectorWithPrerequisites(t *testing.T) do.Injector {
	t.Helper()

	registry, err := appmetering.NewRegistry()
	require.NoError(t, err)

	i := do.New()

	do.ProvideValue[context.Context](i, t.Context())
	do.ProvideValue[*entitlementscfg.Config](i, &entitlementscfg.Config{Plans: DefaultPlans()})
	do.ProvideValue[*platformmetering.Registry](i, registry)
	do.ProvideValue[platformentitlements.PlanSource](i, platformentitlements.NewStaticPlanSource(FreePlan))

	return i
}

func TestRegisterFeatures(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		RegisterFeatures(i)

		features, err := do.Invoke[[]platformentitlements.Feature](i)
		require.NoError(t, err)
		assert.Equal(t, Features(), features)
	})
}

func TestRegisterCatalog(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := injectorWithPrerequisites(t)
		RegisterFeatures(i)
		RegisterCatalog(i)

		catalog, err := do.Invoke[*platformentitlements.Catalog](i)
		require.NoError(t, err)

		assert.Equal(t, []string{UploadedMediaBytesFeature}, catalog.FeatureKeys())
		assert.ElementsMatch(t, []string{FreePlan, SubscriberPlan}, catalog.PlanNames())
	})
}

func TestRegisterQuotaSource(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := injectorWithPrerequisites(t)
		RegisterFeatures(i)
		RegisterCatalog(i)
		RegisterQuotaSource(i)

		source, err := do.Invoke[*platformentitlements.QuotaSource](i)
		require.NoError(t, err)
		assert.NotNil(t, source)
	})

	T.Run("the metering enforcer is handed the catalog's own source", func(t *testing.T) {
		t.Parallel()

		// The whole point of the arrangement, asserted as identity: metering enforces
		// whatever its QuotaSource says, and this is what makes that object the catalog.
		// Two separately built sources would answer the same today and drift the first time
		// one of them was given a cache or a different plan source.
		i := injectorWithPrerequisites(t)
		RegisterFeatures(i)
		RegisterCatalog(i)
		RegisterQuotaSource(i)

		catalogSource, err := do.Invoke[*platformentitlements.QuotaSource](i)
		require.NoError(t, err)

		meteringSource, err := do.Invoke[platformmetering.QuotaSource](i)
		require.NoError(t, err)

		assert.Same(t, catalogSource, meteringSource)
	})
}

func TestRegisterChecker(T *testing.T) {
	T.Parallel()

	T.Run("with no enforcer registered", func(t *testing.T) {
		t.Parallel()

		// The catalog has a quota feature, so a container without a metering enforcer
		// refuses to build the checker rather than building one that errors on the only
		// feature that costs money.
		i := injectorWithPrerequisites(t)
		RegisterFeatures(i)
		RegisterCatalog(i)
		RegisterChecker(i)

		_, err := do.Invoke[platformentitlements.Checker](i)
		require.ErrorIs(t, err, platformentitlements.ErrEnforcerRequired)
	})
}
