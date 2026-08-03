package dataprivacy

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	mealplanningprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/privacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	issue_reports "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"

	"github.com/primandproper/platform-go/v9/database"
	mockdatabase "github.com/primandproper/platform-go/v9/database/mock"
	"github.com/primandproper/platform-go/v9/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/require"
)

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from, so that a test costs a database
// clone rather than a container start plus a migration replay. See
// pgtesting.RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

func buildDatabaseClientForTest(t *testing.T) (repo *repository, auditRepo audit.Repository, idRepo identity.Repository) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)
	identityRepo := identityrepo.ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	issueReportsRepo := issue_reports.ProvideIssueReportsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	mealPlanningRepo := mealplanning.ProvideMealPlanningRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, identityRepo, pgc, nil)
	notificationsRepo := notifications.ProvideNotificationsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, &config.Config, pgc, nil)
	settingsRepo := settings.ProvideSettingsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	uploadedMediaRepo := uploadedmedia.ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	waitlistsRepo := waitlists.ProvideWaitlistsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	webhooksRepo := webhooks.ProvideWebhooksRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	commentsRepo := commentsrepo.ProvideCommentsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	paymentsRepo := paymentsrepo.ProvidePaymentsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)

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
		commentsRepo,
		paymentsRepo,
		pgc,
		[]dataprivacy.UserDataCollector{mealPlanningCollector},
	)

	return c.(*repository), auditLogEntryRepo, identityRepo
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
		nil, // commentsRepo
		nil, // paymentsRepo
		&mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }},
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
