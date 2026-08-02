package main

import (
	"context"
	"fmt"
	"log"

	emaildeliverabilitytestbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/email_deliverability_test"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	emaildeliverabilitytest "github.com/primandproper/dinnerdonebetter/backend/internal/services/email/workers/email_deliverability_test"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.EmailDeliverabilityTestConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	i := emaildeliverabilitytestbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
	defer telemetry.Flush(ctx, i)

	job := do.MustInvoke[*emaildeliverabilitytest.Job](i)

	if err = job.Do(ctx); err != nil {
		return fmt.Errorf("running email deliverability test: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
