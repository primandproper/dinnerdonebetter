/*
Package passwordreset mounts platform-go's PasswordResetService: a link mailed to somebody who
cannot sign in, and the new password they choose with it.

Every RPC on it is anonymous, because a caller who could prove who they are would be changing
their password through SignInService instead.

What is this application's here is three things. The mail goes through the outbox, as an
event the data change message handler renders into the reset email — see
authentication.SignInMailers. The password is held to authentication.PasswordPolicy, the rule
every other door that writes a password answers to. And the password write queues the "your
password was reset" mail on the transaction that spends the link — see
authentication.PasswordResetDirectory.

The token store is the audited one internal/repositories/postgres/auth builds, so issuing,
spending and revoking a link is recorded as it was before the move.
*/
package passwordreset

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformpasswordreset "github.com/primandproper/platform-go/v14/authentication/passwordreset"
	passwordresetgrpc "github.com/primandproper/platform-go/v14/authentication/passwordreset/grpc"
	"github.com/primandproper/platform-go/v14/authentication/passwordreset/passwordresetpb"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
)

// tokenLifetime is how long a reset link is good for: long enough to walk to a phone, and short
// enough that a link left in an inbox is not a standing key to the account.
const tokenLifetime = 30 * time.Minute

// RegisterPasswordResetService registers the reset flow and platform's PasswordResetService
// server with the injector.
func RegisterPasswordResetService(i do.Injector) {
	do.Provide[*platformpasswordreset.Service](i, func(i do.Injector) (*platformpasswordreset.Service, error) {
		logger := do.MustInvoke[logging.Logger](i)
		db := do.MustInvoke[database.Client](i)
		emitter := do.MustInvoke[*events.Emitter](i)

		return platformpasswordreset.NewService(
			db,
			do.MustInvoke[platformpasswordreset.Store](i),
			authentication.NewPasswordResetDirectory(logger, do.MustInvoke[platformidentity.Store](i), emitter),
			do.MustInvoke[authentication.Authenticator](i),
			authentication.NewSignInMailers(logger, db, emitter),
			platformpasswordreset.WithTokenLifetime(tokenLifetime),
			platformpasswordreset.WithPasswordPolicy(authentication.PasswordPolicy),
			platformpasswordreset.WithServiceLogger(logger),
			platformpasswordreset.WithServiceTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformpasswordreset.WithServiceMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[passwordresetpb.PasswordResetServiceServer](i, func(i do.Injector) (passwordresetpb.PasswordResetServiceServer, error) {
		return passwordresetgrpc.NewServer(
			do.MustInvoke[*platformpasswordreset.Service](i),
			// The server's default is the global scope, which is this directory's scope too. It
			// is named anyway, so that the day this application's directory is scoped, a reset
			// follows it.
			passwordresetgrpc.WithScopeResolver(func(context.Context) (tenancy.Scope, error) {
				return tenancy.Global(), nil
			}),
			passwordresetgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			passwordresetgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			passwordresetgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// AnonymousMethods are every RPC on PasswordResetService.
func AnonymousMethods() []string {
	return passwordresetgrpc.AnonymousMethods()
}
