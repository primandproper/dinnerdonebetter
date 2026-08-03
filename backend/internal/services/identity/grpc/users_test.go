package grpc

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/uploads"
	mockuploads "github.com/primandproper/platform-go/v9/uploads/mock"

	"github.com/stretchr/testify/assert"
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

		identityDataManager.CreateUserFunc = func(_ context.Context, input *identity.UserRegistrationInput) (*identity.UserCreationResponse, error) {
			assert.True(t, input.Username == exampleInput.Username && input.EmailAddress == exampleInput.EmailAddress && input.FirstName == exampleInput.FirstName && input.LastName == exampleInput.LastName)

			return exampleResponse, nil
		}

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

		result, err := service.CreateUser(buildSessionContextForTest(t), request)

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

		identityDataManager.CreateUserFunc = func(_ context.Context, _ *identity.UserRegistrationInput) (*identity.UserCreationResponse, error) {
			return nil, errors.New("database error")
		}

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

		result, err := service.CreateUser(buildSessionContextForTest(t), request)

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

		identityDataManager.ArchiveUserFunc = func(_ context.Context, userID string) error {
			assert.Equal(t, exampleUserID, userID)

			return nil
		}

		request := &identitysvc.ArchiveUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.ArchiveUser(buildSessionContextForTest(t), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.ArchiveUserFunc = func(_ context.Context, userID string) error {
			assert.Equal(t, exampleUserID, userID)

			return errors.New("database error")
		}

		request := &identitysvc.ArchiveUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.ArchiveUser(buildSessionContextForTest(t), request)

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

		identityDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, exampleUser.ID, userID)

			return exampleUser, nil
		}

		request := &identitysvc.GetUserRequest{
			UserId: exampleUser.ID,
		}

		result, err := service.GetUser(buildSessionContextForTest(t), request)

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

		identityDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, exampleUserID, userID)

			return nil, errors.New("database error")
		}

		request := &identitysvc.GetUserRequest{
			UserId: exampleUserID,
		}

		result, err := service.GetUser(buildSessionContextForTest(t), request)

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

		identityDataManager.GetUsersFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
			return exampleUsers, nil
		}

		pageSize := uint32(25)
		request := &identitysvc.GetUsersRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetUsers(buildSessionContextForTest(t), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Len(t, result.Results, len(exampleUsers.Data))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, identityDataManager := buildTestService(t)

		identityDataManager.GetUsersFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
			return nil, errors.New("database error")
		}

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

		ctx := buildSessionContextForTest(t)
		service, identityDataManager := buildTestService(t)

		exampleUsers := &filtering.QueryFilteredResult[identity.User]{
			Data: []*identity.User{
				identityfakes.BuildFakeUser(),
				identityfakes.BuildFakeUser(),
			},
		}
		exampleQuery := "test search"

		identityDataManager.SearchForUsersFunc = func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
			assert.Equal(t, exampleQuery, query)
			assert.False(t, useSearchService)

			return exampleUsers, nil
		}

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
		assert.Len(t, result.Results, len(exampleUsers.Data))
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

		identityDataManager.SearchForUsersFunc = func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
			assert.Equal(t, exampleQuery, query)
			assert.True(t, useSearchService)

			return exampleUsers, nil
		}

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: true,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(buildSessionContextForTest(t), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Len(t, result.Results, len(exampleUsers.Data))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleUsers.Data[i].ID)
		}
	})

	T.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleQuery := "test search"

		identityDataManager.SearchForUsersFunc = func(_ context.Context, query string, useSearchService bool, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
			assert.Equal(t, exampleQuery, query)
			assert.False(t, useSearchService)

			return nil, errors.New("search error")
		}

		pageSize := uint32(25)
		request := &identitysvc.SearchForUsersRequest{
			Query:            exampleQuery,
			UseSearchService: false,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.SearchForUsers(buildSessionContextForTest(t), request)

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

		identityDataManager.UpdateUserDetailsFunc = func(_ context.Context, _ string, _ *identity.UserDetailsUpdateRequestInput) error {
			return nil
		}

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName:       "Updated First",
				LastName:        "Updated Last",
				CurrentPassword: "password",
			},
		}

		result, err := service.UpdateUserDetails(buildSessionContextForTest(t), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

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

		identityDataManager.UpdateUserDetailsFunc = func(_ context.Context, _ string, _ *identity.UserDetailsUpdateRequestInput) error {
			return errors.New("update error")
		}

		request := &identitysvc.UpdateUserDetailsRequest{
			Input: &identitysvc.UserDetailsUpdateRequestInput{
				FirstName: "Updated First",
			},
		}

		result, err := service.UpdateUserDetails(buildSessionContextForTest(t), request)

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

		identityDataManager.UpdateUserEmailAddressFunc = func(_ context.Context, _ string, actualNewEmail string) error {
			assert.Equal(t, newEmail, actualNewEmail)

			return nil
		}

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: newEmail,
		}

		result, err := service.UpdateUserEmailAddress(buildSessionContextForTest(t), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

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

		identityDataManager.UpdateUserEmailAddressFunc = func(_ context.Context, _ string, newEmail string) error {
			assert.Equal(t, "new@example.com", newEmail)

			return errors.New("update error")
		}

		request := &identitysvc.UpdateUserEmailAddressRequest{
			NewEmailAddress: "new@example.com",
		}

		result, err := service.UpdateUserEmailAddress(buildSessionContextForTest(t), request)

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

		identityDataManager.UpdateUserUsernameFunc = func(_ context.Context, _ string, actualNewUsername string) error {
			assert.Equal(t, newUsername, actualNewUsername)

			return nil
		}

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: newUsername,
		}

		result, err := service.UpdateUserUsername(buildSessionContextForTest(t), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

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

		identityDataManager.UpdateUserUsernameFunc = func(_ context.Context, _ string, newUsername string) error {
			assert.Equal(t, "newusername", newUsername)

			return errors.New("update error")
		}

		request := &identitysvc.UpdateUserUsernameRequest{
			NewUsername: "newusername",
		}

		result, err := service.UpdateUserUsername(buildSessionContextForTest(t), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

// mockAvatarUploadStream mocks the client streaming interface for UploadUserAvatar.
// mockAvatarUploadStream is a fake avatar upload stream. RecvMsg yields queued messages in
// order and then returns recvErr (io.EOF for a clean end of stream), which models a stream more
// directly than a call-expectation mock does.
type mockAvatarUploadStream struct {
	ctx       context.Context
	recvErr   error
	sendErr   error
	recvQueue []*uploadedmediasvc.UploadRequest
	sent      []any
	recvIndex int
}

func (m *mockAvatarUploadStream) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}

	return m.ctx
}

// next pops the next queued request, or reports the terminal error.
func (m *mockAvatarUploadStream) next() (*uploadedmediasvc.UploadRequest, error) {
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

func (m *mockAvatarUploadStream) RecvMsg(msg any) error {
	req, err := m.next()
	if err != nil {
		return err
	}

	if msg != nil {
		proto.Merge(msg.(proto.Message), req)
	}

	return nil
}

func (m *mockAvatarUploadStream) SendMsg(msg any) error {
	m.sent = append(m.sent, msg)

	return m.sendErr
}

func (m *mockAvatarUploadStream) Recv() (*uploadedmediasvc.UploadRequest, error) {
	return m.next()
}

func (m *mockAvatarUploadStream) SendAndClose(response *identitysvc.UploadUserAvatarResponse) error {
	m.sent = append(m.sent, response)

	return m.sendErr
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

		mockStream := &mockAvatarUploadStream{ctx: buildSessionContextForTest(t)}
		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq, chunkReq}

		uploadManager := service.uploadManager.(*mockuploads.UploadManagerMock)
		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }
		uploadedMediaRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return &uploadedmedia.UploadedMedia{ID: identityfakes.BuildFakeID()}, nil
		}
		identityDataManager.SetUserAvatarFunc = func(_ context.Context, _ string, _ string) error {
			return nil
		}

		stream := &grpc.GenericServerStream[uploadedmediasvc.UploadRequest, identitysvc.UploadUserAvatarResponse]{ServerStream: mockStream}
		err := service.UploadUserAvatar(stream)

		assert.NoError(t, err)
		assert.Len(t, identityDataManager.SetUserAvatarCalls(), 1)
		assert.Len(t, uploadedMediaRepo.CreateUploadedMediaCalls(), 1)
	})

	T.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

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

		mockStream := &mockAvatarUploadStream{ctx: buildSessionContextForTest(t)}
		mockStream.recvQueue = []*uploadedmediasvc.UploadRequest{metadataReq, chunkReq}

		uploadManager := service.uploadManager.(*mockuploads.UploadManagerMock)
		uploadManager.SaveFunc = func(_ context.Context, _ string, _ io.Reader, _ ...uploads.SaveOption) error { return nil }
		uploadedMediaRepo.CreateUploadedMediaFunc = func(_ context.Context, _ *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
			return &uploadedmedia.UploadedMedia{ID: identityfakes.BuildFakeID()}, nil
		}
		identityDataManager.SetUserAvatarFunc = func(_ context.Context, _ string, _ string) error {
			return errors.New("set avatar error")
		}

		stream := &grpc.GenericServerStream[uploadedmediasvc.UploadRequest, identitysvc.UploadUserAvatarResponse]{ServerStream: mockStream}
		err := service.UploadUserAvatar(stream)

		assert.Error(t, err)
		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
		assert.Len(t, identityDataManager.SetUserAvatarCalls(), 1)
		assert.Len(t, uploadedMediaRepo.CreateUploadedMediaCalls(), 1)
	})
}
