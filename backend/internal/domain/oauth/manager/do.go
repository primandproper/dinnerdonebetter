package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v4/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"
	"github.com/primandproper/platform-go/v4/random"

	"github.com/samber/do/v2"
)

// RegisterOAuth2Manager registers the OAuth2 manager with the injector.
func RegisterOAuth2Manager(i do.Injector) {
	do.Provide[OAuth2Manager](i, func(i do.Injector) (OAuth2Manager, error) {
		return NewOAuth2Manager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[random.Generator](i),
			do.MustInvoke[func(context.Context) (*sessions.ContextData, error)](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[oauth.Repository](i),
			do.MustInvoke[*msgconfig.QueuesConfig](i),
		)
	})
}
