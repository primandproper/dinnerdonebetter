package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"

	"google.golang.org/grpc/codes"
)

// listQuery is what every list method needs before it can read: the caller's
// account as a tenancy scope, and the page they asked for.
//
// It is one helper rather than the same fifteen lines in five methods because
// the two failures it can produce — an unauthenticated caller and a malformed
// filter — must be told apart in every one of them, and a copy of that
// distinction is a copy that can drift.
func (s *serviceImpl) listQuery(ctx context.Context, span tracing.Span, protoFilter *filteringpb.QueryFilter) (tenancy.Scope, *filtering.QueryFilter, error) {
	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return tenancy.Scope{}, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	accountID := sessionContextData.GetActiveAccountID()
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	filter, err := filteringgrpc.FromProto(protoFilter)
	if err != nil {
		return tenancy.Scope{}, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	return ddbissuereports.Scope(accountID), filter, nil
}

// responseDetails is the envelope every response carries.
func responseDetails(span tracing.Span, scope tenancy.Scope) *types.ResponseDetails {
	return &types.ResponseDetails{
		TraceId:          span.SpanContext().TraceID().String(),
		CurrentAccountId: scope.Owner(),
	}
}

// convertPage renders a page of reports for the wire.
func convertPage(page *filtering.QueryFilteredResult[issuereports.Report]) []*issuereportssvc.IssueReport {
	results := make([]*issuereportssvc.IssueReport, 0, len(page.Data))
	for _, report := range page.Data {
		results = append(results, converters.ConvertIssueReportToGRPCIssueReport(report))
	}

	return results
}

func (s *serviceImpl) CreateIssueReport(ctx context.Context, request *issuereportssvc.CreateIssueReportRequest) (*issuereportssvc.CreateIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	accountID := sessionContextData.GetActiveAccountID()
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID).WithValue(identitykeys.UserIDKey, sessionContextData.GetUserID())

	report := converters.ConvertGRPCIssueReportCreationRequestInputToIssueReport(request.GetInput(), sessionContextData.GetUserID(), accountID)
	if report == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	// The store validates the reporter, the kind and the details, and refuses a
	// report missing any of them — see internal/services/issuereports/errors for
	// how each refusal reaches the client. There is no second validation here,
	// because a second one is one that can disagree.
	if err = s.issueReports.CreateReport(ctx, report); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating issue report")
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, report.ID)

	return &issuereportssvc.CreateIssueReportResponse{
		ResponseDetails: responseDetails(span, report.Scope),
		Created:         converters.ConvertIssueReportToGRPCIssueReport(report),
	}, nil
}

// GetIssueReport reads one of the caller's account's reports.
//
// A report belonging to another account reads as absent, which is what it is
// from here: the alternative — a permission denial — tells the caller which
// report IDs exist in accounts they cannot see.
func (s *serviceImpl) GetIssueReport(ctx context.Context, request *issuereportssvc.GetIssueReportRequest) (*issuereportssvc.GetIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, request.GetIssueReportId())

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	scope := ddbissuereports.Scope(sessionContextData.GetActiveAccountID())

	report, err := s.issueReports.GetReport(ctx, scope, request.GetIssueReportId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching issue report")
	}

	return &issuereportssvc.GetIssueReportResponse{
		ResponseDetails: responseDetails(span, scope),
		Result:          converters.ConvertIssueReportToGRPCIssueReport(report),
	}, nil
}

func (s *serviceImpl) GetIssueReports(ctx context.Context, request *issuereportssvc.GetIssueReportsRequest) (*issuereportssvc.GetIssueReportsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	scope, filter, err := s.listQuery(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.issueReports.ListReports(ctx, scope, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching issue reports")
	}

	return &issuereportssvc.GetIssueReportsResponse{
		ResponseDetails: responseDetails(span, scope),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         convertPage(page),
	}, nil
}

// GetIssueReportsByStatus is the triage queue.
func (s *serviceImpl) GetIssueReportsByStatus(ctx context.Context, request *issuereportssvc.GetIssueReportsByStatusRequest) (*issuereportssvc.GetIssueReportsByStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportStatusKey, request.GetStatus())

	scope, filter, err := s.listQuery(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportStatusKey, request.GetStatus())

	// Parsed here rather than handed to the store as typed-in text, so a queue
	// asked for by a name nothing serves is an invalid argument rather than an
	// empty page — which is exactly what a misspelled queue looks like.
	status, ok := issuereports.ParseStatus(request.GetStatus())
	if !ok {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.Wrapf(issuereports.ErrUnknownStatus, "status %q", request.GetStatus()), logger, span, codes.InvalidArgument, "invalid issue report status")
	}

	page, err := s.issueReports.ListReportsByStatus(ctx, scope, status, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching issue reports by status")
	}

	return &issuereportssvc.GetIssueReportsByStatusResponse{
		ResponseDetails: responseDetails(span, scope),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         convertPage(page),
	}, nil
}

func (s *serviceImpl) GetIssueReportsBySubjectType(ctx context.Context, request *issuereportssvc.GetIssueReportsBySubjectTypeRequest) (*issuereportssvc.GetIssueReportsBySubjectTypeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportSubjectTypeKey, request.GetSubjectType())

	scope, filter, err := s.listQuery(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.issueReports.ListReportsBySubjectType(ctx, scope, request.GetSubjectType(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching issue reports by subject type")
	}

	return &issuereportssvc.GetIssueReportsBySubjectTypeResponse{
		ResponseDetails: responseDetails(span, scope),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         convertPage(page),
	}, nil
}

func (s *serviceImpl) GetIssueReportsForSubject(ctx context.Context, request *issuereportssvc.GetIssueReportsForSubjectRequest) (*issuereportssvc.GetIssueReportsForSubjectResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportSubjectTypeKey, request.GetSubjectType())
	tracing.AttachToSpan(span, issuereportkeys.IssueReportSubjectIDKey, request.GetSubjectId())

	scope, filter, err := s.listQuery(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	page, err := s.issueReports.ListReportsForSubject(ctx, scope, request.GetSubjectType(), request.GetSubjectId(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Internal, "fetching issue reports for subject")
	}

	return &issuereportssvc.GetIssueReportsForSubjectResponse{
		ResponseDetails: responseDetails(span, scope),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         convertPage(page),
	}, nil
}

// UpdateIssueReport revises what the reporter said.
//
// It reads the report first because platform's UpdateReport takes a whole
// Report, and the read is also the authorization check: a report in another
// account is not there.
func (s *serviceImpl) UpdateIssueReport(ctx context.Context, request *issuereportssvc.UpdateIssueReportRequest) (*issuereportssvc.UpdateIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, request.GetIssueReportId())

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	scope := ddbissuereports.Scope(sessionContextData.GetActiveAccountID())

	report, err := s.issueReports.GetReport(ctx, scope, request.GetIssueReportId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching issue report")
	}

	converters.ApplyGRPCIssueReportUpdateRequestInput(report, request.GetInput())

	if err = s.issueReports.UpdateReport(ctx, report); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating issue report")
	}

	return &issuereportssvc.UpdateIssueReportResponse{
		ResponseDetails: responseDetails(span, scope),
		Updated:         converters.ConvertIssueReportToGRPCIssueReport(report),
	}, nil
}

// TransitionIssueReport moves a report through the triage lifecycle.
//
// The caller names the status it believed the report was in, and the store's
// statement requires the row to still hold it. That is what makes a queue two
// people can work: the second of two triagers resolving the same report is told
// the report moved rather than silently overwriting the first one's note.
func (s *serviceImpl) TransitionIssueReport(ctx context.Context, request *issuereportssvc.TransitionIssueReportRequest) (*issuereportssvc.TransitionIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).
		WithValue(issuereportkeys.IssueReportIDKey, request.GetIssueReportId()).
		WithValue(issuereportkeys.IssueReportStatusKey, request.GetToStatus())

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	scope := ddbissuereports.Scope(sessionContextData.GetActiveAccountID())

	from, ok := issuereports.ParseStatus(request.GetFromStatus())
	if !ok {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.Wrapf(issuereports.ErrUnknownStatus, "from status %q", request.GetFromStatus()), logger, span, codes.InvalidArgument, "invalid issue report status")
	}

	to, ok := issuereports.ParseStatus(request.GetToStatus())
	if !ok {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.Wrapf(issuereports.ErrUnknownStatus, "to status %q", request.GetToStatus()), logger, span, codes.InvalidArgument, "invalid issue report status")
	}

	report, err := s.issueReports.TransitionReport(ctx, scope, request.GetIssueReportId(), from, to, request.GetResolution())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "transitioning issue report")
	}

	return &issuereportssvc.TransitionIssueReportResponse{
		ResponseDetails: responseDetails(span, scope),
		Result:          converters.ConvertIssueReportToGRPCIssueReport(report),
	}, nil
}

func (s *serviceImpl) ArchiveIssueReport(ctx context.Context, request *issuereportssvc.ArchiveIssueReportRequest) (*issuereportssvc.ArchiveIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, request.GetIssueReportId())

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	scope := ddbissuereports.Scope(sessionContextData.GetActiveAccountID())

	// No read first: the store answers an absent, archived, or other-account
	// report as ErrReportNotFound, which is the same refusal one call earlier.
	if err = s.issueReports.ArchiveReport(ctx, scope, request.GetIssueReportId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving issue report")
	}

	return &issuereportssvc.ArchiveIssueReportResponse{
		ResponseDetails: responseDetails(span, scope),
	}, nil
}
