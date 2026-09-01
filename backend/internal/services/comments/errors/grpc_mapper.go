// Package errors maps the comment store's sentinels onto gRPC codes.
//
// Without it every one of them reaches a client as the handler's default, which
// is Internal — and "the server broke" is the wrong thing to tell somebody who
// replied to a comment that had just been archived.
package errors

import (
	"errors"

	comments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/errors/grpc"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(commentsGRPCMapper{})
}

type commentsGRPCMapper struct{}

func (commentsGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// The comment, or the one being replied to, is not in the caller's scope.
	case errors.Is(err, comments.ErrCommentNotFound),
		errors.Is(err, comments.ErrParentNotFound):
		return codes.NotFound, true

	// The target is a kind of thing this application has, and this one is not
	// there. Distinct from an unknown type below: a client shown this has a stale
	// list, and a client shown that has a bug.
	case errors.Is(err, comments.ErrTargetNotFound):
		return codes.NotFound, true

	// Malformed or unregistered input — the request could not have succeeded as
	// written, whatever the state of the database.
	case errors.Is(err, comments.ErrUnknownTargetType),
		errors.Is(err, comments.ErrEmptyTargetType),
		errors.Is(err, comments.ErrEmptyTargetID),
		errors.Is(err, comments.ErrEmptyAuthor),
		errors.Is(err, comments.ErrEmptyBody),
		errors.Is(err, comments.ErrEmptyParent):
		return codes.InvalidArgument, true

	// Well-formed, and refused by the shape of the discussion rather than by the
	// request: threads are one level deep, and a reply belongs to its parent's
	// target.
	case errors.Is(err, comments.ErrNestedReply),
		errors.Is(err, comments.ErrTargetMismatch):
		return codes.FailedPrecondition, true

	default:
		return codes.Unknown, false
	}
}
