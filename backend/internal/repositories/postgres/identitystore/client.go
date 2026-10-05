/*
Package identitystore is this application's audit trail and event stream for
platform-go's identity service.

It is not a store decorator, and that is the whole shape of the thing. Every
other adopted domain here wraps a Store — comments, webhooks, notifications,
waitlists — because each of their writes is one statement, so one decorated
method is one operation. An identity operation is not: registering somebody
writes a user, an account and a membership in one transaction, and a decorator
over the store would record one row of an operation that wrote three, three
times, with no way to say which operation it belonged to.

platform's answer is identity.Hooks: one method per operation, each handed the
operation's own database.Tx after the writes and before the commit. The audit
entry and the outbox row this application owes go in there, so they commit with
the rows they describe or not at all. A spike proved that against a real
database before any of this was written — including the rollback, which is the
half that matters.

The hooks implement identity.Hooks outright rather than embedding
identity.NoopHooks, so an operation added upstream breaks this build until
somebody decides what it records. Embedded, it would compile and record
nothing, which is the one failure an audit log cannot notice — and how
AfterCreateAccount once arrived here unrecorded.
*/
package identitystore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "identity_recording_hooks"

	// The names an audit entry gives what it is about.
	//
	// They are this application's table names from before the adoption, not platform's:
	// an audit log is read across the change, so an investigation asking what happened
	// to a user should find the entries written when the table was called users as well
	// as the ones written since it became ddb_identity_users. webhooksstore and
	// oauth2clientsstore keep their old names for the same reason.
	resourceTypeUsers                  = "users"
	resourceTypeAccounts               = "accounts"
	resourceTypeAccountUserMemberships = "account_user_memberships"
	resourceTypeAccountInvitations     = "account_invitations"
)

// Hooks is platform's identity.Hooks with this application's recording in every method.
type Hooks struct {
	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformidentity.Hooks = (*Hooks)(nil)

// ProvideHooks builds the recording hooks.
func ProvideHooks(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
) *Hooks {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &Hooks{
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}
}
