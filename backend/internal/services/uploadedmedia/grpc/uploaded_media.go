package grpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	uploadedmediakeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/keys"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	errorsgrpc "github.com/primandproper/platform-go/v6/errors/grpc"
	"github.com/primandproper/platform-go/v6/identifiers"
	"github.com/primandproper/platform-go/v6/uploads"

	"google.golang.org/grpc/codes"
)

const (
	maxUploadSize = 100 * 1024 * 1024 // 100 MB
)

func (s *serviceImpl) Upload(stream uploadedmediasvc.UploadedMediaService_UploadServer) error {
	ctx := stream.Context()
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	// Verify authentication
	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.Requester.UserID)

	// Receive first message which should contain metadata
	firstReq, err := stream.Recv()
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive metadata")
	}

	metadata := firstReq.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger,
			span,
			codes.InvalidArgument,
			"first message must contain metadata",
		)
	}

	logger = logger.WithValue("bucket", metadata.Bucket).
		WithValue("object_name", metadata.ObjectName).
		WithValue("content_type", metadata.ContentType)

	// Validate metadata
	if metadata.ObjectName == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("object_name is required"),
			logger,
			span,
			codes.InvalidArgument,
			"object_name is required",
		)
	}

	if metadata.ContentType == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("content_type is required"),
			logger,
			span,
			codes.InvalidArgument,
			"content_type is required",
		)
	}

	// Determine MIME type from content type
	mimeType := metadata.ContentType
	if !uploadedmedia.IsValidMimeType(mimeType) {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("unsupported content type: %s", mimeType),
			logger,
			span,
			codes.InvalidArgument,
			"unsupported content type",
		)
	}

	// Accumulate file chunks
	var fileData bytes.Buffer
	totalSize := int64(0)

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			// All chunks received
			break
		}
		if recvErr != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(recvErr, logger, span, codes.Internal, "failed to receive chunk")
		}

		chunk := req.GetChunk()
		if chunk == nil {
			continue
		}

		chunkSize := int64(len(chunk))
		if totalSize+chunkSize > maxUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxUploadSize),
				logger,
				span,
				codes.InvalidArgument,
				"file too large",
			)
		}

		if _, err = fileData.Write(chunk); err != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to write chunk")
		}

		totalSize += chunkSize
	}

	if totalSize == 0 {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("no file data received"),
			logger,
			span,
			codes.InvalidArgument,
			"no file data received",
		)
	}

	logger = logger.WithValue("size_bytes", totalSize)

	// Generate unique ID for the file
	fileID := identifiers.New()

	// Construct storage path: userID/fileID/objectName
	storagePath := filepath.Join(
		sessionContextData.Requester.UserID,
		fileID,
		metadata.ObjectName,
	)

	// Save file using upload manager
	if err = uploads.SaveFile(ctx, s.uploadManager, storagePath, fileData.Bytes()); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to save file")
	}

	// Create database record
	uploadedMediaInput := &uploadedmedia.UploadedMediaDatabaseCreationInput{
		ID:            fileID,
		StoragePath:   storagePath,
		MimeType:      mimeType,
		CreatedByUser: sessionContextData.Requester.UserID,
	}

	if err = uploadedMediaInput.ValidateWithContext(ctx); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate uploaded media")
	}

	created, err := s.uploadedMediaManager.CreateUploadedMedia(ctx, uploadedMediaInput)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create uploaded media record")
	}

	logger = logger.WithValue(uploadedmediakeys.UploadedMediaIDKey, created.ID)

	// Send response
	response := &uploadedmediasvc.UploadResponse{
		ObjectUrl: storagePath,
		SizeBytes: totalSize,
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("file uploaded successfully")

	return nil
}

func (s *serviceImpl) CreateUploadedMedia(ctx context.Context, request *uploadedmediasvc.CreateUploadedMediaRequest) (*uploadedmediasvc.CreateUploadedMediaResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.Requester.UserID)

	input := converters.ConvertGRPCUploadedMediaCreationRequestInputToUploadedMediaDatabaseCreationInput(request.Input, sessionContextData.Requester.UserID)
	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate uploaded media creation request")
	}

	created, err := s.uploadedMediaManager.CreateUploadedMedia(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create uploaded media")
	}

	x := &uploadedmediasvc.CreateUploadedMediaResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Created: converters.ConvertUploadedMediaToGRPCUploadedMedia(created),
	}

	return x, nil
}

func (s *serviceImpl) GetUploadedMedia(ctx context.Context, request *uploadedmediasvc.GetUploadedMediaRequest) (*uploadedmediasvc.GetUploadedMediaResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(uploadedmediakeys.UploadedMediaIDKey, request.UploadedMediaId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.Requester.UserID)

	uploadedMedia, err := s.uploadedMediaManager.GetUploadedMedia(ctx, request.UploadedMediaId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch uploaded media")
	}

	// Verify the uploaded media belongs to the user
	if uploadedMedia.CreatedByUser != sessionContextData.Requester.UserID {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "uploaded media does not belong to user")
	}

	x := &uploadedmediasvc.GetUploadedMediaResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Result: converters.ConvertUploadedMediaToGRPCUploadedMedia(uploadedMedia),
	}

	return x, nil
}

func (s *serviceImpl) GetUploadedMediaWithIDs(ctx context.Context, request *uploadedmediasvc.GetUploadedMediaWithIDsRequest) (*uploadedmediasvc.GetUploadedMediaWithIDsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.Requester.UserID)

	if len(request.Ids) == 0 {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("no IDs provided"), logger, span, codes.InvalidArgument, "no IDs provided")
	}

	uploadedMediaList, err := s.uploadedMediaManager.GetUploadedMediaWithIDs(ctx, request.Ids)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch uploaded media")
	}

	x := &uploadedmediasvc.GetUploadedMediaWithIDsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
	}

	for _, uploadedMedia := range uploadedMediaList {
		// Only return media that belongs to the user
		if uploadedMedia.CreatedByUser == sessionContextData.Requester.UserID {
			x.Results = append(x.Results, converters.ConvertUploadedMediaToGRPCUploadedMedia(uploadedMedia))
		}
	}

	return x, nil
}

func (s *serviceImpl) GetUploadedMediaForUser(ctx context.Context, request *uploadedmediasvc.GetUploadedMediaForUserRequest) (*uploadedmediasvc.GetUploadedMediaForUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, request.UserId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	// Verify the user is requesting their own media
	if request.UserId != sessionContextData.Requester.UserID {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "cannot access other user's media")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	uploadedMediaList, err := s.uploadedMediaManager.GetUploadedMediaForUser(ctx, request.UserId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch uploaded media for user")
	}

	x := &uploadedmediasvc.GetUploadedMediaForUserResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(uploadedMediaList.Pagination, filter),
	}

	for _, uploadedMedia := range uploadedMediaList.Data {
		x.Results = append(x.Results, converters.ConvertUploadedMediaToGRPCUploadedMedia(uploadedMedia))
	}

	return x, nil
}

func (s *serviceImpl) UpdateUploadedMedia(ctx context.Context, request *uploadedmediasvc.UpdateUploadedMediaRequest) (*uploadedmediasvc.UpdateUploadedMediaResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(uploadedmediakeys.UploadedMediaIDKey, request.UploadedMediaId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.Requester.UserID)

	// Fetch the existing uploaded media
	uploadedMedia, err := s.uploadedMediaManager.GetUploadedMedia(ctx, request.UploadedMediaId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch uploaded media")
	}

	// Verify the uploaded media belongs to the user
	if uploadedMedia.CreatedByUser != sessionContextData.Requester.UserID {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "uploaded media does not belong to user")
	}

	// Apply updates
	updateInput := converters.ConvertGRPCUploadedMediaUpdateRequestInputToUploadedMediaUpdateRequestInput(request.Input)
	if err = updateInput.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate uploaded media update request")
	}

	uploadedMedia.Update(updateInput)

	if err = s.uploadedMediaManager.UpdateUploadedMedia(ctx, uploadedMedia); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update uploaded media")
	}

	x := &uploadedmediasvc.UpdateUploadedMediaResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Updated: converters.ConvertUploadedMediaToGRPCUploadedMedia(uploadedMedia),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveUploadedMedia(ctx context.Context, request *uploadedmediasvc.ArchiveUploadedMediaRequest) (*uploadedmediasvc.ArchiveUploadedMediaResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(uploadedmediakeys.UploadedMediaIDKey, request.UploadedMediaId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(identitykeys.UserIDKey, sessionContextData.Requester.UserID)

	// Fetch the existing uploaded media to verify ownership
	uploadedMedia, err := s.uploadedMediaManager.GetUploadedMedia(ctx, request.UploadedMediaId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch uploaded media")
	}

	// Verify the uploaded media belongs to the user
	if uploadedMedia.CreatedByUser != sessionContextData.Requester.UserID {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("permission denied"), logger, span, codes.PermissionDenied, "uploaded media does not belong to user")
	}

	if err = s.uploadedMediaManager.ArchiveUploadedMedia(ctx, request.UploadedMediaId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive uploaded media")
	}

	x := &uploadedmediasvc.ArchiveUploadedMediaResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
	}

	return x, nil
}
