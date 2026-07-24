package identity

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	domainidentity "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v6/database"
	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIdentityRepository registers the identity repository with the injector.
func RegisterIdentityRepository(i do.Injector) {
	do.Provide[domainidentity.Repository](i, func(i do.Injector) (domainidentity.Repository, error) {
		return ProvideIdentityRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		), nil
	})

	do.Provide[domainidentity.UserDataManager](i, func(i do.Injector) (domainidentity.UserDataManager, error) {
		return ProvideUserDataManager(do.MustInvoke[domainidentity.Repository](i)), nil
	})
}

func ProvideUserDataManager(r domainidentity.Repository) domainidentity.UserDataManager {
	return r
}
