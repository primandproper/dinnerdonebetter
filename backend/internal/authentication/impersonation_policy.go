package authentication

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// ErrImpersonationNotPermitted is an operator asking to act as somebody without the
// permission to.
var ErrImpersonationNotPermitted = platformerrors.New("operator may not impersonate users")

// NewImpersonationPolicy is this application's rule for who may act as somebody else, in the
// shape signin.WithImpersonationPolicy takes: an operator whose service roles grant
// authorization.ImpersonateUserPermission.
//
// The roles are resolved through the same policy every request's session is, so the grant an
// operator is refused here is the grant their session would show. Whether the operator and
// the subject are in good standing is platform's to check, and it does, before this runs.
func NewImpersonationPolicy(policy platformauthz.PolicyResolver) signin.ImpersonationPolicy {
	return func(ctx context.Context, operator, _ *platformidentity.User) error {
		if operator == nil {
			return ErrImpersonationNotPermitted
		}

		granted, err := policy.PermissionsForRoles(ctx, operator.ServiceRoles...)
		if err != nil {
			return platformerrors.Wrap(err, "resolving what an operator's roles grant")
		}

		if !authorization.NewServiceRolePermissionCheckerFromSet(operator.ServiceRoles, granted).CanImpersonateUsers() {
			return ErrImpersonationNotPermitted
		}

		return nil
	}
}
