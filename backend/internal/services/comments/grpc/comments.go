package grpc

import (
	"context"

	commentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	commentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	converters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	errorsgrpc "github.com/primandproper/platform-go/v7/errors/grpc"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateComment(ctx context.Context, request *commentssvc.CreateCommentRequest) (*commentssvc.CreateCommentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if request.Input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), s.logger, span, codes.InvalidArgument, "input is required")
	}

	logger := observability.ObserveValues(map[string]any{
		"target_type":   request.Input.GetTargetType(),
		"referenced_id": request.Input.GetReferencedId(),
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertProtoCommentCreationRequestInputToDomain(
		request.Input,
		"",
		"",
		sessionContextData.GetUserID(),
	)
	if input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}
	if input.TargetType == "" || input.ReferencedID == "" {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("target_type and referenced_id are required"), logger, span, codes.InvalidArgument, "target_type and referenced_id are required")
	}

	comment, err := s.commentsManager.CreateComment(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	return &commentssvc.CreateCommentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: converters.ConvertCommentToGRPCComment(comment),
	}, nil
}

func (s *serviceImpl) GetCommentsForReference(ctx context.Context, request *commentssvc.GetCommentsForReferenceRequest) (*commentssvc.GetCommentsForReferenceResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentIDKey: request.ReferencedId,
		"target_type":             request.TargetType,
	}, span, s.logger)

	if _, err := s.sessionContextDataFetcher(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	result, err := s.commentsManager.GetCommentsForReference(ctx, request.TargetType, request.ReferencedId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching comments")
	}

	x := &commentssvc.GetCommentsForReferenceResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(result.Pagination, filter),
	}
	for _, c := range result.Data {
		x.Data = append(x.Data, converters.ConvertCommentToGRPCComment(c))
	}

	return x, nil
}

func (s *serviceImpl) UpdateComment(ctx context.Context, request *commentssvc.UpdateCommentRequest) (*commentssvc.UpdateCommentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentIDKey: request.CommentId,
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	comment, err := s.commentsManager.GetComment(ctx, request.CommentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching comment")
	}

	if comment.BelongsToUser != sessionContextData.GetUserID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("comment does not belong to user"), logger, span, codes.PermissionDenied, "comment does not belong to user")
	}

	updateInput := converters.ConvertProtoCommentUpdateRequestInputToDomain(request.Input)
	if updateInput == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	if err = s.commentsManager.UpdateComment(ctx, request.CommentId, sessionContextData.GetUserID(), updateInput); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating comment")
	}

	updated, err := s.commentsManager.GetComment(ctx, request.CommentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching updated comment")
	}

	return &commentssvc.UpdateCommentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: converters.ConvertCommentToGRPCComment(updated),
	}, nil
}

func (s *serviceImpl) ArchiveComment(ctx context.Context, request *commentssvc.ArchiveCommentRequest) (*commentssvc.ArchiveCommentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentIDKey: request.CommentId,
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	comment, err := s.commentsManager.GetComment(ctx, request.CommentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching comment")
	}

	if comment.BelongsToUser != sessionContextData.GetUserID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("comment does not belong to user"), logger, span, codes.PermissionDenied, "comment does not belong to user")
	}

	if err = s.commentsManager.ArchiveComment(ctx, request.CommentId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving comment")
	}

	return &commentssvc.ArchiveCommentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}
