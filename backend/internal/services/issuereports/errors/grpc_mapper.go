// Package errors maps the issue report store's sentinels onto gRPC codes.
//
// Without it every one of them reaches a client as the handler's default, which
// is Internal — and "the server broke" is the wrong thing to tell a triager
// whose colleague resolved the report a moment before they did.
package errors

import (
	"errors"

	"github.com/primandproper/platform-go/v13/errors/grpc"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(issueReportsGRPCMapper{})
}

type issueReportsGRPCMapper struct{}

func (issueReportsGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// The report is not in the caller's account. Absent, archived and somebody
	// else's are the same answer from here, deliberately: the alternative tells a
	// caller which report IDs exist in other accounts.
	case errors.Is(err, issuereports.ErrReportNotFound):
		return codes.NotFound, true

	// Malformed input — the request could not have succeeded as written, whatever
	// the state of the database.
	case errors.Is(err, issuereports.ErrUnknownStatus),
		errors.Is(err, issuereports.ErrEmptyReporter),
		errors.Is(err, issuereports.ErrEmptyKind),
		errors.Is(err, issuereports.ErrEmptyDetails):
		return codes.InvalidArgument, true

	// Well-formed, and refused by the shape of the lifecycle rather than by the
	// request: an acknowledged report cannot go back to open, and nothing
	// transitions to itself.
	case errors.Is(err, issuereports.ErrInvalidStatusTransition):
		return codes.FailedPrecondition, true

	// The report moved between the caller reading it and deciding about it. Aborted
	// rather than FailedPrecondition because this one is worth retrying: the caller
	// re-reads the current status and decides again.
	case errors.Is(err, issuereports.ErrStatusConflict):
		return codes.Aborted, true

	default:
		return codes.Unknown, false
	}
}
