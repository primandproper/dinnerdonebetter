package converters_test

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/stretchr/testify/assert"
)

func TestConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		membership := fakes.BuildFakeAccountUserMembership()

		actual := converters.ConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput(membership)

		assert.NotEmpty(t, actual.ID)
		assert.Equal(t, membership.BelongsToUser, actual.UserID)
		// the membership's account, not the membership's own ID.
		assert.Equal(t, membership.BelongsToAccount, actual.AccountID)
		assert.NotEqual(t, membership.ID, actual.AccountID)
	})
}
