package main

import (
	"context"
	"fmt"
	"log"

	searchdataindexscheduler "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/search_data_index_scheduler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v5/search/text/indexing"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.SearchDataIndexSchedulerConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	cfg.Database.RunMigrations = false

	i := searchdataindexscheduler.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
	defer telemetry.Flush(ctx, i)

	scheduler := do.MustInvoke[*indexing.IndexScheduler](i)

	if err = scheduler.IndexTypes(ctx); err != nil {
		return fmt.Errorf("error indexing types: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
