package interceptors

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"

	"github.com/primandproper/platform-go/v8/authentication/tokens"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/samber/do/v2"
)

// RegisterAuthInterceptor registers the auth interceptor with the injector.
func RegisterAuthInterceptor(i do.Injector) {
	do.Provide[*AuthInterceptor](i, func(i do.Injector) (*AuthInterceptor, error) {
		return ProvideAuthInterceptor(
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
			do.MustInvoke[auth.Repository](i),
			do.MustInvoke[*manage.Manager](i),
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[MethodPermissionsMap](i),
		), nil
	})
}
