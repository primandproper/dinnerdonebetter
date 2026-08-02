package migrations

import (
	"testing"

	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"

	"github.com/stretchr/testify/require"
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))
	})
}
