package main

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config/environments"
)

func main() {
	ctx := context.Background()

	// localdev config is generated to two locations:
	// - config_files/ for docker-compose usage
	// - kustomize/configs/ for Kubernetes usage (hostnames overridden via env vars)
	localdevConfig := environments.BuildLocalDevConfig()

	envConfigs := map[string]*config.EnvironmentConfigSet{
		"deploy/environments/localdev/config_files": {
			RootConfig: localdevConfig,
		},
		"deploy/environments/localdev/kustomize/configs": {
			RootConfig: localdevConfig,
		},
		"deploy/environments/testing/config_files": {
			APIServiceConfigPath: "integration-tests-config.json",
			RootConfig:           environments.BuildIntegrationTestsConfig(),
		},
		"deploy/environments/prod/kustomize/configs": {
			RootConfig: environments.BuildProdConfig(),
			ServiceDatabaseUsers: map[string]string{
				"db_cleaner": "db_cleaner",
				// The six interval jobs that used to be one CronJob (and one database user)
				// each now share a process, so they share a user with the union of their
				// grants. Least privilege between them is gone; least privilege against the
				// API server's user is not.
				"scheduler":                     "scheduler",
				"async_message_handler":         "async_message_handler",
				"dinner_done_better_mcp_server": "mcp_server",
			},
		},
	}

	for p, cfg := range envConfigs {
		if err := cfg.Render(ctx, p); err != nil {
			panic(fmt.Errorf("rendering config %s: %w", p, err))
		}
	}
}
