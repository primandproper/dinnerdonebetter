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
	return waitlistsgrpc.SignupAuthorizerFuncs{
		Withdrawal: func(ctx context.Context, caller callers.Principal, scope tenancy.Scope, listID, signupID string) error {
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
		},

		// SubjectRead is the other half, and it is the reason a member can ask
		// where they are in a queue at all.
		//
		// The grant on the method cannot answer this one: the subject comes off
		// the request, so a member holding ReadOwnWaitlistSignupsPermission could
		// name anybody. platform asks here instead, after the subject is read and
		// before any row is, and a refusal is answered as NotFound so that a
		// subject nobody has signed up and one belonging to somebody else are the
		// same answer.
		//
		// A service admin is permitted because the four signup reads are already
		// theirs; this is the fourth, reached under a narrower grant, and refusing
		// them here would make the split subtract from the role it was carved out
		// of.
		SubjectRead: func(ctx context.Context, caller callers.Principal, _ tenancy.Scope, subject platformwaitlists.Subject) error {
			if caller == nil {
				return callers.ErrTargetNotPermitted
			}

			if subject.Type == platformwaitlists.SubjectUser && subject.ID == caller.UserID() {
				return nil
			}

			if data := sessions.FromContext(ctx); data.GetServicePermissions().IsServiceAdmin() {
				return nil
			}

			return callers.ErrTargetNotPermitted
		},
	}
}

// ownContact is this deployment's answer to where a join's address comes from:
// the session's, never the request's.
//
// platform reads it off the wire by default, which is right for the deployment
// its public Join is written for — a pre-launch visitor with no account has no
// session to derive an address from, and double opt-in is what makes that
// honest. This deployment is the other kind: Join is behind a grant (see
// Permissions), so every caller already has an address this deployment
// verified, and one read off the wire would let any of them sign somebody else
// up.
//
// It is not a defense against learning whether an address is already on a list
// or has withdrawn. Nothing here needs to be: platform answers both of those
// refusals as success and records the outcome on the operation instead, so
// naming an address tells a caller nothing either way. This narrows who may
// name one.
//
// An anonymous request is refused rather than passed through, because there is
// no address to substitute and letting the stated one stand would reopen the
// hole for the one caller least accountable for it. ErrTargetNotPermitted is
// what platform reads as a refusal; anything else it reads as an outage.
func ownContact() waitlistsgrpc.ContactResolver {
	return waitlistsgrpc.ContactResolverFunc(
		func(ctx context.Context, _ callers.Principal, _ string) (string, error) {
			if contact := sessions.FromContext(ctx).GetEmailAddress(); contact != "" {
				return contact, nil
			}

			return "", callers.ErrTargetNotPermitted
		})
}

// RegisterWaitlistsService registers platform's waitlists surface with the injector.
func RegisterWaitlistsService(i do.Injector) {
	do.Provide[waitlistspb.WaitlistsServiceServer](i, func(i do.Injector) (waitlistspb.WaitlistsServiceServer, error) {
		db := do.MustInvoke[database.Client](i)
		store := do.MustInvoke[platformwaitlists.Store](i)

		return waitlistsgrpc.NewServer(
			store,
			db,
			sessions.PrincipalFromContext,
			ownSignupOrAdmin(store, db),
			waitlistsgrpc.WithContactResolver(ownContact()),
			waitlistsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			waitlistsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			waitlistsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			waitlistsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
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

	// And one of platform's fourteen is re-declared under a narrower grant than
	// the one platform assigns it.
	//
	// platform puts four signup reads behind PermissionReadSignups and says in
	// as many words that it is the grant to think hardest about, because
	// GetSignupByContact turns it into an oracle over every address in the
	// deployment. That grant is a service admin's here. But the fourth read is a
	// member asking about themselves, and holding it hostage to the other three
	// is what the SubjectRead authorizer above exists to undo: the grant says
	// this caller may make this kind of call, and the authorizer says whose
	// signups these are.
	out[waitlistspb.WaitlistsService_ListSignupsForSubject_FullMethodName] = []authorization.Permission{
		authorization.ReadOwnWaitlistSignupsPermission,
	}

	return out
}
