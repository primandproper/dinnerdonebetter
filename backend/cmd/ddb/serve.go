package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	apiserver "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcpserver"

	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the API server (HTTP + gRPC)",
		Args:  cobra.NoArgs,
		RunE:  runServe,
	}

	cmd.AddCommand(serveMCPCmd())

	return cmd
}

func runServe(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

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

func serveMCPCmd() *cobra.Command {
	var (
		transport string
		baseURL   string
	)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run the MCP server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// MCP_BASE_URL outranks the flag, as it did when this was its own binary: the
			// deployment sets the env var, and the flag default is only ever right for localdev.
			if envBase := os.Getenv("MCP_BASE_URL"); envBase != "" {
				baseURL = envBase
			}

			return mcpserver.Run(cmd.Context(), transport, baseURL)
		},
	}

	cmd.Flags().StringVar(&transport, "transport", mcpserver.TransportHTTP, fmt.Sprintf("Transport method: one of %s", strings.Join(mcpserver.ValidTransports(), ", ")))
	cmd.Flags().StringVar(&baseURL, "base-url", mcpserver.DefaultBaseURL, "Public base URL of the MCP server (used for OAuth2 metadata)")

	return cmd
}
