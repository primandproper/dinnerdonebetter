package errors

import (
	"testing"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/errors/grpc"
	textsearch "github.com/primandproper/platform-go/v13/search/text"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

// This repo used to carry a searchGRPCMapper for these two sentinels. Platform maps them
// centrally now, and MapToGRPC consults PlatformMapper before any registered domain mapper,
// so ours was already unreachable. These cases stay as a guard that the statuses a client
// sees have not moved: the deleted mapper answered OutOfRange and InvalidArgument, and so
// does platform.
func TestPlatformSearchErrorMapping(T *testing.T) {
	T.Parallel()

	T.Run("with pagination past the result window", func(t *testing.T) {
		t.Parallel()

		// Wrapped as the managers wrap it, since that is how it arrives.
		code := grpc.MapToGRPC(platformerrors.Wrap(textsearch.ErrResultWindowExceeded, "searching for recipes"), codes.Unknown)
		assert.Equal(t, codes.OutOfRange, code)
	})

	T.Run("with a cursor the index did not issue", func(t *testing.T) {
		t.Parallel()

		code := grpc.MapToGRPC(platformerrors.Wrap(textsearch.ErrInvalidCursor, "searching for recipes"), codes.Unknown)
		assert.Equal(t, codes.InvalidArgument, code)
	})

	T.Run("with an unrelated error", func(t *testing.T) {
		t.Parallel()

		code := grpc.MapToGRPC(platformerrors.New("connection refused"), codes.Unknown)
		assert.Equal(t, codes.Unknown, code)
	})
}
