package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	// Embeds the zoneinfo database. Cron schedules name IANA zones, and both the scheduler's
	// own Timezone and a job's CRON_TZ= prefix are resolved at startup — so without this, a
	// zone the base image happens not to ship is a crash loop rather than a missed run. The
	// image is Debian and does ship one today; this makes the binary not care.
	_ "time/tzdata"

	schedulerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/outbox"
	"github.com/primandproper/platform-go/v9/saga"
	"github.com/primandproper/platform-go/v9/webhooks"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

// drainTimeout bounds how long shutdown waits for a job that is mid-execution. It is generous
// because a job killed partway through has already done some of its work and will redo it on
// the next tick.
const drainTimeout = 60 * time.Second

func main() {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.SchedulerConfig]()
	if err != nil {
		log.Fatalf("error getting config: %v", err)
	}
	cfg.Database.RunMigrations = false

	// run owns every defer (telemetry flush, scheduler close), so a fatal error can exit
	// non-zero here without skipping them.
	if err = run(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg *config.SchedulerConfig) error {
	i := schedulerbuild.BuildInjector(ctx, cfg)

	defer telemetry.Flush(ctx, i)

	scheduler := do.MustInvoke[*jobs.Scheduler](i)
	relay := do.MustInvoke[*outbox.Relay](i)
	sagaWorker := do.MustInvoke[*saga.Worker](i)
	webhookWorker := do.MustInvoke[*webhooks.Worker](i)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	// None of these Runs takes a context, on purpose: tied to a server context they would stop
	// mid-job, mid-publish, mid-saga, and mid-delivery the instant that context was canceled,
	// which is the worst moment to stop. Close is the stop signal, and it lets in-flight work
	// finish.
	go scheduler.Run()
	go relay.Run()
	go sagaWorker.Run()
	go webhookWorker.Run()

	<-signalChan

	closeCtx, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()

	// All four are closed even if an earlier one fails, so a scheduler that will not drain
	// cannot leave the relay holding claims it is never going to publish, the saga worker
	// holding leases on instances it is never going to advance, or the webhook worker holding
	// leases on dispatches it is never going to deliver.
	return errors.Join(
		wrapClose("scheduler", scheduler.Close(closeCtx)),
		wrapClose("saga worker", sagaWorker.Close(closeCtx)),
		wrapClose("outbox relay", relay.Close(closeCtx)),
		wrapClose("webhook worker", webhookWorker.Close(closeCtx)),
	)
}

func wrapClose(name string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("shutting down %s: %w", name, err)
}
