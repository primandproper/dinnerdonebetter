package email

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMealPlanCreatedEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = new(time.Now())
		mealPlan := mealplanningfakes.BuildFakeMealPlan()

		actual, err := BuildMealPlanCreatedEmail(user, mealPlan, "https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Contains(t, actual.HTMLContent, branding.LogoURL)
	})
}
