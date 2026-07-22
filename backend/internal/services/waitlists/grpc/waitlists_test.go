package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	waitlistmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/filtering"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func buildTestService(t *testing.T) (*serviceImpl, *waitlistmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	waitlistRepo := &waitlistmock.Repository{}

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
		waitlistsManager: waitlistRepo,
	}

	return service, waitlistRepo
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
		waitlistsManager: &waitlistmock.Repository{},
	}

	return service
}

func TestServiceImpl_CreateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		fakeInput := waitlistfakes.BuildFakeWaitlistCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWaitlist), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.WaitlistDatabaseCreationInput")).Return(fakeWaitlist, nil)

		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        fakeInput.Name,
				Description: fakeInput.Description,
				ValidUntil:  timestamppb.New(fakeInput.ValidUntil),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWaitlist.ID, response.Created.Id)
		assert.Equal(t, fakeWaitlist.Name, response.Created.Name)
		assert.Equal(t, fakeWaitlist.Description, response.Created.Description)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        "test waitlist",
				Description: "test description",
				ValidUntil:  timestamppb.New(time.Now().Add(24 * time.Hour)),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
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

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeInput := waitlistfakes.BuildFakeWaitlistCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWaitlist), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.WaitlistDatabaseCreationInput")).Return((*waitlists.Waitlist)(nil), errors.New("repository error"))

		request := &waitlistssvc.CreateWaitlistRequest{
			Input: &waitlistssvc.WaitlistCreationRequestInput{
				Name:        fakeInput.Name,
				Description: fakeInput.Description,
				ValidUntil:  timestamppb.New(fakeInput.ValidUntil),
			},
		}

		response, err := service.CreateWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlist), testutils.ContextMatcher, waitlistID).Return(fakeWaitlist, nil)

		request := &waitlistssvc.GetWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.GetWaitlist(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWaitlist.ID, response.Result.Id)
		assert.Equal(t, fakeWaitlist.Name, response.Result.Name)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlist), testutils.ContextMatcher, waitlistID).Return((*waitlists.Waitlist)(nil), errors.New("repository error"))

		request := &waitlistssvc.GetWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.GetWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWaitlists := waitlistfakes.BuildFakeWaitlistsList()

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(fakeWaitlists, nil)

		request := &waitlistssvc.GetWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlists(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWaitlists.Data))
		if len(fakeWaitlists.Data) > 0 {
			assert.Equal(t, fakeWaitlists.Data[0].ID, response.Results[0].Id)
		}

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[waitlists.Waitlist])(nil), errors.New("repository error"))

		request := &waitlistssvc.GetWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlists(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetActiveWaitlists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWaitlists := waitlistfakes.BuildFakeWaitlistsList()

		mockRepo.On(reflection.GetMethodName(mockRepo.GetActiveWaitlists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(fakeWaitlists, nil)

		request := &waitlistssvc.GetActiveWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetActiveWaitlists(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWaitlists.Data))
		if len(fakeWaitlists.Data) > 0 {
			assert.Equal(t, fakeWaitlists.Data[0].ID, response.Results[0].Id)
		}

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetActiveWaitlists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[waitlists.Waitlist])(nil), errors.New("repository error"))

		request := &waitlistssvc.GetActiveWaitlistsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetActiveWaitlists(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_UpdateWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		waitlistID := "test-waitlist-id"
		newName := "updated name"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlist), testutils.ContextMatcher, waitlistID).Return(fakeWaitlist, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateWaitlist), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.Waitlist")).Return(nil)

		request := &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: waitlistID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name: &newName,
			},
		}

		response, err := service.UpdateWaitlist(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("get waitlist error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlist), testutils.ContextMatcher, waitlistID).Return((*waitlists.Waitlist)(nil), errors.New("repository error"))

		request := &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: waitlistID,
			Input:      &waitlistssvc.WaitlistUpdateRequestInput{},
		}

		response, err := service.UpdateWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeWaitlist := waitlistfakes.BuildFakeWaitlist()
		waitlistID := "test-waitlist-id"
		newName := "updated name"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlist), testutils.ContextMatcher, waitlistID).Return(fakeWaitlist, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateWaitlist), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.Waitlist")).Return(errors.New("update error"))

		request := &waitlistssvc.UpdateWaitlistRequest{
			WaitlistId: waitlistID,
			Input: &waitlistssvc.WaitlistUpdateRequestInput{
				Name: &newName,
			},
		}

		response, err := service.UpdateWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWaitlist), testutils.ContextMatcher, waitlistID).Return(nil)

		request := &waitlistssvc.ArchiveWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.ArchiveWaitlist(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWaitlist), testutils.ContextMatcher, waitlistID).Return(errors.New("repository error"))

		request := &waitlistssvc.ArchiveWaitlistRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.ArchiveWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_WaitlistIsNotExpired(t *testing.T) {
	t.Parallel()

	t.Run("success - not expired", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.WaitlistIsNotExpired), testutils.ContextMatcher, waitlistID).Return(true, nil)

		request := &waitlistssvc.WaitlistIsNotExpiredRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.WaitlistIsNotExpired(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.True(t, response.IsNotExpired)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("success - expired", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.WaitlistIsNotExpired), testutils.ContextMatcher, waitlistID).Return(false, nil)

		request := &waitlistssvc.WaitlistIsNotExpiredRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.WaitlistIsNotExpired(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.False(t, response.IsNotExpired)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.WaitlistIsNotExpired), testutils.ContextMatcher, waitlistID).Return(false, errors.New("repository error"))

		request := &waitlistssvc.WaitlistIsNotExpiredRequest{
			WaitlistId: waitlistID,
		}

		response, err := service.WaitlistIsNotExpired(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_CreateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeInput := waitlistfakes.BuildFakeWaitlistSignupCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWaitlistSignup), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.WaitlistSignupDatabaseCreationInput")).Return(fakeSignup, nil)

		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             fakeInput.Notes,
				BelongsToWaitlist: fakeInput.BelongsToWaitlist,
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeSignup.ID, response.Created.Id)
		assert.Equal(t, fakeSignup.Notes, response.Created.Notes)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             "test notes",
				BelongsToWaitlist: "test-waitlist-id",
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _ := buildTestService(t)

		// Invalid request with empty notes
		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             "", // Invalid empty notes
				BelongsToWaitlist: "test-waitlist-id",
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeInput := waitlistfakes.BuildFakeWaitlistSignupCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWaitlistSignup), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.WaitlistSignupDatabaseCreationInput")).Return((*waitlists.WaitlistSignup)(nil), errors.New("repository error"))

		request := &waitlistssvc.CreateWaitlistSignupRequest{
			Input: &waitlistssvc.WaitlistSignupCreationRequestInput{
				Notes:             fakeInput.Notes,
				BelongsToWaitlist: fakeInput.BelongsToWaitlist,
			},
		}

		response, err := service.CreateWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = "test-user-id"
		signupID := "test-signup-id"
		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignup), testutils.ContextMatcher, signupID, waitlistID).Return(fakeSignup, nil)

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeSignup.ID, response.Result.Id)
		assert.Equal(t, fakeSignup.Notes, response.Result.Notes)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		signupID := "test-signup-id"
		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignup), testutils.ContextMatcher, signupID, waitlistID).Return((*waitlists.WaitlistSignup)(nil), errors.New("repository error"))

		request := &waitlistssvc.GetWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
		}

		response, err := service.GetWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWaitlistSignupsForWaitlist(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeSignups := waitlistfakes.BuildFakeWaitlistSignupsList()
		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignupsForWaitlist), testutils.ContextMatcher, waitlistID, testutils.QueryFilterMatcher).Return(fakeSignups, nil)

		request := &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlistID,
			Filter:     &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlistSignupsForWaitlist(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeSignups.Data))
		if len(fakeSignups.Data) > 0 {
			assert.Equal(t, fakeSignups.Data[0].ID, response.Results[0].Id)
		}

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignupsForWaitlist), testutils.ContextMatcher, waitlistID, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[waitlists.WaitlistSignup])(nil), errors.New("repository error"))

		request := &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
			WaitlistId: waitlistID,
			Filter:     &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWaitlistSignupsForWaitlist(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_UpdateWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = "test-user-id"
		signupID := "test-signup-id"
		waitlistID := "test-waitlist-id"
		newNotes := "updated notes"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignup), testutils.ContextMatcher, signupID, waitlistID).Return(fakeSignup, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateWaitlistSignup), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.WaitlistSignup")).Return(nil)

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("get signup error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		signupID := "test-signup-id"
		waitlistID := "test-waitlist-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignup), testutils.ContextMatcher, signupID, waitlistID).Return((*waitlists.WaitlistSignup)(nil), errors.New("repository error"))

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input:            &waitlistssvc.WaitlistSignupUpdateRequestInput{},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeSignup := waitlistfakes.BuildFakeWaitlistSignup()
		fakeSignup.BelongsToUser = "test-user-id"
		signupID := "test-signup-id"
		waitlistID := "test-waitlist-id"
		newNotes := "updated notes"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWaitlistSignup), testutils.ContextMatcher, signupID, waitlistID).Return(fakeSignup, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateWaitlistSignup), testutils.ContextMatcher, mock.AnythingOfType("*waitlists.WaitlistSignup")).Return(errors.New("update error"))

		request := &waitlistssvc.UpdateWaitlistSignupRequest{
			WaitlistSignupId: signupID,
			WaitlistId:       waitlistID,
			Input: &waitlistssvc.WaitlistSignupUpdateRequestInput{
				Notes: &newNotes,
			},
		}

		response, err := service.UpdateWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWaitlistSignup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		signupID := "test-signup-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWaitlistSignup), testutils.ContextMatcher, signupID).Return(nil)

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		signupID := "test-signup-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWaitlistSignup), testutils.ContextMatcher, signupID).Return(errors.New("repository error"))

		request := &waitlistssvc.ArchiveWaitlistSignupRequest{
			WaitlistSignupId: signupID,
		}

		response, err := service.ArchiveWaitlistSignup(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
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
