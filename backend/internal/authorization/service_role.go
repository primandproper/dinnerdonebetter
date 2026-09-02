package authorization

import (
	"encoding/gob"
	"slices"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
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

	invalidServiceRoleWarning = "INVALID_SERVICE_ROLE"

	// invalidServiceRole is a service role to apply for non-admin users to have one.
	invalidServiceRole ServiceRole = iota
	// ServiceUserRole is a service role to apply for non-admin users to have one.
	ServiceUserRole ServiceRole = iota
	// ServiceAdminRole is a role that allows a user to do basically anything.
	ServiceAdminRole ServiceRole = iota
)

type (
	// ServiceRole describes a role a user has for the Service context.
	ServiceRole role

	// ServiceRolePermissionChecker checks permissions for one or more service Roles.
	ServiceRolePermissionChecker interface {
		HasPermission(Permission) bool

		AsAccountRolePermissionChecker() AccountRolePermissionsChecker
		IsServiceAdmin() bool
		CanUpdateUserAccountStatuses() bool
		CanImpersonateUsers() bool
		CanManageUserSessions() bool
	}

	serviceRoleCollection struct {
		// A nil set is a valid empty one, so a principal with no service-wide
		// authority needs no special case at any call site.
		Permissions *platformauthz.PermissionSet
		RoleNames   []string
	}
)

func (r ServiceRole) String() string {
	switch r {
	case invalidServiceRole:
		return invalidServiceRoleWarning
	case ServiceUserRole:
		return ServiceUserRoleName
	case ServiceAdminRole:
		return ServiceAdminRoleName
	default:
		return ""
	}
}

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

func (r serviceRoleCollection) AsAccountRolePermissionChecker() AccountRolePermissionsChecker {
	return NewAccountRolePermissionCheckerFromSet(r.RoleNames, r.Permissions)
}

// HasPermission returns whether a user can do something or not.
func (r serviceRoleCollection) HasPermission(p Permission) bool {
	return r.Permissions.Has(platformauthz.Permission(p))
}

// IsServiceAdmin returns if a role is an admin.
func (r serviceRoleCollection) IsServiceAdmin() bool {
	return slices.Contains(r.RoleNames, ServiceAdminRoleName)
}

// CanUpdateUserAccountStatuses returns whether a user can update user account statuses.
func (r serviceRoleCollection) CanUpdateUserAccountStatuses() bool {
	return r.HasPermission(UpdateUserStatusPermission)
}

// CanImpersonateUsers returns whether a user can impersonate others.
func (r serviceRoleCollection) CanImpersonateUsers() bool {
	return r.HasPermission(ImpersonateUserPermission)
}

// CanManageUserSessions returns whether a user can manage other users' sessions.
func (r serviceRoleCollection) CanManageUserSessions() bool {
	return r.HasPermission(ManageUserSessionsPermission)
}
