package errors

import (
	"testing"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
	"github.com/primandproper/platform-go/v9/search/text/elasticsearch"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestSearchGRPCMapper_Map(T *testing.T) {
	T.Parallel()

	T.Run("with pagination past the result window", func(t *testing.T) {
		t.Parallel()

		// Wrapped as the managers wrap it, since that is how it arrives.
		code, ok := searchGRPCMapper{}.Map(platformerrors.Wrap(elasticsearch.ErrResultWindowExceeded, "searching for recipes"))
		assert.True(t, ok)
		assert.Equal(t, codes.OutOfRange, code)
	})

	T.Run("with a cursor the index did not issue", func(t *testing.T) {
		t.Parallel()

		code, ok := searchGRPCMapper{}.Map(platformerrors.Wrap(textsearch.ErrInvalidCursor, "searching for recipes"))
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, code)
	})

	T.Run("with an unrelated error", func(t *testing.T) {
		t.Parallel()

		_, ok := searchGRPCMapper{}.Map(platformerrors.New("connection refused"))
		assert.False(t, ok)
	})

	T.Run("with nil error", func(t *testing.T) {
		t.Parallel()

		_, ok := searchGRPCMapper{}.Map(nil)
		assert.False(t, ok)
	})
}
