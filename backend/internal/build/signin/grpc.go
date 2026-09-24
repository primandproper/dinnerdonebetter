/*
Package signin mounts platform-go's sign-in surface, SignInService, beside this
application's own AuthService.

It is the door @primandproper/platform-client's Session signs in and refreshes
through. The sign-in it runs is the same signin.Service AuthService proves passwords
with, built in internal/authentication, so the two doors check the same credentials
in the same order and publish the same "logged in" event; what differs is the token
they hand back. AuthService's names a row in the session store. This one's names a
login, rotates a refresh token through the store in
internal/repositories/postgres/auth, and is checked by signature alone — see
AuthInterceptor.signInSessionContextData for what that trades away.

Not every RPC on the surface is reachable. The interceptor denies a method no
permission table names, so a method is exposed by being named below and withheld by
being left out. Withheld, and why:

  - Register, UpdatePassword and AttachPassword write a password, and platform's
    server applies no password policy and offers no seam for one. This application
    refuses a weak password on each of those writes today, through AuthService, and
    would stop refusing it here.
  - RefreshTOTPSecret, VerifyTOTPSecret and VerifyEmailAddress are credential writes
    whose events AuthService publishes around the call. They move once the hooks in
    internal/authentication publish them for both doors, as they already do for
    sign-in.
  - RequestMagicLink and RedeemMagicLink: this deployment names no magic link store,
    so signin refuses both anyway.

Password reset is not mounted at all, for the first of those reasons: its last step
writes a password.
*/
package signin

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformsignin "github.com/primandproper/platform-go/v14/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v14/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
)

// RegisterSignInService registers platform's sign-in surface with the injector.
func RegisterSignInService(i do.Injector) {
	do.Provide[signinpb.SignInServiceServer](i, func(i do.Injector) (signinpb.SignInServiceServer, error) {
		return signingrpc.NewServer(
			do.MustInvoke[*platformsignin.Service](i),
			sessions.PrincipalFromContext,
			// The server's default is the global scope, which is this directory's scope
			// too. It is named anyway, so that the day this application's directory is
			// scoped, sign-in follows it rather than signing people into another one.
			signingrpc.WithScopeResolver(func(context.Context) (tenancy.Scope, error) {
				return ddbidentity.Scope(), nil
			}),
			signingrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			signingrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			signingrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// AnonymousMethods are the exposed RPCs a caller reaches without a token: the two sign-in
// doors, the refresh exchange, and sign-out, whose authority is the refresh token it carries.
func AnonymousMethods() []string {
	return []string{
		signinpb.SignInService_LoginForToken_FullMethodName,
		signinpb.SignInService_AdminLoginForToken_FullMethodName,
		signinpb.SignInService_ExchangeRefreshToken_FullMethodName,
		signinpb.SignInService_SignOut_FullMethodName,
	}
}

// OptionallyAuthenticatedMethods answer an anonymous caller and a signed-in one differently.
//
// GetAuthStatus is the only one. It is anonymous on platform's list, and it is also the call
// a client makes to learn what a signed-in user still owes — a verified address, a second
// factor, a new password — which it can only answer if the token that came with it is read.
func OptionallyAuthenticatedMethods() []string {
	return []string{
		signinpb.SignInService_GetAuthStatus_FullMethodName,
	}
}

// Permissions maps the exposed RPCs that need a signed-in caller.
//
// Each maps to no permission, which is how this application's interceptor spells "any
// signed-in caller": they act on the caller alone, and platform's own list calls them
// self-service for that reason.
func Permissions() map[string][]authorization.Permission {
	return map[string][]authorization.Permission{
		signinpb.SignInService_GetSelf_FullMethodName:           {},
		signinpb.SignInService_SignOutEverywhere_FullMethodName: {},
	}
}
