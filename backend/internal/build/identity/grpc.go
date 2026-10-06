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
is SetUserRequiresPasswordChange under another name — the fourth operator write,
which platform had built at every layer but the surface until this adoption asked
for it.

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
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/succession"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitygrpc "github.com/primandproper/platform-go/v15/identity/grpc"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
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

		// No WithInvitationTTL. platform's defaults stand — seven days for an invitation a
		// client sends no expiry with, and the cap above it — and that is a ruling rather
		// than an omission: identity/config exists to make both environment-settable and
		// is refused here, because a knob this application configures belongs in its own
		// config tree beside the others rather than in a second one. See
		// docs/configuration.md.
		server, err := identitygrpc.NewServer(
			do.MustInvoke[*platformidentity.Service](i),
			store,
			client,
			sessions.PrincipalFromContext,
			identitygrpc.WithTargetAuthorizer(memberships),
			// Operators share no account with the people they act on, so the membership rule
			// refuses them. These let a holder of identity.directory.read_any or act_any past
			// it, and record every such admission in the audit log before the call proceeds.
			// The same grants decide whether a page honors include_archived.
			identitygrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			identitygrpc.WithOperatorRecorder(do.MustInvoke[platformaudit.Recorder](i)),
			// Named rather than left to platform's defaults, though today they are the same two
			// strings. The operator grants are the ones IdentityOperatorPermissions gives a
			// service admin, and naming them here is what keeps the permission the role grid
			// grants and the one this server checks a single declaration: a rename upstream, or
			// one here, moves both rather than leaving every operator refused by a server
			// reading a name nobody holds.
			identitygrpc.WithOperatorPermission(authorization.PermissionOperatorRead, authorization.PermissionOperatorAct),
			// GetPrincipal answers what the caller may do in the account it resolved, off the
			// same policy the session's grants are resolved from, so a client can shape its
			// screens without guessing at roles.
			identitygrpc.WithPermissionResolver(do.MustInvoke[platformauthz.PolicyResolver](i)),
			identitygrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			identitygrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			identitygrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
		if err != nil {
			return nil, err
		}

		rule, err := succession.New(store)
		if err != nil {
			return nil, err
		}

		// Archiving a user settles the households they own first — see archival.go.
		return &settlesAccountsOnArchival{
			IdentityServiceServer: server,
			directory:             do.MustInvoke[*platformidentity.Service](i),
			rule:                  rule,
			db:                    client,
			logger:                do.MustInvoke[logging.Logger](i),
			tracer:                tracing.NewNamedTracer(do.MustInvoke[tracing.Provider](i), "identity_archival"),
		}, nil
	})
}

// Permissions is platform's fragment as this application's table spells it: platform's map, and
// the eight RPCs it deliberately leaves out of it declared public, which is what platform's own
// Require declares them as.
//
// platform gates every RPC whose subject is somebody else and declines to gate the ones whose
// subject is the caller — a grant on the method cannot say "only about yourself", so it leaves
// that decision to the consumer rather than inventing a permission that would be a lie. An
// empty slice is how this application's table says "no permission", and the authorization
// enforcer reads it as public.
//
// It is the fragment unamended. What this deployment changes about it is PermissionOverrides.
func Permissions() map[string][]authorization.Permission {
	out := identitygrpc.Permissions()

	for _, method := range identitygrpc.SelfServiceMethods() {
		out[method] = []authorization.Permission{}
	}

	return out
}

// PermissionOverrides gates the eight self-service RPCs behind grants of this application's own,
// where platform's fragment leaves them public.
//
// They are granted to an account member, which is everybody — registration mints an account.
// See internal/authorization for the grants and waitlists' build package for the same
// arrangement. The registration path is not here at all: signing up happens on the auth
// surface, which is where a caller with no session can reach it.
//
// platform's own suggestion is to leave them Public, which this application could also do.
// Naming them instead is the arrangement waitlists took for ListSignupsForSubject, and it keeps
// the policy tables answering "what may a member do" in one place rather than two.
//
// The eight are identitygrpc.SelfServiceMethods(), and grpc_test.go asserts that this map
// overrides exactly that set. They go through RequirementsBuilder.Override rather than into
// Permissions, so a method platform stops declaring fails the requirements build instead of
// being quietly declared here under a name nothing serves.
func PermissionOverrides() map[string][]authorization.Permission {
	return map[string][]authorization.Permission{
		identitypb.IdentityService_UpdateProfile_FullMethodName: {
			authorization.UpdateOwnProfilePermission,
		},
		identitypb.IdentityService_RecordAgreement_FullMethodName: {
			authorization.RecordOwnAgreementPermission,
		},
		identitypb.IdentityService_GetPrincipal_FullMethodName: {
			authorization.ReadOwnPrincipalPermission,
		},
		identitypb.IdentityService_SetDefaultAccount_FullMethodName: {
			authorization.SetOwnDefaultAccountPermission,
		},
		identitypb.IdentityService_AcceptInvitation_FullMethodName: {
			authorization.AnswerOwnInvitationsPermission,
		},
		identitypb.IdentityService_RejectInvitation_FullMethodName: {
			authorization.AnswerOwnInvitationsPermission,
		},
		identitypb.IdentityService_ListInvitationsFromUser_FullMethodName: {
			authorization.ReadOwnInvitationsPermission,
		},
		identitypb.IdentityService_ListInvitationsForEmailAddress_FullMethodName: {
			authorization.ReadOwnInvitationsPermission,
		},
	}
}
