package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/mock"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	"github.com/primandproper/platform-go/v10/filtering"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testSessionUserID    = identityfakes.BuildFakeID()
	testSessionAccountID = identityfakes.BuildFakeID()
)

func buildTestService(t *testing.T) (*serviceImpl, *waitlistmock.RepositoryMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	waitlistRepo := &waitlistmock.RepositoryMock{}

	service := &serviceImpl{
		tracer:           tracer,
		logger:           logger,
		waitlistsManager: waitlistRepo,
	}

	return service, waitlistRepo
}

func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: testSessionAccountID,
		Requester: sessions.RequesterInfo{
			UserID:             testSessionUserID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
		},
	})
}

func buildAdminSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: testSessionAccountID,
		Requester: sessions.RequesterInfo{
			UserID:             testSessionUserID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceAdminRole.String()}, authorization.ServiceAdminPermissions),
		},
	})
}

func TestServiceImpl_CreateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		fakeInput := waitlistfakes.BuildFakeWaitlistCreationRequestInput()

		mockRepo.CreateWaitlistFunc = func(_ context.Context, _ *waitlists.WaitlistDatabaseCreationInput) (*waitlists.Waitlist, error) {
			return fakeWaitlist, nil
		}

		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        fakeInput.Name,
				Description: fakeInput.Description,
				ValidUntil:  timestamppb.New(fakeInput.ValidUntil),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWaitlist.ID, response.Created.Id)
		assert.Equal(t, fakeWaitlist.Name, response.Created.Name)
		assert.Equal(t, fakeWaitlist.Description, response.Created.Description)

		assert.Len(t, mockRepo.CreateWaitlistCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        "test waitlist",
				Description: "test description",
				ValidUntil:  timestamppb.New(time.Now().Add(24 * time.Hour)),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _ := buildTestService(t)

		// Invalid request with empty name
		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        "", // Invalid empty name
				Description: "test description",
				ValidUntil:  timestamppb.New(time.Now().Add(24 * time.Hour)),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeInput := waitlistfakes.BuildFakeWaitlistCreationRequestInput()

		mockRepo.CreateWaitlistFunc = func(_ context.Context, _ *waitlists.WaitlistDatabaseCreationInput) (*waitlists.Waitlist, error) {
			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        fakeInput.Name,
				Description: fakeInput.Description,
				ValidUntil:  timestamppb.New(fakeInput.ValidUntil),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.CreateWaitlistCalls(), 1)
	})
}

func TestServiceImpl_GetWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		waitlistID := "test-waitlist-id"

		mockRepo.GetWaitlistFunc = func(_ context.Context, actualWaitlistID string) (*waitlists.Waitlist, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeWaitlist, nil
		}

		request := &waitlistssvc.GetWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.GetWaitlist(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWaitlist.ID, response.Result.Id)
		assert.Equal(t, fakeWaitlist.Name, response.Result.Name)

		assert.Len(t, mockRepo.GetWaitlistCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.GetWaitlistFunc = func(_ context.Context, actualWaitlistID string) (*waitlists.Waitlist, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.GetWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.GetWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistCalls(), 1)
	})
}

func TestServiceImpl_GetWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWaitlists := waitlistfakes.BuildFakeWaitlistsList()

		mockRepo.GetWaitlistsFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
			return fakeWaitlists, nil
		}

		request := &waitlistssvc.GetWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlists(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWaitlists.Data))
		if len(fakeWaitlists.Data) > 0 {
			assert.Equal(t, fakeWaitlists.Data[0].ID, response.Results[0].Id)
		}

		assert.Len(t, mockRepo.GetWaitlistsCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		mockRepo.GetWaitlistsFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.GetWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlists(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistsCalls(), 1)
	})
}

func TestServiceImpl_GetActiveWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWaitlists := waitlistfakes.BuildFakeWaitlistsList()

		mockRepo.GetActiveWaitlistsFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
			return fakeWaitlists, nil
		}

		request := &waitlistssvc.GetActiveWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetActiveWaitlists(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWaitlists.Data))
		if len(fakeWaitlists.Data) > 0 {
			assert.Equal(t, fakeWaitlists.Data[0].ID, response.Results[0].Id)
		}

		assert.Len(t, mockRepo.GetActiveWaitlistsCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		mockRepo.GetActiveWaitlistsFunc = func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.Waitlist], error) {
			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.GetActiveWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetActiveWaitlists(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetActiveWaitlistsCalls(), 1)
	})
}

func TestServiceImpl_UpdateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		waitlistID := "test-waitlist-id"
		newName := "updated name"

		mockRepo.GetWaitlistFunc = func(_ context.Context, actualWaitlistID string) (*waitlists.Waitlist, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeWaitlist, nil
		}
		mockRepo.UpdateWaitlistFunc = func(_ context.Context, _ *waitlists.Waitlist) error {
			return nil
		}

		request := &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: waitlistID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name: &newName,
			},
		}

		response, err := service.UpdateWaitlist(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockRepo.GetWaitlistCalls(), 1)
		assert.Len(t, mockRepo.UpdateWaitlistCalls(), 1)
	})

	t.Run("get waitlist error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.GetWaitlistFunc = func(_ context.Context, actualWaitlistID string) (*waitlists.Waitlist, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: waitlistID,
			Input:      &waitlistssvc.WaitlistUpdateRequestInput{},
		}

		response, err := service.UpdateWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistCalls(), 1)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		waitlistID := "test-waitlist-id"
		newName := "updated name"

		mockRepo.GetWaitlistFunc = func(_ context.Context, actualWaitlistID string) (*waitlists.Waitlist, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeWaitlist, nil
		}
		mockRepo.UpdateWaitlistFunc = func(_ context.Context, _ *waitlists.Waitlist) error {
			return errors.New("update error")
		}

		request := &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: waitlistID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name: &newName,
			},
		}

		response, err := service.UpdateWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistCalls(), 1)
		assert.Len(t, mockRepo.UpdateWaitlistCalls(), 1)
	})
}

func TestServiceImpl_ArchiveWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.ArchiveWaitlistFunc = func(_ context.Context, actualWaitlistID string) error {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return nil
		}

		request := &waitlistssvc.ArchiveWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.ArchiveWaitlist(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockRepo.ArchiveWaitlistCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.ArchiveWaitlistFunc = func(_ context.Context, actualWaitlistID string) error {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return errors.New("repository error")
		}

		request := &waitlistssvc.ArchiveWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.ArchiveWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.ArchiveWaitlistCalls(), 1)
	})
}

func TestServiceImpl_WaitlistIsNotExpired(t *testing.T) {
	t.Parallel()

	t.Run("success - not expired", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.WaitlistIsNotExpiredFunc = func(_ context.Context, actualWaitlistID string) (bool, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return true, nil
		}

		request := &waitlistssvc.WaitlistIsNotExpiredRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.WaitlistIsNotExpired(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.True(t, response.IsNotExpired)

		assert.Len(t, mockRepo.WaitlistIsNotExpiredCalls(), 1)
	})

	t.Run("success - expired", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.WaitlistIsNotExpiredFunc = func(_ context.Context, actualWaitlistID string) (bool, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return false, nil
		}

		request := &waitlistssvc.WaitlistIsNotExpiredRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.WaitlistIsNotExpired(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.False(t, response.IsNotExpired)

		assert.Len(t, mockRepo.WaitlistIsNotExpiredCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.WaitlistIsNotExpiredFunc = func(_ context.Context, actualWaitlistID string) (bool, error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return false, errors.New("repository error")
		}

		request := &waitlistssvc.WaitlistIsNotExpiredRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.WaitlistIsNotExpired(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.WaitlistIsNotExpiredCalls(), 1)
	})
}

func TestServiceImpl_CreateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeInput := waitlistfakes.BuildFakeWaitlistSignupCreationRequestInput()

		mockRepo.CreateWaitlistSignupFunc = func(_ context.Context, _ *waitlists.WaitlistSignupDatabaseCreationInput) (*waitlists.WaitlistSignup, error) {
			return fakeSignup, nil
		}

		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             fakeInput.Notes,
				BelongsToWaitlist: fakeInput.BelongsToWaitlist,
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeSignup.ID, response.Created.Id)
		assert.Equal(t, fakeSignup.Notes, response.Created.Notes)

		assert.Len(t, mockRepo.CreateWaitlistSignupCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             "test notes",
				BelongsToWaitlist: "test-waitlist-id",
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _ := buildTestService(t)

		// Invalid request with empty notes
		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             "", // Invalid empty notes
				BelongsToWaitlist: "test-waitlist-id",
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeInput := waitlistfakes.BuildFakeWaitlistSignupCreationRequestInput()

		mockRepo.CreateWaitlistSignupFunc = func(_ context.Context, _ *waitlists.WaitlistSignupDatabaseCreationInput) (*waitlists.WaitlistSignup, error) {
			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             fakeInput.Notes,
				BelongsToWaitlist: fakeInput.BelongsToWaitlist,
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.CreateWaitlistSignupCalls(), 1)
	})
}

func TestServiceImpl_GetWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = testSessionUserID
		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignup, nil
		}

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeSignup.ID, response.Result.Id)
		assert.Equal(t, fakeSignup.Notes, response.Result.Notes)

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
	})

	t.Run("as service admin for another user's signup", func(t *testing.T) {
		t.Parallel()

		ctx := buildAdminSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignup, nil
		}

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
	})

	t.Run("as another user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignup, nil
		}

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: identityfakes.BuildFakeID(),
			WaitlistId:       identityfakes.BuildFakeID(),
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
	})
}

func TestServiceImpl_GetWaitlistSignupsForWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success as service admin", func(t *testing.T) {
		t.Parallel()

		ctx := buildAdminSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignups := waitlistfakes.BuildFakeWaitlistSignupsList()
		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupsForWaitlistFunc = func(_ context.Context, actualWaitlistID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignups, nil
		}

		request := &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlistID,
			Filter:     &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlistSignupsForWaitlist(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeSignups.Data))
		if len(fakeSignups.Data) > 0 {
			assert.Equal(t, fakeSignups.Data[0].ID, response.Results[0].Id)
		}

		assert.Len(t, mockRepo.GetWaitlistSignupsForWaitlistCalls(), 1)
	})

	t.Run("as regular user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, _ := buildTestService(t)

		request := &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: identityfakes.BuildFakeID(),
			Filter:     &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlistSignupsForWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: identityfakes.BuildFakeID(),
			Filter:     &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlistSignupsForWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildAdminSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupsForWaitlistFunc = func(_ context.Context, actualWaitlistID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[waitlists.WaitlistSignup], error) {
			assert.Equal(t, waitlistID, actualWaitlistID)

			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlistID,
			Filter:     &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlistSignupsForWaitlist(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupsForWaitlistCalls(), 1)
	})
}

func TestServiceImpl_UpdateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = testSessionUserID
		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()
		newNotes := "updated notes"

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignup, nil
		}
		mockRepo.UpdateWaitlistSignupFunc = func(_ context.Context, _ *waitlists.WaitlistSignup) error {
			return nil
		}

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
		assert.Len(t, mockRepo.UpdateWaitlistSignupCalls(), 1)
	})

	t.Run("as another user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()
		newNotes := "updated notes"

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignup, nil
		}

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: identityfakes.BuildFakeID(),
			WaitlistId:       identityfakes.BuildFakeID(),
			Input:            &waitlistssvc.WaitlistSignupUpdateRequestInput{},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("get signup error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input:            &waitlistssvc.WaitlistSignupUpdateRequestInput{},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = testSessionUserID
		signupID := identityfakes.BuildFakeID()
		waitlistID := identityfakes.BuildFakeID()
		newNotes := "updated notes"

		mockRepo.GetWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string, actualWaitlistID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)
			assert.Equal(t, waitlistID, actualWaitlistID)

			return fakeSignup, nil
		}
		mockRepo.UpdateWaitlistSignupFunc = func(_ context.Context, _ *waitlists.WaitlistSignup) error {
			return errors.New("update error")
		}

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupCalls(), 1)
		assert.Len(t, mockRepo.UpdateWaitlistSignupCalls(), 1)
	})
}

func TestServiceImpl_ArchiveWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = testSessionUserID
		signupID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupByIDFunc = func(_ context.Context, waitlistSignupID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)

			return fakeSignup, nil
		}
		mockRepo.ArchiveWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string) error {
			assert.Equal(t, signupID, waitlistSignupID)

			return nil
		}

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		assert.Len(t, mockRepo.GetWaitlistSignupByIDCalls(), 1)
		assert.Len(t, mockRepo.ArchiveWaitlistSignupCalls(), 1)
	})

	t.Run("as service admin for another user's signup", func(t *testing.T) {
		t.Parallel()

		ctx := buildAdminSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		signupID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupByIDFunc = func(_ context.Context, waitlistSignupID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)

			return fakeSignup, nil
		}
		mockRepo.ArchiveWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string) error {
			assert.Equal(t, signupID, waitlistSignupID)

			return nil
		}

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)

		assert.Len(t, mockRepo.GetWaitlistSignupByIDCalls(), 1)
		assert.Len(t, mockRepo.ArchiveWaitlistSignupCalls(), 1)
	})

	t.Run("as another user", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		signupID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupByIDFunc = func(_ context.Context, waitlistSignupID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)

			return fakeSignup, nil
		}

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupByIDCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: identityfakes.BuildFakeID(),
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("get signup error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		signupID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupByIDFunc = func(_ context.Context, waitlistSignupID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)

			return nil, errors.New("repository error")
		}

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupByIDCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = testSessionUserID
		signupID := identityfakes.BuildFakeID()

		mockRepo.GetWaitlistSignupByIDFunc = func(_ context.Context, waitlistSignupID string) (*waitlists.WaitlistSignup, error) {
			assert.Equal(t, signupID, waitlistSignupID)

			return fakeSignup, nil
		}
		mockRepo.ArchiveWaitlistSignupFunc = func(_ context.Context, waitlistSignupID string) error {
			assert.Equal(t, signupID, waitlistSignupID)

			return errors.New("repository error")
		}

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetWaitlistSignupByIDCalls(), 1)
		assert.Len(t, mockRepo.ArchiveWaitlistSignupCalls(), 1)
	})
}

func TestServiceImpl_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	t.Run("implements WaitlistsServiceServer", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		assert.Implements(t, (*waitlistssvc.WaitlistsServiceServer)(nil), service)
	})
}
