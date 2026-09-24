package authentication

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// signInHooks is what this application does when platform's sign-in proves somebody.
//
// There are two doors into a password sign-in now: this repository's AuthService, through
// Manager.ProcessLogin, and platform's SignInService, which this server registers as well.
// Both prove the password through the same signin.Service, and AfterAuthenticate runs for
// both, so it is the one place the "logged in" event can be published without one door
// forgetting it. It used to be published by ProcessLogin after the call returned, which is
// what a second door would have skipped.
//
// It runs inside the transaction signin opens around the proof, and an error from it refuses
// the sign-in. That is the arrangement ProcessLogin had — a publish that failed failed the
// login — kept on purpose: the event feeds the audit trail and outbound webhooks, and a
// sign-in nobody has a record of is the one an investigation would ask about.
//
// The other hooks are NoopHooks'. The credential writes they follow are still made through
// this repository's AuthService, which records them itself; platform's doors for them are
// not reachable on this server yet. See internal/build/signin.
type signInHooks struct {
	signin.NoopHooks

	dataChangesPublisher messagequeue.Publisher
}

var _ signin.Hooks = (*signInHooks)(nil)

// ErrNilSignInPublisher is returned when the sign-in hooks are built without a publisher.
var ErrNilSignInPublisher = platformerrors.New("sign-in hooks need a data changes publisher")

// NewSignInHooks builds the hooks platform's sign-in runs for this application.
func NewSignInHooks(dataChangesPublisher messagequeue.Publisher) (signin.Hooks, error) {
	if dataChangesPublisher == nil {
		return nil, ErrNilSignInPublisher
	}

	return &signInHooks{dataChangesPublisher: dataChangesPublisher}, nil
}

// AfterAuthenticate publishes the "logged in" event for a proven sign-in, through either door
// and whether or not it was administrative.
func (h *signInHooks) AfterAuthenticate(ctx context.Context, _ database.Tx, _ tenancy.Scope, authentication *signin.Authentication) error {
	if authentication == nil || authentication.Principal == nil || authentication.Principal.User == nil {
		return platformerrors.New("sign-in hook called with no principal")
	}

	principal := authentication.Principal

	if err := h.dataChangesPublisher.Publish(ctx, &audit.DataChangeMessage{
		EventType: ddbidentity.UserLoggedInServiceEventType,
		AccountID: principal.ActiveAccountID,
		UserID:    principal.User.ID,
	}); err != nil {
		return platformerrors.Wrap(err, "publishing a sign-in")
	}

	return nil
}
