package datachangemessagehandler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	notificationsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/manager"
	settingsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/manager"
	waitlistsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issue_reports "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	paymentsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"

	analyticscfg "github.com/primandproper/platform-go/v6/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v6/database/config"
	"github.com/primandproper/platform-go/v6/database/postgres"
	emailcfg "github.com/primandproper/platform-go/v6/email/config"
	"github.com/primandproper/platform-go/v6/encoding"
	"github.com/primandproper/platform-go/v6/httpclient"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v6/notifications/mobile/config"
	"github.com/primandproper/platform-go/v6/observability"
	loggingcfg "github.com/primandproper/platform-go/v6/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v6/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v6/observability/tracing/config"
	"github.com/primandproper/platform-go/v6/uploads/objectstorage"

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
	objectstorage.RegisterUploadManager(i)
	notificationscfg.RegisterPushSender(i)

	// Domain: mealplanning
	mealplanningregistration.RegisterForDataChangeHandler(i)

	// repos
	auditlogentries.RegisterAuditLogRepository(i)
	auth.RegisterAuthRepository(i)
	commentsrepo.RegisterCommentsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	dataprivacy.RegisterDataPrivacyRepository(i)
	identity.RegisterIdentityRepository(i)
	issue_reports.RegisterIssueReportsRepository(i)
	uploadedmedia.RegisterUploadedMediaRepository(i)
	webhooks.RegisterWebhooksRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)

	// managers
	notificationsmanager.RegisterNotificationsDataManager(i)
	settingsmanager.RegisterSettingsDataManager(i)
	waitlistsmanager.RegisterWaitlistDataManager(i)

	// indexing
	identityindexing.RegisterCoreDataIndexer(i)

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
