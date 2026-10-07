package authentication

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/devices"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformdevices "github.com/primandproper/platform-go/v15/authentication/signin/devices"
	"github.com/primandproper/platform-go/v15/authentication/signin/refreshtokens"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/notifications/mail"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
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
		// The one mailer every door here hands its mail to: platform's, which queues it on the
		// outbox for the mail Drainer the async message handler runs. See
		// internal/build/queuedmail.
		mailer := do.MustInvoke[*mail.QueuedMailer](i)

		// platform's own recording of every sign-in door's writes: the authentication, the
		// credential writes, the account switch, the revocations — each as an audit entry and
		// an event on the door's own transaction, filed under the user. An impersonation is
		// recorded as the subject's authentication with the operator as its Impersonator.
		// The Recorder is the one recordingcfg.Register provides, which is what every other
		// adopted store's hooks here are built over.
		recordingHooks, err := signin.NewRecordingHooks(do.MustInvoke[*platformrecording.Recorder](i))
		if err != nil {
			return nil, err
		}

		// And where each token was issued to, on the same transaction — what "where you're
		// signed in" shows beside each login — deleted again on the transaction that ends the
		// login. The table and the hooks are platform's; which parts of a request to trust is
		// this application's, and is internal/authentication/devices.
		hooks, err := platformdevices.NewHooks(recordingHooks, do.MustInvoke[platformdevices.Store](i), devices.Extract)
		if err != nil {
			return nil, err
		}

		return signin.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[Authenticator](i),
			do.MustInvoke[tokens.Issuer](i),
			// The role a registrant owns their account with. RegistrationPolicy names it too;
			// the service requires a default of its own whatever a policy does.
			[]string{authorization.AccountAdminRoleName},
			signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
			// Who the administrative door admits: every operator role, since a token from any
			// other door carries none of them — see authorization.OrdinaryServiceRoles.
			signin.WithAdminServiceRoles(authorization.AdministrativeServiceRoleNames()...),
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
			// And the half of it that needs the account: a change may not keep the password
			// it is changing. Platform enforces nothing here by default.
			signin.WithAccountPasswordPolicy(AccountPasswordPolicy),
			// And the rest of this application's registration — standing, roles, the second
			// factor, the agreements — applied to every registration SignInService.Register
			// writes, which is open to anybody: this policy is what stands in front of it.
			signin.WithRegistrationPolicy(RegistrationPolicy),
			// Where a sign-in through platform's SignInService keeps the refresh token that
			// outlives its access token. With a store named, LoginForToken mints a rotating
			// pair and the three refresh doors work; without one they refuse.
			//
			// The lifetimes are platform's defaults, deliberately: an hour for an access token
			// and thirty days for a sign-in, fifteen minutes and twelve hours for an
			// administrative one. No environment wants different ones, so there is no config
			// for them; WithTokenTTL and its siblings are where one would go. A sign-out does
			// not wait out a lifetime either way: every request checks the login its token
			// names through signin.Service.CheckSignIn.
			signin.WithRefreshTokenStore(do.MustInvoke[*refreshtokens.SQLStore](i)),
			// The two mails platform's own doors send — a verification link, at registration
			// and on request, and a reminder of somebody's username — are queued on the
			// outbox and rendered in this application's words by the mail Drainer.
			signin.WithVerificationMailer(mailer),
			// The profile write the two handle doors make. identity's service rather than its
			// store, so identity's AfterUpdateProfile hook records the change as it does for
			// every other profile write; each door asks the current password first.
			signin.WithProfileUpdater(do.MustInvoke[*platformidentity.Service](i)),
			signin.WithHandleReminderMailer(mailer),
			// Who may act as somebody else. Without a policy every impersonation is refused.
			signin.WithImpersonationPolicy(NewImpersonationPolicy(do.MustInvoke[platformauthz.PolicyResolver](i))),
			signin.WithHooks(hooks),
			signin.WithLogger(do.MustInvoke[logging.Logger](i)),
			signin.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			signin.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

func ProvideHasher(authenticator Authenticator) Hasher {
	return authenticator
}
