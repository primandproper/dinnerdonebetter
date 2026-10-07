package payments

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
)

// TestMain starts the one postgres container this package's webhook tests share and
// migrates the template database each of them is cloned from. Without
// RUN_CONTAINER_TESTS=true it starts nothing, the container-backed tests skip themselves,
// and the rest run as they always did. See pgtesting.RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}
