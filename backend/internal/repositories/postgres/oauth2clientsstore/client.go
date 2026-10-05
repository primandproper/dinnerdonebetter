package oauth2clientsstore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformoauth2clients "github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const o11yName = "oauth2_clients_db_client"

// ProvideHooks builds what the registry Service records each write with.
func ProvideHooks(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
) platformoauth2clients.Hooks {
	return &hooks{
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracing.NewNamedTracer(tracerProvider, o11yName), auditLogEntryRepo, eventEmitter),
	}
}
