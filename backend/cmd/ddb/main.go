package main

import (
	"context"
	"fmt"
	"os"
	// Embeds the zoneinfo database. Cron schedules name IANA zones, and both the scheduler's
	// own Timezone and a job's CRON_TZ= prefix are resolved at startup — so without this, a
	// zone the base image happens not to ship is a crash loop rather than a missed run. The
	// image is Debian and does ship one today; this makes the binary not care.
	_ "time/tzdata"

	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"
)

func main() {
	root := &cobra.Command{
		Use:   "ddb",
		Short: "Dinner Done Better",
		Long: `Every Dinner Done Better workload lives in this one binary, selected by subcommand.

Each subcommand loads its own config from the file named by CONFIGURATION_FILEPATH.`,
		// A failing workload is a runtime problem, not a usage problem; printing the whole
		// usage text after it would bury the error that actually matters.
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE:         helpAndFail,
	}

	root.AddCommand(
		serveCmd(),
		workerCmd(),
		jobCmd(),
		migrateCmd(),
		versionCmd(),
	)

	if err := root.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}

// helpAndFail prints a command's help and then fails. Cobra's own behavior for a command that
// only groups subcommands is to print help and exit 0, which as a container entrypoint means a
// manifest naming no workload — or half of one, like `worker` with no worker — reports success
// having done nothing. Every grouping command here uses this instead.
func helpAndFail(cmd *cobra.Command, _ []string) error {
	if err := cmd.Help(); err != nil {
		return err
	}

	return fmt.Errorf("%q requires a subcommand", cmd.CommandPath())
}
