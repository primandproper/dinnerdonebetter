package authorization

import (
	webhooksgrpc "github.com/primandproper/platform-go/v14/webhooks/grpc"
)

// The webhook surface's permissions are platform's, re-exported under this
// application's names so that the roles read the way the rest of them do.
//
// They are finer than the ten this repository's own webhooks service used, and
// deliberately so: an endpoint, a subscription and a secret rotation are
// separately grantable now, where "update.webhooks" covered all three. The
// trigger-config and trigger-event permissions have no successor at all — those
// tables went with the store migration, and what a subscriber hears about is a
// subscription to an event type rather than a config row.
const (
	// SaveWebhookEndpointsPermission allows creating or rewriting an endpoint.
	SaveWebhookEndpointsPermission = webhooksgrpc.PermissionSaveEndpoints
	// ReadWebhookEndpointsPermission allows reading an account's endpoints.
	ReadWebhookEndpointsPermission = webhooksgrpc.PermissionReadEndpoints
	// ArchiveWebhookEndpointsPermission allows retiring an endpoint.
	ArchiveWebhookEndpointsPermission = webhooksgrpc.PermissionArchiveEndpoints

	// RotateWebhookSecretPermission allows minting an endpoint a new signing
	// secret. Separate from saving the endpoint because it is the one write that
	// invalidates every signature a subscriber has already learned to check.
	RotateWebhookSecretPermission = webhooksgrpc.PermissionRotateEndpointSecrets

	// AddWebhookSubscriptionsPermission allows subscribing an endpoint to an event type.
	AddWebhookSubscriptionsPermission = webhooksgrpc.PermissionAddSubscriptions
	// ReadWebhookSubscriptionsPermission allows reading what an endpoint hears about.
	ReadWebhookSubscriptionsPermission = webhooksgrpc.PermissionReadSubscriptions
	// ArchiveWebhookSubscriptionsPermission allows unsubscribing an endpoint.
	ArchiveWebhookSubscriptionsPermission = webhooksgrpc.PermissionArchiveSubscriptions

	// ReadWebhookAttemptsPermission allows reading the delivery log, which is how
	// a subscriber finds out why they never heard.
	ReadWebhookAttemptsPermission = webhooksgrpc.PermissionReadAttempts

	// ReadWebhookEventTypesPermission allows reading the catalog of what can be
	// subscribed to. It is held by anybody who may manage a subscription, because
	// a subscription cannot be made without it.
	ReadWebhookEventTypesPermission = webhooksgrpc.PermissionReadEventTypes
)

var (
	// WebhooksPermissions contains all webhook-related permissions.
	WebhooksPermissions = []Permission{
		SaveWebhookEndpointsPermission,
		ReadWebhookEndpointsPermission,
		ArchiveWebhookEndpointsPermission,
		RotateWebhookSecretPermission,
		AddWebhookSubscriptionsPermission,
		ReadWebhookSubscriptionsPermission,
		ArchiveWebhookSubscriptionsPermission,
		ReadWebhookAttemptsPermission,
		ReadWebhookEventTypesPermission,
	}
)
