// Package errors maps the waitlist store's sentinels onto gRPC codes.
//
// Without it every one of them reaches a client as the handler's default, which
// is Internal — and "the server broke" is the wrong thing to tell somebody whose
// real problem is that they already unsubscribed.
package errors

import (
	"errors"

	"github.com/primandproper/platform-go/v13/errors/grpc"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(waitlistsGRPCMapper{})
}

type waitlistsGRPCMapper struct{}

func (waitlistsGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// No such live list or signup here. An archived one and one that never
	// existed are the same answer, deliberately.
	case errors.Is(err, waitlists.ErrListNotFound),
		errors.Is(err, waitlists.ErrSignupNotFound):
		return codes.NotFound, true

	// Malformed input — the request could not have succeeded as written, whatever
	// the state of the database.
	case errors.Is(err, waitlists.ErrEmptyListName),
		errors.Is(err, waitlists.ErrEmptyClosesAt),
		errors.Is(err, waitlists.ErrEmptyContact),
		errors.Is(err, waitlists.ErrEmptySubjectType),
		errors.Is(err, waitlists.ErrEmptySubjectID):
		return codes.InvalidArgument, true

	// Well-formed, and refused by the state of the list or the signup rather than
	// by the request. A closed list is a page that says so; a signup that is not
	// where the transition needs it to be is a queue somebody else has moved.
	case errors.Is(err, waitlists.ErrListClosed),
		errors.Is(err, waitlists.ErrWrongStatus),
		errors.Is(err, waitlists.ErrAlreadyWithdrawn):
		return codes.FailedPrecondition, true

	// The address is already spoken for on this list. AlreadyExists rather than
	// FailedPrecondition because the two lead somewhere different: one is "you are
	// already on this list", and the withdrawal below is "you asked not to be, and
	// we are honoring that".
	case errors.Is(err, waitlists.ErrAlreadySignedUp):
		return codes.AlreadyExists, true

	// A contact that has withdrawn cannot be re-added by filling the form in
	// again. PermissionDenied says the refusal is deliberate and not about this
	// request's shape — putting somebody back on a list they left is an
	// administrator's act, not a retry.
	case errors.Is(err, waitlists.ErrContactWithdrawn):
		return codes.PermissionDenied, true

	default:
		return codes.Unknown, false
	}
}
