package authentication

import (
	"context"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/messagequeue"
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

	// platform's sign-in orchestration, which the Manager below proves passwords through.
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
		return signin.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[Authenticator](i),
			do.MustInvoke[tokens.Issuer](i),
			signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
			signin.WithAdminServiceRoles(authorization.ServiceAdminRoleName),
			signin.WithTOTPVerifier(do.MustInvoke[totp.Verifier](i)),
			signin.WithLogger(do.MustInvoke[logging.Logger](i)),
			signin.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			signin.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[Manager](i, func(i do.Injector) (Manager, error) {
		return NewManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*queuescfg.Config](i),
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[*signin.Service](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[auth.SessionStore](i),
			do.MustInvoke[*authcfg.TokensConfig](i),
		)
	})
}

func ProvideHasher(authenticator Authenticator) Hasher {
	return authenticator
}
