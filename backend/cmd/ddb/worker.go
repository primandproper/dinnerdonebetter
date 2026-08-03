package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	datachangemessagehandlerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/functions/data_change_message_handler"
	schedulerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"

	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/outbox"
	"github.com/primandproper/platform-go/v9/saga"
	"github.com/primandproper/platform-go/v9/webhooks"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// schedulerDrainTimeout bounds how long shutdown waits for a job that is mid-execution. It is
// generous because a job killed partway through has already done some of its work and will redo
// it on the next tick.
const schedulerDrainTimeout = 60 * time.Second

func workerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Run a long-lived background worker",
		Args:  cobra.NoArgs,
		RunE:  helpAndFail,
	}

	cmd.AddCommand(
		workerAsyncMessagesCmd(),
		workerSchedulerCmd(),
	)

	return cmd
}

// notifyShutdown returns a channel that receives the first shutdown signal the process is sent.
func notifyShutdown() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	return signalChan
}

func workerAsyncMessagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "async-messages",
		Short: "Consume data change messages and run their handlers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config.ConditionallyCease()

			cfg, err := config.LoadConfigFromEnvironment[config.AsyncMessageHandlerConfig]()
			if err != nil {
				return fmt.Errorf("error getting config: %w", err)
			}
			cfg.Database.RunMigrations = false

			return runAsyncMessages(cmd.Context(), cfg)
		},
	}
}

// runAsyncMessages owns every defer (telemetry flush, drain cancel), so a fatal consume error
// can be returned without skipping them.
func runAsyncMessages(ctx context.Context, cfg *config.AsyncMessageHandlerConfig) error {
	injector := datachangemessagehandlerbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so a short-lived pod exports its spans/metrics before terminating.
	defer telemetry.Flush(ctx, injector)

	dataChangeMessageHandler := do.MustInvoke[*datachangemessagehandler.AsyncDataChangeMessageHandler](injector)

	signalChan := notifyShutdown()

	if err := dataChangeMessageHandler.Start(ctx); err != nil {
		return err
	}

	// Block until the first shutdown signal arrives.
	<-signalChan

	// Close stops each pool's consumer first and only then retires its workers, so a message
	// already being handled finishes. The timeout bounds that drain; past it the pools cancel
	// their handlers so the deferred telemetry flush still runs and the command returns.
	drainCtx, cancelDrain := context.WithTimeout(ctx, 30*time.Second)
	defer cancelDrain()

	return dataChangeMessageHandler.Close(drainCtx)
}

func workerSchedulerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scheduler",
		Short: "Run the interval-shaped periodic jobs, outbox relay, saga, webhook, and data privacy workers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config.ConditionallyCease()

			cfg, err := config.LoadConfigFromEnvironment[config.SchedulerConfig]()
			if err != nil {
				return fmt.Errorf("error getting config: %w", err)
			}
			cfg.Database.RunMigrations = false

			return runScheduler(cmd.Context(), cfg)
		},
	}
}

// runScheduler owns every defer (telemetry flush, scheduler close), so a fatal error can be
// returned without skipping them.
func runScheduler(ctx context.Context, cfg *config.SchedulerConfig) error {
	i := schedulerbuild.BuildInjector(ctx, cfg)

	defer telemetry.Flush(ctx, i)

	scheduler := do.MustInvoke[*jobs.Scheduler](i)
	relay := do.MustInvoke[*outbox.Relay](i)
	sagaWorker := do.MustInvoke[*saga.Worker](i)
	webhookWorker := do.MustInvoke[*webhooks.Worker](i)
	dataPrivacyWorker := do.MustInvoke[*platformdataprivacy.Worker](i)

	signalChan := notifyShutdown()

	// None of these Runs takes a context, on purpose: tied to a server context they would stop
	// mid-job, mid-publish, mid-saga, mid-delivery, and mid-erasure the instant that context
	// was canceled, which is the worst moment to stop. Close is the stop signal, and it lets
	// in-flight work finish.
	go scheduler.Run()
	go relay.Run()
	go sagaWorker.Run()
	go webhookWorker.Run()
	go dataPrivacyWorker.Run()

	<-signalChan

	closeCtx, cancel := context.WithTimeout(ctx, schedulerDrainTimeout)
	defer cancel()

	// All five are closed even if an earlier one fails, so a scheduler that will not drain
	// cannot leave the relay holding claims it is never going to publish, the saga worker
	// holding leases on instances it is never going to advance, the webhook worker holding
	// leases on dispatches it is never going to deliver, or the data privacy worker holding
	// a lease on an erasure it is never going to run.
	//
	// The audit sweeper and the data privacy sweep are not among them: both run as scheduled
	// jobs rather than loops of their own, so the scheduler's own drain is what waits for a
	// sweep in flight.
	return errors.Join(
		wrapClose("scheduler", scheduler.Close(closeCtx)),
		wrapClose("saga worker", sagaWorker.Close(closeCtx)),
		wrapClose("outbox relay", relay.Close(closeCtx)),
		wrapClose("webhook worker", webhookWorker.Close(closeCtx)),
		wrapClose("data privacy worker", dataPrivacyWorker.Close(closeCtx)),
	)
}

func wrapClose(name string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("shutting down %s: %w", name, err)
}
