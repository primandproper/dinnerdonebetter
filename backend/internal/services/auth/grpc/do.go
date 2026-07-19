package grpc

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/webauthn"
	webauthncfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/webauthn/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/branding"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/managers"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/featureflags"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthService registers the auth gRPC service with the injector.
func RegisterAuthService(i do.Injector) {
	do.Provide[*webauthncfg.Config](i, func(i do.Injector) (*webauthncfg.Config, error) {
		return ProvideSessionStoreConfig(do.MustInvoke[*config.APIServiceConfig](i)), nil
	})

	do.Provide[webauthn.Config](i, func(i do.Injector) (webauthn.Config, error) {
		return ProvidePasskeyConfig(do.MustInvoke[*config.APIServiceConfig](i)), nil
	})

	do.Provide[webauthn.SessionStore](i, func(i do.Injector) (webauthn.SessionStore, error) {
		return webauthncfg.ProvideSessionStore(
			do.MustInvoke[*webauthncfg.Config](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
		)
	})

	do.Provide[*webauthn.Service](i, func(i do.Injector) (*webauthn.Service, error) {
		return ProvidePasskeyService(
			do.MustInvoke[webauthn.Config](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[webauthn.SessionStore](i),
		)
	})

	do.Provide[AuthMethodPermissions](i, func(i do.Injector) (AuthMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[authsvc.AuthServiceServer](i, func(i do.Injector) (authsvc.AuthServiceServer, error) {
		return NewAuthService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
			do.MustInvoke[managers.AuthManagerInterface](i),
			do.MustInvoke[authentication.Manager](i),
			do.MustInvoke[featureflags.FeatureFlagManager](i),
			do.MustInvoke[*webauthn.Service](i),
		), nil
	})
}

// ProvideSessionStoreConfig extracts the session store config from the API service config.
func ProvideSessionStoreConfig(cfg *config.APIServiceConfig) *webauthncfg.Config {
	return &cfg.Auth.SessionStore
}

// ProvidePasskeyConfig extracts passkey config from the API service config.
// When RPID is empty (e.g. local dev), returns localhost defaults.
func ProvidePasskeyConfig(cfg *config.APIServiceConfig) webauthn.Config {
	p := cfg.Auth.Passkey
	if p.RPID == "" {
		return webauthn.Config{
			RPID:          "localhost",
			RPDisplayName: branding.CompanyName,
			RPOrigins:     []string{"https://localhost:8080", "http://localhost:8080"},
		}
	}
	origins := p.RPOrigins
	if len(origins) == 0 {
		origins = []string{}
	}
	return webauthn.Config{
		RPID:          p.RPID,
		RPDisplayName: orDefault(p.RPDisplayName, branding.CompanyName),
		RPOrigins:     origins,
	}
}

func orDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
