/*
Package passkeys mounts platform-go's PasskeysService: a passkey enrolled, listed and archived
by a signed-in user, and a sign-in made with one.

The ceremony is platform's — the relying party, the challenge in flight between the two
requests of each ceremony, and the credential it produces. What is here is this application's
half: which user a WebAuthn handle names, whether somebody without a passkey can still sign
in, and the door a finished login is minted through, which is the same signin.Service every
other sign-in goes through. A passkey login therefore gets the token a password login gets,
with the same refresh token, the same hooks, and the same "logged in" event.

A handle is the user's ID, as it was before the move, so a credential registered then is
still found.

Enrollment is admitted for any signed-in caller, as it was before the move: a live session is
taken as proof enough to add a passkey. Archiving the last passkey is refused only for
somebody who has no password to fall back on, which in this application is nobody yet.
*/
package passkeys

import (
	"context"
	"errors"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	platformpasskeys "github.com/primandproper/platform-go/v15/authentication/passkeys"
	passkeysgrpc "github.com/primandproper/platform-go/v15/authentication/passkeys/grpc"
	"github.com/primandproper/platform-go/v15/authentication/passkeys/passkeyspb"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	webauthncfg "github.com/primandproper/platform-go/v15/authentication/webauthnsessions/config"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	platformwebauthn "github.com/primandproper/primitives-go/v2/authentication/webauthn"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
)

// RegisterPasskeysService registers the relying party, the passkey service and platform's
// PasskeysService server with the injector.
//
// The credential store is registered with the identity store, which every container holding
// this service also registers — see identitystore.RegisterPasskeyStore for why it lives there.
func RegisterPasskeysService(i do.Injector) {
	do.Provide[*webauthncfg.Config](i, func(i do.Injector) (*webauthncfg.Config, error) {
		return ProvidePasskeyConfig(do.MustInvoke[*config.APIServiceConfig](i)), nil
	})

	do.Provide[*platformwebauthn.RelyingParty](i, func(i do.Injector) (*platformwebauthn.RelyingParty, error) {
		// The container's context, not a request's: it bounds the ceremony table's sweeper,
		// which lives as long as the process does.
		return webauthncfg.NewRelyingParty(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*webauthncfg.Config](i),
			do.MustInvoke[database.Client](i),
			webauthncfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			webauthncfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			webauthncfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[*platformpasskeys.Service](i, func(i do.Injector) (*platformpasskeys.Service, error) {
		store := do.MustInvoke[platformpasskeys.Store](i)
		directory := do.MustInvoke[platformidentity.Store](i)
		db := do.MustInvoke[database.Client](i)

		users, err := platformpasskeys.NewUserSource(store, resolveHandle(directory, db))
		if err != nil {
			return nil, err
		}

		// A passkey added or removed is a credential write, recorded on its own transaction
		// as a password change is; a failed login records nothing, as a failed password
		// sign-in does not. Both decisions are platform's RecordingHooks'.
		hooks, err := platformpasskeys.NewRecordingHooks(do.MustInvoke[*platformrecording.Recorder](i))
		if err != nil {
			return nil, err
		}

		return platformpasskeys.NewService(
			db,
			store,
			do.MustInvoke[*platformwebauthn.RelyingParty](i),
			users,
			platformpasskeys.WithEnrollmentGate(platformpasskeys.AdmitEveryEnrollment),
			platformpasskeys.WithUsernameResolver(resolveUsername(directory, db)),
			platformpasskeys.WithAlternativeSignIn(holdsPassword(directory)),
			platformpasskeys.WithHooks(hooks),
			platformpasskeys.WithServiceLogger(do.MustInvoke[logging.Logger](i)),
			platformpasskeys.WithServiceTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformpasskeys.WithServiceMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[passkeyspb.PasskeysServiceServer](i, func(i do.Injector) (passkeyspb.PasskeysServiceServer, error) {
		return passkeysgrpc.NewServer(
			do.MustInvoke[*platformpasskeys.Service](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*signin.Service](i),
			sessions.PrincipalFromContext,
			// The server's default is the global scope, which is this directory's scope too. It
			// is named anyway, so that the day this application's directory is scoped, passkeys
			// follow it.
			passkeysgrpc.WithScopeResolver(func(context.Context) (tenancy.Scope, error) {
				return tenancy.Global(), nil
			}),
			passkeysgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			passkeysgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			passkeysgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// ProvidePasskeyConfig extracts the passkey config from the API service config.
//
// When no relying party is configured — local dev, where the rendered config has no origins to
// name — it fills in localhost defaults so the ceremony is buildable. It deliberately does not
// fill in the store: an omitted provider is the platform's default, which is the table, and the
// in-memory option that used to be reachable by leaving this blank no longer exists.
func ProvidePasskeyConfig(cfg *config.APIServiceConfig) *webauthncfg.Config {
	passkey := cfg.Auth.Passkey

	if passkey.RelyingParty.RPID == "" {
		passkey.RelyingParty.RPID = branding.LocalDevRPID
		passkey.RelyingParty.RPOrigins = branding.LocalDevWebAppOrigins()
	}

	if passkey.RelyingParty.RPDisplayName == "" {
		passkey.RelyingParty.RPDisplayName = branding.CompanyName
	}

	return &passkey
}

// resolveHandle answers which user a WebAuthn handle names: the handle is their ID.
func resolveHandle(directory platformidentity.Store, db database.Client) platformpasskeys.UserResolver {
	return func(ctx context.Context, handle []byte) (platformpasskeys.UserIdentity, error) {
		user, err := directory.GetUser(ctx, db.Reader(), tenancy.Global(), string(handle))
		if err != nil {
			return platformpasskeys.UserIdentity{}, err
		}

		return platformpasskeys.UserIdentity{
			UserID:      user.ID,
			Name:        user.Username,
			DisplayName: displayName(user),
		}, nil
	}
}

// displayName is what an authenticator shows beside a passkey: the person's name, or their
// username when they gave none.
func displayName(user *platformidentity.User) string {
	if name := strings.TrimSpace(user.FirstName + " " + user.LastName); name != "" {
		return name
	}

	return user.Username
}

// resolveUsername answers which handle a username names, for a login that names one.
func resolveUsername(directory platformidentity.Store, db database.Client) platformpasskeys.UsernameResolver {
	return func(ctx context.Context, scope tenancy.Scope, username string) ([]byte, error) {
		user, err := directory.GetUserByUsername(ctx, db.Reader(), scope, username)
		if err != nil {
			if errors.Is(err, platformidentity.ErrUserNotFound) {
				return nil, platformpasskeys.ErrUnknownUsername
			}

			return nil, err
		}

		return []byte(user.ID), nil
	}
}

// holdsPassword reports whether a user can sign in without a passkey, which is what lets them
// archive their last one.
func holdsPassword(directory platformidentity.Store) platformpasskeys.AlternativeSignIn {
	return func(ctx context.Context, q database.SQLQueryExecutor, scope tenancy.Scope, userID string) (bool, error) {
		user, err := directory.GetUser(ctx, q, scope, userID)
		if err != nil {
			return false, err
		}

		return user.HashedPassword != "", nil
	}
}

// AnonymousMethods are the login ceremony's two halves, which somebody who is not signed in
// makes.
func AnonymousMethods() []string {
	return passkeysgrpc.AnonymousMethods()
}

// Permissions maps the RPCs a signed-in caller makes about their own passkeys to no permission,
// which is how this application's interceptor spells "any signed-in caller".
func Permissions() map[string][]authorization.Permission {
	out := map[string][]authorization.Permission{}
	for _, method := range passkeysgrpc.SelfServiceMethods() {
		out[method] = []authorization.Permission{}
	}

	return out
}
