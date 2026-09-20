package waitlists

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v14/waitlists/waitlistspb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ownContactOnly is platform's waitlists server with one RPC narrowed: a signup
// carries the session's address, never the one the request stated.
//
// This is the rule the service it replaced enforced by construction — it built
// the signup from sessionContextData.GetEmailAddress() and never read a contact
// off the wire — and platform's Join takes the address from the request
// instead. That is the right default for the deployment platform's public Join
// is written for, where somebody with nothing to sign in to types their address
// into a pre-launch form. It is the wrong one here, where Join sits behind a
// grant and every caller already has an address this deployment verified.
//
// Two things go wrong without it, and the second is the one that matters:
//
//   - A caller can sign somebody else up. The address gets waitlist mail it
//     never asked for, and the person who typed it is not the person the signup
//     names, so the subject and the contact disagree about who this is.
//   - A caller can ask whether a given address has withdrawn. Joining with it
//     answers ErrContactWithdrawn or does not, and a withdrawal is exactly the
//     decision a person made about being contacted — so the refusal that exists
//     to honor an opt-out becomes the oracle that discloses it.
//
// The anonymous case is refused rather than passed through. A request with no
// session has no address to substitute, and passing the stated one along would
// reopen both holes for the one caller least accountable for them. Joining here
// requires a session — see Permissions, which puts Join behind a grant — so
// this cannot be reached by a caller who should have been allowed in.
//
// Upstream has no seam for this: NewServer takes a SignupAuthorizer for
// withdrawal and a ScopeResolver for tenancy, and nothing that decides where a
// contact comes from. A ContactResolver alongside SignupAuthorizer would delete
// this file.
type ownContactOnly struct {
	waitlistspb.WaitlistsServiceServer
}

// Join overwrites the stated contact with the session's before the handler sees it.
func (s ownContactOnly) Join(ctx context.Context, request *waitlistspb.JoinRequest) (*waitlistspb.JoinResponse, error) {
	contact := sessions.FromContext(ctx).GetEmailAddress()
	if contact == "" {
		return nil, status.Error(codes.Unauthenticated, "joining a waitlist requires a session")
	}

	// Copied rather than mutated in place: the request belongs to the caller's
	// codec, and a handler that edits it is a handler whose retry sends something
	// different from what it was given.
	narrowed := &waitlistspb.JoinRequest{
		ListId:  request.GetListId(),
		Contact: contact,
	}

	return s.WaitlistsServiceServer.Join(ctx, narrowed)
}
