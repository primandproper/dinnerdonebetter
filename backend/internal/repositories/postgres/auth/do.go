package auth

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v13/authentication/passwordreset"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthRepository registers the auth repository with the injector.
func RegisterAuthRepository(i do.Injector) {
	do.Provide[domainauth.Repository](i, func(i do.Injector) (domainauth.Repository, error) {
		return ProvideAuthRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		), nil
	})

	RegisterPasswordResetTokenSQLStore(i)

	do.Provide[passwordreset.Store](i, func(i do.Injector) (passwordreset.Store, error) {
		return ProvidePasswordResetTokenStore(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		)
	})

	do.Provide[domainauth.UserSessionDataManager](i, func(i do.Injector) (domainauth.UserSessionDataManager, error) {
		return ProvideUserSessionDataManager(do.MustInvoke[domainauth.Repository](i)), nil
	})
}

func ProvideUserSessionDataManager(r domainauth.Repository) domainauth.UserSessionDataManager {
	return r
}

// RegisterPasswordResetTokenSQLStore registers the unwrapped platform store with the injector.
//
// It is separate from RegisterAuthRepository because the db-cleaner job wants this and nothing
// else in that package: its container has a database client and no audit repository, and a
// sweep is not an auditable event — there is no actor and no subject, only rows past their
// deadline.
func RegisterPasswordResetTokenSQLStore(i do.Injector) {
	do.Provide[*passwordreset.SQLStore](i, func(i do.Injector) (*passwordreset.SQLStore, error) {
		return ProvidePasswordResetTokenSQLStore(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
