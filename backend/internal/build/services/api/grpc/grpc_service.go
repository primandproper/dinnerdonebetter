package grpcapi

import (
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	auditsvc "github.com/primandproper/platform-go/v14/audit/auditpb"
	"github.com/primandproper/platform-go/v14/authentication/oauth2clients/oauth2clientspb"
	paymentssvc "github.com/primandproper/platform-go/v14/billing/billingpb"
	issuereportssvc "github.com/primandproper/platform-go/v14/issuereports/issuereportspb"
	notificationssvc "github.com/primandproper/platform-go/v14/notifications/notificationspb"
	settingssvc "github.com/primandproper/platform-go/v14/settings/settingspb"
	waitlistssvc "github.com/primandproper/platform-go/v14/waitlists/waitlistspb"
	webhookssvc "github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/server/grpc"
)

type GRPCService struct {
	auditsvc.AuditServiceServer
	authsvc.AuthServiceServer
	dataprivacysvc.DataPrivacyServiceServer
	identitysvc.IdentityServiceServer
	internalopssvc.InternalOperationsServer
	issuereportssvc.IssueReportsServiceServer
	mealplanningsvc.MealPlanningServiceServer
	notificationssvc.NotificationsServiceServer
	oauth2clientspb.OAuth2ClientsServiceServer
	paymentssvc.BillingServiceServer
	settingssvc.SettingsServiceServer
	uploadedmediasvc.UploadedMediaServiceServer
	waitlistssvc.WaitlistsServiceServer
	webhookssvc.WebhooksServiceServer
	*grpc.Server
}

func NewGRPCService(
	auditServiceServer auditsvc.AuditServiceServer,
	authServiceServer authsvc.AuthServiceServer,
	dataPrivacyServiceServer dataprivacysvc.DataPrivacyServiceServer,
	identityServiceServer identitysvc.IdentityServiceServer,
	internalOperationsServer internalopssvc.InternalOperationsServer,
	issueReportsServiceServer issuereportssvc.IssueReportsServiceServer,
	mealPlanningServiceServer mealplanningsvc.MealPlanningServiceServer,
	userNotificationsServiceServer notificationssvc.NotificationsServiceServer,
	oauth2ClientsServiceServer oauth2clientspb.OAuth2ClientsServiceServer,
	paymentsServiceServer paymentssvc.BillingServiceServer,
	settingsServiceServer settingssvc.SettingsServiceServer,
	uploadedMediaServiceServer uploadedmediasvc.UploadedMediaServiceServer,
	webhooksServiceServer webhookssvc.WebhooksServiceServer,
	waitlistsServiceServer waitlistssvc.WaitlistsServiceServer,
	server *grpc.Server,
) *GRPCService {
	return &GRPCService{
		Server:                     server,
		AuditServiceServer:         auditServiceServer,
		AuthServiceServer:          authServiceServer,
		DataPrivacyServiceServer:   dataPrivacyServiceServer,
		IdentityServiceServer:      identityServiceServer,
		InternalOperationsServer:   internalOperationsServer,
		IssueReportsServiceServer:  issueReportsServiceServer,
		MealPlanningServiceServer:  mealPlanningServiceServer,
		NotificationsServiceServer: userNotificationsServiceServer,
		OAuth2ClientsServiceServer: oauth2ClientsServiceServer,
		BillingServiceServer:       paymentsServiceServer,
		SettingsServiceServer:      settingsServiceServer,
		UploadedMediaServiceServer: uploadedMediaServiceServer,
		WebhooksServiceServer:      webhooksServiceServer,
		WaitlistsServiceServer:     waitlistsServiceServer,
	}
}
