package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateIssueReport(ctx context.Context, request *issuereportssvc.CreateIssueReportRequest) (*issuereportssvc.CreateIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID()).WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCIssueReportCreationRequestInputToIssueReportDatabaseCreationInput(request.Input, sessionContextData.GetUserID(), sessionContextData.GetActiveAccountID())
	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate issue report creation request")
	}

	created, err := s.issueReportsManager.CreateIssueReport(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create issue report")
	}

	x := &issuereportssvc.CreateIssueReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Created: converters.ConvertIssueReportToGRPCIssueReport(created),
	}

	return x, nil
}

func (s *serviceImpl) GetIssueReport(ctx context.Context, request *issuereportssvc.GetIssueReportRequest) (*issuereportssvc.GetIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, request.IssueReportId)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	issueReport, err := s.issueReportsManager.GetIssueReport(ctx, request.IssueReportId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue report")
	}

	// Verify the issue report belongs to the user's account
	if issueReport.BelongsToAccount != sessionContextData.GetActiveAccountID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "issue report does not belong to account")
	}

	x := &issuereportssvc.GetIssueReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Result: converters.ConvertIssueReportToGRPCIssueReport(issueReport),
	}

	return x, nil
}

func (s *serviceImpl) GetIssueReports(ctx context.Context, request *issuereportssvc.GetIssueReportsRequest) (*issuereportssvc.GetIssueReportsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	issueReports, err := s.issueReportsManager.GetIssueReports(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue reports")
	}

	x := &issuereportssvc.GetIssueReportsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Pagination: filteringgrpc.PaginationToProto(issueReports.Pagination),
	}

	for _, issueReport := range issueReports.Data {
		x.Results = append(x.Results, converters.ConvertIssueReportToGRPCIssueReport(issueReport))
	}

	return x, nil
}

func (s *serviceImpl) GetIssueReportsForAccount(ctx context.Context, request *issuereportssvc.GetIssueReportsForAccountRequest) (*issuereportssvc.GetIssueReportsForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	issueReports, err := s.issueReportsManager.GetIssueReportsForAccount(ctx, sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue reports for account")
	}

	x := &issuereportssvc.GetIssueReportsForAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Pagination: filteringgrpc.PaginationToProto(issueReports.Pagination),
	}

	for _, issueReport := range issueReports.Data {
		x.Results = append(x.Results, converters.ConvertIssueReportToGRPCIssueReport(issueReport))
	}

	return x, nil
}

func (s *serviceImpl) GetIssueReportsForTable(ctx context.Context, request *issuereportssvc.GetIssueReportsForTableRequest) (*issuereportssvc.GetIssueReportsForTableResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue("table_name", request.TableName)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	issueReports, err := s.issueReportsManager.GetIssueReportsForTable(ctx, request.TableName, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue reports for table")
	}

	x := &issuereportssvc.GetIssueReportsForTableResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Pagination: filteringgrpc.PaginationToProto(issueReports.Pagination),
	}

	for _, issueReport := range issueReports.Data {
		x.Results = append(x.Results, converters.ConvertIssueReportToGRPCIssueReport(issueReport))
	}

	return x, nil
}

func (s *serviceImpl) GetIssueReportsForRecord(ctx context.Context, request *issuereportssvc.GetIssueReportsForRecordRequest) (*issuereportssvc.GetIssueReportsForRecordResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue("table_name", request.TableName).WithValue("record_id", request.RecordId)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	issueReports, err := s.issueReportsManager.GetIssueReportsForRecord(ctx, request.TableName, request.RecordId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue reports for record")
	}

	x := &issuereportssvc.GetIssueReportsForRecordResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Pagination: filteringgrpc.PaginationToProto(issueReports.Pagination),
	}

	for _, issueReport := range issueReports.Data {
		x.Results = append(x.Results, converters.ConvertIssueReportToGRPCIssueReport(issueReport))
	}

	return x, nil
}

func (s *serviceImpl) UpdateIssueReport(ctx context.Context, request *issuereportssvc.UpdateIssueReportRequest) (*issuereportssvc.UpdateIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, request.IssueReportId)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	// Fetch the existing issue report
	issueReport, err := s.issueReportsManager.GetIssueReport(ctx, request.IssueReportId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue report")
	}

	// Verify the issue report belongs to the user's account
	if issueReport.BelongsToAccount != sessionContextData.GetActiveAccountID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "issue report does not belong to account")
	}

	// Apply updates
	updateInput := converters.ConvertGRPCIssueReportUpdateRequestInputToIssueReportUpdateRequestInput(request.Input)
	if err = updateInput.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate issue report update request")
	}

	issueReport.Update(updateInput)

	if err = s.issueReportsManager.UpdateIssueReport(ctx, issueReport); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update issue report")
	}

	x := &issuereportssvc.UpdateIssueReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
		Updated: converters.ConvertIssueReportToGRPCIssueReport(issueReport),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveIssueReport(ctx context.Context, request *issuereportssvc.ArchiveIssueReportRequest) (*issuereportssvc.ArchiveIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, request.IssueReportId)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	// Fetch the existing issue report to verify ownership
	issueReport, err := s.issueReportsManager.GetIssueReport(ctx, request.IssueReportId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch issue report")
	}

	// Verify the issue report belongs to the user's account
	if issueReport.BelongsToAccount != sessionContextData.GetActiveAccountID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "issue report does not belong to account")
	}

	if err = s.issueReportsManager.ArchiveIssueReport(ctx, request.IssueReportId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive issue report")
	}

	x := &issuereportssvc.ArchiveIssueReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.GetActiveAccountID(),
		},
	}

	return x, nil
}
