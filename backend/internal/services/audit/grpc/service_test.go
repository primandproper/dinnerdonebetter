package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	auditfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/fakes"
	auditmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	auditsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *auditmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	auditManager := &auditmock.Repository{}

	service := &serviceImpl{
		tracer:       tracer,
		logger:       logger,
		auditManager: auditManager,
	}

	return service, auditManager
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		auditManager := &auditmock.Repository{}

		service := NewService(logger, tracerProvider, auditManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*auditsvc.AuditServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, auditManager, impl.auditManager)
	})
}

func TestServiceImpl_GetAuditLogEntriesForAccount(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeAuditLogEntries := auditfakes.BuildFakeAuditLogEntriesList()
		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		accountID := identifiers.New()

		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: accountID,
		})

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForAccount), testutils.ContextMatcher, accountID, testutils.QueryFilterMatcher).Return(fakeAuditLogEntries, nil)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeAuditLogEntries.Data))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		accountID := identifiers.New()

		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: accountID,
		})

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForAccount), testutils.ContextMatcher, accountID, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[audit.AuditLogEntry])(nil), errors.New("repository error"))

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetAuditLogEntriesForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeAuditLogEntries := auditfakes.BuildFakeAuditLogEntriesList()
		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		userID := identifiers.New()

		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: identifiers.New(),
		})

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForUser), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return(fakeAuditLogEntries, nil)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeAuditLogEntries.Data))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		userID := identifiers.New()

		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: identifiers.New(),
		})

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForUser), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[audit.AuditLogEntry])(nil), errors.New("repository error"))

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetAuditLogEntryByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		userID := identifiers.New()
		sessionContextData := &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: identifiers.New(),
		}
		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, sessionContextData)

		fakeAuditLogEntry := auditfakes.BuildFakeAuditLogEntry()
		fakeAuditLogEntry.BelongsToUser = userID
		entryID := fakeAuditLogEntry.ID

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntry), testutils.ContextMatcher, entryID).Return(fakeAuditLogEntry, nil)

		request := &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		}

		response, err := service.GetAuditLogEntryByID(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeAuditLogEntry.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: identifiers.New(),
		})

		entryID := "nonexistent-entry"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntry), testutils.ContextMatcher, entryID).Return((*audit.AuditLogEntry)(nil), errors.New("repository error"))

		request := &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		}

		response, err := service.GetAuditLogEntryByID(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}
