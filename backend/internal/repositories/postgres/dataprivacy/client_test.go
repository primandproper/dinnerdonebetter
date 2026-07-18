package dataprivacy

import (
	"context"
	"database/sql"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	mealplanningprivacy "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/privacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	issue_reports "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"

	mockdatabase "github.com/primandproper/platform-go/v5/database/mock"
	"github.com/primandproper/platform-go/v5/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"

	"github.com/stretchr/testify/require"
	pgcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func buildDatabaseClientForTest(t *testing.T) (repo *repository, auditRepo audit.Repository, idRepo identity.Repository, pgContainer *pgcontainers.PostgresContainer) {
	t.Helper()

	ctx := t.Context()
	container, db, config := pgtesting.BuildDatabaseContainerForTest(t)
	require.NoError(t, migrations.NewMigrator(loggingnoop.NewLogger()).Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	identityRepo := identityrepo.ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	issueReportsRepo := issue_reports.ProvideIssueReportsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	mealPlanningRepo := mealplanning.ProvideMealPlanningRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, identityRepo, pgc)
	notificationsRepo := notifications.ProvideNotificationsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, config, pgc)
	settingsRepo := settings.ProvideSettingsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	uploadedMediaRepo := uploadedmedia.ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	waitlistsRepo := waitlists.ProvideWaitlistsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	webhooksRepo := webhooks.ProvideWebhooksRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)

	mealPlanningCollector := mealplanningprivacy.NewCollector(mealPlanningRepo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

	c := ProvideDataPrivacyRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		auditLogEntryRepo,
		identityRepo,
		issueReportsRepo,
		notificationsRepo,
		settingsRepo,
		uploadedMediaRepo,
		waitlistsRepo,
		webhooksRepo,
		pgc,
		[]dataprivacy.UserDataCollector{mealPlanningCollector},
	)

	return c.(*repository), auditLogEntryRepo, identityRepo, container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideDataPrivacyRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		nil, // auditLogRepo
		nil, // identityRepo
		nil, // issueReportsRepo
		nil, // notificationsRepo
		nil, // settingsRepo
		nil, // uploadedMediaRepo
		nil, // waitlistsRepo
		nil, // webhooksRepo
		&mockdatabase.ClientMock{ReadDBFunc: func() *sql.DB { return nil }, WriteDBFunc: func() *sql.DB { return nil }},
		nil, // dataCollectors
	)

	return c.(*repository)
}

func createUserForTest(t *testing.T, ctx context.Context, exampleUser *identity.User, identityRepo identity.Repository) *identity.User {
	t.Helper()

	if exampleUser == nil {
		exampleUser = fakes.BuildFakeUser()
	}
	exampleUser.TwoFactorSecretVerifiedAt = nil
	dbInput := converters.ConvertUserToUserDatabaseCreationInput(exampleUser)

	created, err := identityRepo.CreateUser(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	return created
}
