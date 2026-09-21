/*
Package oauth2clients mounts platform-go's registered-client surface.

Four RPCs, and no service of this application's own. All four were renames of
the deleted service's — CreateOAuth2Client, GetOAuth2Client, ListOAuth2Clients
and ArchiveOAuth2Client — and the store behind them gained two columns this
application had no way to express: the registry a client is in, and the person
it belongs to.

Both are left empty here, which is not a gap. A registration this deployment
mints is an operator's: minted to let an application speak for the service on
behalf of whoever signs in, admitting any subject, which is exactly what
oauth2clients.Client.Admits calls the global registry naming no owner. That is
the arrangement the deleted table had — it had no column for anything else — so
adopting the wider one preserves the behavior and leaves the narrower
arrangements available to whoever wants them later.

There is no Update on the wire, and platform's store has one. The four RPCs are
the four the deleted service had, and a revision cannot rotate a secret anyway
— UpdateInput has no field for one — so the RPC that is missing is the one a
console would want least.
*/
package oauth2clients

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	oauth2clientsgrpc "github.com/primandproper/platform-go/v14/authentication/oauth2clients/grpc"
	"github.com/primandproper/platform-go/v14/authentication/oauth2clients/oauth2clientspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterOAuth2ClientsService registers platform's client registry surface with the injector.
//
// The global principal, not the account-scoped one. A registration is not an account's
// object — it is the deployment's, which is what the empty scope means — so handing this
// the account-scoped principal would file every client under whichever account the operator
// happened to be looking at, and make it invisible from the next one.
func RegisterOAuth2ClientsService(i do.Injector) {
	do.Provide[oauth2clientspb.OAuth2ClientsServiceServer](i, func(i do.Injector) (oauth2clientspb.OAuth2ClientsServiceServer, error) {
		return oauth2clientsgrpc.NewServer(
			do.MustInvoke[*platformoauth2clients.Service](i),
			do.MustInvoke[platformoauth2clients.Store](i),
			do.MustInvoke[database.Client](i),
			sessions.PrincipalFromContext,
			oauth2clientsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			oauth2clientsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			oauth2clientsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// Permissions is platform's map, unamended.
//
// Three grants over four methods, and this deployment has nothing to add to them: there are
// no public RPCs here — every one of the four is an administrative act — so unlike waitlists
// there is no method platform deliberately omits that this application has to declare.
func Permissions() map[string][]authorization.Permission {
	return oauth2clientsgrpc.Permissions()
}
