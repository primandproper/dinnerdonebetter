/*
Package webhooks mounts platform-go's webhooks surface.

Eleven RPCs, and no service of this application's own. Ten of them were renames
of the deleted service's — CreateWebhook is SaveEndpoint, AddWebhookTriggerConfig
is AddSubscription — and the eleventh, ListEventTypes, is the one the deleted
service had that platform did not: the catalog a subscription is judged against,
which a client building a subscription form cannot ask for anywhere else.

Endpoints belong to an account, so this takes the account-scoped principal. See
sessions.AccountScopedPrincipal for why handing it the global one would be quiet
and wrong.
*/
package webhooks

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformwebhooks "github.com/primandproper/platform-go/v14/webhooks"
	webhooksgrpc "github.com/primandproper/platform-go/v14/webhooks/grpc"
	"github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWebhooksService registers platform's webhooks surface with the injector.
func RegisterWebhooksService(i do.Injector) {
	do.Provide[webhookspb.WebhooksServiceServer](i, func(i do.Injector) (webhookspb.WebhooksServiceServer, error) {
		return webhooksgrpc.NewServer(
			do.MustInvoke[platformwebhooks.Dispatcher](i),
			do.MustInvoke[platformwebhooks.Store](i),
			do.MustInvoke[database.Client](i),
			// Account-scoped: an endpoint is its account's.
			sessions.AccountScopedPrincipalFromContext,
			// Gates include_archived, which this surface gained with billing and
			// notifications. Retired endpoint URLs are the delivery targets a
			// deployment stopped trusting, so the grant matters here more than most.
			webhooksgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			webhooksgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			webhooksgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			webhooksgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
