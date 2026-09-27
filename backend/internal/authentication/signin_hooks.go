package authentication

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
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
// The four credential writes — a password change, a second factor refreshed or proven, and
// an email address proven — are recorded here for the same reason. Both doors make them
// through signin.Service: AuthService's UpdatePassword, NewTOTPSecret, TOTPSecretVerification
// and the two email verifications, and platform's SignInService RPCs of the same names. Each
// used to be published by AuthService after the call returned, which is exactly what the
// second door would have skipped, and published to the broker after the write had committed,
// so an outage between the two lost the event. On the hook's transaction it commits with the
// write or not at all, and a failed enqueue refuses the write — the audit trail is what a
// "who changed my password" investigation reads.
//
// The remaining hooks are NoopHooks'. See internal/build/signin for which doors are mounted.
type signInHooks struct {
	signin.NoopHooks

	logger   logging.Logger
	emitter  *events.Emitter
	recorder *recording.Recorder
}

var _ signin.Hooks = (*signInHooks)(nil)

// NewSignInHooks builds the hooks platform's sign-in runs for this application.
//
// The emitter may be nil, which is inert: a process with no data changes topic records no
// sign-ins, as it records no other change. See events.NewEmitter. The recorder writes the audit
// entry and the event a credential write leaves; nil records nothing, for a process that makes
// no credential writes.
func NewSignInHooks(logger logging.Logger, emitter *events.Emitter, recorder *recording.Recorder) signin.Hooks {
	return &signInHooks{
		logger:   logging.NewNamedLogger(logger, "signin_hooks"),
		emitter:  emitter,
		recorder: recorder,
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

// AfterUpdatePassword records a password change, through either door.
func (h *signInHooks) AfterUpdatePassword(ctx context.Context, tx database.Tx, _ tenancy.Scope, user *identity.User) error {
	return h.recordCredentialWrite(ctx, tx, user, auth.PasswordChangedEventType, "recording a password change")
}

// AfterRefreshTOTPSecret records a second factor being replaced, through either door. The
// secret itself is not on the user and is not recorded.
func (h *signInHooks) AfterRefreshTOTPSecret(ctx context.Context, tx database.Tx, _ tenancy.Scope, user *identity.User) error {
	return h.recordCredentialWrite(ctx, tx, user, auth.TwoFactorSecretChangedServiceEventType, "recording a second factor refresh")
}

// AfterVerifyTOTPSecret records a second factor being proven, through either door.
func (h *signInHooks) AfterVerifyTOTPSecret(ctx context.Context, tx database.Tx, _ tenancy.Scope, user *identity.User) error {
	return h.recordCredentialWrite(ctx, tx, user, auth.TwoFactorSecretVerifiedServiceEventType, "recording a second factor verification")
}

// AfterVerify records an email address proven by its mailed link, through either door.
//
// Only the link door is an email verification. CompleteVerification promotes somebody on a
// proof this application does not ask for, and says nothing about the address, so it records
// nothing here; and this application registers people in good standing, so Promoted is not
// what distinguishes the event.
func (h *signInHooks) AfterVerify(ctx context.Context, tx database.Tx, _ tenancy.Scope, verification *signin.Verification) error {
	if verification == nil || !verification.EmailAddressProven {
		return nil
	}

	return h.recordCredentialWrite(ctx, tx, verification.User, auth.UserEmailAddressVerifiedEventType, "recording a verified email address")
}

// recordCredentialWrite records a credential write for user on the write's own transaction: an
// audit entry naming the user as updated, and eventType on the outbox beside it.
//
// Both, and in the transaction, because this is where the write happens for both doors. An
// event alone left no audit trail, and a credential change is exactly the entry an
// investigation reads first.
//
// The user is named rather than read from the context: an email verification arrives on a
// link with nobody signed in, and the other three are about the caller anyway.
func (h *signInHooks) recordCredentialWrite(ctx context.Context, tx database.Tx, user *identity.User, eventType, describing string) error {
	if user == nil {
		return platformerrors.Newf("%s: hook called with no user", describing)
	}

	if h.recorder == nil {
		return nil
	}

	logger := h.logger.WithValue(identitykeys.UserIDKey, user.ID)

	entry := &audit.AuditLogEntry{
		ID:            identifiers.New(),
		ResourceType:  usersResourceType,
		RelevantID:    user.ID,
		EventType:     audit.AuditLogEventTypeUpdated,
		BelongsToUser: user.ID,
	}

	if err := h.recorder.RecordAndEmit(ctx, tx, logger, entry, eventType, "",
		map[string]any{identitykeys.UserIDKey: user.ID}, events.WithUserID(user.ID)); err != nil {
		return platformerrors.Wrap(err, describing)
	}

	return nil
}

// usersResourceType is what an audit entry about a user names, as the identity store's own
// hooks name it.
const usersResourceType = "users"
