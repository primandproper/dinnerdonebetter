package authentication

import (
	"context"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	"github.com/primandproper/primitives-go/v2/observability/logging"
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

	do.Provide[Manager](i, func(i do.Injector) (Manager, error) {
		return NewManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*queuescfg.Config](i),
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[Authenticator](i),
			do.MustInvoke[totp.Verifier](i),
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
