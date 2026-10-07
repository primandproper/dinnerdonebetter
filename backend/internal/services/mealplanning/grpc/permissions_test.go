package grpc

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/stretchr/testify/assert"
)

func TestProvideMethodPermissions(T *testing.T) {
	T.Parallel()

	T.Run("GetRecipeLists requires the recipe list read grant", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t,
			[]authorization.Permission{authorization.ReadRecipeListsPermission},
			ProvideMethodPermissions()[mealplanningsvc.MealPlanningService_GetRecipeLists_FullMethodName],
		)
	})
}
