package recipeanalysis

import (
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterRecipeAnalyzer registers the recipe analyzer with the injector.
func RegisterRecipeAnalyzer(i do.Injector) {
	do.Provide[RecipeAnalyzer](i, func(i do.Injector) (RecipeAnalyzer, error) {
		return NewRecipeAnalyzer(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
		), nil
	})
}
