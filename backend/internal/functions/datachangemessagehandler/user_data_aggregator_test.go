package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/primandproper/platform-go/v5/identifiers"
	"github.com/primandproper/platform-go/v5/reflection"
	"github.com/primandproper/platform-go/v5/uploads"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsyncDataChangeMessageHandler_UserDataAggregationEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

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

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, userID).Return(&dataprivacy.UserDataCollection{}, nil)

		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dataPrivacyRepo)
	})

	t.Run("marks disclosure completed when request carries a disclosure ID", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

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

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, userID).Return(&dataprivacy.UserDataCollection{}, nil)
		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.MarkUserDataDisclosureCompleted), mock.Anything, disclosureID, reportID).Return(nil)

		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dataPrivacyRepo)
	})

	t.Run("marks disclosure failed when aggregation fails", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

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
		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, userID).Return((*dataprivacy.UserDataCollection)(nil), expectedError)
		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.MarkUserDataDisclosureFailed), mock.Anything, disclosureID).Return(nil)

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetching user data collection")

		mock.AssertExpectationsForObjects(t, dataPrivacyRepo)
	})

	t.Run("with decode error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, decoder, _ := buildTestAsyncDataChangeMessageHandler(t)

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

		handler, _, _, _, _, _, _, _, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

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
		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, userID).Return((*dataprivacy.UserDataCollection)(nil), expectedError)

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetching user data collection")

		mock.AssertExpectationsForObjects(t, dataPrivacyRepo)
	})

	t.Run("with upload error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

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

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, userID).Return(&dataprivacy.UserDataCollection{}, nil)

		expectedError := errors.New("upload error")
		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return expectedError }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "saving collection")

		mock.AssertExpectationsForObjects(t, dataPrivacyRepo)
	})

	t.Run("with empty report ID", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, uploadManager, _, decoder, dataPrivacyRepo := buildTestAsyncDataChangeMessageHandler(t)

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

		dataPrivacyRepo.On(reflection.GetMethodName(dataPrivacyRepo.FetchUserDataCollection), mock.Anything, userID).Return(&dataprivacy.UserDataCollection{}, nil)

		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		err = handler.UserDataAggregationEventHandler("user_data_aggregation")(ctx, rawMsg)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dataPrivacyRepo)
	})
}
