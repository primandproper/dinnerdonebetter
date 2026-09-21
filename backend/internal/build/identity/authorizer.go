package identity

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v14/callers"
	identitygrpc "github.com/primandproper/platform-go/v14/identity/grpc"
)

// operatorOrMember is platform's MembershipAuthorizer with a service administrator
// carve-out around it.
//
// The default answers from memberships alone and says in as many words why: it cannot see
// grants, the permission is on the method, and "a consumer whose operators need the
// directory implements this interface with their grants in hand". This deployment has such
// operators — banning a user, reading the directory, impersonating somebody to reproduce a
// report — and none of them shares an account with the person they are acting on.
//
// It composes rather than reimplements, which is the arrangement platform recommends and
// the one that keeps the three membership rules in one place. An administrator is permitted
// first and the default is asked otherwise, so the rules a member is subject to are still
// exactly platform's and there is no second copy of them to drift.
//
// The grant is still the first gate. Every one of these methods carries a permission that
// the authorization interceptor has already enforced by the time this runs, so this decides
// which rows an allowed call may touch, never whether the call is allowed — the division
// waitlists draws between PermissionReadSignups and AuthorizeSubjectRead, and audit draws
// between ReadAuditLogEntriesPermission and the operator read.
type operatorOrMember struct {
	inner identitygrpc.TargetAuthorizer
}

var _ identitygrpc.TargetAuthorizer = (*operatorOrMember)(nil)

// AuthorizeAccount permits an administrator any account, and otherwise defers.
func (a *operatorOrMember) AuthorizeAccount(ctx context.Context, caller callers.Principal, accountID string) error {
	if isServiceAdmin(ctx) {
		return nil
	}

	return a.inner.AuthorizeAccount(ctx, caller, accountID)
}

// AuthorizeUser permits an administrator any user, and otherwise defers.
func (a *operatorOrMember) AuthorizeUser(ctx context.Context, caller callers.Principal, userID string) error {
	if isServiceAdmin(ctx) {
		return nil
	}

	return a.inner.AuthorizeUser(ctx, caller, userID)
}

// AuthorizeInvitation permits an administrator any invitation, and otherwise defers.
func (a *operatorOrMember) AuthorizeInvitation(ctx context.Context, caller callers.Principal, invitationID string) error {
	if isServiceAdmin(ctx) {
		return nil
	}

	return a.inner.AuthorizeInvitation(ctx, caller, invitationID)
}

// isServiceAdmin reports whether the session holds the service administrator role.
func isServiceAdmin(ctx context.Context) bool {
	return sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin()
}
