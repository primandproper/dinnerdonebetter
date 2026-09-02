package webhooks

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainwebhooks "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/webhooks"
)

const (
	o11yName = "webhook_db_client"
)

// repository is the webhook repository client.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
	recorder          *recording.Recorder
	// dispatcher registers endpoints and fans events out to them. It is required rather than
	// optional, because a webhook that is stored and not registered is one the account was
	// told exists and that will never fire.
	dispatcher webhooks.Dispatcher
	// endpoints is the same store the dispatcher was built over, held for the two endpoint
	// operations Dispatcher does not offer: replacing a subscription set, and retiring an
	// endpoint. See endpoints.go.
	endpoints webhooks.Store
	readDB    database.SQLQueryExecutor
	writeDB   database.SQLQueryExecutor
}

// ProvideWebhooksRepository provides a new repository.
func ProvideWebhooksRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
	dispatcher webhooks.Dispatcher,
	endpoints webhooks.Store,
) domainwebhooks.Repository {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	c := &repository{
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracer,
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		events:            eventEmitter,
		recorder:          recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
		dispatcher:        dispatcher,
		endpoints:         endpoints,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
