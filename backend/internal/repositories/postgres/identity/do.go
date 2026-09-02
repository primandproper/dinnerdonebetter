package identity

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	"github.com/samber/do/v2"
)

// RegisterPolicyResolver registers what a role grants with the injector.
//
// It reads the policy tables the migrator seeds, so it is registered wherever the
// identity repository is: a session build resolves the roles a principal holds into the
// permissions they carry, and without this it has nothing to resolve them against.
func RegisterPolicyResolver(i do.Injector) {
	do.Provide[platformauthz.PolicyResolver](i, func(i do.Injector) (platformauthz.PolicyResolver, error) {
		return authorization.NewDatabaseResolver(
			do.MustInvoke[database.Client](i).Reader(),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}

// RegisterIdentityRepository registers the identity repository with the injector.
func RegisterIdentityRepository(i do.Injector) {
	do.Provide[domainidentity.Repository](i, func(i do.Injector) (domainidentity.Repository, error) {
		return ProvideIdentityRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
			do.MustInvoke[registry.Store](i),
			do.MustInvoke[platformauthz.PolicyResolver](i),
		), nil
	})

	do.Provide[domainidentity.UserDataManager](i, func(i do.Injector) (domainidentity.UserDataManager, error) {
		return ProvideUserDataManager(do.MustInvoke[domainidentity.Repository](i)), nil
	})
}

func ProvideUserDataManager(r domainidentity.Repository) domainidentity.UserDataManager {
	return r
}
