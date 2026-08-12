package metering

import (
	"testing"

	"github.com/primandproper/platform-go/v10/identifiers"
	platformmetering "github.com/primandproper/platform-go/v10/metering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProviderMapper(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		ref, err := NewProviderMapper().ProviderRefFor(ctx, identifiers.New(), UploadedMediaBytesMeter)

		// ErrNoProviderRef specifically, and not some other error: the flusher settles a
		// total on that one and retries forever on anything else. A mapper that reported
		// "not billable" as a generic failure would make every account the permanent head
		// of the flush queue.
		require.ErrorIs(t, err, platformmetering.ErrNoProviderRef)
		assert.Equal(t, platformmetering.ProviderRef{}, ref)
	})
}
