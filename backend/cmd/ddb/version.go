package main

import (
	"github.com/primandproper/platform-go/v12/version"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info as JSON (commit hash and build times)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return version.WriteJSONToStdout()
		},
	}
}
