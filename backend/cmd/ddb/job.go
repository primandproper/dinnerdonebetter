package main

import (
	"fmt"

	dbcleanerbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/db_cleaner"
	emaildeliverabilitytestbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/email_deliverability_test"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	emaildeliverabilitytest "github.com/primandproper/dinnerdonebetter/backend/internal/services/email/workers/email_deliverability_test"
	dbcleaner "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/workers/db_cleaner"

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
			cfg.Database.RunMigrations = false

			i := dbcleanerbuild.BuildInjector(ctx, cfg)

			// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
			defer telemetry.Flush(ctx, i)

			if err = do.MustInvoke[*dbcleaner.Job](i).Do(ctx); err != nil {
				return fmt.Errorf("cleaning database: %w", err)
			}

			return nil
		},
	}
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
