/*
Package identity mounts platform-go's identity surface.

Users, accounts, memberships and invitations, and no service of this
application's own. The one that used to live in internal/services/identity
forwarded twenty-nine RPCs to a repository of eight thousand lines, and every
one of them has a counterpart here — under platform's names rather than this
application's, which is the part a client notices: CreateAccountInvitation is
Invite, GetUsersForAccount is ListAccountMembers, and the three RPCs that each
changed one field of a user are one UpdateProfile.

Two of the deleted service's had no counterpart and did not want one.
UploadUserAvatar is an upload, and avatars are not in platform's User at all —
correctly, since what this application stores is a row in the uploads registry
and a reference, so it belongs on the media surface. AdminSetPasswordChangeRequired
is Service.SetUserRequiresPasswordChange under another name.

What this application still owns is the recording. An identity operation writes
several rows in one transaction, so the audit entry and the outbox row go in
through identity.Hooks rather than around a store — see
internal/repositories/postgres/identitystore, and the spike that proved it
commits and rolls back as one.
*/
package identity

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	identitygrpc "github.com/primandproper/platform-go/v14/identity/grpc"
	"github.com/primandproper/platform-go/v14/identity/identitypb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIdentityService registers platform's identity surface with the injector.
//
// The global principal, not the account-scoped one. A user is not an account's object and
// neither is an invitation into one; the account a request acts on is named by the request
// and checked by the authorizer below, which is a live membership rather than whatever the
// session says is active.
func RegisterIdentityService(i do.Injector) {
	do.Provide[identitypb.IdentityServiceServer](i, func(i do.Injector) (identitypb.IdentityServiceServer, error) {
		client := do.MustInvoke[database.Client](i)
		store := do.MustInvoke[platformidentity.Store](i)

		memberships, err := identitygrpc.NewMembershipAuthorizer(client, store)
		if err != nil {
			return nil, err
		}

		return identitygrpc.NewServer(
			do.MustInvoke[*platformidentity.Service](i),
			store,
			client,
			sessions.PrincipalFromContext,
			identitygrpc.WithTargetAuthorizer(&operatorOrMember{inner: memberships}),
			identitygrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			identitygrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			identitygrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// Permissions is platform's map, unamended.
//
// Every one of these RPCs is behind a grant upstream, so unlike waitlists there is no
// method platform deliberately leaves public that this application has to declare. The
// registration path is not here at all — signing up happens on the auth surface, which is
// where a caller with no session can reach it.
func Permissions() map[string][]authorization.Permission {
	return identitygrpc.Permissions()
}
