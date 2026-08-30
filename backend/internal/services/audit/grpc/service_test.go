package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/fakes"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	auditsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"

	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildTestService builds a service backed by the given repository mock. A nil repo gets an
// unconfigured mock, which panics if any of its methods are called.
func buildTestService(t *testing.T, auditManager *auditmock.RepositoryMock) *serviceImpl {
	t.Helper()

	if auditManager == nil {
		auditManager = &auditmock.RepositoryMock{}
	}

	return &serviceImpl{
		tracer:       tracing.NewTracerForTest(t.Name()),
		logger:       loggingnoop.NewLogger(),
		auditManager: auditManager,
	}
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		auditManager := &auditmock.RepositoryMock{}

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

		fakeAuditLogEntries := auditfakes.BuildFakeAuditLogEntriesList()
		pageSize := uint16(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		accountID := identifiers.New()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: accountID,
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
				accountID: authorization.NewAccountRolePermissionChecker(nil),
			},
		})

		mockRepo := &auditmock.RepositoryMock{
			GetAuditLogEntriesForAccountFunc: func(_ context.Context, actualAccountID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.NotNil(t, actualFilter)

				return fakeAuditLogEntries, nil
			},
		}
		service := buildTestService(t, mockRepo)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForAccount(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeAuditLogEntries.Data))

		assert.Len(t, mockRepo.GetAuditLogEntriesForAccountCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		pageSize := uint16(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		accountID := identifiers.New()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: accountID,
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
				accountID: authorization.NewAccountRolePermissionChecker(nil),
			},
		})

		mockRepo := &auditmock.RepositoryMock{
			GetAuditLogEntriesForAccountFunc: func(_ context.Context, actualAccountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
				assert.Equal(t, accountID, actualAccountID)

				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForAccount(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetAuditLogEntriesForAccountCalls(), 1)
	})
}

func TestServiceImpl_GetAuditLogEntriesForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeAuditLogEntries := auditfakes.BuildFakeAuditLogEntriesList()
		pageSize := uint16(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		userID := identifiers.New()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: identifiers.New(),
		})

		mockRepo := &auditmock.RepositoryMock{
			GetAuditLogEntriesForUserFunc: func(_ context.Context, actualUserID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
				assert.Equal(t, userID, actualUserID)
				assert.NotNil(t, actualFilter)

				return fakeAuditLogEntries, nil
			},
		}
		service := buildTestService(t, mockRepo)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForUser(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeAuditLogEntries.Data))

		assert.Len(t, mockRepo.GetAuditLogEntriesForUserCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		pageSize := uint16(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		userID := identifiers.New()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: identifiers.New(),
		})

		mockRepo := &auditmock.RepositoryMock{
			GetAuditLogEntriesForUserFunc: func(_ context.Context, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
				assert.Equal(t, userID, actualUserID)

				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForUser(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetAuditLogEntriesForUserCalls(), 1)
	})
}

func TestServiceImpl_GetAuditLogEntryByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := identifiers.New()
		sessionContextData := &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: identifiers.New(),
		}
		ctx = sessions.AttachToContext(ctx, sessionContextData)

		fakeAuditLogEntry := auditfakes.BuildFakeAuditLogEntry()
		fakeAuditLogEntry.BelongsToUser = userID
		entryID := fakeAuditLogEntry.ID

		mockRepo := &auditmock.RepositoryMock{
			GetAuditLogEntryFunc: func(_ context.Context, auditLogID string) (*audit.AuditLogEntry, error) {
				assert.Equal(t, entryID, auditLogID)

				return fakeAuditLogEntry, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		}

		response, err := service.GetAuditLogEntryByID(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeAuditLogEntry.ID, response.Result.Id)

		assert.Len(t, mockRepo.GetAuditLogEntryCalls(), 1)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		ctx = sessions.AttachToContext(ctx, &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: identifiers.New(),
		})

		entryID := "nonexistent-entry"

		mockRepo := &auditmock.RepositoryMock{
			GetAuditLogEntryFunc: func(_ context.Context, auditLogID string) (*audit.AuditLogEntry, error) {
				assert.Equal(t, entryID, auditLogID)

				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		request := &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		}

		response, err := service.GetAuditLogEntryByID(ctx, request)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetAuditLogEntryCalls(), 1)
	})
}
