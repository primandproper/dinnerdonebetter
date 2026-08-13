package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v10/messagequeue"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	"github.com/primandproper/platform-go/v10/random"

	"github.com/samber/do/v2"
)

// RegisterOAuth2Manager registers the OAuth2 manager with the injector.
func RegisterOAuth2Manager(i do.Injector) {
	do.Provide[OAuth2Manager](i, func(i do.Injector) (OAuth2Manager, error) {
		return NewOAuth2Manager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[random.Generator](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[oauth.Repository](i),
			do.MustInvoke[*queuescfg.Config](i),
		)
	})
}
