package authorization

import (
	platformauthz "github.com/primandproper/platform-go/v13/authorization"
)

// This file bridges this package's permission model onto platform-go/v13's
// authorization package.
//
// The permissions themselves stay here: what this service can be asked to do is
// its own vocabulary, and no platform package supplies it. What moved out is
// resolution — which permissions a role name carries — because that was declared
// twice, once as the slices in permissions.go and once as INSERT statements in a
// migration, with nothing but a string-matching test between them. PlatformPolicy
// below is now the single declaration: the migrator seeds it, and
// authorization/database resolves against what it seeded.

// TablePrefix namespaces the policy tables authorization/database renders.
//
// The platform's default prefix is empty, which would render authz_roles and
// authz_permissions — names generic enough to collide in a database this
// application shares, and the same reason every other adopted store here carries
// one. Changing it renames tables, so it moves only with a migration.
const TablePrefix = "ddb"

// PermissionLister is implemented by this package's permission checkers so a
// resolved set can be handed to the platform without re-deriving it from roles.
//
// It is a separate interface rather than a method on the checker interfaces
// because those are mocked across the service tests, and widening them would
// churn every mock for a method only this bridge calls.
type PermissionLister interface {
	GrantedPermissions() *platformauthz.PermissionSet
}

// GrantedPermissions returns every permission this checker grants.
//
// It is not called Permissions because both concrete checkers already export a
// field by that name — the set the check itself reads.
func (r serviceRoleCollection) GrantedPermissions() *platformauthz.PermissionSet {
	return r.Permissions
}

// GrantedPermissions returns every permission this checker grants.
func (r accountRoleCollection) GrantedPermissions() *platformauthz.PermissionSet {
	return r.Permissions
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
// It is the declaration the policy tables are seeded from — see
// postgres/migrations — so it is not a second copy of the policy but the only
// one. Everything below is derived from the permission slices in this package
// and from the inheritance the roles have always had.
//
// Roles inherit. This used to be spelled out flat, on the reasoning that
// expressing inheritance would be a behavioral change disguised as a refactor,
// and that had it exactly backwards: the database has inherited since #1215
// seeded user_role_hierarchy, so flatness was the disguised change. It stayed
// invisible because the only consumer validated this table through
// static.NewResolver and then discarded the resolver. Expanded, the flat
// version understated service_admin by 210 permissions and account_admin by
// 126, and overstated service_user by 131.
//
// Two edges, both matching what the hierarchy rows said:
//
//   - account_admin inherits account_member. permissions_test.go has always
//     built an account admin as the concatenation of the two.
//   - service_admin inherits account_admin and service_data_admin. The second
//     is not hierarchy in the old schema — the meal planning migration granted
//     service_admin the data admin set directly, row for row — but inheritance
//     and a duplicated grant list resolve to the same set, and only one of them
//     can drift.
//
// service_user holds nothing, which is not an oversight. Every user is assigned
// it at signup and it is a service-wide assignment; the authority an ordinary
// user has is account_member, held per account. Granting account_member here
// would hand every account-scoped permission to every user unscoped by account.
func PlatformPolicy() []platformauthz.Role {
	return []platformauthz.Role{
		{
			Name:        ServiceAdminRoleName,
			Description: "service-wide administrator",
			Permissions: ToPlatformPermissions(ServiceAdminPermissions),
			Inherits:    []string{AccountAdminRoleName, ServiceDataAdminRoleName},
		},
		{
			Name:        ServiceDataAdminRoleName,
			Description: "administrator of the service's reference data",
			Permissions: ToPlatformPermissions(ServiceDataAdminPermissions),
		},
		{
			Name:        ServiceUserRoleName,
			Description: "an ordinary user of the service",
		},
		{
			Name:        AccountAdminRoleName,
			Description: "administrator of a single account",
			Permissions: ToPlatformPermissions(AccountAdminPermissions),
			Inherits:    []string{AccountMemberRoleName},
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

		// NewGrants drops nil and empty sets, so a principal with authority in
		// only one scope needs no special case here.
		sets = append(sets, lister.GrantedPermissions())
	}

	return platformauthz.NewGrants(sets...)
}
