package grpc

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediafakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	uploadedmediamock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	"github.com/primandproper/platform-go/v6/filtering"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	"github.com/primandproper/platform-go/v6/observability/tracing"
	"github.com/primandproper/platform-go/v6/uploads"
	mockuploads "github.com/primandproper/platform-go/v6/uploads/mock"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *uploadedmediamock.RepositoryMock, *mockuploads.UploadManagerMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	uploadedMediaRepo := &uploadedmediamock.RepositoryMock{}
	uploadManager := &mockuploads.UploadManagerMock{}

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: "test-account-id",
				Requester: sessions.RequesterInfo{
					UserID: "test-user-id",
				},
			}, nil
		},
		uploadedMediaManager: uploadedMediaRepo,
		uploadManager:        uploadManager,
	}

	return service, uploadedMediaRepo, uploadManager
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
		uploadedMediaManager: &uploadedmediamock.RepositoryMock{},
		uploadManager:        &mockuploads.UploadManagerMock{},
	}

	return service
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

func TestServiceImpl_CreateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.NoError(t, err)
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
		service := buildTestServiceWithSessionError(t)

		request := &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: &uploadedmediasvc.UploadedMediaCreationRequestInput{},
		}

		response, err := service.CreateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeInput := uploadedmediafakes.BuildFakeUploadedMediaCreationRequestInput()

		mockRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: converters.ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(fakeInput),
		}

		response, err := service.CreateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)
	})
}

func TestServiceImpl_GetUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, fakeUploadedMedia.ID, uploadedMediaID)

			return fakeUploadedMedia, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.GetUploadedMedia(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeUploadedMedia.ID, response.Result.Id)

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.GetUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, "some-id", uploadedMediaID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.GetUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})
}

func TestServiceImpl_GetUploadedMediaWithIDs(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia1 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia1.CreatedByUser = "test-user-id"
		fakeUploadedMedia2 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia2.CreatedByUser = "test-user-id"

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

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		assert.Len(t, mockRepo.GetUploadedMediaWithIDsCalls(), 1)
	})

	t.Run("filters out other users' media", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia1 := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia1.CreatedByUser = "test-user-id"
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

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)
		assert.Equal(t, fakeUploadedMedia1.ID, response.Results[0].Id)

		assert.Len(t, mockRepo.GetUploadedMediaWithIDsCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{"id1", "id2"},
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("no IDs provided", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{},
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaWithIDsCalls(), 1)
	})
}

func TestServiceImpl_GetUploadedMediaForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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
			assert.Equal(t, "test-user-id", userID)

			return fakeUploadedMediaList, nil
		}

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "test-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		assert.Len(t, mockRepo.GetUploadedMediaForUserCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "test-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t)

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "different-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaForUserFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
			assert.Equal(t, "test-user-id", userID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "test-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaForUserCalls(), 1)
	})
}

func TestServiceImpl_UpdateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

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

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
		assert.Len(t, mockRepo.UpdateUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: "some-id",
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on update", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

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

		assert.Error(t, err)
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

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

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

		assert.NoError(t, err)
		assert.NotNil(t, response)

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
		assert.Len(t, mockRepo.ArchiveUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.GetUploadedMediaFunc = func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
			assert.Equal(t, "some-id", uploadedMediaID)

			return nil, errors.New("repository error")
		}

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUploadedMediaCalls(), 1)
	})

	t.Run("repository error on archive", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

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

		assert.Error(t, err)
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

		service, mockRepo, mockUploadMgr := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.MimeType = uploadedmedia.MimeTypeImagePNG

		// Create mock stream
		mockStream := &mockUploadStream{
			ctx: t.Context(),
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
		assert.NoError(t, err)
		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
		}

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("missing metadata in first message", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
		}

		// First message is a chunk instead of metadata
		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: []byte("some data"),
			},
		}

		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{chunkReq}

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing object_name", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing content_type", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("unsupported MIME type", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("file too large", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("no file data", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("upload manager error", func(t *testing.T) {
		t.Parallel()

		service, _, mockUploadMgr := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		service, mockRepo, mockUploadMgr := buildTestService(t)

		mockStream := &mockUploadStream{
			ctx: t.Context(),
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

		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Len(t, mockRepo.CreateUploadedMediaCalls(), 1)
	})
}
