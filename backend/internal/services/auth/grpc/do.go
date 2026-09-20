package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/managers"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/authentication/passkeys"
	webauthncfg "github.com/primandproper/platform-go/v14/authentication/webauthnsessions/config"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformwebauthn "github.com/primandproper/primitives-go/v2/authentication/webauthn"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/featureflags"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthService registers the auth gRPC service with the injector.
func RegisterAuthService(i do.Injector) {
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

	// The passkey credential store, under the same namespace as the directory: platform
	// names its table webauthn_credentials too, and this deployment holds both.
	do.Provide[passkeys.Store](i, func(i do.Injector) (passkeys.Store, error) {
		return passkeys.NewSQLStore(
			do.MustInvoke[database.Client](i),
			passkeys.WithTablePrefix(identity.TablePrefix),
			passkeys.WithLogger(do.MustInvoke[logging.Logger](i)),
			passkeys.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			passkeys.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[*webauthn.Service](i, func(i do.Injector) (*webauthn.Service, error) {
		return ProvidePasskeyService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*platformwebauthn.RelyingParty](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[passkeys.Store](i),
		)
	})

	do.Provide[AuthMethodPermissions](i, func(i do.Injector) (AuthMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[authsvc.AuthServiceServer](i, func(i do.Injector) (authsvc.AuthServiceServer, error) {
		return NewAuthService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[managers.AuthManagerInterface](i),
			do.MustInvoke[authentication.Manager](i),
			do.MustInvoke[featureflags.FeatureFlagManager](i),
			do.MustInvoke[*webauthn.Service](i),
		), nil
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

// ProvidePasskeyService creates a WebAuthn passkey service.
func ProvidePasskeyService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	relyingParty *platformwebauthn.RelyingParty,
	directory platformidentity.Store,
	db database.Client,
	credentials passkeys.Store,
) (*webauthn.Service, error) {
	return webauthn.NewService(
		logger,
		tracerProvider,
		relyingParty,
		credentials,
		db,
		&passkeyUserStore{directory: directory, db: db},
	)
}
