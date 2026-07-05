package grpc

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/reflection"
	"github.com/primandproper/platform-go/v3/uploads"
	mockuploads "github.com/primandproper/platform-go/v3/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestServiceImpl_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInput := identityfakes.BuildFakeUserCreationInput()
		exampleResponse := &identity.UserCreationResponse{
			CreatedUserID:   identityfakes.BuildFakeID(),
			Username:        exampleInput.Username,
			EmailAddress:    exampleInput.EmailAddress,
			FirstName:       exampleInput.FirstName,
			LastName:        exampleInput.LastName,
			TwoFactorSecret: "secret",
			TwoFactorQRCode: "qr_code",
			AccountStatus:   identity.UnverifiedAccountStatus.String(),
			CreatedAt:       identityfakes.BuildFakeTime(),
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateUser), testutils.ContextMatcher, mock.MatchedBy(func(input *identity.UserRegistrationInput) bool {
			return input.Username == exampleInput.Username &&
				input.EmailAddress == exampleInput.EmailAddress &&
				input.FirstName == exampleInput.FirstName &&
				input.LastName == exampleInput.LastName
		})).Return(exampleResponse, nil)

		request := &identitysvc.CreateUserRequest{
			Input: &identitysvc.UserRegistrationInput{
				Username:              exampleInput.Username,
				EmailAddress:          exampleInput.EmailAddress,
				FirstName:             exampleInput.FirstName,
				LastName:              exampleInput.LastName,
				Password:              exampleInput.Password,
				AccountName:           exampleInput.AccountName,
				AcceptedTos:           exampleInput.AcceptedTOS,
				AcceptedPrivacyPolicy: exampleInput.AcceptedPrivacyPolicy,
			},
		}

		result, err := service.CreateUser(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Created)
		assert.Equal(t, exampleResponse.CreatedUserID, result.Created.CreatedUserId)
		assert.Equal(t, exampleResponse.Username, result.Created.Username)
		assert.Equal(t, exampleResponse.EmailAddress, result.Created.EmailAddress)
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInput := identityfakes.BuildFakeUserCreationInput()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateUser), testutils.ContextMatcher, mock.AnythingOfType("*identity.UserRegistrationInput")).Return((*identity.UserCreationResponse)(nil), errors.New("database error"))

		request := &identitysvc.CreateUserRequest{
			Input: &identitysvc.UserRegistrationInput{
				Username:              exampleInput.Username,
				EmailAddress:          exampleInput.EmailAddress,
				FirstName:             exampleInput.FirstName,
				LastName:              exampleInput.LastName,
				Password:              exampleInput.Password,
				AccountName:           exampleInput.AccountName,
				AcceptedTos:           exampleInput.AcceptedTOS,
				AcceptedPrivacyPolicy: exampleInput.AcceptedPrivacyPolicy,
			},
		}

		result, err := service.CreateUser(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveUser), testutils.ContextMatcher, exampleUserID).Return(nil)

		request := &identitysvc.ArchiveUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.ArchiveUser(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveUser), testutils.ContextMatcher, exampleUserID).Return(errors.New("database error"))

		request := &identitysvc.ArchiveUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.ArchiveUser(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUser := identityfakes.BuildFakeUser()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUser), testutils.ContextMatcher, exampleUser.ID).Return(exampleUser, nil)

		request := &identitysvc.GetUserRequest{
			UserId: exampleUser.ID,
		}

		result, err := service.GetUser(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Result)
		assert.Equal(t, exampleUser.ID, result.Result.Id)
		assert.Equal(t, exampleUser.Username, result.Result.Username)
		assert.Equal(t, exampleUser.EmailAddress, result.Result.EmailAddress)
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUser), testutils.ContextMatcher, exampleUserID).Return((*identity.User)(nil), errors.New("database error"))

		request := &identitysvc.GetUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.GetUser(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
				identityfakes.BuildFakeUser(),
			},
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUsers), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(exampleUsers, nil)

		pageSize := uint32(25)
		request := &identitysvc.GetUsersRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetUsers(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleUsers.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetUsers), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.User])(nil), errors.New("database error"))

		pageSize := uint32(25)
		request := &identitysvc.GetUsersRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetUsers(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_SearchForUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
				identityfakes.BuildFakeUser(),
			},
		}
		exampleQuery := "test search"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SearchForUsers), testutils.ContextMatcher, exampleQuery, false, testutils.QueryFilterMatcher).Return(exampleUsers, nil)

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: false,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleUsers.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	T.Run("with search service enabled", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
			},
		}
		exampleQuery := "search query"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SearchForUsers), testutils.ContextMatcher, exampleQuery, true, testutils.QueryFilterMatcher).Return(exampleUsers, nil)

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: true,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleUsers.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleQuery := "test search"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SearchForUsers), testutils.ContextMatcher, exampleQuery, false, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.User])(nil), errors.New("search error"))

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: false,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateUserDetails(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserDetails), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.UserDetailsUpdateRequestInput")).Return(nil)

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName:       "Updated First",
				LastName:        "Updated Last",
				CurrentPassword: "password",
			},
		}

		result, err := service.UpdateUserDetails(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName: "Updated First",
			},
		}

		result, err := service.UpdateUserDetails(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserDetails), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.UserDetailsUpdateRequestInput")).Return(errors.New("update error"))

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName: "Updated First",
			},
		}

		result, err := service.UpdateUserDetails(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateUserEmailAddress(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		newEmail := "new@example.com"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserEmailAddress), testutils.ContextMatcher, mock.AnythingOfType("string"), newEmail).Return(nil)

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: newEmail,
		}

		result, err := service.UpdateUserEmailAddress(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: "new@example.com",
		}

		result, err := service.UpdateUserEmailAddress(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserEmailAddress), testutils.ContextMatcher, mock.AnythingOfType("string"), "new@example.com").Return(errors.New("update error"))

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: "new@example.com",
		}

		result, err := service.UpdateUserEmailAddress(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateUserUsername(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		newUsername := "newusername"

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserUsername), testutils.ContextMatcher, mock.AnythingOfType("string"), newUsername).Return(nil)

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: newUsername,
		}

		result, err := service.UpdateUserUsername(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: "newusername",
		}

		result, err := service.UpdateUserUsername(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateUserUsername), testutils.ContextMatcher, mock.AnythingOfType("string"), "newusername").Return(errors.New("update error"))

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: "newusername",
		}

		result, err := service.UpdateUserUsername(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

// mockAvatarUploadStream mocks the client streaming interface for UploadUserAvatar.
type mockAvatarUploadStream struct {
	ctx context.Context
	mock.Mock
}

func (m *mockAvatarUploadStream) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *mockAvatarUploadStream) RecvMsg(msg any) error {
	args := m.Called()
	if args.Get(0) != nil && msg != nil {
		proto.Merge(msg.(proto.Message), args.Get(0).(*uploadedmediasvc.UploadRequest))
	}
	return args.Error(1)
}

func (m *mockAvatarUploadStream) SendMsg(msg any) error {
	args := m.Called(msg)
	if len(args) == 0 {
		return nil
	}
	return args.Error(0)
}

func (m *mockAvatarUploadStream) Recv() (*uploadedmediasvc.UploadRequest, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uploadedmediasvc.UploadRequest), args.Error(1)
}

func (m *mockAvatarUploadStream) SendAndClose(response *identitysvc.UploadUserAvatarResponse) error {
	args := m.Called(response)
	return args.Error(0)
}

func (m *mockAvatarUploadStream) SendHeader(_ metadata.MD) error {
	return nil
}

func (m *mockAvatarUploadStream) SetHeader(_ metadata.MD) error {
	return nil
}

func (m *mockAvatarUploadStream) SetTrailer(_ metadata.MD) {
}

func TestServiceImpl_UploadUserAvatar(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager, uploadedMediaRepo := buildTestServiceWithUploadMocks(t)

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  "avatar.png",
					ContentType: "image/png",
				},
			},
		}
		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: []byte("image-data"),
			},
		}

		mockStream := &mockAvatarUploadStream{ctx: t.Context()}
		mockStream.On("RecvMsg").Return(metadataReq, nil).Once()
		mockStream.On("RecvMsg").Return(chunkReq, nil).Once()
		mockStream.On("RecvMsg").Return(nil, io.EOF).Once()
		mockStream.On("SendMsg", mock.AnythingOfType("*identity.UploadUserAvatarResponse")).Return(nil).Once()

		uploadManager := service.uploadManager.(*mockuploads.UploadManagerMock)
		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }
		uploadedMediaRepo.On(reflection.GetMethodName(uploadedMediaRepo.CreateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMediaDatabaseCreationInput")).Return(&uploadedmedia.UploadedMedia{ID: identityfakes.BuildFakeID()}, nil)
		identityDataManager.On(reflection.GetMethodName(identityDataManager.SetUserAvatar), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

		stream := &grpc.GenericServerStream[uploadedmediasvc.UploadRequest, identitysvc.UploadUserAvatarResponse]{ServerStream: mockStream}
		err := service.UploadUserAvatar(stream)

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mockStream, identityDataManager, uploadedMediaRepo)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		mockStream := &mockAvatarUploadStream{ctx: t.Context()}

		stream := &grpc.GenericServerStream[uploadedmediasvc.UploadRequest, identitysvc.UploadUserAvatarResponse]{ServerStream: mockStream}
		err := service.UploadUserAvatar(stream)

		assert.Error(t, err)
		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager, uploadedMediaRepo := buildTestServiceWithUploadMocks(t)

		metadataReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Metadata{
				Metadata: &uploadedmediasvc.UploadMetadata{
					ObjectName:  "avatar.png",
					ContentType: "image/png",
				},
			},
		}
		chunkReq := &uploadedmediasvc.UploadRequest{
			Payload: &uploadedmediasvc.UploadRequest_Chunk{
				Chunk: []byte("image-data"),
			},
		}

		mockStream := &mockAvatarUploadStream{ctx: t.Context()}
		mockStream.On("RecvMsg").Return(metadataReq, nil).Once()
		mockStream.On("RecvMsg").Return(chunkReq, nil).Once()
		mockStream.On("RecvMsg").Return(nil, io.EOF).Once()

		uploadManager := service.uploadManager.(*mockuploads.UploadManagerMock)
		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }
		uploadedMediaRepo.On(reflection.GetMethodName(uploadedMediaRepo.CreateUploadedMedia), testutils.ContextMatcher, mock.AnythingOfType("*uploadedmedia.UploadedMediaDatabaseCreationInput")).Return(&uploadedmedia.UploadedMedia{ID: identityfakes.BuildFakeID()}, nil)
		identityDataManager.On(reflection.GetMethodName(identityDataManager.SetUserAvatar), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errors.New("set avatar error"))

		stream := &grpc.GenericServerStream[uploadedmediasvc.UploadRequest, identitysvc.UploadUserAvatarResponse]{ServerStream: mockStream}
		err := service.UploadUserAvatar(stream)

		assert.Error(t, err)
		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
		mock.AssertExpectationsForObjects(t, mockStream, identityDataManager, uploadedMediaRepo)
	})
}
