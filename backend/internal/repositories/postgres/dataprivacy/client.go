package dataprivacy

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/dataprivacy/generated"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "dataprivacy_db_client"
)

// disclosureRepository is the half of the data privacy repository that manages disclosure
// records. It is separated from the whole because the two halves have wildly different
// appetites: gathering a user's data reaches into every domain in the application, while
// tracking the request that asked for it needs a database handle and an audit log. The
// scheduler reaps expired disclosures and nothing else, and would otherwise have to construct
// every repository in the system to do it.
type disclosureRepository struct {
	logger           logging.Logger
	tracer           tracing.Tracer
	generatedQuerier generated.Querier
	auditLogRepo     audit.Repository
	database.Client
	readDB  database.SQLQueryExecutor
	writeDB database.SQLQueryExecutor
}

// repository is the data privacy repository client.
type repository struct {
	*disclosureRepository
	issueReportsRepo  issuereports.Repository
	uploadedMediaRepo uploadedmedia.Repository
	identityRepo      identity.Repository
	webhooksRepo      webhooks.Repository
	settingsRepo      settings.Repository
	notificationsRepo notifications.Repository
	waitlistsRepo     waitlists.Repository
	commentsRepo      comments.Repository
	paymentsRepo      payments.Repository
	dataCollectors    []dataprivacy.UserDataCollector
}

// ProvideDataPrivacyRepository provides a new repository.
func ProvideDataPrivacyRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogRepo audit.Repository,
	identityRepo identity.Repository,
	issueReportsRepo issuereports.Repository,
	notificationsRepo notifications.Repository,
	settingsRepo settings.Repository,
	uploadedMediaRepo uploadedmedia.Repository,
	waitlistsRepo waitlists.Repository,
	webhooksRepo webhooks.Repository,
	commentsRepo comments.Repository,
	paymentsRepo payments.Repository,
	client database.Client,
	dataCollectors []dataprivacy.UserDataCollector,
) dataprivacy.Repository {
	c := &repository{
		disclosureRepository: provideDisclosureRepository(logger, tracerProvider, auditLogRepo, client),
		identityRepo:         identityRepo,
		issueReportsRepo:     issueReportsRepo,
		notificationsRepo:    notificationsRepo,
		settingsRepo:         settingsRepo,
		uploadedMediaRepo:    uploadedMediaRepo,
		waitlistsRepo:        waitlistsRepo,
		webhooksRepo:         webhooksRepo,
		commentsRepo:         commentsRepo,
		paymentsRepo:         paymentsRepo,
		dataCollectors:       dataCollectors,
	}

	return c
}

// ProvideUserDataDisclosureRepository provides just the disclosure-record half of the data
// privacy repository, for processes that manage disclosure requests without ever gathering the
// data behind them.
func ProvideUserDataDisclosureRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogRepo audit.Repository,
	client database.Client,
) dataprivacy.UserDataDisclosureDataManager {
	return provideDisclosureRepository(logger, tracerProvider, auditLogRepo, client)
}

func provideDisclosureRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogRepo audit.Repository,
	client database.Client,
) *disclosureRepository {
	return &disclosureRepository{
		Client:           client,
		readDB:           client.Reader(),
		writeDB:          client.Writer(),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:           logging.NewNamedLogger(logger, o11yName),
		generatedQuerier: generated.New(),
		auditLogRepo:     auditLogRepo,
	}
}
