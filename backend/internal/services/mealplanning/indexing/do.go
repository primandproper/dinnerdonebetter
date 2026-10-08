package indexing

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"

	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	syncsource "github.com/primandproper/platform-go/v15/searchsync/source"
	textsearch "github.com/primandproper/primitives-go/v2/search/text"

	"github.com/samber/do/v2"
)

var _ searchindexes.Registrar = RegisterIndexes

// RegisterIndexes adds all eight meal planning indexes to registry, each with its Syncer, its
// Reindexer and the stamp buffer behind the Syncer. It is a searchindexes.Registrar.
//
// Which of them a given process actually runs is still that process's business — the consumer
// drains the Syncers' topics, the scheduler runs the Reindexers — but they are registered
// together, because a Syncer with no Reindexer has no way back from an index that was already
// wrong, and a Reindexer with no Syncer is the sampler this replaced, just on a longer timer.
func RegisterIndexes(i do.Injector, registry *searchsync.Registry) error {
	repo := do.MustInvoke[mealplanning.Repository](i)

	for _, register := range []func() error{
		func() error {
			return registerIndex(registry, NewMealSource, repo, do.MustInvoke[searchindex.MealTextSearcher](i), repo.MarkMealsAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewRecipeSource, repo, do.MustInvoke[searchindex.RecipeTextSearcher](i), repo.MarkRecipesAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewValidIngredientSource, repo, do.MustInvoke[searchindex.ValidIngredientTextSearcher](i), repo.MarkValidIngredientsAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewValidInstrumentSource, repo, do.MustInvoke[searchindex.ValidInstrumentTextSearcher](i), repo.MarkValidInstrumentsAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewValidMeasurementUnitSource, repo, do.MustInvoke[searchindex.ValidMeasurementUnitTextSearcher](i), repo.MarkValidMeasurementUnitsAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewValidPreparationSource, repo, do.MustInvoke[searchindex.ValidPreparationTextSearcher](i), repo.MarkValidPreparationsAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewValidIngredientStateSource, repo, do.MustInvoke[searchindex.ValidIngredientStateTextSearcher](i), repo.MarkValidIngredientStatesAsIndexed)
		},
		func() error {
			return registerIndex(registry, NewValidVesselSource, repo, do.MustInvoke[searchindex.ValidVesselTextSearcher](i), repo.MarkValidVesselsAsIndexed)
		},
	} {
		if err := register(); err != nil {
			return err
		}
	}

	return nil
}

// registerIndex builds one entity's source over the repository and registers its text index.
func registerIndex[E, T any](
	registry *searchsync.Registry,
	source func(mealplanning.Repository) (*syncsource.Source[E, T], error),
	repo mealplanning.Repository,
	index textsearch.IndexManager,
	stamp func(ctx context.Context, ids []string) error,
) error {
	src, err := source(repo)
	if err != nil {
		return err
	}

	return searchindexes.RegisterTextIndex(registry, src, index, stamp)
}
