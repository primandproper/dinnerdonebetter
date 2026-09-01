package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	commentsconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	comments "github.com/primandproper/platform-go/v13/comments"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

// AddCommentToIssueReport files a comment against one of the caller's account's
// issue reports.
//
// The report is read first, and that read does two jobs. It is the existence
// check the comment store's target catalog cannot run for this type — the
// catalog's hook is handed the comment's scope, and comments are filed globally
// while issue reports are filed per account, so there is no scope the hook could
// read a report in (see internal/build/comments). And it is the authorization
// check: a report in another account is not there, so a comment cannot be filed
// against one.
func (s *serviceImpl) AddCommentToIssueReport(ctx context.Context, request *issuereportssvc.AddCommentToIssueReportRequest) (*issuereportssvc.AddCommentToIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		issuereportkeys.IssueReportIDKey: request.GetIssueReportId(),
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	scope := ddbissuereports.Scope(sessionContextData.GetActiveAccountID())

	if _, err = s.issueReports.GetReport(ctx, scope, request.GetIssueReportId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching issue report")
	}

	comment := commentsconverters.ConvertProtoCommentCreationRequestInputToDomain(
		request.GetInput(),
		comments.Target{
			Type: ddbissuereports.CommentTargetTypeIssueReports,
			ID:   request.GetIssueReportId(),
		},
		sessionContextData.GetUserID(),
	)
	if comment == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment.Scope = ddbcomments.Scope()

	if err = s.comments.CreateComment(ctx, comment); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, comment.ID)

	return &issuereportssvc.AddCommentToIssueReportResponse{
		ResponseDetails: responseDetails(span, scope),
		Comment:         commentsconverters.ConvertCommentToGRPCComment(comment),
	}, nil
}
