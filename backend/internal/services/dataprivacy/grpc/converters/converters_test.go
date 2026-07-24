package converters

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/primandproper/platform-go/v5/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertUserDataCollectionToGRPCUserDataCollection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		reportID := identifiers.New()
		userID := identifiers.New()

		collection := &dataprivacy.UserDataCollection{
			Comments: []comments.Comment{
				{ID: identifiers.New(), BelongsToUser: userID, Content: "first"},
				{ID: identifiers.New(), BelongsToUser: userID, Content: "second"},
			},
		}

		result := ConvertUserDataCollectionToGRPCUserDataCollection(collection, reportID)

		require.NotNil(t, result)
		assert.Equal(t, reportID, result.ReportId)
		// Meal planning and payments collections are always populated (no longer nil, closing the historical TODO).
		assert.NotNil(t, result.MealPlanningDataCollection)
		assert.NotNil(t, result.PaymentsDataCollection)
		assert.Len(t, result.Comments, 2)
	})
}

func TestConvertUserDataDisclosureToGRPCUserDataDisclosure(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC()
		completedAt := now
		disclosure := &dataprivacy.UserDataDisclosure{
			ID:            identifiers.New(),
			BelongsToUser: identifiers.New(),
			Status:        dataprivacy.UserDataDisclosureStatusCompleted,
			ReportID:      identifiers.New(),
			CreatedAt:     now,
			ExpiresAt:     now.Add(time.Hour),
			CompletedAt:   &completedAt,
		}

		result := ConvertUserDataDisclosureToGRPCUserDataDisclosure(disclosure)

		require.NotNil(t, result)
		assert.Equal(t, disclosure.ID, result.Id)
		assert.Equal(t, disclosure.BelongsToUser, result.BelongsToUser)
		assert.Equal(t, string(disclosure.Status), result.Status)
		assert.Equal(t, disclosure.ReportID, result.ReportId)
		assert.NotNil(t, result.CreatedAt)
		assert.NotNil(t, result.ExpiresAt)
		assert.NotNil(t, result.CompletedAt)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, ConvertUserDataDisclosureToGRPCUserDataDisclosure(nil))
	})
}
