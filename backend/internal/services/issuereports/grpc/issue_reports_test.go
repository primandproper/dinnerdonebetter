package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	issuereportmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/issuereports/grpc/converters"

	"github.com/primandproper/platform-go/v7/filtering"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	"github.com/primandproper/platform-go/v7/observability/tracing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildTestService builds a service backed by the given repository mock. A nil repo gets an
// unconfigured mock, which panics if any of its methods are called.
func buildTestService(t *testing.T, issueReportRepo *issuereportmock.RepositoryMock) *serviceImpl {
	t.Helper()

	if issueReportRepo == nil {
		issueReportRepo = &issuereportmock.RepositoryMock{}
	}

	return &serviceImpl{
		tracer: tracing.NewTracerForTest(t.Name()),
		logger: loggingnoop.NewLogger(),
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: "test-account-id",
				Requester: sessions.RequesterInfo{
					UserID: "test-user-id",
				},
			}, nil
		},
		issueReportsManager: issueReportRepo,
	}
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	return &serviceImpl{
		tracer: tracing.NewTracerForTest(t.Name()),
		logger: loggingnoop.NewLogger(),
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
		issueReportsManager: &issuereportmock.RepositoryMock{},
	}
}

func TestServiceImpl_CreateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeInput := issuereportfakes.BuildFakeIssueReportCreationRequestInput()

		mockRepo := &issuereportmock.RepositoryMock{
			CreateIssueReportFunc: func(_ context.Context, input *issuereports.IssueReportDatabaseCreationInput) (*issuereports.IssueReport, error) {
				assert.NotNil(t, input)

				return fakeIssueReport, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.CreateIssueReportRequest{
			Input: converters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(fakeInput),
		}

		response, err := service.CreateIssueReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeIssueReport.ID, response.Created.Id)
		assert.Equal(t, fakeIssueReport.IssueType, response.Created.IssueType)
		assert.Equal(t, fakeIssueReport.Details, response.Created.Details)

		assert.Len(t, mockRepo.CreateIssueReportCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.CreateIssueReportRequest{
			Input: &issuereportssvc.IssueReportCreationRequestInput{},
		}

		response, err := service.CreateIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeInput := issuereportfakes.BuildFakeIssueReportCreationRequestInput()

		mockRepo := &issuereportmock.RepositoryMock{
			CreateIssueReportFunc: func(_ context.Context, _ *issuereports.IssueReportDatabaseCreationInput) (*issuereports.IssueReport, error) {
				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.CreateIssueReportRequest{
			Input: converters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(fakeInput),
		}

		response, err := service.CreateIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.CreateIssueReportCalls(), 1)
	})
}

func TestServiceImpl_GetIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "test-account-id"

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*issuereports.IssueReport, error) {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return fakeIssueReport, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.GetIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.GetIssueReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeIssueReport.ID, response.Result.Id)

		assert.Len(t, mockRepo.GetIssueReportCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.GetIssueReportRequest{
			IssueReportId: "some-id",
		}

		response, err := service.GetIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "different-account-id"

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*issuereports.IssueReport, error) {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return fakeIssueReport, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.GetIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.GetIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetIssueReportCalls(), 1)
	})
}

func TestServiceImpl_GetIssueReports(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReports := &filtering.QueryFilteredResult[issuereports.IssueReport]{
			Data: []*issuereports.IssueReport{
				issuereportfakes.BuildFakeIssueReport(),
				issuereportfakes.BuildFakeIssueReport(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    2,
				FilteredCount: 2,
			},
		}

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportsFunc: func(_ context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
				assert.NotNil(t, filter)

				return fakeIssueReports, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.GetIssueReportsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReports(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		assert.Len(t, mockRepo.GetIssueReportsCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.GetIssueReportsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReports(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestServiceImpl_GetIssueReportsForAccount(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReports := &filtering.QueryFilteredResult[issuereports.IssueReport]{
			Data: []*issuereports.IssueReport{
				issuereportfakes.BuildFakeIssueReport(),
				issuereportfakes.BuildFakeIssueReport(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    2,
				FilteredCount: 2,
			},
		}

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportsForAccountFunc: func(_ context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
				assert.Equal(t, "test-account-id", accountID)
				assert.NotNil(t, filter)

				return fakeIssueReports, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.GetIssueReportsForAccountRequest{
			AccountId: "test-account-id",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		assert.Len(t, mockRepo.GetIssueReportsForAccountCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.GetIssueReportsForAccountRequest{
			AccountId: "test-account-id",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestServiceImpl_GetIssueReportsForTable(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReports := &filtering.QueryFilteredResult[issuereports.IssueReport]{
			Data: []*issuereports.IssueReport{
				issuereportfakes.BuildFakeIssueReport(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    1,
				FilteredCount: 1,
			},
		}

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportsForTableFunc: func(_ context.Context, tableName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
				assert.Equal(t, "recipes", tableName)
				assert.NotNil(t, filter)

				return fakeIssueReports, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.GetIssueReportsForTableRequest{
			TableName: "recipes",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForTable(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)

		assert.Len(t, mockRepo.GetIssueReportsForTableCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.GetIssueReportsForTableRequest{
			TableName: "recipes",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForTable(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestServiceImpl_GetIssueReportsForRecord(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReports := &filtering.QueryFilteredResult[issuereports.IssueReport]{
			Data: []*issuereports.IssueReport{
				issuereportfakes.BuildFakeIssueReport(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    1,
				FilteredCount: 1,
			},
		}

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportsForRecordFunc: func(_ context.Context, tableName, recordID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
				assert.Equal(t, "recipes", tableName)
				assert.Equal(t, "some-record-id", recordID)
				assert.NotNil(t, filter)

				return fakeIssueReports, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.GetIssueReportsForRecordRequest{
			TableName: "recipes",
			RecordId:  "some-record-id",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForRecord(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)

		assert.Len(t, mockRepo.GetIssueReportsForRecordCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.GetIssueReportsForRecordRequest{
			TableName: "recipes",
			RecordId:  "some-record-id",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForRecord(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestServiceImpl_UpdateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "test-account-id"

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*issuereports.IssueReport, error) {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return fakeIssueReport, nil
			},
			UpdateIssueReportFunc: func(_ context.Context, issueReport *issuereports.IssueReport) error {
				assert.NotNil(t, issueReport)

				return nil
			},
		}
		service := buildTestService(t, mockRepo)

		newDetails := "Updated details"
		request := &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
			Input: &issuereportssvc.IssueReportUpdateRequestInput{
				Details: &newDetails,
			},
		}

		response, err := service.UpdateIssueReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Updated)

		assert.Len(t, mockRepo.GetIssueReportCalls(), 1)
		assert.Len(t, mockRepo.UpdateIssueReportCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: "some-id",
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{},
		}

		response, err := service.UpdateIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "different-account-id"

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*issuereports.IssueReport, error) {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return fakeIssueReport, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{},
		}

		response, err := service.UpdateIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetIssueReportCalls(), 1)
		assert.Empty(t, mockRepo.UpdateIssueReportCalls())
	})
}

func TestServiceImpl_ArchiveIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "test-account-id"

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*issuereports.IssueReport, error) {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return fakeIssueReport, nil
			},
			ArchiveIssueReportFunc: func(_ context.Context, issueReportID string) error {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.ArchiveIssueReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		assert.Len(t, mockRepo.GetIssueReportCalls(), 1)
		assert.Len(t, mockRepo.ArchiveIssueReportCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: "some-id",
		}

		response, err := service.ArchiveIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("permission denied - different account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "different-account-id"

		mockRepo := &issuereportmock.RepositoryMock{
			GetIssueReportFunc: func(_ context.Context, issueReportID string) (*issuereports.IssueReport, error) {
				assert.Equal(t, fakeIssueReport.ID, issueReportID)

				return fakeIssueReport, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.ArchiveIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		assert.Len(t, mockRepo.GetIssueReportCalls(), 1)
		assert.Empty(t, mockRepo.ArchiveIssueReportCalls())
	})
}
