package auth

import (
	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/platform-go/v14/authentication/signin/refreshtokens"
	sessionsdatabase "github.com/primandproper/platform-go/v14/sessions/database"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthRepository registers the auth repository with the injector.
func RegisterAuthRepository(i do.Injector) {
	RegisterPasswordResetTokenSQLStore(i)
	RegisterRefreshTokenSQLStore(i)
	RegisterUserSessionBackend(i)

	do.Provide[passwordreset.Store](i, func(i do.Injector) (passwordreset.Store, error) {
		return ProvidePasswordResetTokenStore(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		)
	})

	do.Provide[domainauth.SessionStore](i, func(i do.Injector) (domainauth.SessionStore, error) {
		return ProvideUserSessionStore(
			do.MustInvoke[*authcfg.SessionsConfig](i),
			do.MustInvoke[*sessionsdatabase.Backend[domainauth.SessionPayload]](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
		)
	})
}

// RegisterUserSessionBackend registers the unwrapped platform backend with the injector.
//
// It is separate from RegisterAuthRepository for the same reason the password reset store
// is: the db-cleaner job wants this and nothing else in this package. Its container has a
// database client and no audit repository, and a sweep is not an auditable event — there
// is no actor and no subject, only rows past their deadline.
func RegisterUserSessionBackend(i do.Injector) {
	do.Provide[*sessionsdatabase.Backend[domainauth.SessionPayload]](i, func(i do.Injector) (*sessionsdatabase.Backend[domainauth.SessionPayload], error) {
		return ProvideUserSessionBackend(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
	})
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

// RegisterRefreshTokenSQLStore registers the platform's refresh token store with the injector.
//
// It is separate from RegisterAuthRepository for the reason the password reset store is: the
// db-cleaner job sweeps it, and that job's container has nothing else from this package.
func RegisterRefreshTokenSQLStore(i do.Injector) {
	do.Provide[*refreshtokens.SQLStore](i, func(i do.Injector) (*refreshtokens.SQLStore, error) {
		return ProvideRefreshTokenSQLStore(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
