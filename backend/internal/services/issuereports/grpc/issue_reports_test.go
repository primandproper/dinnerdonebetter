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
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/filtering"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *issuereportmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	issueReportRepo := &issuereportmock.Repository{}

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
		issueReportsManager: issueReportRepo,
	}

	return service, issueReportRepo
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
		issueReportsManager: &issuereportmock.Repository{},
	}

	return service
}

func TestServiceImpl_CreateIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeInput := issuereportfakes.BuildFakeIssueReportCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateIssueReport), testutils.ContextMatcher, mock.AnythingOfType("*issuereports.IssueReportDatabaseCreationInput")).Return(fakeIssueReport, nil)

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

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeInput := issuereportfakes.BuildFakeIssueReportCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateIssueReport), testutils.ContextMatcher, mock.AnythingOfType("*issuereports.IssueReportDatabaseCreationInput")).Return(nil, errors.New("repository error"))

		request := &issuereportssvc.CreateIssueReportRequest{
			Input: converters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(fakeInput),
		}

		response, err := service.CreateIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "test-account-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(fakeIssueReport, nil)

		request := &issuereportssvc.GetIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.GetIssueReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeIssueReport.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "different-account-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(fakeIssueReport, nil)

		request := &issuereportssvc.GetIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.GetIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetIssueReports(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReports), testutils.ContextMatcher, mock.AnythingOfType("*filtering.QueryFilter")).Return(fakeIssueReports, nil)

		request := &issuereportssvc.GetIssueReportsRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReports(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

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

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReportsForAccount), testutils.ContextMatcher, "test-account-id", mock.AnythingOfType("*filtering.QueryFilter")).Return(fakeIssueReports, nil)

		request := &issuereportssvc.GetIssueReportsForAccountRequest{
			AccountId: "test-account-id",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 2)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeIssueReports := &filtering.QueryFilteredResult[issuereports.IssueReport]{
			Data: []*issuereports.IssueReport{
				issuereportfakes.BuildFakeIssueReport(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    1,
				FilteredCount: 1,
			},
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReportsForTable), testutils.ContextMatcher, "recipes", mock.AnythingOfType("*filtering.QueryFilter")).Return(fakeIssueReports, nil)

		request := &issuereportssvc.GetIssueReportsForTableRequest{
			TableName: "recipes",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForTable(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeIssueReports := &filtering.QueryFilteredResult[issuereports.IssueReport]{
			Data: []*issuereports.IssueReport{
				issuereportfakes.BuildFakeIssueReport(),
			},
			Pagination: filtering.Pagination{
				TotalCount:    1,
				FilteredCount: 1,
			},
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReportsForRecord), testutils.ContextMatcher, "recipes", "some-record-id", mock.AnythingOfType("*filtering.QueryFilter")).Return(fakeIssueReports, nil)

		request := &issuereportssvc.GetIssueReportsForRecordRequest{
			TableName: "recipes",
			RecordId:  "some-record-id",
			Filter:    &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetIssueReportsForRecord(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Results, 1)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "test-account-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(fakeIssueReport, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateIssueReport), testutils.ContextMatcher, mock.AnythingOfType("*issuereports.IssueReport")).Return(nil)

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

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "different-account-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(fakeIssueReport, nil)

		request := &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{},
		}

		response, err := service.UpdateIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "test-account-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(fakeIssueReport, nil)
		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(nil)

		request := &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.ArchiveIssueReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		fakeIssueReport := issuereportfakes.BuildFakeIssueReport()
		fakeIssueReport.BelongsToAccount = "different-account-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetIssueReport), testutils.ContextMatcher, fakeIssueReport.ID).Return(fakeIssueReport, nil)

		request := &issuereportssvc.ArchiveIssueReportRequest{
			IssueReportId: fakeIssueReport.ID,
		}

		response, err := service.ArchiveIssueReport(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}
