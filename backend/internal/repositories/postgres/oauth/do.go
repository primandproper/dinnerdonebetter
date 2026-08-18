package oauth

import (
	"context"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainoauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterOAuthRepository registers the OAuth repository with the injector.
func RegisterOAuthRepository(i do.Injector) {
	do.Provide[domainoauth.Repository](i, func(i do.Injector) (domainoauth.Repository, error) {
		return ProvideOAuthRepository(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*dbcfg.Config](i),
			do.MustInvoke[database.Client](i),
		), nil
	})
}
