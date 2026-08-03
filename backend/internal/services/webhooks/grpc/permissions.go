package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
)

// WebhooksMethodPermissions is a named type for Wire dependency injection.
type WebhooksMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the webhooks service's method permissions.
func ProvideMethodPermissions() WebhooksMethodPermissions {
	return WebhooksMethodPermissions{
		webhookssvc.WebhooksService_GetWebhook_FullMethodName: {
			authorization.ReadWebhooksPermission,
		},
		webhookssvc.WebhooksService_GetWebhooks_FullMethodName: {
			authorization.ReadWebhooksPermission,
		},
		webhookssvc.WebhooksService_CreateWebhook_FullMethodName: {
			authorization.CreateWebhooksPermission,
		},
		webhookssvc.WebhooksService_ArchiveWebhook_FullMethodName: {
			authorization.ArchiveWebhooksPermission,
		},
		webhookssvc.WebhooksService_AddWebhookTriggerConfig_FullMethodName: {
			authorization.CreateWebhookTriggerConfigsPermission,
		},
		webhookssvc.WebhooksService_ArchiveWebhookTriggerConfig_FullMethodName: {
			authorization.ArchiveWebhookTriggerConfigsPermission,
		},
		// Rotating a signing secret invalidates the old one for every delivery once the
		// window closes, so it is gated as a webhook mutation rather than a read.
		webhookssvc.WebhooksService_RotateWebhookSecret_FullMethodName: {
			authorization.UpdateWebhooksPermission,
		},
		// The catalog is generated Go, identical for every account, and names no account's
		// data — so reading it needs only the permission that lets someone see webhooks at
		// all, and there is no create/update/archive counterpart to gate.
		webhookssvc.WebhooksService_GetWebhookEventTypes_FullMethodName: {
			authorization.ReadWebhooksPermission,
		},
	}
}
