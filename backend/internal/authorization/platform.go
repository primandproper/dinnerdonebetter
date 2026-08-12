package authorization

import (
	platformauthz "github.com/primandproper/platform-go/v10/authorization"
)

// This file bridges this package's hand-rolled permission model onto
// platform-go/v10's authorization package.
//
// The bridge exists so the platform enforcer can be run in audit-only mode
// beside the existing checks: it evaluates every call against the same policy
// and records where the two disagree, without denying anything. That is the
// package's documented way to move enforcement onto it — turning it on across a
// service that already has a large hand-written permission table is otherwise a
// coin flip.
//
// Until the audit is clean, this package remains the source of truth. The Roles
// table below is derived from the same permission slices the checkers use, so
// the two cannot drift.

// Role names as the platform policy knows them. They match the strings this
// package already assigns to roles, so a session's role names resolve without
// translation.
const (
	ServiceAdminRoleName     = serviceAdminRoleName
	ServiceDataAdminRoleName = serviceDataAdminRoleName
	ServiceUserRoleName      = serviceUserRoleName
)

// PermissionLister is implemented by this package's permission checkers so a
// resolved set can be handed to the platform without re-deriving it from roles.
//
// It is a separate interface rather than a method on the checker interfaces
// because those are mocked across the service tests, and widening them would
// churn every mock for a method only this bridge calls.
type PermissionLister interface {
	GrantedPermissions() []Permission
}

// GrantedPermissions returns every permission this checker grants.
//
// It is not called Permissions because both concrete checkers already export a
// field by that name — the map the check itself reads.
func (r serviceRoleCollection) GrantedPermissions() []Permission {
	out := make([]Permission, 0, len(r.Permissions))
	for p := range r.Permissions {
		out = append(out, p)
	}

	return out
}

// GrantedPermissions returns every permission this checker grants.
func (r accountRoleCollection) GrantedPermissions() []Permission {
	out := make([]Permission, 0, len(r.Permissions))
	for p := range r.Permissions {
		out = append(out, p)
	}

	return out
}

// ToPlatformPermissions converts this package's permissions to the platform's.
//
// The two types are both string aliases and the values are identical; they are
// distinct types only because each package declares its own.
func ToPlatformPermissions(perms []Permission) []platformauthz.Permission {
	out := make([]platformauthz.Permission, len(perms))
	for i, p := range perms {
		out[i] = platformauthz.Permission(p)
	}

	return out
}

// PlatformPolicy returns the role table as the platform's authorization package
// models it.
//
// It is built from the same slices NewServiceRolePermissionChecker and
// NewAccountRolePermissionChecker are handed, so a role's meaning cannot differ
// between the two systems. Roles are flat rather than inheriting: this package
// has always spelled each role's permissions out in full, and expressing that
// as inheritance would be a behavioral change disguised as a refactor.
func PlatformPolicy() []platformauthz.Role {
	return []platformauthz.Role{
		{
			Name:        ServiceAdminRoleName,
			Description: "service-wide administrator",
			Permissions: ToPlatformPermissions(ServiceAdminPermissions),
		},
		{
			Name:        ServiceDataAdminRoleName,
			Description: "administrator of the service's reference data",
			Permissions: ToPlatformPermissions(ServiceDataAdminPermissions),
		},
		{
			Name:        ServiceUserRoleName,
			Description: "an ordinary user of the service",
			Permissions: ToPlatformPermissions(AccountMemberPermissions),
		},
		{
			Name:        AccountAdminRoleName,
			Description: "administrator of a single account",
			Permissions: ToPlatformPermissions(AccountAdminPermissions),
		},
		{
			Name:        AccountMemberRoleName,
			Description: "member of a single account",
			Permissions: ToPlatformPermissions(AccountMemberPermissions),
		},
	}
}

// PlatformGrants converts a session's checkers into the platform's Grants.
//
// Service-wide and per-account authority are handed over as separate sets and
// unioned by the platform, which is exactly how this package already treats
// them: a permission held in either place is held.
func PlatformGrants(service, account any) platformauthz.Grants {
	sets := make([]*platformauthz.PermissionSet, 0, 2)

	for _, checker := range []any{service, account} {
		lister, ok := checker.(PermissionLister)
		if !ok || lister == nil {
			continue
		}

		if perms := lister.GrantedPermissions(); len(perms) > 0 {
			sets = append(sets, platformauthz.NewPermissionSet(ToPlatformPermissions(perms)...))
		}
	}

	return platformauthz.NewGrants(sets...)
}
