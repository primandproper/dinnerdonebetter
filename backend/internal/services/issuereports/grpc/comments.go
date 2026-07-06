package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	commentsconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v3/errors"
	errorsgrpc "github.com/primandproper/platform-go/v3/errors/grpc"
	"github.com/primandproper/platform-go/v3/observability"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) AddCommentToIssueReport(ctx context.Context, request *issuereportssvc.AddCommentToIssueReportRequest) (*issuereportssvc.AddCommentToIssueReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		issuereportkeys.IssueReportIDKey: request.IssueReportId,
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if _, err = s.issueReportsManager.GetIssueReport(ctx, request.IssueReportId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "validating issue report exists")
	}

	input := commentsconverters.ConvertProtoCommentCreationRequestInputToDomain(
		request.Input,
		issuereports.CommentTargetTypeIssueReports,
		request.IssueReportId,
		sessionContextData.GetUserID(),
	)
	if input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment, err := s.commentsManager.CreateComment(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	return &issuereportssvc.AddCommentToIssueReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: commentsconverters.ConvertCommentToGRPCComment(comment),
	}, nil
}
