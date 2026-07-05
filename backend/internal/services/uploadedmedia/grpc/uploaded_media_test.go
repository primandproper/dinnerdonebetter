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
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/database/filtering"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	"github.com/primandproper/platform-go/v3/reflection"
	"github.com/primandproper/platform-go/v3/uploads"
	mockuploads "github.com/primandproper/platform-go/v3/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *uploadedmediamock.Repository, *mockuploads.UploadManagerMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	uploadedMediaRepo := &uploadedmediamock.Repository{}
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
		uploadedMediaManager: &uploadedmediamock.Repository{},
		uploadManager:        &mockuploads.UploadManagerMock{},
	}

	return service
}

// mockUploadStream is a mock implementation of the upload stream.
type mockUploadStream struct {
	ctx context.Context
	mock.Mock
}

func (m *mockUploadStream) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *mockUploadStream) SendMsg(msg any) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *mockUploadStream) RecvMsg(msg any) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *mockUploadStream) Recv() (*uploadedmediasvc.UploadRequest, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uploadedmediasvc.UploadRequest), args.Error(1)
}

func (m *mockUploadStream) SendAndClose(response *uploadedmediasvc.UploadResponse) error {
	args := m.Called(response)
	return args.Error(0)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMediaDatabaseCreationInput")).Return(fakeUploadedMedia, nil)

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

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMediaDatabaseCreationInput")).Return(nil, errors.New("repository error"))

		request := &uploadedmediasvc.CreateUploadedMediaRequest{
			Input: converters.ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(fakeInput),
		}

		response, err := service.CreateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.GetUploadedMedia(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeUploadedMedia.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.GetUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, "some-id").Return(nil, errors.New("repository error"))

		request := &uploadedmediasvc.GetUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.GetUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMediaWithIDs), testutils.ContextMatcher, ids).Return(fakeUploadedMediaList, nil)

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMediaWithIDs), testutils.ContextMatcher, ids).Return(fakeUploadedMediaList, nil)

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)
		assert.Equal(t, fakeUploadedMedia1.ID, response.Results[0].Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMediaWithIDs), testutils.ContextMatcher, ids).Return(nil, errors.New("repository error"))

		request := &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		}

		response, err := service.GetUploadedMediaWithIDs(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMediaForUser), testutils.ContextMatcher, "test-user-id", mock.AnythingOfType("*filtering.QueryFilter")).Return(fakeUploadedMediaList, nil)

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "test-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMediaForUser), testutils.ContextMatcher, "test-user-id", mock.AnythingOfType("*filtering.QueryFilter")).Return(nil, errors.New("repository error"))

		request := &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: "test-user-id",
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetUploadedMediaForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMedia")).Return(nil)

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

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, "some-id").Return(nil, errors.New("repository error"))

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: "some-id",
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error on update", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMedia")).Return(errors.New("repository error"))

		request := &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
			Input:           &uploadedmediasvc.UploadedMediaUpdateRequestInput{},
		}

		response, err := service.UpdateUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(nil)

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		mock.AssertExpectationsForObjects(t, mockRepo)
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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, "some-id").Return(nil, errors.New("repository error"))

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: "some-id",
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error on archive", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		fakeUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
		fakeUploadedMedia.CreatedByUser = "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(fakeUploadedMedia, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveUploadedMedia), testutils.ContextMatcher, fakeUploadedMedia.ID).Return(errors.New("repository error"))

		request := &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: fakeUploadedMedia.ID,
		}

		response, err := service.ArchiveUploadedMedia(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(chunkReq1, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(chunkReq2, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(nil, io.EOF).Once()
		mockStream.On(reflection.GetMethodName(mockStream.SendAndClose), mock.AnythingOfType("*uploaded_media.UploadResponse")).Return(nil).Once()

		// Setup mock upload manager expectation
		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		// Setup mock repo expectation
		mockRepo.On(reflection.GetMethodName(mockRepo.CreateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMediaDatabaseCreationInput")).Return(fakeUploadedMedia, nil)

		// Execute
		err := service.Upload(mockStream)

		// Assert
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mockStream, mockRepo)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(chunkReq, nil).Once()

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(chunkReq, nil).Once()

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(nil, io.EOF).Once()

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(chunkReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(nil, io.EOF).Once()

		// Mock upload manager to return error
		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error {
			return errors.New("storage error")
		}

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream)
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

		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(metadataReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(chunkReq, nil).Once()
		mockStream.On(reflection.GetMethodName(mockStream.Recv)).Return(nil, io.EOF).Once()

		mockUploadMgr.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }

		// Mock repo to return error
		mockRepo.On(reflection.GetMethodName(mockRepo.CreateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMediaDatabaseCreationInput")).Return(nil, errors.New("database error"))

		err := service.Upload(mockStream)

		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		mock.AssertExpectationsForObjects(t, mockStream, mockRepo)
	})
}
