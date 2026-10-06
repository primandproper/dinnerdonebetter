package main

import (
	"context"
	"errors"
	"fmt"

	dbcleanerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/db_cleaner"
	emaildeliverabilitytestbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/email_deliverability_test"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	emaildeliverabilitytest "github.com/primandproper/dinnerdonebetter/backend/internal/services/email/workers/email_deliverability_test"
	dbcleaner "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/workers/db_cleaner"

	"github.com/primandproper/platform-go/v15/service"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

func jobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Run a one-shot job and exit",
		Args:  cobra.NoArgs,
		RunE:  helpAndFail,
	}

	cmd.AddCommand(
		jobDBCleanerCmd(),
		jobEmailDeliverabilityCmd(),
	)

	return cmd
}

func jobDBCleanerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "db-cleaner",
		Short: "Delete expired OAuth2 tokens and other stale rows",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			config.ConditionallyCease()

			cfg, err := config.LoadConfigFromEnvironment[config.DBCleanerConfig]()
			if err != nil {
				return fmt.Errorf("error getting config: %w", err)
			}
			if cfg.Service.Database != nil {
				cfg.Service.Database.RunMigrations = false
			}

			return runDBCleaner(ctx, cfg)
		},
	}
}

// runDBCleaner composes the database cleaner, runs its one pass, and shuts it down.
//
// The job is not a loop, so the service is never Run: New builds what the config names — a
// database that cannot be reached fails here rather than partway through the sweep — the pass runs
// against it, and Shutdown releases the client and flushes the pillars, so this short-lived
// CronJob pod exports its spans and metrics before it exits.
func runDBCleaner(ctx context.Context, cfg *config.DBCleanerConfig) error {
	i, err := dbcleanerbuild.BuildInjector(ctx, cfg)
	if err != nil {
		return err
	}

	svc, err := service.New(i)
	if err != nil {
		return fmt.Errorf("could not assemble the db cleaner: %w", err)
	}

	jobErr := do.MustInvoke[*dbcleaner.Job](i).Do(ctx)
	if jobErr != nil {
		jobErr = fmt.Errorf("cleaning database: %w", jobErr)
	}

	return errors.Join(
		jobErr,
		svc.Shutdown(context.WithoutCancel(ctx)),
		releaseContainer(ctx, i, cfg.Service.ShutdownTimeout),
	)
}

func jobEmailDeliverabilityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "email-deliverability",
		Short: "Send the periodic deliverability probe email",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			config.ConditionallyCease()

			cfg, err := config.LoadConfigFromEnvironment[config.EmailDeliverabilityTestConfig]()
			if err != nil {
				return fmt.Errorf("error getting config: %w", err)
			}

			i := emaildeliverabilitytestbuild.BuildInjector(ctx, cfg)

			// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
			defer telemetry.Flush(ctx, i)

			if err = do.MustInvoke[*emaildeliverabilitytest.Job](i).Do(ctx); err != nil {
				return fmt.Errorf("running email deliverability test: %w", err)
			}

			return nil
		},
	}
}
