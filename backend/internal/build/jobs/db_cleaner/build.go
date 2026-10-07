package dbcleaner

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	dbcleaner "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/workers/db_cleaner"

	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"
	"github.com/primandproper/platform-go/v15/service"

	"github.com/samber/do/v2"
)

// BuildInjector validates cfg and composes the database cleaner from it.
//
// service.Register builds the pillars and the database from cfg.Service; the stores the job
// sweeps, and the job, are registered after it. Validation comes first for the reason the
// scheduler's BuildInjector gives.
func BuildInjector(
	ctx context.Context,
	cfg *config.DBCleanerConfig,
) (*do.RootScope, error) {
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating db cleaner config: %w", err)
	}

	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	service.Register(i, &cfg.Service)

	RegisterConfigs(i)

	internalops.RegisterInternalOpsRepository(i)
	oauth2servercfg.RegisterStore(i)
	authrepo.RegisterPasswordResetTokenSQLStore(i)
	authrepo.RegisterRefreshTokenSQLStore(i)
	authrepo.RegisterSignInDevicesSQLStore(i)
	dbcleaner.RegisterDBCleaner(i)

	return i, nil
}
