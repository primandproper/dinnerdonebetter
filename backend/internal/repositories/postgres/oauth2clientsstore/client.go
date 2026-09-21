/*
Package oauth2clientsstore is platform-go's registered-client store with this
application's recording around it.

It replaces internal/repositories/postgres/oauth, which owned the oauth2_clients
table. The table is platform's now — see migrate.go's
adopt_oauth2_registered_clients — and what is left here is the part that was
never platform's: the audit entry and the data change event a registration owes,
written as further statements in the transaction that performed it.

Both used to be written a layer apart. The repository recorded the audit entry
inside its own transaction and the manager emitted the event after that
transaction had committed, so a registration whose event failed to enqueue was a
client that exists and that nothing downstream was told about. They are one
write now, which is the property the Store interface made available: platform
opens the transaction and hands it in, so an entry refused rolls the credential
back with it.

The reads are embedded. Five of the eight methods add nothing, and five
forwarding stubs are five chances to drift from the store they forward to.
*/
package oauth2clientsstore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "oauth2_clients_db_client"

	// resourceTypeOAuth2Clients is what an audit entry about a registration names.
	//
	// It keeps the old table's name rather than taking the new one's, for the reason
	// webhooksstore keeps webhooks: an audit log is read across the change, and an
	// investigation asking what happened to a client should find the entries written
	// before the store moved as well as the ones after.
	resourceTypeOAuth2Clients = "oauth2_clients"
)

// store is platform's client registry with this application's recording around its writes.
type store struct {
	platformoauth2clients.Store

	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

// ProvideStore wraps a platform client registry in this application's recording.
func ProvideStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	inner platformoauth2clients.Store,
) (platformoauth2clients.Store, error) {
	if inner == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil oauth2 client store")
	}

	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &store{
		Store:    inner,
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}, nil
}
