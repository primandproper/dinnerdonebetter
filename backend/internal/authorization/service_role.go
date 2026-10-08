package authorization

import (
	"context"
	"encoding/gob"
	"slices"

	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
)

func init() {
	gob.Register(serviceRoleCollection{})
}

const (
	// ServiceUserRoleName is the role every user is assigned at signup. It is
	// service-wide and grants nothing; an ordinary user's authority is
	// AccountMemberRoleName, held per account.
	ServiceUserRoleName = "service_user"
	// ServiceAdminRoleName is the role that can do essentially anything.
	ServiceAdminRoleName = "service_admin"
	// ServiceDataAdminRoleName administers the service's reference data.
	ServiceDataAdminRoleName = "service_data_admin"
)

type (
	// ServiceRolePermissionChecker checks permissions for one or more service Roles.
	ServiceRolePermissionChecker interface {
		HasPermission(Permission) bool

		IsServiceAdmin() bool
		CanImpersonateUsers() bool
	}

	serviceRoleCollection struct {
		// A nil set is a valid empty one, so a principal with no service-wide
		// authority needs no special case at any call site.
		Permissions *platformauthz.PermissionSet
		RoleNames   []string
	}
)

// NewServiceRolePermissionChecker returns a new checker from role names and a set of permissions.
func NewServiceRolePermissionChecker(roleNames []string, perms []Permission) ServiceRolePermissionChecker {
	return NewServiceRolePermissionCheckerFromSet(roleNames, platformauthz.NewPermissionSet(ToPlatformPermissions(perms)...))
}

// NewServiceRolePermissionCheckerFromSet returns a checker over an already-resolved
// permission set.
//
// This is what the session build uses: the policy resolver answers in a PermissionSet, so
// taking one avoids flattening it to a slice and rebuilding a map per request. The
// slice-taking constructor above remains for tests and for callers that hold a literal
// list.
func NewServiceRolePermissionCheckerFromSet(roleNames []string, perms *platformauthz.PermissionSet) ServiceRolePermissionChecker {
	return &serviceRoleCollection{
		Permissions: perms,
		RoleNames:   roleNames,
	}
}

// HasPermission returns whether a user can do something or not.
func (r serviceRoleCollection) HasPermission(p Permission) bool {
	return r.Permissions.Has(p)
}

// IsServiceAdmin returns if a role is an admin.
func (r serviceRoleCollection) IsServiceAdmin() bool {
	return slices.Contains(r.RoleNames, ServiceAdminRoleName)
}

// CanImpersonateUsers returns whether a user can impersonate others.
func (r serviceRoleCollection) CanImpersonateUsers() bool {
	return r.HasPermission(ImpersonateUserPermission)
}

// AdministrativeServiceRoleNames are the service roles that make somebody an operator: the ones
// the administrative sign-in door admits, and the ones a token from any other door does not
// carry.
func AdministrativeServiceRoleNames() []string {
	return []string{ServiceAdminRoleName, ServiceDataAdminRoleName}
}

// OrdinaryServiceRoles is what a token minted through an ordinary door carries of the service
// roles its holder has, in the shape platform's sign-in extractor takes (WithOrdinaryServiceRoles).
//
// The administrative roles are dropped: an operator's grants ride only on a token from the
// administrative door, which demands a proven second factor whatever the account's own policy
// says. What is kept is service_user — what a person holds about themselves, their privacy
// requests and their uploads — and anybody holding any service role holds that, because an
// operator is also a person, which is why service_admin inherits it. Dropping it with the rest
// would leave an operator signed in to the consumer app unable to export their own data.
//
// Somebody holding no service role at all is given none.
func OrdinaryServiceRoles(_ context.Context, held []string) []string {
	if len(held) == 0 {
		return nil
	}

	return []string{ServiceUserRoleName}
}
