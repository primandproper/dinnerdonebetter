package authentication

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/platform-go/v14/authentication/signin/refreshtokens"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuth registers authentication providers with the injector.
func RegisterAuth(i do.Injector) {
	do.Provide[Authenticator](i, func(i do.Injector) (Authenticator, error) {
		return NewArgon2Authenticator(
			argon2.WithLogger(do.MustInvoke[logging.Logger](i)),
			argon2.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		), nil
	})

	do.Provide[Hasher](i, func(i do.Injector) (Hasher, error) {
		return ProvideHasher(do.MustInvoke[Authenticator](i)), nil
	})

	do.Provide[totp.Verifier](i, func(i do.Injector) (totp.Verifier, error) {
		return totp.NewVerifier(totp.WithTracerProvider(do.MustInvoke[tracing.Provider](i))), nil
	})

	// platform's sign-in orchestration, which every door that signs somebody in goes through.
	//
	// It takes the identity store as its Directory and this application's token issuer as
	// its TokenIssuer, both without an adapter: identity.Store satisfies signin.Directory
	// outright, and tokens.Issuer already has IssueToken's shape. Nothing here is a
	// translation layer, which is most of why the adoption was worth making.
	//
	// The second-factor policy is the default, SecondFactorWhenEnrolled, and is stated
	// rather than left implicit because it is this application's rule and not an accident
	// of a zero value: every user is issued a TOTP secret at registration and is asked for
	// a code only once they have proven it. The administrative door ignores that policy
	// and demands a proven second factor whatever it says, which is platform's rule and
	// the one this application already enforced by hand.
	do.Provide[*signin.Service](i, func(i do.Injector) (*signin.Service, error) {
		mailers := NewSignInMailers(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		)

		return signin.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[Authenticator](i),
			do.MustInvoke[tokens.Issuer](i),
			// The role a registrant owns their account with. RegistrationPolicy names it too;
			// the service requires a default of its own whatever a policy does.
			[]string{authorization.AccountAdminRoleName},
			signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
			signin.WithAdminServiceRoles(authorization.ServiceAdminRoleName),
			signin.WithTOTPVerifier(do.MustInvoke[totp.Verifier](i)),
			// The label an authenticator app shows beside the code, which has to match
			// what the QR builder encodes or a re-enrollment renames the entry.
			signin.WithTOTPIssuer(branding.CompanyName),
			// The registrar signing somebody up runs through. identity.Service satisfies
			// signin.Registrar outright, so registration is one call that hashes the
			// password, mints the verification token and writes the user, the account and
			// the membership on one transaction.
			signin.WithRegistrar(do.MustInvoke[*platformidentity.Service](i)),
			// And the verification reads and writes an emailed link is answered through.
			// The identity store satisfies this one outright too.
			signin.WithVerifications(do.MustInvoke[platformidentity.Store](i)),
			// This application's password rule, applied to every password the sign-in
			// service writes, so no door mounted from platform accepts a password this
			// application refuses.
			signin.WithPasswordPolicy(PasswordPolicy),
			// And the rest of this application's registration — standing, roles, the second
			// factor, the agreements — applied to every registration SignInService.Register
			// writes, which is open to anybody: this policy is what stands in front of it.
			signin.WithRegistrationPolicy(RegistrationPolicy),
			// Where a sign-in through platform's SignInService keeps the refresh token that
			// outlives its access token. With a store named, LoginForToken mints a rotating
			// pair and the three refresh doors work; without one they refuse.
			//
			// The lifetimes are platform's defaults: an hour for an access token and thirty
			// days for a sign-in, fifteen minutes and twelve hours for an administrative one.
			// This application's own TokensConfig lifetimes are not reused — they are unset in
			// every environment, which issues tokens with no lifetime at all, and an access
			// token here is not checked against anything but its signature, so its lifetime
			// is how long a sign-out takes to take effect.
			signin.WithRefreshTokenStore(do.MustInvoke[*refreshtokens.SQLStore](i)),
			// The two mails platform's own doors send — another verification link, and a
			// reminder of somebody's username — go through this application's outbox and
			// its data change message handler, which renders the email.
			signin.WithVerificationMailer(mailers),
			// The profile write the two handle doors make. identity's service rather than its
			// store, so identity's AfterUpdateProfile hook records the change as it does for
			// every other profile write; each door asks the current password first.
			signin.WithProfileUpdater(do.MustInvoke[*platformidentity.Service](i)),
			signin.WithHandleReminderMailer(mailers),
			// Who may act as somebody else. Without a policy every impersonation is refused.
			signin.WithImpersonationPolicy(NewImpersonationPolicy(do.MustInvoke[platformauthz.PolicyResolver](i))),
			signin.WithHooks(NewSignInHooks(
				do.MustInvoke[logging.Logger](i),
				do.MustInvoke[*events.Emitter](i),
				recording.NewRecorder(
					tracing.NewNamedTracer(do.MustInvoke[tracing.Provider](i), "signin_hooks"),
					do.MustInvoke[audit.Repository](i),
					do.MustInvoke[*events.Emitter](i),
				),
			)),
			signin.WithLogger(do.MustInvoke[logging.Logger](i)),
			signin.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			signin.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

func ProvideHasher(authenticator Authenticator) Hasher {
	return authenticator
}
