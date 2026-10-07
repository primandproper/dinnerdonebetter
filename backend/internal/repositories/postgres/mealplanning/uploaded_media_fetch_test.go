package mealplanning

import (
	"testing"

	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerier_Integration_GetUploadedMediaWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("answers in the order asked, without the ids that name nothing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dbc, _ := buildDatabaseClientForTest(t)

		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

		recorded := make([]*mediaregistry.Object, 3)
		for i := range recorded {
			objectID := identifiers.New()

			require.NoError(t, dbc.WithTransaction(ctx, func(tx database.Tx) error {
				var err error
				recorded[i], err = dbc.uploads.RecordObject(ctx, tx, tenancy.Global(), mediaregistry.ObjectInput{
					ID:          objectID,
					Key:         user.ID + "/" + objectID + "/" + fake.BuildFakeID() + ".png",
					ContentType: "image/png",
					OwnerID:     user.ID,
					Size:        int64(i + 1),
				})

				return err
			}))
		}

		// Out of id order, with an id that names nothing in the middle: the result follows
		// the ids as given and leaves the missing one out.
		actual, err := dbc.GetUploadedMediaWithIDs(ctx, []string{recorded[2].ID, identifiers.New(), recorded[0].ID, recorded[1].ID})
		require.NoError(t, err)

		require.Len(t, actual, 3)
		assert.Equal(t, recorded[2].ID, actual[0].ID)
		assert.Equal(t, recorded[0].ID, actual[1].ID)
		assert.Equal(t, recorded[1].ID, actual[2].ID)
	})

	T.Run("with no ids", func(t *testing.T) {
		t.Parallel()

		dbc, _ := buildDatabaseClientForTest(t)

		actual, err := dbc.GetUploadedMediaWithIDs(t.Context(), nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}
