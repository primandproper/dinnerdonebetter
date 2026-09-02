package authorization

import (
	"encoding/gob"
	"slices"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
)

type (
	// AccountRole describes a role a user has for an account context.
	AccountRole role

	// AccountRolePermissionsChecker checks permissions for one or more account Roles.
	AccountRolePermissionsChecker interface {
		HasPermission(Permission) bool
	}
)

const (
	// AccountMemberRole is a role for a plain account participant.
	AccountMemberRole AccountRole = iota
	// AccountAdminRole is a role for someone who can manipulate the specifics of an account.
	AccountAdminRole AccountRole = iota

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

func (r AccountRole) String() string {
	switch r {
	case AccountMemberRole:
		return AccountMemberRoleName
	case AccountAdminRole:
		return AccountAdminRoleName
	default:
		return ""
	}
}

// HasPermission returns whether a user can do something or not.
func (r accountRoleCollection) HasPermission(p Permission) bool {
	return r.Permissions.Has(platformauthz.Permission(p))
}

// IsAccountAdmin returns whether a user is an account admin.
func (r accountRoleCollection) IsAccountAdmin() bool {
	return slices.Contains(r.RoleNames, AccountAdminRoleName)
}
