package grpc

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediafakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	uploadedmediamock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/metering"
	meteringmock "github.com/primandproper/platform-go/v9/metering/mock"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/uploads"
	mockuploads "github.com/primandproper/platform-go/v9/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	testAccountID = identifiers.New()
	testUserID    = identifiers.New()
)

func buildTestService(t *testing.T) (*serviceImpl, *uploadedmediamock.RepositoryMock, *mockuploads.UploadManagerMock) {
	t.Helper()

	service, repo, uploadManager, _ := buildTestServiceWithRecorder(t)

	return service, repo, uploadManager
}

// buildTestServiceWithRecorder is buildTestService with the usage recorder handed back, for the
// tests that care what got metered.
//
// The recorder accepts everything by default. Upload counts bytes after the file and its row are
// already committed, so a recorder that failed would be testing the swallow rather than the
// count — see recordUploadUsage.
func buildTestServiceWithRecorder(t *testing.T) (*serviceImpl, *uploadedmediamock.RepositoryMock, *mockuploads.UploadManagerMock, *meteringmock.RecorderMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	uploadedMediaRepo := &uploadedmediamock.RepositoryMock{}
	uploadManager := &mockuploads.UploadManagerMock{}
	usageRecorder := &meteringmock.RecorderMock{
		RecordFunc: func(context.Context, ...metering.Usage) error { return nil },
	}

	service := &serviceImpl{
		tracer:               tracer,
		logger:               logger,
		uploadedMediaManager: uploadedMediaRepo,
		uploadManager:        uploadManager,
		usageRecorder:        usageRecorder,
	}

	return service, uploadedMediaRepo, uploadManager, usageRecorder
}

// mockUploadStream is a fake upload stream. Recv yields queued messages in order and then
// returns recvErr (io.EOF for a clean end of stream), which models a stream far more directly
// than a call-expectation mock does.
type mockUploadStream struct {
	ctx             context.Context
	recvErr         error
	sendAndCloseErr error
	recvQueue       []*uploadedmediasvc.UploadRequest
	closedWith      []*uploadedmediasvc.UploadResponse
	recvIndex       int
}

func (m *mockUploadStream) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}

	return m.ctx
}

func (m *mockUploadStream) SendMsg(any) error { return nil }

func (m *mockUploadStream) RecvMsg(any) error { return nil }

func (m *mockUploadStream) Recv() (*uploadedmediasvc.UploadRequest, error) {
	if m.recvIndex < len(m.recvQueue) {
		req := m.recvQueue[m.recvIndex]
		m.recvIndex++

		return req, nil
	}

	if m.recvErr != nil {
		return nil, m.recvErr
	}

	return nil, io.EOF
}

func (m *mockUploadStream) SendAndClose(response *uploadedmediasvc.UploadResponse) error {
	m.closedWith = append(m.closedWith, response)

	return m.sendAndCloseErr
}

func (m *mockUploadStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockUploadStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockUploadStream) SetTrailer(md metadata.MD) {
}

func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: testAccountID,
		Requester:       sessions.RequesterInfo{UserID: testUserID},
	})
}

func TestServiceImpl_CreateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeInput := uploadedmediafakes.BuildFakeUploadedMediaCreationRequestInput()

		mockRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return fakeUploadedMedia, nil
		}

		request := &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: converters.ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(fakeInput),
		}

		response, err := service.CreateUploadedMedia(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeUploadedMedia.ID, response.Created.Id)
		assert.Equal(t, fakeUploadedMedia.StoragePath, response.Created.StoragePath)

		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{},
		}

		response, err := service.CreateUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeInput := uploadedmediafakes.BuildFakeUploadedMediaCreationRequestInput()

		mockRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: converters.ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(fakeInput),
		}

		response, err := service.CreateUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)
	})
}

func TestServiceImpl_GetUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = testUserID

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.GetUploadedMedia(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeUploadedMedia.ID, response.Result.Id)

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.GetUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "different-user-id"

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.GetUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, "some-id", uploadedMediaID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.GetUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})
}

func TestServiceImpl_GetUploadedMediaWithIDs(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia1 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia1.CreatedByUser = testUserID
		fakeUploadedMedia2 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia2.CreatedByUser = testUserID

		fakeUploadedMediaList := []*uploadedmedia.UploadedMedia{
			fakeUploadedMedia1,
			fakeUploadedMedia2,
		}

		ids := []string{fakeUploadedMedia1.ID, fakeUploadedMedia2.ID}

		mockRepo.GetUploadedMediaWithIDsFunc = func(_ context.Context, actualIds []string) ([]*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, ids, actualIds)

			return fakeUploadedMediaList, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		assert.Len(t, mockRepo.GetUploadedMediaWithIDsCalls(), 1)
	})

	t.Run("filters out other users' media", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia1 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia1.CreatedByUser = testUserID
		fakeUploadedMedia2 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia2.CreatedByUser = "other-user-id"

		fakeUploadedMediaList := []*uploadedmedia.UploadedMedia{
			fakeUploadedMedia1,
			fakeUploadedMedia2,
		}

		ids := []string{fakeUploadedMedia1.ID, fakeUploadedMedia2.ID}

		mockRepo.GetUploadedMediaWithIDsFunc = func(_ context.Context, actualIds []string) ([]*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, ids, actualIds)

			return fakeUploadedMediaList, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)
		assert.Equal(t, fakeUploadedMedia1.ID, response.Results[0].Id)

		assert.Len(t, mockRepo.GetUploadedMediaWithIDsCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{"id1", "id2"},
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("no IDs provided", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{},
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		ids := []string{"id1", "id2"}

		mockRepo.GetUploadedMediaWithIDsFunc = func(_ context.Context, actualIds []string) ([]*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, ids, actualIds)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaWithIDsCalls(), 1)
	})
}

func TestServiceImpl_GetUploadedMediaForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMediaList := &filtering.QueryFilteredResult[uploadedmedia.UploadedMedia]{
			Data: []*uploadedmedia.UploadedMedia{
				uploadedmediafakes.BuildFakeUploadedMedia(),
				uploadedmediafakes.BuildFakeUploadedMedia(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    2,
				FilteredCount: 2,
			},
		}

		mockRepo.GetUploadedMediaForUserFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
			assert.Equal(t, testUserID, userID)

			return fakeUploadedMediaList, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: testUserID,
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		assert.Len(t, mockRepo.GetUploadedMediaForUserCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: testUserID,
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "different-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaForUserFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
			assert.Equal(t, testUserID, userID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: testUserID,
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaForUserCalls(), 1)
	})
}

func TestServiceImpl_UpdateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = testUserID

		newStoragePath := "updated/path.jpg"

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}
		mockRepo.UpdateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMedia) error {
			return nil
		}

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
			Input: &uploadedmediasvc.UploadedMediaUpdateRequestInput{
				StoragePath: &newStoragePath,
			},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
		assert.Len(t, mockRepo.UpdateUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: "some-id",
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "different-user-id"

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, "some-id", uploadedMediaID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: "some-id",
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on update", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = testUserID

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}
		mockRepo.UpdateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMedia) error {
			return errors.New("repository error")
		}

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
		assert.Len(t, mockRepo.UpdateUploadedMediaCalls(), 1)
	})
}

func TestServiceImpl_ArchiveUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = testUserID

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}
		mockRepo.ArchiveUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) error {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return nil
		}

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
		assert.Len(t, mockRepo.ArchiveUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "different-user-id"

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, "some-id", uploadedMediaID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on archive", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = testUserID

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}
		mockRepo.ArchiveUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) error {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return errors.New("repository error")
		}

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
		assert.Len(t, mockRepo.ArchiveUploadedMediaCalls(), 1)
	})
}

func TestServiceImpl_Upload(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, mockRepo, mockUploadMgr, usageRecorder := buildTestServiceWithRecorder(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.MimeType = uploadedmedia.MimeTypeImagePNG

		// Create mock stream
		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		// Setup metadata message
		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "test-file.png",
			ContentType: uploadedmedia.MimeTypeImagePNG,
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		// Setup chunk messages
		chunk1 := []byte("test file content part 1")
		chunk2 := []byte("test file content part 2")

		chunkReq1 := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: chunk1,
			},
		}

		chunkReq2 := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: chunk2,
			},
		}

		// Setup mock stream expectations
		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq, chunkReq1, chunkReq2}

		// Setup mock upload manager expectation
		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		// Setup mock repo expectation
		mockRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return fakeUploadedMedia, nil
		}

		// Execute
		err := service.Upload(mockStream)

		// Assert
		require.NoError(t, err)
		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)

		// The bytes are counted against the account, keyed by the row that was created —
		// not by anything request-scoped, because a retried upload stores a second object
		// and is genuinely a second charge.
		require.Len(t, usageRecorder.RecordCalls(), 1)
		require.Len(t, usageRecorder.RecordCalls()[0].U, 1)

		usage := usageRecorder.RecordCalls()[0].U[0]
		assert.Equal(t, appmetering.UploadedMediaBytesMeter, usage.Meter)
		assert.Equal(t, testAccountID, usage.Subject)
		assert.Equal(t, fakeUploadedMedia.ID, usage.IdempotencyKey)
		assert.Equal(t, int64(len(chunk1)+len(chunk2)), usage.Quantity)
		assert.Equal(t, uploadedmedia.MimeTypeImagePNG, usage.Dimensions["mime_type"])
	})

	t.Run("a failed usage record does not fail the upload", func(t *testing.T) {
		t.Parallel()

		// The file is in the bucket and its row is in the database before the meter is
		// touched, so failing here would tell the client an upload did not happen that did.
		service, mockRepo, mockUploadMgr, usageRecorder := buildTestServiceWithRecorder(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()

		mockStream := &mockUploadStream{ctx: buildSessionContextForTest(t)}
		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{
			{Payload: &uploadedmediasvc.UploadRequest_Metadata{Metadata: &uploadedmediasvc.UploadMetadata{
				Bucket:      "test-bucket",
				ObjectName:  "test-file.png",
				ContentType: uploadedmedia.MimeTypeImagePNG,
			}}},
			{Payload: &uploadedmediasvc.UploadRequest_Chunk{Chunk: []byte("test file content")}},
		}

		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }
		mockRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return fakeUploadedMedia, nil
		}
		usageRecorder.RecordFunc = func(context.Context, ...metering.Usage) error {
			return platformerrors.New("blah")
		}

		require.NoError(t, service.Upload(mockStream))
		assert.Len(t, usageRecorder.RecordCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
		}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("missing metadata in first message", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		// First message is a chunk instead of metadata
		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: []byte("some data"),
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{chunkReq}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing object_name", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "", // Missing
			ContentType: uploadedmedia.MimeTypeImagePNG,
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing content_type", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "test-file.png",
			ContentType: "", // Missing
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("unsupported MIME type", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "test-file.pdf",
			ContentType: "application/pdf", // Unsupported
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("file too large", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "large-file.png",
			ContentType: uploadedmedia.MimeTypeImagePNG,
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		// Create a chunk that's larger than maxUploadSize (100 MB)
		largeChunk := make([]byte, 101*1024*1024)

		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: largeChunk,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq, chunkReq}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("no file data", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "empty-file.png",
			ContentType: uploadedmedia.MimeTypeImagePNG,
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("upload manager error", func(t *testing.T) {
		t.Parallel()

		service, _, mockUploadMgr := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "test-file.png",
			ContentType: uploadedmedia.MimeTypeImagePNG,
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		chunk := []byte("test file content")
		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: chunk,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq, chunkReq}

		// Mock upload manager to return error
		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error {
			return errors.New("storage error")
		}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		service, mockRepo, mockUploadMgr := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: buildSessionContextForTest(t),
		}

		uploadMetadata := &uploadedmediasvc.UploadMetadata{
			Bucket:      "test-bucket",
			ObjectName:  "test-file.png",
			ContentType: uploadedmedia.MimeTypeImagePNG,
		}

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: uploadMetadata,
			},
		}

		chunk := []byte("test file content")
		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: chunk,
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq, chunkReq}

		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		// Mock repo to return error
		mockRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return nil, errors.New("database error")
		}

		err := service.Upload(mockStream)

		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)
	})
}
