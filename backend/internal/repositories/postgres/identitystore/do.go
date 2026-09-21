package identitystore

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v14/authentication/passkeys"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIdentityStore registers platform's identity store and the service over it.
//
// The store is registered undecorated here, unlike every other adopted domain, and that is
// the seam being different rather than the rule being broken. A store decorator cannot
// record an identity operation — registering somebody is three writes in one transaction —
// so the recording is the Hooks the service is built with, and anything holding the store
// directly is holding it for reads.
func RegisterIdentityStore(i do.Injector) {
	RegisterPasskeyStore(i)

	do.Provide[platformidentity.Store](i, func(i do.Injector) (platformidentity.Store, error) {
		return platformidentity.NewSQLStore(
			do.MustInvoke[database.Client](i),
			platformidentity.WithTablePrefix(ddbidentity.TablePrefix),
			platformidentity.WithStoreLogger(do.MustInvoke[logging.Logger](i)),
			platformidentity.WithStoreTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformidentity.WithStoreMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[*Hooks](i, func(i do.Injector) (*Hooks, error) {
		return ProvideHooks(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*events.Emitter](i),
		), nil
	})

	do.Provide[*platformidentity.Service](i, func(i do.Injector) (*platformidentity.Service, error) {
		return platformidentity.NewService(
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformidentity.Store](i),
			// Without this every identity write is a write nothing recorded, and the
			// failure is silent: NoopHooks satisfies the interface.
			platformidentity.WithHooks(do.MustInvoke[*Hooks](i)),
			platformidentity.WithServiceLogger(do.MustInvoke[logging.Logger](i)),
			platformidentity.WithServiceTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformidentity.WithServiceMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// RegisterPasskeyStore registers the passkey credential store, under the same namespace as
// the directory: platform names its table webauthn_credentials too, and this deployment
// holds both.
//
// It lives here, beside the identity store, rather than with the passkey service that was
// its only caller. Two processes need it and only one of them serves a ceremony: the
// scheduler fulfills subject access requests, and an export owes the subject the list of
// devices that can sign in as them. A store registered where its service happens to be
// wired is a store the other process resolves at runtime and does not find — which `do`
// reports when the worker starts rather than when this file compiles.
func RegisterPasskeyStore(i do.Injector) {
	do.Provide[passkeys.Store](i, func(i do.Injector) (passkeys.Store, error) {
		return passkeys.NewSQLStore(
			do.MustInvoke[database.Client](i),
			passkeys.WithTablePrefix(ddbidentity.TablePrefix),
			passkeys.WithLogger(do.MustInvoke[logging.Logger](i)),
			passkeys.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			passkeys.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
