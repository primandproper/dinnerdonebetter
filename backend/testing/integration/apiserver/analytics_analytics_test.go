package integration

import (
	"testing"

	analyticsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalytics_TrackEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.TrackEvent(ctx, &analyticsgrpc.TrackEventRequest{
			Source: "ios",
			Event:  "test_event",
			Properties: map[string]string{
				"key": "value",
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.TrackEvent(ctx, &analyticsgrpc.TrackEventRequest{
			Source: "ios",
			Event:  "test_event",
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestAnalytics_TrackAnonymousEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.TrackAnonymousEvent(ctx, &analyticsgrpc.TrackAnonymousEventRequest{
			Source:      "ios",
			Event:       "anonymous_event",
			AnonymousId: "anon-test-id-123",
			Properties: map[string]string{
				"key": "value",
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, res)
	})
}
