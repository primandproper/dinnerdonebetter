package main

import (
	"context"
	"fmt"
	"log"

	mobilenotificationscheduler "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.MobileNotificationSchedulerConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	cfg.Database.RunMigrations = false

	i := mobilenotificationscheduler.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
	defer telemetry.Flush(ctx, i)

	scheduler := do.MustInvoke[*mobilenotificationscheduler.Scheduler](i)

	if err = scheduler.ScheduleNotifications(ctx); err != nil {
		return fmt.Errorf("error scheduling notifications: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
