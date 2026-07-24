package dataprivacy

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/dataprivacy/generated"

	"github.com/primandproper/platform-go/v6/database"
	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

const (
	o11yName = "dataprivacy_db_client"
)

// repository is the data privacy repository client.
type repository struct {
	issueReportsRepo  issuereports.Repository
	uploadedMediaRepo uploadedmedia.Repository
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogRepo      audit.Repository
	identityRepo      identity.Repository
	tracer            tracing.Tracer
	webhooksRepo      webhooks.Repository
	database.Client
	settingsRepo      settings.Repository
	notificationsRepo notifications.Repository
	waitlistsRepo     waitlists.Repository
	commentsRepo      comments.Repository
	paymentsRepo      payments.Repository
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
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
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:            logging.NewNamedLogger(logger, o11yName),
		generatedQuerier:  generated.New(),
		auditLogRepo:      auditLogRepo,
		identityRepo:      identityRepo,
		issueReportsRepo:  issueReportsRepo,
		notificationsRepo: notificationsRepo,
		settingsRepo:      settingsRepo,
		uploadedMediaRepo: uploadedMediaRepo,
		waitlistsRepo:     waitlistsRepo,
		webhooksRepo:      webhooksRepo,
		commentsRepo:      commentsRepo,
		paymentsRepo:      paymentsRepo,
		dataCollectors:    dataCollectors,
	}

	return c
}
