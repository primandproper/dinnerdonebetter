package comments

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	platformcomments "github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithExistenceCheck(T *testing.T) {
	T.Parallel()

	T.Run("a read that succeeds is present", func(t *testing.T) {
		t.Parallel()

		var seen string
		definition := WithExistenceCheck(platformcomments.TargetDefinition{Description: "A thing."}, func(_ context.Context, targetID string) error {
			seen = targetID

			return nil
		})

		exists, err := definition.Exists(t.Context(), tenancy.Global(), t.Name())
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, t.Name(), seen)
		assert.Equal(t, "A thing.", definition.Description, "the description is kept")
	})

	T.Run("a read that finds no row is absent, not an error", func(t *testing.T) {
		t.Parallel()

		definition := WithExistenceCheck(platformcomments.TargetDefinition{}, func(context.Context, string) error {
			return sql.ErrNoRows
		})

		exists, err := definition.Exists(t.Context(), tenancy.Global(), t.Name())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	T.Run("any other failure is an error, not absence", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("the table is unreachable")
		definition := WithExistenceCheck(platformcomments.TargetDefinition{}, func(context.Context, string) error {
			return boom
		})

		exists, err := definition.Exists(t.Context(), tenancy.Global(), t.Name())
		require.ErrorIs(t, err, boom)
		assert.False(t, exists)
	})
}
