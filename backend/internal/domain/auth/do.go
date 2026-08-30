package auth

import "github.com/samber/do/v2"

// RegisterProviders registers auth domain providers with the injector.
func RegisterProviders(i do.Injector) {
	do.Provide[UserSessionDataManager](i, func(i do.Injector) (UserSessionDataManager, error) {
		return ProvideUserSessionDataManagerFromRepository(do.MustInvoke[Repository](i)), nil
	})
}

func ProvideUserSessionDataManagerFromRepository(r Repository) UserSessionDataManager {
	return r
}
