package authentication

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v8/authentication/argon2"
	"github.com/primandproper/platform-go/v8/authentication/tokens"
	tokenscfg "github.com/primandproper/platform-go/v8/authentication/tokens/config"
	"github.com/primandproper/platform-go/v8/authentication/totp"
	"github.com/primandproper/platform-go/v8/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuth registers authentication providers with the injector.
func RegisterAuth(i do.Injector) {
	do.Provide[Authenticator](i, func(i do.Injector) (Authenticator, error) {
		return NewArgon2Authenticator(
			argon2.WithLogger(do.MustInvoke[logging.Logger](i)),
			argon2.WithTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
		), nil
	})

	do.Provide[Hasher](i, func(i do.Injector) (Hasher, error) {
		return ProvideHasher(do.MustInvoke[Authenticator](i)), nil
	})

	do.Provide[totp.Verifier](i, func(i do.Injector) (totp.Verifier, error) {
		return totp.NewVerifier(totp.WithTracerProvider(do.MustInvoke[tracing.TracerProvider](i))), nil
	})

	do.Provide[Manager](i, func(i do.Injector) (Manager, error) {
		return NewManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*msgconfig.QueuesConfig](i),
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[Authenticator](i),
			do.MustInvoke[totp.Verifier](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[auth.Repository](i),
			do.MustInvoke[*tokenscfg.Config](i),
		)
	})
}

func ProvideHasher(authenticator Authenticator) Hasher {
	return authenticator
}
