package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
)

func TestAsyncDataChangeMessageHandler_UserDataAggregationEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, reportArtifacts, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userID := identifiers.New()
		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: identifiers.New(),
			UserID:   userID,
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			arg := dest.(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
			return nil
		}

		dataPrivacyRepo.FetchUserDataCollectionFunc = func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
			assert.Equal(t, userID, actualUserID)

			return &dataprivacy.UserDataCollection{}, nil
		}

		reportArtifacts.SaveFunc = func(_ context.Context, _ string, _ []byte) error { return nil }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.NoError(t, err)

		assert.Len(t, dataPrivacyRepo.FetchUserDataCollectionCalls(), 1)
	})

	t.Run("marks disclosure completed when request carries a disclosure ID", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, reportArtifacts, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userID := identifiers.New()
		disclosureID := identifiers.New()
		reportID := identifiers.New()
		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			RequestID: disclosureID,
			ReportID:  reportID,
			UserID:    userID,
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			arg := dest.(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
			return nil
		}

		dataPrivacyRepo.FetchUserDataCollectionFunc = func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
			assert.Equal(t, userID, actualUserID)

			return &dataprivacy.UserDataCollection{}, nil
		}
		dataPrivacyRepo.MarkUserDataDisclosureCompletedFunc = func(_ context.Context, actualDisclosureID string, actualReportID string) error {
			assert.Equal(t, disclosureID, actualDisclosureID)
			assert.Equal(t, reportID, actualReportID)

			return nil
		}

		reportArtifacts.SaveFunc = func(_ context.Context, _ string, _ []byte) error { return nil }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.NoError(t, err)

		assert.Len(t, dataPrivacyRepo.FetchUserDataCollectionCalls(), 1)
		assert.Len(t, dataPrivacyRepo.MarkUserDataDisclosureCompletedCalls(), 1)
	})

	t.Run("marks disclosure failed when aggregation fails", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userID := identifiers.New()
		disclosureID := identifiers.New()
		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			RequestID: disclosureID,
			ReportID:  identifiers.New(),
			UserID:    userID,
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			arg := dest.(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
			return nil
		}

		expectedError := errors.New("fetch error")
		dataPrivacyRepo.FetchUserDataCollectionFunc = func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
			assert.Equal(t, userID, actualUserID)

			return nil, expectedError
		}
		dataPrivacyRepo.MarkUserDataDisclosureFailedFunc = func(_ context.Context, actualDisclosureID string) error {
			assert.Equal(t, disclosureID, actualDisclosureID)

			return nil
		}

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetching user data collection")

		assert.Len(t, dataPrivacyRepo.FetchUserDataCollectionCalls(), 1)
		assert.Len(t, dataPrivacyRepo.MarkUserDataDisclosureFailedCalls(), 1)
	})

	t.Run("with decode error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, decoder, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		rawMsg := []byte(`{"invalid": "json"}`)

		expectedError := errors.New("decode error")
		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, _ any) error { return expectedError }

		err := handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decoding JSON body")
	})

	t.Run("with fetch user data collection error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userID := identifiers.New()
		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: identifiers.New(),
			UserID:   userID,
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			arg := dest.(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
			return nil
		}

		expectedError := errors.New("fetch error")
		dataPrivacyRepo.FetchUserDataCollectionFunc = func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
			assert.Equal(t, userID, actualUserID)

			return nil, expectedError
		}

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetching user data collection")

		assert.Len(t, dataPrivacyRepo.FetchUserDataCollectionCalls(), 1)
	})

	t.Run("with upload error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, reportArtifacts, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userID := identifiers.New()
		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: identifiers.New(),
			UserID:   userID,
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			arg := dest.(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
			return nil
		}

		dataPrivacyRepo.FetchUserDataCollectionFunc = func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
			assert.Equal(t, userID, actualUserID)

			return &dataprivacy.UserDataCollection{}, nil
		}

		expectedError := errors.New("upload error")
		reportArtifacts.SaveFunc = func(_ context.Context, _ string, _ []byte) error { return expectedError }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "saving collection")

		assert.Len(t, dataPrivacyRepo.FetchUserDataCollectionCalls(), 1)
	})

	t.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, reportArtifacts, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		userID := identifiers.New()
		userDataCollectionRequest := &dataprivacy.UserDataAggregationRequest{
			ReportID: "", // Empty report ID
			UserID:   userID,
		}

		rawMsg, err := json.Marshal(userDataCollectionRequest)
		assert.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			arg := dest.(*dataprivacy.UserDataAggregationRequest)
			*arg = *userDataCollectionRequest
			return nil
		}

		dataPrivacyRepo.FetchUserDataCollectionFunc = func(_ context.Context, actualUserID string) (*dataprivacy.UserDataCollection, error) {
			assert.Equal(t, userID, actualUserID)

			return &dataprivacy.UserDataCollection{}, nil
		}

		reportArtifacts.SaveFunc = func(_ context.Context, _ string, _ []byte) error { return nil }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.NoError(t, err)

		assert.Len(t, dataPrivacyRepo.FetchUserDataCollectionCalls(), 1)
	})
}
