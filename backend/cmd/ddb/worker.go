package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	datachangemessagehandlerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/functions/data_change_message_handler"
	schedulerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v15/service"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

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
			if cfg.Service.Database != nil {
				cfg.Service.Database.RunMigrations = false
			}

			return runAsyncMessages(cmd.Context(), cfg)
		},
	}
}

// runAsyncMessages composes the async message handler from its config and runs it until it is
// signaled.
//
// The handler's pools are subscribed before the service runs, because a subscription the broker
// refuses is a process that would otherwise report healthy and drain nothing. On SIGINT or
// SIGTERM service.Run closes the handler first — each pool's consumer stops and the messages
// already being handled finish — then flushes what they stamped, and only then releases the
// database and the broker.
func runAsyncMessages(ctx context.Context, cfg *config.AsyncMessageHandlerConfig) error {
	i, err := datachangemessagehandlerbuild.BuildInjector(ctx, cfg)
	if err != nil {
		return err
	}

	handler, err := datachangemessagehandlerbuild.NewHandlerRunner(i)
	if err != nil {
		return fmt.Errorf("building the data change message handler: %w", err)
	}

	svc, err := service.New(i, service.WithRunners(handler))
	if err != nil {
		return fmt.Errorf("could not assemble the async message handler: %w", err)
	}

	if err = handler.Start(ctx); err != nil {
		return errors.Join(err, svc.Shutdown(context.WithoutCancel(ctx)))
	}

	runErr := svc.Run(ctx)

	return errors.Join(runErr, releaseContainer(ctx, i, cfg.Service.ShutdownTimeout))
}

// releaseContainer retires the DI container once the service has shut down.
//
// service.Service holds an ordering, not the injector, so whatever the container built that the
// service's own walk does not name is still the container's to release. It runs on a context free
// of the cancellation that ended Run, because draining on a cancelled context cancels every drain
// it is made of.
func releaseContainer(ctx context.Context, i *do.RootScope, timeout time.Duration) error {
	releaseCtx, cancelRelease := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancelRelease()

	if report := i.ShutdownWithContext(releaseCtx); report != nil && !report.Succeed {
		return report
	}

	return nil
}

func workerSchedulerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scheduler",
		Short: "Run the interval-shaped periodic jobs, outbox relay, saga, webhook, operations, and data privacy workers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config.ConditionallyCease()

			cfg, err := config.LoadConfigFromEnvironment[config.SchedulerConfig]()
			if err != nil {
				return fmt.Errorf("error getting config: %w", err)
			}

			if cfg.Service.Database != nil {
				cfg.Service.Database.RunMigrations = false
			}

			return runScheduler(cmd.Context(), cfg)
		},
	}
}

// runScheduler composes the scheduler process from its config and runs it until it is signaled.
//
// service.New assembles every loop the process runs — the job scheduler, the outbox relay, the
// saga, webhook and operations workers — and service.Run starts them and, on SIGINT or SIGTERM,
// takes them down in the order their drains need: the loops in reverse, then the final flushes
// (the operations queue's pending enqueues and the metering flusher's last pass), then the
// clients those flushes write through, then the pillars. That ordering used to be written out
// here by hand, and it had two holes: the operations queue was never flushed, so an enqueue
// batched in the moment before shutdown was lost, and the operations worker was stopped by
// cancelling its context while every other loop was asked to drain.
//
// None of the loops is tied to the command's context, on purpose: tied to it they would stop
// mid-job, mid-publish, mid-saga and mid-delivery the instant it was cancelled, which is the
// worst moment to stop. Run's signal is the stop, and each loop's Close lets in-flight work
// finish inside the config's ShutdownTimeout.
func runScheduler(ctx context.Context, cfg *config.SchedulerConfig) error {
	i, err := schedulerbuild.BuildInjector(ctx, cfg)
	if err != nil {
		return err
	}

	notificationQueue, err := schedulerbuild.NewNotificationQueue(i)
	if err != nil {
		return fmt.Errorf("building the meal plan task notification queue: %w", err)
	}

	svc, err := service.New(i, service.WithRunners(notificationQueue))
	if err != nil {
		return fmt.Errorf("could not assemble the scheduler: %w", err)
	}

	runErr := svc.Run(ctx)

	return errors.Join(runErr, releaseContainer(ctx, i, cfg.Service.ShutdownTimeout))
}
