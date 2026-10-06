/*
Package signin mounts platform-go's sign-in surface: SignInService, and the operator half beside
it, SignInAdministrationService.

It is the door every client signs in, refreshes and signs out through, and the one
@primandproper/platform-client's Session calls. The signin.Service it runs is built in
internal/authentication, with this application's password, registration, second-factor and
impersonation policies, and its hooks and mailers. A token it issues names a login, rotates a
refresh token through the store in internal/repositories/postgres/auth, and is checked against
that login on every request — see AuthInterceptor.signInSessionContextData.

Not every RPC on the surface is reachable. The interceptor denies a method no permission table
names, so a method is exposed by being named below and withheld by being left out.

Exposed without a caller, beyond the sign-in doors themselves:

  - VerifyEmailAddress, whose authority is the mailed link it carries. Its event is recorded by
    the AfterVerify hook in internal/authentication.
  - AttachPassword furnishes an account that has no password, once, with the link it was mailed.
    This application has no passwordless arrival, so today it only ever answers that a password
    is already set; it is reachable so that the refusal is platform's rather than this
    interceptor's, and it writes nothing this application's rule would refuse, because the
    sign-in service is built with authentication.PasswordPolicy.
  - RequestMagicLink and RedeemMagicLink. This deployment names no magic link store, so platform
    refuses both; what reaching them buys is that the refusal says so, rather than claiming a
    caller was missing.
  - RequestVerificationEmailByAddress and RequestHandleReminder, which answer every address the
    same way and mail only its holder, through this application's outbox — see
    authentication.SignInMailers.

Exposed to any signed-in caller, about themselves: GetSelf, SignOutEverywhere, the sign-ins they
hold (ListSignIns, EndSignIn, EndOtherSignIns), another verification link
(RequestVerificationEmail), their handles (UpdateEmailAddress and UpdateUsername, each of which
asks the current password again), and the three credential writes UpdatePassword,
RefreshTOTPSecret and VerifyTOTPSecret. Each write's event is recorded by a hook in
internal/authentication on the write's own transaction, and UpdatePassword answers to
authentication.PasswordPolicy.

Open to anybody: Register, the sign-up door. Who may sign up, and as what — good standing, the
service role, a second factor, ownership of their account, and the terms and privacy agreements,
refused without them — is authentication.RegistrationPolicy's, which the sign-in service runs on
every registration. It reads a token when one is sent, so a signed-in operator registering
somebody is still somebody to the policy.

Reserved to an operator: SignInAdministrationService, which lists and ends somebody else's
sign-ins. Its two permissions are a service administrator's.

Password reset is PasswordResetService, mounted in internal/build/passwordreset; passkeys are
PasskeysService, mounted in internal/build/passkeys.
*/
package signin

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformsignin "github.com/primandproper/platform-go/v15/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
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
				return tenancy.Global(), nil
			}),
			signingrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			signingrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			signingrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// AnonymousMethods are the exposed RPCs a caller reaches without a token: the two sign-in
// doors, the refresh exchange and the account switch, and sign-out, whose authority is the
// refresh token it carries;
// and the doors whose authority is a mailed link, which are refused by platform rather than
// here — see the package documentation.
func AnonymousMethods() []string {
	return []string{
		signinpb.SignInService_LoginForToken_FullMethodName,
		signinpb.SignInService_AdminLoginForToken_FullMethodName,
		signinpb.SignInService_ExchangeRefreshToken_FullMethodName,
		signinpb.SignInService_SwitchAccount_FullMethodName,
		signinpb.SignInService_SignOut_FullMethodName,
		signinpb.SignInService_AttachPassword_FullMethodName,
		signinpb.SignInService_VerifyEmailAddress_FullMethodName,
		signinpb.SignInService_RequestMagicLink_FullMethodName,
		signinpb.SignInService_RedeemMagicLink_FullMethodName,
		signinpb.SignInService_RequestVerificationEmailByAddress_FullMethodName,
		signinpb.SignInService_RequestHandleReminder_FullMethodName,
	}
}

// OptionallyAuthenticatedMethods answer an anonymous caller and a signed-in one differently.
//
// GetAuthStatus is anonymous on platform's list, and it is also the call a client makes to learn
// what a signed-in user still owes — a verified address, a second factor, a new password — which
// it can only answer if the token that came with it is read.
//
// Register is anonymous too, and platform declares it optional for the same reason: an operator
// registering somebody is resolved, so RegistrationPolicy can read who is asking.
func OptionallyAuthenticatedMethods() []string {
	return []string{
		signinpb.SignInService_GetAuthStatus_FullMethodName,
		signinpb.SignInService_Register_FullMethodName,
	}
}

// Permissions maps the exposed RPCs that need a signed-in caller.
//
// The self-service ones map to no permission, which is how this application's
// interceptor spells "any signed-in caller": they act on the caller alone, and platform's
// own list calls them self-service for that reason.
func Permissions() map[string][]authorization.Permission {
	return map[string][]authorization.Permission{
		signinpb.SignInService_GetSelf_FullMethodName:                  {},
		signinpb.SignInService_SignOutEverywhere_FullMethodName:        {},
		signinpb.SignInService_UpdatePassword_FullMethodName:           {},
		signinpb.SignInService_RefreshTOTPSecret_FullMethodName:        {},
		signinpb.SignInService_VerifyTOTPSecret_FullMethodName:         {},
		signinpb.SignInService_ListSignIns_FullMethodName:              {},
		signinpb.SignInService_EndSignIn_FullMethodName:                {},
		signinpb.SignInService_EndOtherSignIns_FullMethodName:          {},
		signinpb.SignInService_RequestVerificationEmail_FullMethodName: {},
		signinpb.SignInService_UpdateEmailAddress_FullMethodName:       {},
		signinpb.SignInService_UpdateUsername_FullMethodName:           {},
		signinpb.SignInAdministrationService_ListSignInsForUser_FullMethodName: {
			authorization.ReadAnySignInsPermission,
		},
		signinpb.SignInAdministrationService_EndSignInForUser_FullMethodName: {
			authorization.EndAnySignInsPermission,
		},
		signinpb.SignInAdministrationService_EndAllSignInsForUser_FullMethodName: {
			authorization.EndAnySignInsPermission,
		},
	}
}
