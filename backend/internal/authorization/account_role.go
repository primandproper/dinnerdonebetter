package authorization

import (
	"encoding/gob"

	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
)

type (
	// AccountRolePermissionsChecker checks permissions for one or more account Roles.
	AccountRolePermissionsChecker interface {
		HasPermission(Permission) bool
	}
)

const (
	// AccountAdminRoleName administers a single account.
	AccountAdminRoleName = "account_admin"
	// AccountMemberRoleName is ordinary membership of a single account.
	AccountMemberRoleName = "account_member"
)

type accountRoleCollection struct {
	// A nil set is a valid empty one, so an account a user is not a member of
	// needs no special case at the call sites that index this map.
	Permissions *platformauthz.PermissionSet
	RoleNames   []string
}

func init() {
	gob.Register(accountRoleCollection{})
}

// NewAccountRolePermissionChecker returns a new checker from a set of permissions.
func NewAccountRolePermissionChecker(perms []Permission) AccountRolePermissionsChecker {
	return NewAccountRolePermissionCheckerFromSet(nil, platformauthz.NewPermissionSet(ToPlatformPermissions(perms)...))
}

// NewAccountRolePermissionCheckerFromSet returns a checker over an already-resolved
// permission set. See NewServiceRolePermissionCheckerFromSet.
func NewAccountRolePermissionCheckerFromSet(roleNames []string, perms *platformauthz.PermissionSet) AccountRolePermissionsChecker {
	return &accountRoleCollection{
		Permissions: perms,
		RoleNames:   roleNames,
	}
}

// HasPermission returns whether a user can do something or not.
func (r accountRoleCollection) HasPermission(p Permission) bool {
	return r.Permissions.Has(p)
}
