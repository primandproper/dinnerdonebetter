package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	comments "github.com/primandproper/platform-go/v13/comments"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/filtering"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateComment(ctx context.Context, request *commentssvc.CreateCommentRequest) (*commentssvc.CreateCommentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), s.logger, span, codes.InvalidArgument, "input is required")
	}

	target := converters.ConvertProtoCommentTargetToDomain(request.GetInput().GetTarget())

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentTargetTypeKey: target.Type.String(),
		commentskeys.CommentTargetIDKey:   target.ID,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	// The shape check before the store's catalog check, so a request naming half a
	// target is answered as the malformed request it is rather than as an unknown
	// target type.
	if err = target.Validate(); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid comment target")
	}

	return s.createComment(ctx, span, request.GetInput(), target, sessionContextData.GetUserID())
}

// createComment is the half of CreateComment the AddCommentTo* methods share:
// everything after the caller has established which target this comment is about
// and that they may comment on it.
func (s *serviceImpl) createComment(ctx context.Context, span tracing.Span, input *commentssvc.CommentCreationRequestInput, target comments.Target, author string) (*commentssvc.CreateCommentResponse, error) {
	logger := s.logger.WithSpan(span)

	comment := converters.ConvertProtoCommentCreationRequestInputToDomain(input, target, author)
	if comment == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment.Scope = ddbcomments.Scope()

	if err := s.comments.CreateComment(ctx, comment); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, comment.ID)

	return &commentssvc.CreateCommentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: converters.ConvertCommentToGRPCComment(comment),
	}, nil
}

// GetRootComments pages the top level of one target's discussion.
//
// It does not check that the target is live, and that is a known gap rather than
// an oversight: see #1362. The store cannot check — the target's row is in a
// table it has never seen — and the catalog it does hold gates writes rather than
// reads, deliberately, so that an operator withdrawing a target type can still
// reach the rows they stranded.
func (s *serviceImpl) GetRootComments(ctx context.Context, request *commentssvc.GetRootCommentsRequest) (*commentssvc.GetRootCommentsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	target := converters.ConvertProtoCommentTargetToDomain(request.GetTarget())

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentTargetTypeKey: target.Type.String(),
		commentskeys.CommentTargetIDKey:   target.ID,
	}, span, s.logger)

	if _, err := sessions.RequireFromContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err := target.Validate(); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid comment target")
	}

	filter, err := filteringgrpc.FromProto(request.GetFilter())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	result, err := s.comments.ListRootComments(ctx, ddbcomments.Scope(), target, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching root comments")
	}

	return &commentssvc.GetRootCommentsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Data:       convertPage(result),
		Pagination: filteringgrpc.PaginationToProto(result.Pagination),
	}, nil
}

// GetCommentReplies pages one root comment's replies.
//
// A parent that is no longer there is an empty page rather than an error: a reply
// outlives the comment it replies to — archived, or erased with its author — and
// is still a reply. See the platform package's documentation.
func (s *serviceImpl) GetCommentReplies(ctx context.Context, request *commentssvc.GetCommentRepliesRequest) (*commentssvc.GetCommentRepliesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	target := converters.ConvertProtoCommentTargetToDomain(request.GetTarget())

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentTargetTypeKey: target.Type.String(),
		commentskeys.CommentTargetIDKey:   target.ID,
		commentskeys.CommentIDKey:         request.GetParentId(),
	}, span, s.logger)

	if _, err := sessions.RequireFromContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err := target.Validate(); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid comment target")
	}

	filter, err := filteringgrpc.FromProto(request.GetFilter())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	result, err := s.comments.ListReplies(ctx, ddbcomments.Scope(), target, request.GetParentId(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching comment replies")
	}

	return &commentssvc.GetCommentRepliesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Data:       convertPage(result),
		Pagination: filteringgrpc.PaginationToProto(result.Pagination),
	}, nil
}

func (s *serviceImpl) UpdateComment(ctx context.Context, request *commentssvc.UpdateCommentRequest) (*commentssvc.UpdateCommentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		commentskeys.CommentIDKey: request.GetCommentId(),
	}, span, s.logger)

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment, err := s.ownedComment(ctx, span, request.GetCommentId())
	if err != nil {
		return nil, err
	}

	// Only the body. The store writes only the body too, but naming it here keeps
	// the read above from being the thing that decides what an edit may touch.
	comment.Body = request.GetInput().GetBody()

	if err = s.comments.UpdateComment(ctx, comment); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating comment")
	}

	updated, err := s.comments.GetComment(ctx, ddbcomments.Scope(), request.GetCommentId())
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
		commentskeys.CommentIDKey: request.GetCommentId(),
	}, span, s.logger)

	if _, err := s.ownedComment(ctx, span, request.GetCommentId()); err != nil {
		return nil, err
	}

	if err := s.comments.ArchiveComment(ctx, ddbcomments.Scope(), request.GetCommentId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving comment")
	}

	return &commentssvc.ArchiveCommentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}, nil
}

// ownedComment reads the comment the caller named and refuses it if somebody else
// wrote it.
//
// The check is here rather than in the store because the store does not know who
// is asking: its writes are keyed on the scope, and this deployment files every
// comment in one. Editing and archiving are both the author's acts, so both go
// through this.
func (s *serviceImpl) ownedComment(ctx context.Context, span tracing.Span, commentID string) (*comments.Comment, error) {
	logger := s.logger.WithSpan(span).WithValue(commentskeys.CommentIDKey, commentID)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	comment, err := s.comments.GetComment(ctx, ddbcomments.Scope(), commentID)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching comment")
	}

	if comment.Author != sessionContextData.GetUserID() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("comment does not belong to user"), logger, span, codes.PermissionDenied, "comment does not belong to user")
	}

	return comment, nil
}

// convertPage converts a page of stored comments to proto.
func convertPage(result *filtering.QueryFilteredResult[comments.Comment]) []*commentssvc.Comment {
	converted := make([]*commentssvc.Comment, 0, len(result.Data))
	for _, c := range result.Data {
		converted = append(converted, converters.ConvertCommentToGRPCComment(c))
	}

	return converted
}
