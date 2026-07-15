package migrations

import (
	"testing"
	"time"

	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	loggingnoop "github.com/primandproper/platform-go/v4/observability/logging/noop"
	"github.com/primandproper/platform-go/v4/pointer"

	"github.com/stretchr/testify/require"
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		container, db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		require.NoError(t, NewMigrator(loggingnoop.NewLogger()).Migrate(ctx, db))

		if err := container.Stop(ctx, pointer.To(time.Second*10)); err != nil {
			t.Log(err)
		}
	})
}
