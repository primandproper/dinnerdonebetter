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

# The signup page is public, and a double opt-in

Joining a list used to need a session, and the address a join wrote was the
session's rather than the request's. That was the only honest arrangement while
anybody could state an address: a public form that wrote whatever it was handed
would let one person sign another up.

The surface is built with WithConfirmation now, and that changes what stating
an address can do. A join is written pending, and the address is mailed a link
that confirms it and a link that takes it off the list; until the first is
followed, nobody is waiting on the signup and nothing will be sent to it but
that one message. So the five doors platform makes public — the open catalog,
Join, Confirm, Withdraw and Unsubscribe — are public here too, and a signed-in
caller's join is still attributed to them (the session is read when it is
sent), but the address is whatever they typed, and it is the mailbox that
vouches for it.

What it costs is recorded in internal/domain/waitlists: a visitor's signup has
no subject, so it is reached by its address — the unsubscribe link — rather than
by a subject access request.
*/
package waitlists

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v14/callers"
	"github.com/primandproper/platform-go/v14/links"
	linkscfg "github.com/primandproper/platform-go/v14/links/config"
	"github.com/primandproper/platform-go/v14/outbox"
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
// A nil caller is an anonymous request naming a signup by its id, and it is
// refused. The way off a list for somebody who is not signed in is the
// unsubscribe link their confirmation mail carried, which Unsubscribe redeems
// without asking this authorizer at all — the link is the authorization. An id
// is not: it is not a secret, and a withdrawal by id alone would let anybody who
// had seen one take its owner off a list.
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

// RegisterWaitlistsService registers platform's waitlists surface with the injector, and the
// action-link minter its confirmation loop mints through.
func RegisterWaitlistsService(i do.Injector) {
	// The minter over ddb_action_links. Its registry is the rendered configuration's — see
	// config.DefaultLinksConfig — and it runs the table's sweeper in this process, on the
	// injector's context, so expired and spent rows are reclaimed wherever links are minted.
	//
	// The table prefix is overwritten with the one the migration renders the table under,
	// whatever the configuration says: a prefix is not configuration here, because one that
	// disagreed with the migration would point the store at a table that does not exist. See
	// docs/configuration.md.
	do.Provide[*links.Minter](i, func(i do.Injector) (*links.Minter, error) {
		cfg := *do.MustInvoke[*linkscfg.Config](i)
		cfg.Database.TablePrefix = branding.TablePrefix

		return linkscfg.NewMinter(
			do.MustInvoke[context.Context](i),
			&cfg,
			do.MustInvoke[database.Client](i),
			linkscfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			linkscfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			linkscfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[waitlistspb.WaitlistsServiceServer](i, func(i do.Injector) (waitlistspb.WaitlistsServiceServer, error) {
		db := do.MustInvoke[database.Client](i)
		store := do.MustInvoke[platformwaitlists.Store](i)

		mailer, err := newConfirmationMailer(
			db,
			do.MustInvoke[*outbox.Writer](i),
			do.MustInvoke[*queuescfg.Config](i).OutboundEmailsTopicName,
			do.MustInvoke[*config.APIServiceConfig](i).BaseURL,
		)
		if err != nil {
			return nil, err
		}

		return waitlistsgrpc.NewServer(
			store,
			db,
			sessions.PrincipalFromContext,
			ownSignupOrAdmin(store, db),
			waitlistsgrpc.WithConfirmation(do.MustInvoke[*links.Minter](i), mailer),
			waitlistsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			waitlistsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			waitlistsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			waitlistsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// PublicMethods are the signup page: the five RPCs platform leaves out of its map because a
// visitor with nothing to sign in to has to reach them.
//
// The interceptor serves them optionally authenticated rather than unauthenticated — a call
// with no token is a visitor, and a call with one is held to it — because a signed-in
// caller's join is attributed to them and their withdrawal is authorized as theirs, and both
// need the session read.
func PublicMethods() []string {
	return waitlistsgrpc.PublicMethods()
}

// Permissions is platform's map, the signup page declared public, and one read re-declared.
//
// Declaring the public five is not optional. The interceptor is fail-closed, so a method named
// nowhere is denied: taking platform's map alone would leave the signup page refusing
// everybody. An empty slice is how this application's table says "no permission", and the
// authorization enforcer reads it as public.
func Permissions() map[string][]authorization.Permission {
	out := waitlistsgrpc.Permissions()

	for _, method := range PublicMethods() {
		out[method] = []authorization.Permission{}
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
