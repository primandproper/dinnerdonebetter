/*
Package waitlists mounts platform-go's waitlists surface.

The service this replaces forwarded fifteen RPCs, converted between two
spellings of the same fields, and applied one rule of its own: a signup is its
subject's, or a service administrator's. That rule is the SignupAuthorizer
below.

One RPC has no platform counterpart and needs none. WaitlistIsOpen read a list
and answered whether it was taking signups; openness is
`archived_at == null && closes_at > now`, and platform's GetList puts both
fields on the wire, so a client computes it from a read it was already making.
*/
package waitlists

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v14/callers"
	platformwaitlists "github.com/primandproper/platform-go/v14/waitlists"
	waitlistsgrpc "github.com/primandproper/platform-go/v14/waitlists/grpc"
	"github.com/primandproper/platform-go/v14/waitlists/waitlistspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
)

// ownSignupOrAdmin is this deployment's withdrawal rule: the person who joined
// may leave, and so may a service administrator.
//
// A nil caller is an anonymous request, which platform documents as the
// ordinary case rather than an error — it is what an unsubscribe link looks
// like. This deployment has no such link: joining a list requires a session, so
// an anonymous withdrawal is refused rather than allowed, and the day there is
// an unsubscribe link this is the one place that changes.
func ownSignupOrAdmin(store platformwaitlists.SignupStore, db database.Client) waitlistsgrpc.SignupAuthorizer {
	return waitlistsgrpc.SignupAuthorizerFunc(
		func(ctx context.Context, caller callers.Principal, scope tenancy.Scope, listID, signupID string) error {
			if caller == nil {
				return callers.ErrTargetNotPermitted
			}

			signup, err := store.GetSignup(ctx, db.Reader(), scope, listID, signupID)
			if err != nil {
				// Not a refusal: a store that would not answer is a failure to
				// decide, and platform reads anything but the sentinel as
				// codes.Internal for exactly that reason.
				return err
			}

			if signup.Subject.ID == caller.UserID() {
				return nil
			}

			if data := sessions.FromContext(ctx); data.GetServicePermissions().IsServiceAdmin() {
				return nil
			}

			return callers.ErrTargetNotPermitted
		})
}

// RegisterWaitlistsService registers platform's waitlists surface with the injector.
func RegisterWaitlistsService(i do.Injector) {
	do.Provide[waitlistspb.WaitlistsServiceServer](i, func(i do.Injector) (waitlistspb.WaitlistsServiceServer, error) {
		db := do.MustInvoke[database.Client](i)
		store := do.MustInvoke[platformwaitlists.Store](i)

		server, err := waitlistsgrpc.NewServer(
			store,
			db,
			sessions.PrincipalFromContext,
			ownSignupOrAdmin(store, db),
			waitlistsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			waitlistsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			waitlistsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			waitlistsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
		if err != nil {
			return nil, err
		}

		// Wrapped, not returned bare: a signup carries the session's address rather
		// than the one the request stated. See ownContactOnly.
		return ownContactOnly{WaitlistsServiceServer: server}, nil
	})
}

// Permissions is platform's map plus the three RPCs it leaves public.
//
// platform omits ListOpenLists, Join and Withdraw from its map deliberately:
// they are the signup page, and a deployment with a pre-launch list wants them
// reachable by somebody who has nothing to sign in to. This deployment does
// not — see authorization.JoinWaitlistsPermission — so all three are declared
// here, and the interceptor refuses a caller without the grant.
//
// Declaring them is not optional either way. The interceptor is fail-closed, so
// a method named nowhere is denied: taking platform's map alone would leave the
// signup page returning PermissionDenied to everybody.
func Permissions() map[string][]authorization.Permission {
	out := waitlistsgrpc.Permissions()

	out[waitlistspb.WaitlistsService_ListOpenLists_FullMethodName] = []authorization.Permission{
		authorization.ReadWaitlistsPermission,
	}
	out[waitlistspb.WaitlistsService_Join_FullMethodName] = []authorization.Permission{
		authorization.JoinWaitlistsPermission,
	}
	out[waitlistspb.WaitlistsService_Withdraw_FullMethodName] = []authorization.Permission{
		authorization.JoinWaitlistsPermission,
	}

	return out
}
