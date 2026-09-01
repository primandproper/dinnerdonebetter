package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	commentsconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	comments "github.com/primandproper/platform-go/v13/comments"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

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

	comment := commentsconverters.ConvertProtoCommentCreationRequestInputToDomain(
		request.GetInput(),
		comments.Target{
			Type: issuereports.CommentTargetTypeIssueReports,
			ID:   request.GetIssueReportId(),
		},
		sessionContextData.GetUserID(),
	)
	if comment == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment.Scope = ddbcomments.Scope()

	// The issue report is no longer read here first. The target catalog registers
	// an existence check for this type, so the store refuses a comment about a
	// report that is not there — one read instead of two, and the refusal arrives
	// as ErrTargetNotFound rather than as whatever the read happened to fail with.
	if err = s.comments.CreateComment(ctx, comment); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, comment.ID)

	return &issuereportssvc.AddCommentToIssueReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: commentsconverters.ConvertCommentToGRPCComment(comment),
	}, nil
}
