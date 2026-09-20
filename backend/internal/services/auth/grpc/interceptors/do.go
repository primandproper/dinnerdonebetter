package interceptors

import (
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"

	oauth2servercfg "github.com/primandproper/platform-go/v14/authentication/oauth2serverstore/config"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthInterceptor registers the auth interceptor with the injector.
func RegisterAuthInterceptor(i do.Injector) {
	do.Provide[*AuthInterceptor](i, func(i do.Injector) (*AuthInterceptor, error) {
		return ProvideAuthInterceptor(
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*identitybuild.SessionBuilder](i),
			do.MustInvoke[auth.SessionStore](i),
			do.MustInvoke[*oauth2server.Server](i),
			resourceIdentifier(do.MustInvoke[*oauth2servercfg.Config](i)),
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[MethodPermissionsMap](i),
		), nil
	})
}

// resourceIdentifier is the RFC 8707 name this server answers to, for the audience check.
//
// The first configured resource, falling back to the issuer: a deployment where the
// authorization server and the resource server are the same process — which this one is —
// names itself once, and Resources exists for the case where they are not.
func resourceIdentifier(cfg *oauth2servercfg.Config) string {
	if len(cfg.Resources) > 0 {
		return cfg.Resources[0]
	}

	return cfg.Issuer
}
