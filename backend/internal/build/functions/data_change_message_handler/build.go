package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	settingsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/manager"
	waitlistsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issue_reports "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	analyticscfg "github.com/primandproper/platform-go/v11/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v11/database/config"
	"github.com/primandproper/platform-go/v11/database/postgres"
	emailcfg "github.com/primandproper/platform-go/v11/email/config"
	"github.com/primandproper/platform-go/v11/encoding"
	"github.com/primandproper/platform-go/v11/httpclient"
	msgconfig "github.com/primandproper/platform-go/v11/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v11/notifications/mobile/config"
	"github.com/primandproper/platform-go/v11/observability"
	loggingcfg "github.com/primandproper/platform-go/v11/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v11/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v11/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.AsyncMessageHandlerConfig,
) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	// config field extraction
	RegisterConfigs(i)

	// platform providers
	observability.RegisterO11yConfigs(i)
	tracingcfg.RegisterTracerProvider(i)
	loggingcfg.RegisterLogger(i)
	metricscfg.RegisterMetricsProvider(i)
	msgconfig.RegisterMessageQueue(i)
	httpclient.RegisterHTTPClient(i)
	encoding.RegisterServerEncoderDecoder(i)
	analyticscfg.RegisterEventReporter(i)
	emailcfg.RegisterEmailer(i)
	databasecfg.RegisterClientConfig(i)
	postgres.RegisterDatabaseClient(i)
	notificationscfg.RegisterPushSender(i)

	// Domain: mealplanning
	mealplanningregistration.RegisterForDataChangeHandler(i)

	// repos
	auditlogentries.RegisterAuditLogRepository(i)
	auth.RegisterAuthRepository(i)
	commentsrepo.RegisterCommentsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	identity.RegisterIdentityRepository(i)
	issue_reports.RegisterIssueReportsRepository(i)
	uploadedmedia.RegisterUploadedMediaRepository(i)
	webhookdispatch.RegisterWebhookDispatch(i)
	webhooks.RegisterWebhooksRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)

	// managers
	notificationsmanager.RegisterNotificationsDataManager(i)
	settingsmanager.RegisterSettingsDataManager(i)
	waitlistsmanager.RegisterWaitlistDataManager(i)

	// indexing
	identityindexing.RegisterUserSyncer(i)

	// searchers
	RegisterSearchers(i)

	// main handler
	datachangemessagehandler.RegisterAsyncDataChangeMessageHandler(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.AsyncMessageHandlerConfig,
) (*datachangemessagehandler.AsyncDataChangeMessageHandler, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*datachangemessagehandler.AsyncDataChangeMessageHandler](i), nil
}
