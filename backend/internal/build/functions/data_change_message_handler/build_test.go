package datachangemessagehandler

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInjector_RegistersAllProviders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &config.AsyncMessageHandlerConfig{}

	i := BuildInjector(ctx, cfg)

	services := i.ListProvidedServices()
	assert.NotEmpty(t, services, "expected providers to be registered")
	assert.Greater(t, len(services), 10, "expected many providers to be registered")
}

// TestBuildInjector_RegistersAStamperPerIndex guards the one thing about the search sync
// wiring that a compiler cannot: the Stamper each Syncer writes last_indexed_at through is
// registered under its index's name, and this process resolves it by that name in order to
// close it. A name registered in one place and looked up in another is a runtime panic during
// startup, which is exactly the failure a build test is for.
func TestBuildInjector_RegistersAStamperPerIndex(t *testing.T) {
	t.Parallel()

	i := BuildInjector(context.Background(), &config.AsyncMessageHandlerConfig{})

	registered := map[string]struct{}{}
	for _, service := range i.ListProvidedServices() {
		registered[service.Service] = struct{}{}
	}

	for _, index := range []string{
		identityindexing.IndexTypeUsers,
		mealplanningindexing.IndexTypeMeals,
		mealplanningindexing.IndexTypeRecipes,
		mealplanningindexing.IndexTypeValidIngredients,
		mealplanningindexing.IndexTypeValidInstruments,
		mealplanningindexing.IndexTypeValidMeasurementUnits,
		mealplanningindexing.IndexTypeValidPreparations,
		mealplanningindexing.IndexTypeValidIngredientStates,
		mealplanningindexing.IndexTypeValidVessels,
	} {
		assert.Contains(t, registered, index, "no stamper registered for the %s index", index)
	}
}
