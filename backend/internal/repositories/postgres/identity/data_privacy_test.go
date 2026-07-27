package identity

import (
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
)

func TestQuerier_Integration_DataPrivacy(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = fmt.Sprintf("%d", pgtesting.HashStringToNumberForTest(t, exampleUser.Username))
	exampleUser.TwoFactorSecretVerifiedAt = nil

	// create
	createdUser := createUserForTest(t, ctx, exampleUser, dbc)

	assert.NoError(t, dbc.DeleteUser(ctx, createdUser.ID))
}

func TestQuerier_DeleteUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.DeleteUser(ctx, ""))
	})
}
