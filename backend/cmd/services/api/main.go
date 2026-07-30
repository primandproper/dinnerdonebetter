package main

import (
	"context"
	"fmt"
	"os"

	apiserver "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/services/api"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v8/version"

	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"
)

func main() {
	root := &cobra.Command{
		Use:   "server",
		Short: "API server CLI",
	}
	root.AddCommand(serveCmd())
	root.AddCommand(versionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the API server (HTTP + gRPC)",
		RunE:  runServe,
	}
}

func runServe(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	cfg, err := config.LoadConfigFromEnvironment[config.APIServiceConfig]()
	if err != nil {
		return fmt.Errorf("could not load config from environment: %w", err)
	}

	buildCtx, cancel := context.WithTimeout(ctx, cfg.HTTPServer.StartupDeadline)
	defer cancel()

	pillars, err := cfg.Observability.NewPillars(buildCtx)
	if err != nil {
		return fmt.Errorf("could not create observability pillars: %w", err)
	}

	server, err := apiserver.NewServer(buildCtx, pillars, cfg)
	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}

	return server.Run(ctx)
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info as JSON (commit hash and build times)",
		RunE:  runVersion,
	}
}

func runVersion(_ *cobra.Command, _ []string) error {
	return version.WriteJSONToStdout()
}
