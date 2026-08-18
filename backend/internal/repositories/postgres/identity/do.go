package identity

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIdentityRepository registers the identity repository with the injector.
func RegisterIdentityRepository(i do.Injector) {
	do.Provide[domainidentity.Repository](i, func(i do.Injector) (domainidentity.Repository, error) {
		return ProvideIdentityRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		), nil
	})

	do.Provide[domainidentity.UserDataManager](i, func(i do.Injector) (domainidentity.UserDataManager, error) {
		return ProvideUserDataManager(do.MustInvoke[domainidentity.Repository](i)), nil
	})
}

func ProvideUserDataManager(r domainidentity.Repository) domainidentity.UserDataManager {
	return r
}
