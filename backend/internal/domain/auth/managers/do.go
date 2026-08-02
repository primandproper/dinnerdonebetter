package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	queuescfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v9/authentication/totp"
	"github.com/primandproper/platform-go/v9/messagequeue"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/qrcodes"
	"github.com/primandproper/platform-go/v9/random"

	"github.com/samber/do/v2"
)

// RegisterAuthManager registers the auth manager with the injector.
func RegisterAuthManager(i do.Injector) {
	do.Provide[AuthManagerInterface](i, func(i do.Injector) (AuthManagerInterface, error) {
		return ProvideAuthManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[auth.PasswordResetTokenDataManager](i),
			do.MustInvoke[auth.UserSessionDataManager](i),
			do.MustInvoke[identity.UserDataManager](i),
			do.MustInvoke[authentication.Authenticator](i),
			do.MustInvoke[totp.Verifier](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[random.Generator](i),
			do.MustInvoke[qrcodes.Builder](i),
			do.MustInvoke[*queuescfg.Config](i),
		)
	})
}
