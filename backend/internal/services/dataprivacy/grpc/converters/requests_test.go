package converters

import (
	"testing"
	"time"

	platformdataprivacy "github.com/primandproper/platform-go/v10/dataprivacy"
	"github.com/primandproper/platform-go/v10/identifiers"
	"github.com/primandproper/platform-go/v10/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertRequestToGRPCRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC().Truncate(time.Second)
		requestID := identifiers.New()
		subjectID := identifiers.New()

		input := &platformdataprivacy.Request{
			ID:            requestID,
			Subject:       platformdataprivacy.Subject{ID: subjectID, Type: platformdataprivacy.SubjectUser},
			Type:          platformdataprivacy.RequestExport,
			Status:        platformdataprivacy.StatusCompleted,
			RequestedAt:   now,
			DueAt:         now.Add(30 * 24 * time.Hour),
			ExpiresAt:     now.Add(7 * 24 * time.Hour),
			CompletedAt:   pointer.To(now.Add(time.Minute)),
			ArtifactBytes: 4096,
		}

		actual := ConvertRequestToGRPCRequest(input)

		require.NotNil(t, actual)
		assert.Equal(t, requestID, actual.GetId())
		assert.Equal(t, subjectID, actual.GetSubjectId())
		assert.Equal(t, "export", actual.GetRequestType())
		assert.Equal(t, "completed", actual.GetStatus())
		assert.Equal(t, now, actual.GetRequestedAt().AsTime())
		assert.Equal(t, int64(4096), actual.GetArtifactBytes())
		// Attempts is deliberately unset: v10 moved the claim count onto the operation
		// fulfilling the request, so there is nothing on the request to copy.
		assert.Zero(t, actual.GetAttempts())
		assert.NotNil(t, actual.GetExpiresAt())
		assert.NotNil(t, actual.GetCompletedAt())
	})

	T.Run("a partial export carries the failures that made it partial", func(t *testing.T) {
		t.Parallel()

		// Dropping these would render as an unqualified "completed" over an artifact with
		// three sections missing, told to somebody who has thirty days to complain.
		input := &platformdataprivacy.Request{
			ID:       identifiers.New(),
			Type:     platformdataprivacy.RequestExport,
			Status:   platformdataprivacy.StatusCompleted,
			Failures: map[string]string{"payments": "context deadline exceeded"},
		}

		actual := ConvertRequestToGRPCRequest(input)

		require.NotNil(t, actual)
		assert.Equal(t, map[string]string{"payments": "context deadline exceeded"}, actual.GetFailures())
	})

	T.Run("an unfulfilled request has no timestamps for what has not happened", func(t *testing.T) {
		t.Parallel()

		// Zero rather than absent would render as the Unix epoch, which reads as a date
		// rather than as "never happened".
		input := &platformdataprivacy.Request{
			ID:     identifiers.New(),
			Type:   platformdataprivacy.RequestErasure,
			Status: platformdataprivacy.StatusInProgress,
		}

		actual := ConvertRequestToGRPCRequest(input)

		require.NotNil(t, actual)
		assert.Nil(t, actual.GetExpiresAt())
		assert.Nil(t, actual.GetCompletedAt())
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, ConvertRequestToGRPCRequest(nil))
	})
}
