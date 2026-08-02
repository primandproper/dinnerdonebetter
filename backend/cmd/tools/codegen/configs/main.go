package main

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v9/encoding"
)

const (
	defaultHTTPPort = 8000
	defaultGRPCPort = 8001
	maxAttempts     = 50
	otelServiceName = "api_server"

	/* #nosec G101 */
	debugCookieHashKey = "HEREISA32CHARSECRETWHICHISMADEUP"

	// run modes.
	developmentEnv = "development"
	testingEnv     = "testing"

	// Universal Link identifiers, mirrored from the iOS project's DEVELOPMENT_TEAM and
	// PRODUCT_BUNDLE_IDENTIFIER. They are public identifiers, not secrets — the
	// apple-app-site-association document Apple fetches contains both.
	appleTeamID   = "K8R2Q5UWQS"
	appleBundleID = "com.dinnerdonebetter.ios"

	// message provider topics.
	dataChangesTopicName              = "data_changes"
	outboundEmailsTopicName           = "outbound_emails"
	searchIndexRequestsTopicName      = "search_index_requests"
	mobileNotificationsTopicName      = "mobile_notifications"
	userDataAggregationTopicName      = "user_data_aggregation_requests"
	webhookExecutionRequestsTopicName = "webhook_execution_requests"
)

var (
	contentTypeJSON = encoding.ContentTypeJSON.String()
)

func main() {
	// localdev config is generated to two locations:
	// - config_files/ for docker-compose usage
	// - kustomize/configs/ for Kubernetes usage (hostnames overridden via env vars)
	localdevConfig := buildLocalDevConfig()

	envConfigs := map[string]*config.EnvironmentConfigSet{
		"deploy/environments/localdev/config_files": {
			RootConfig: localdevConfig,
		},
		"deploy/environments/localdev/kustomize/configs": {
			RootConfig: localdevConfig,
		},
		"deploy/environments/testing/config_files": {
			APIServiceConfigPath: "integration-tests-config.json",
			RootConfig:           buildIntegrationTestsConfig(),
		},
		"deploy/environments/prod/kustomize/configs": {
			RootConfig: buildProdConfig(),
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
		if err := cfg.Render(p, true, true); err != nil {
			panic(fmt.Errorf("validating config %s: %w", p, err))
		}
	}
}
