package auth

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	domainauth "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthRepository registers the auth repository with the injector.
func RegisterAuthRepository(i do.Injector) {
	do.Provide[domainauth.Repository](i, func(i do.Injector) (domainauth.Repository, error) {
		return ProvideAuthRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		), nil
	})

	do.Provide[domainauth.PasswordResetTokenDataManager](i, func(i do.Injector) (domainauth.PasswordResetTokenDataManager, error) {
		return ProvidePasswordResetTokenDataManager(do.MustInvoke[domainauth.Repository](i)), nil
	})

	do.Provide[domainauth.UserSessionDataManager](i, func(i do.Injector) (domainauth.UserSessionDataManager, error) {
		return ProvideUserSessionDataManager(do.MustInvoke[domainauth.Repository](i)), nil
	})
}

func ProvidePasswordResetTokenDataManager(r domainauth.Repository) domainauth.PasswordResetTokenDataManager {
	return r
}

func ProvideUserSessionDataManager(r domainauth.Repository) domainauth.UserSessionDataManager {
	return r
}
