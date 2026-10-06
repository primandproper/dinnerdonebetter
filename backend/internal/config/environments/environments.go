/*
Package environments holds the configuration every environment this application is deployed
to actually ships with — localdev, the integration test environment, and prod.

They live here rather than in the generator that writes them out so that they are reachable
from a test. A test that needs a valid configuration set should get one from these builders,
because then the config under test is valid by virtue of being the one that is shipped, rather
than a fixture in a second place that has to be kept valid as the config structs grow.
*/
package environments

import (
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/primitives-go/v2/encoding"
	"github.com/primandproper/primitives-go/v2/httpclient"
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
	dataChangesTopicName         = queuescfg.DefaultDataChangesTopicName
	outboundEmailsTopicName      = "outbound_emails"
	searchIndexRequestsTopicName = "search_index_requests"
	mobileNotificationsTopicName = "mobile_notifications"
)

var (
	contentTypeJSON = encoding.ContentTypeJSON.String()
)

// defaultHTTPClientConfig returns the outbound HTTP client every environment's API server builds
// its third-party clients over — the feature flag provider's among them.
//
// primitives' own defaults, asked for rather than copied, and written out rather than left
// zero: a block with nothing in it is one service.Register treats as unconfigured, and an API
// server with no HTTP client is one whose feature flag manager cannot be built.
func defaultHTTPClientConfig() *httpclient.Config {
	cfg := &httpclient.Config{EnableTracing: true}
	cfg.EnsureDefaults()

	return cfg
}
