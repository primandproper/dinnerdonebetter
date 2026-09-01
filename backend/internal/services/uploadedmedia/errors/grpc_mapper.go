// Package errors maps the upload registry's sentinels onto gRPC codes.
//
// Without it every one of them reaches a client as the handler's default, which
// is Internal — and "the server broke" is the wrong thing to tell somebody who
// asked for a picture that had just been archived, or who registered a key
// somebody else already holds.
package errors

import (
	"errors"

	"github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(uploadedMediaGRPCMapper{})
}

type uploadedMediaGRPCMapper struct{}

func (uploadedMediaGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// No row in the caller's scope: absent, archived, or another tenant's. All
	// three read the same from here, deliberately — an answer that distinguished
	// them would be an oracle for which keys exist elsewhere.
	case errors.Is(err, registry.ErrObjectNotFound):
		return codes.NotFound, true

	// The key names bytes that are already registered to somebody. The client's
	// remedy is a new key, which is a different thing to be told than "retry".
	case errors.Is(err, registry.ErrObjectKeyTaken):
		return codes.AlreadyExists, true

	// A subject with a type and no id, or an id and no type, and the zero subject
	// handed to a read that lists by one. Neither could have succeeded as
	// written, whatever the state of the database.
	case errors.Is(err, registry.ErrPartialSubject),
		errors.Is(err, registry.ErrUnattachedSubject):
		return codes.InvalidArgument, true

	default:
		return codes.Unknown, false
	}
}
