package main

import (
	"context"
	"fmt"
	"log"

	dbcleanerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/db_cleaner"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcleaner "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/workers/db_cleaner"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.DBCleanerConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	cfg.Database.RunMigrations = false

	i := dbcleanerbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
	defer telemetry.Flush(ctx, i)

	dbCleaner := do.MustInvoke[*dbcleaner.Job](i)

	if err = dbCleaner.Do(ctx); err != nil {
		return fmt.Errorf("cleaning database: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
