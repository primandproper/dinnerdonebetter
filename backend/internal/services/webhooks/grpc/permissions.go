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
		webhookssvc.WebhooksService_ArchiveWebhookTriggerEvent_FullMethodName: {
			authorization.ArchiveWebhookTriggerEventsPermission,
		},
		webhookssvc.WebhooksService_CreateWebhookTriggerEvent_FullMethodName: {
			authorization.CreateWebhookTriggerEventsPermission,
		},
		webhookssvc.WebhooksService_GetWebhookTriggerEvent_FullMethodName: {
			authorization.ReadWebhookTriggerEventsPermission,
		},
		webhookssvc.WebhooksService_GetWebhookTriggerEvents_FullMethodName: {
			authorization.ReadWebhookTriggerEventsPermission,
		},
		webhookssvc.WebhooksService_UpdateWebhookTriggerEvent_FullMethodName: {
			authorization.UpdateWebhookTriggerEventsPermission,
		},
	}
}
