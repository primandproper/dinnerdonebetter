package grpcapi

import (
	auditsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	"github.com/primandproper/platform-go/v9/server/grpc"
)

type GRPCService struct {
	auditsvc.AuditServiceServer
	authsvc.AuthServiceServer
	dataprivacysvc.DataPrivacyServiceServer
	identitysvc.IdentityServiceServer
	internalopssvc.InternalOperationsServer
	issuereportssvc.IssueReportsServiceServer
	mealplanningsvc.MealPlanningServiceServer
	notificationssvc.UserNotificationsServiceServer
	oauthsvc.OAuthServiceServer
	paymentssvc.PaymentsServiceServer
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
	userNotificationsServiceServer notificationssvc.UserNotificationsServiceServer,
	oauthServiceServer oauthsvc.OAuthServiceServer,
	paymentsServiceServer paymentssvc.PaymentsServiceServer,
	settingsServiceServer settingssvc.SettingsServiceServer,
	uploadedMediaServiceServer uploadedmediasvc.UploadedMediaServiceServer,
	webhooksServiceServer webhookssvc.WebhooksServiceServer,
	waitlistsServiceServer waitlistssvc.WaitlistsServiceServer,
	server *grpc.Server,
) *GRPCService {
	return &GRPCService{
		Server:                         server,
		AuditServiceServer:             auditServiceServer,
		AuthServiceServer:              authServiceServer,
		DataPrivacyServiceServer:       dataPrivacyServiceServer,
		IdentityServiceServer:          identityServiceServer,
		InternalOperationsServer:       internalOperationsServer,
		IssueReportsServiceServer:      issueReportsServiceServer,
		MealPlanningServiceServer:      mealPlanningServiceServer,
		UserNotificationsServiceServer: userNotificationsServiceServer,
		OAuthServiceServer:             oauthServiceServer,
		PaymentsServiceServer:          paymentsServiceServer,
		SettingsServiceServer:          settingsServiceServer,
		UploadedMediaServiceServer:     uploadedMediaServiceServer,
		WebhooksServiceServer:          webhooksServiceServer,
		WaitlistsServiceServer:         waitlistsServiceServer,
	}
}
