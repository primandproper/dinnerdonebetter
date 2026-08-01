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

	schedulerbuild "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/scheduler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v8/jobs"
	"github.com/primandproper/platform-go/v8/outbox"

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	// Neither Run takes a context, on purpose: tied to a server context they would stop
	// mid-job and mid-publish the instant that context was canceled, which is the worst
	// moment to stop. Close is the stop signal, and it lets in-flight work finish.
	go scheduler.Run()
	go relay.Run()

	<-signalChan

	closeCtx, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()

	// Both are closed even if the first fails, so a scheduler that will not drain cannot
	// leave the relay holding claims it is never going to publish.
	return errors.Join(
		wrapClose("scheduler", scheduler.Close(closeCtx)),
		wrapClose("outbox relay", relay.Close(closeCtx)),
	)
}

func wrapClose(name string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("shutting down %s: %w", name, err)
}
