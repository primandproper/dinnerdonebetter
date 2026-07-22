package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mobilenotificationscheduler "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v5/observability/metrics"
	"github.com/primandproper/platform-go/v5/observability/tracing"

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
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if shutdownErr := do.MustInvoke[metrics.Provider](i).Shutdown(shutdownCtx); shutdownErr != nil {
			log.Printf("error shutting down metrics: %v", shutdownErr)
		}
		if flushErr := do.MustInvoke[tracing.TracerProvider](i).ForceFlush(shutdownCtx); flushErr != nil {
			log.Printf("error flushing traces: %v", flushErr)
		}
	}()

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
