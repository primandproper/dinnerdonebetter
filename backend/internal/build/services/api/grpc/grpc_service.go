package grpcapi

import (
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	auditsvc "github.com/primandproper/platform-go/v15/audit/auditpb"
	"github.com/primandproper/platform-go/v15/authentication/oauth2clients/oauth2clientspb"
	paymentssvc "github.com/primandproper/platform-go/v15/billing/billingpb"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	issuereportssvc "github.com/primandproper/platform-go/v15/issuereports/issuereportspb"
	notificationssvc "github.com/primandproper/platform-go/v15/notifications/notificationspb"
	settingssvc "github.com/primandproper/platform-go/v15/settings/settingspb"
	waitlistssvc "github.com/primandproper/platform-go/v15/waitlists/waitlistspb"
	webhookssvc "github.com/primandproper/platform-go/v15/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/server/grpc"
)

type GRPCService struct {
	auditsvc.AuditServiceServer
	identitypb.IdentityServiceServer
	internalopssvc.InternalOperationsServer
	issuereportssvc.IssueReportsServiceServer
	mealplanningsvc.MealPlanningServiceServer
	notificationssvc.NotificationsServiceServer
	oauth2clientspb.OAuth2ClientsServiceServer
	paymentssvc.BillingServiceServer
	settingssvc.SettingsServiceServer
	waitlistssvc.WaitlistsServiceServer
	webhookssvc.WebhooksServiceServer
	*grpc.Server
}

func NewGRPCService(
	auditServiceServer auditsvc.AuditServiceServer,
	identityServiceServer identitypb.IdentityServiceServer,
	internalOperationsServer internalopssvc.InternalOperationsServer,
	issueReportsServiceServer issuereportssvc.IssueReportsServiceServer,
	mealPlanningServiceServer mealplanningsvc.MealPlanningServiceServer,
	userNotificationsServiceServer notificationssvc.NotificationsServiceServer,
	oauth2ClientsServiceServer oauth2clientspb.OAuth2ClientsServiceServer,
	paymentsServiceServer paymentssvc.BillingServiceServer,
	settingsServiceServer settingssvc.SettingsServiceServer,
	webhooksServiceServer webhookssvc.WebhooksServiceServer,
	waitlistsServiceServer waitlistssvc.WaitlistsServiceServer,
	server *grpc.Server,
) *GRPCService {
	return &GRPCService{
		Server:                     server,
		AuditServiceServer:         auditServiceServer,
		IdentityServiceServer:      identityServiceServer,
		InternalOperationsServer:   internalOperationsServer,
		IssueReportsServiceServer:  issueReportsServiceServer,
		MealPlanningServiceServer:  mealPlanningServiceServer,
		NotificationsServiceServer: userNotificationsServiceServer,
		OAuth2ClientsServiceServer: oauth2ClientsServiceServer,
		BillingServiceServer:       paymentsServiceServer,
		SettingsServiceServer:      settingsServiceServer,
		WebhooksServiceServer:      webhooksServiceServer,
		WaitlistsServiceServer:     waitlistsServiceServer,
	}
}
