package grpcconverters

import (
	"testing"

	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertGRPCQueryFilterToQueryFilter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := ConvertGRPCQueryFilterToQueryFilter(&grpcfiltering.QueryFilter{
			MaxResponseSize: new(uint32(25)),
		})

		require.NotNil(t, actual.MaxResponseSize)
		assert.Equal(t, uint16(25), *actual.MaxResponseSize)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		actual := ConvertGRPCQueryFilterToQueryFilter(nil)

		assert.Equal(t, filtering.DefaultQueryFilter(), actual)
	})

	T.Run("with oversized page size clamped instead of truncated", func(t *testing.T) {
		t.Parallel()

		// 300 used to truncate mod 256 to 44; 256 used to truncate to 0.
		for _, oversized := range []uint32{256, 300, 1_000_000} {
			actual := ConvertGRPCQueryFilterToQueryFilter(&grpcfiltering.QueryFilter{
				MaxResponseSize: new(oversized),
			})

			require.NotNil(t, actual.MaxResponseSize)
			assert.Equal(t, filtering.MaxQueryFilterLimit, *actual.MaxResponseSize)
		}
	})

	T.Run("with no page size uses default", func(t *testing.T) {
		t.Parallel()

		actual := ConvertGRPCQueryFilterToQueryFilter(&grpcfiltering.QueryFilter{})

		require.NotNil(t, actual.MaxResponseSize)
		assert.Equal(t, uint16(50), *actual.MaxResponseSize)
	})
}
