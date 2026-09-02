// Package errors maps the settings store's sentinels onto gRPC codes.
//
// Without it every one of them reaches a client as the handler's default, which
// is Internal — and "the server broke" is the wrong thing to tell somebody whose
// real problem is that they picked a unit the setting does not offer.
package errors

import (
	"errors"

	"github.com/primandproper/platform-go/v13/errors/grpc"
	settings "github.com/primandproper/platform-go/v13/settings"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(settingsGRPCMapper{})
}

type settingsGRPCMapper struct{}

func (settingsGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// No such live setting, or nobody has answered it. An archived definition
	// and one that never existed are the same answer, deliberately.
	case errors.Is(err, settings.ErrDefinitionNotFound),
		errors.Is(err, settings.ErrValueNotFound):
		return codes.NotFound, true

	// Malformed input — the request could not have succeeded as written,
	// whatever the state of the database. The value ones are here rather than
	// under FailedPrecondition because what refuses them is the setting's own
	// definition, which the caller could have read first: "celsius" for a
	// boolean is a wrong request, not a wrong moment.
	case errors.Is(err, settings.ErrEmptyDefinitionName),
		errors.Is(err, settings.ErrEmptySubjectType),
		errors.Is(err, settings.ErrEmptySubjectID),
		errors.Is(err, settings.ErrEmptyEnumerationValue),
		errors.Is(err, settings.ErrDuplicateEnumerationValue),
		errors.Is(err, settings.ErrUnknownKind),
		errors.Is(err, settings.ErrMalformedValue),
		errors.Is(err, settings.ErrNotEnumerated),
		// A typed read of the wrong kind is a mistake in the calling code.
		// Nothing this service does reads a resolution as a type, so it reaches
		// a client only through a caller of its own; it is mapped anyway,
		// because an unmapped sentinel is reported as an outage.
		errors.Is(err, settings.ErrKindMismatch):
		return codes.InvalidArgument, true

	// The name is already defined in this catalog.
	case errors.Is(err, settings.ErrDefinitionNameTaken):
		return codes.AlreadyExists, true

	// Well-formed, and refused by what is already stored rather than by the
	// request. Narrowing an enumeration or changing a kind is an edit somebody
	// can make once they have cleared or migrated the values it would strand,
	// and the wrapped message names the first of them — so this is a state to
	// fix rather than a request to rewrite.
	case errors.Is(err, settings.ErrStrandedValues):
		return codes.FailedPrecondition, true

	default:
		return codes.Unknown, false
	}
}
