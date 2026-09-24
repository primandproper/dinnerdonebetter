package authentication

import (
	"context"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// signInHooks is what this application does when platform's sign-in proves somebody.
//
// There are two doors into a password sign-in now: this repository's AuthService, through
// Manager.ProcessLogin, and platform's SignInService, which this server registers as well.
// Both prove the password through the same signin.Service, and AfterAuthenticate runs for
// both, so it is the one place the "logged in" event can be recorded without one door
// forgetting it. It used to be published by ProcessLogin after the call returned, which is
// what a second door would have skipped.
//
// The event goes on the outbox, on the transaction signin hands the hook, rather than to the
// broker. On SignInService's door that transaction also writes the refresh token, so the
// event and the login commit together: a sign-in that fails after proving the password —
// the refresh token write, or a later hook — rolls the event back with it, where a publish
// from here would already have told the broker about a login that never happened. And a
// failed enqueue refuses the sign-in, which is the arrangement ProcessLogin had — a publish
// that failed failed the login — kept on purpose: the event feeds the audit trail, and a
// sign-in nobody has a record of is the one an investigation would ask about.
//
// AuthService's door is narrower and not fully closed. It proves the password through
// signin.Authenticate, whose transaction holds this hook alone, and issues its session
// afterwards; a session write that fails leaves the event committed. What the event says is
// then still true — the password was proven — which is what platform means by this hook,
// and it is the only door that can be read that way.
//
// The other hooks are NoopHooks'. The credential writes they follow are still made through
// this repository's AuthService, which records them itself; platform's doors for them are
// not reachable on this server yet. See internal/build/signin.
type signInHooks struct {
	signin.NoopHooks

	logger  logging.Logger
	emitter *events.Emitter
}

var _ signin.Hooks = (*signInHooks)(nil)

// NewSignInHooks builds the hooks platform's sign-in runs for this application.
//
// The emitter may be nil, which is inert: a process with no data changes topic records no
// sign-ins, as it records no other change. See events.NewEmitter.
func NewSignInHooks(logger logging.Logger, emitter *events.Emitter) signin.Hooks {
	return &signInHooks{
		logger:  logging.NewNamedLogger(logger, "signin_hooks"),
		emitter: emitter,
	}
}

// AfterAuthenticate records the "logged in" event for a proven sign-in, through either door
// and whether or not it was administrative.
func (h *signInHooks) AfterAuthenticate(ctx context.Context, tx database.Tx, _ tenancy.Scope, authentication *signin.Authentication) error {
	if authentication == nil || authentication.Principal == nil || authentication.Principal.User == nil {
		return platformerrors.New("sign-in hook called with no principal")
	}

	principal := authentication.Principal
	logger := h.logger.WithValue(identitykeys.UserIDKey, principal.User.ID)

	if err := h.emitter.Emit(ctx, tx, logger,
		ddbidentity.UserLoggedInServiceEventType,
		principal.ActiveAccountID,
		nil,
		events.WithUserID(principal.User.ID),
	); err != nil {
		return platformerrors.Wrap(err, "recording a sign-in")
	}

	return nil
}
