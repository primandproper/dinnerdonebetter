package identity

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/random"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

const (
	o11yName = "identity_db_client"
)

var _ identity.Repository = (*repository)(nil)

// repository is the identity repository implementation.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
	recorder          *recording.Recorder
	secretGenerator   random.Generator

	// uploads answers what a user_avatars row points at. The avatar itself lives
	// in platform-go's upload registry, whose table this repository's statements
	// cannot join — see avatarFor.
	uploads registry.Store

	readDB  database.SQLQueryExecutor
	writeDB database.SQLQueryExecutor
}

// ProvideIdentityRepository provides a new repository.
func ProvideIdentityRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
	uploads registry.Store,
) identity.Repository {
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
		uploads:           uploads,
		secretGenerator:   random.NewGenerator(random.WithLogger(logger), random.WithTracerProvider(tracerProvider)),
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}

// withEvent runs a write and the data change event describing it in one transaction, so the
// event cannot survive a write that rolled back — nor be lost after one that committed.
//
// The user writes that change the search index do not go through this — they emit inside a
// larger transaction of their own — so this takes no events.EmitOption. See users.go.
func (r *repository) withEvent(
	ctx context.Context,
	logger logging.Logger,
	eventType, accountID string,
	metadata map[string]any,
	write func(tx database.Tx) error,
) error {
	return r.WithTransaction(ctx, func(tx database.Tx) error {
		if err := write(tx); err != nil {
			return err
		}

		return r.events.Emit(ctx, tx, logger, eventType, accountID, metadata)
	})
}
