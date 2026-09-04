package datachangemessagehandler

import (
	"context"

	commentstargets "github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	"github.com/primandproper/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issue_reports "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	analyticscfg "github.com/primandproper/platform-go/v13/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/database/postgres"
	emailcfg "github.com/primandproper/platform-go/v13/email/config"
	"github.com/primandproper/platform-go/v13/encoding"
	"github.com/primandproper/platform-go/v13/httpclient"
	msgconfig "github.com/primandproper/platform-go/v13/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v13/notifications/mobile/config"
	"github.com/primandproper/platform-go/v13/observability"
	loggingcfg "github.com/primandproper/platform-go/v13/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v13/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v13/observability/tracing/config"

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
	// No existence checks on the catalog: this process reads and erases comments
	// but never writes one, and the catalog gates writes rather than reads.
	commentstargets.RegisterReadOnlyTargets(i)
	commentsrepo.RegisterCommentsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	// What a role grants, read from the policy tables the migrator seeds. The
	// identity repository resolves a principal's role names through it when it
	// builds a session.
	identity.RegisterPolicyResolver(i)
	identity.RegisterIdentityRepository(i)
	issue_reports.RegisterIssueReportsRepository(i)
	uploadedmedia.RegisterUploadedMediaRepository(i)
	webhooks.RegisterWebhooksRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)

	// managers
	notificationsmanager.RegisterNotificationsDataManager(i)
	// The push fan-out over that manager. The scheduler builds the same one for the prep task
	// reminders it now sends itself, so both processes deliver through one component.
	push.RegisterFanout(i)
	settingsrepo.RegisterSettingsRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)

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
