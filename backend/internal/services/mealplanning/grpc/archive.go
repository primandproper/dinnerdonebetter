package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/filtering/filteringpb"
	filteringgrpc "github.com/primandproper/primitives-go/v2/filtering/grpc"
	"github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// Every list RPC here decodes its filter under one of two archive decisions:
//
//   - archivedIfHeld(ctx, <the noun's archive grant>), for a read whose page the grant covers in
//     full. That is the valid-* catalogs, whose archive grant is an operator's over every row, and
//     the reads confined to the caller's own account or the caller's own rows (meal plans and
//     everything under them, meal lists, ingredient preferences, instrument ownerships).
//   - filteringgrpc.ArchivedDenied, for a read across everybody's rows whose archive grant only
//     reaches the caller's own: recipes, meals, recipe lists, everything under a recipe, and
//     ratings. A member holds archive.recipes, but archiving is confined to recipes they own, so
//     the grant says nothing about anybody else's archived recipe. Meal plan tasks deny too: they
//     have no archive, so there is nothing for an allowance to buy.
//
// No surface passes ArchivedAllowed, so no caller's include_archived is honored without the
// noun's archive grant. archive_test.go holds every list RPC to its decision, and fails for one
// that is not listed.

// archivedIfHeld allows include_archived for a caller holding permission, from the service role
// or the active account's membership, which is the same reading the authorization interceptor
// gives a method's required permissions. A missing session holds nothing and is denied.
func archivedIfHeld(ctx context.Context, permission authorization.Permission) filteringgrpc.ArchiveDecision {
	session := sessions.FromContext(ctx)

	return filteringgrpc.ArchivedIf(
		session.ServiceRolePermissionChecker().HasPermission(permission) ||
			session.AccountRolePermissionsChecker().HasPermission(permission),
	)
}

// decodeQueryFilter decodes a list RPC's filter under the caller's archive decision, and records on
// the span when the decision took away archived rows the client asked for.
func decodeQueryFilter(span tracing.Span, in *filteringpb.QueryFilter, archived filteringgrpc.ArchiveDecision) (*filtering.QueryFilter, error) {
	filter, cleared, err := filteringgrpc.QueryFilterFromProto(in, archived)
	if cleared {
		tracing.AttachToSpan(span, keys.FilterIncludeArchivedClearedKey, true)
	}

	return filter, err
}
