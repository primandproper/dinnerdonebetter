package grpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc/converters"
	uploadedmediaconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	platformkeys "github.com/primandproper/platform-go/v13/observability/keys"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const maxAvatarUploadSize = 5 * 1024 * 1024 // 5 MB for avatars

// avatarSubjectType is what an avatar hangs off in the registry's belongs-to
// pair. The vocabulary is this application's — the registry neither knows nor
// validates it — so the word is declared once here rather than spelled at each
// write.
const avatarSubjectType = "user"

// avatarObjectName returns a unique object name (UUID + extension) from the MIME type, matching iOS behavior.
func avatarObjectName(mimeType string) string {
	ext := ".jpg"
	switch strings.ToLower(mimeType) {
	case uploadedmedia.MimeTypeImagePNG:
		ext = ".png"
	case uploadedmedia.MimeTypeImageJPEG:
		ext = ".jpg"
	case uploadedmedia.MimeTypeImageGIF:
		ext = ".gif"
	}
	return uuid.New().String() + ext
}

func (s *serviceImpl) CreateUser(ctx context.Context, request *identitysvc.CreateUserRequest) (*identitysvc.CreateUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertGRPCUserRegistrationInputToUserRegistrationInput(request.Input)

	created, err := s.identityDataManager.CreateUser(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to create user")
	}

	x := &identitysvc.CreateUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Created:         converters.ConvertUserCreationResponseToGRPCUserCreationResponse(created),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveUser(ctx context.Context, request *identitysvc.ArchiveUserRequest) (*identitysvc.ArchiveUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.identityDataManager.ArchiveUser(ctx, request.UserId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to archive user")
	}

	x := &identitysvc.ArchiveUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) GetUser(ctx context.Context, request *identitysvc.GetUserRequest) (*identitysvc.GetUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.UserIDKey: request.UserId,
	}, span, s.logger)

	user, err := s.identityDataManager.GetUser(ctx, request.UserId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch users from database")
	}

	x := &identitysvc.GetUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Result:          converters.ConvertUserToGRPCUser(user),
	}

	return x, nil
}

func (s *serviceImpl) GetUsers(ctx context.Context, request *identitysvc.GetUsersRequest) (*identitysvc.GetUsersResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	users, err := s.identityDataManager.GetUsers(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch users from database")
	}

	x := &identitysvc.GetUsersResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      filteringgrpc.PaginationToProto(users.Pagination),
	}

	for _, user := range users.Data {
		x.Results = append(x.Results, converters.ConvertUserToGRPCUser(user))
	}

	return x, nil
}

func (s *serviceImpl) GetUsersForAccount(ctx context.Context, request *identitysvc.GetUsersForAccountRequest) (*identitysvc.GetUsersForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	users, err := s.identityDataManager.GetUsersForAccount(ctx, request.AccountId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch users from database")
	}

	x := &identitysvc.GetUsersForAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      filteringgrpc.PaginationToProto(users.Pagination),
	}

	for _, user := range users.Data {
		x.Results = append(x.Results, converters.ConvertUserToGRPCUser(user))
	}

	return x, nil
}

func (s *serviceImpl) SearchForUsers(ctx context.Context, request *identitysvc.SearchForUsersRequest) (*identitysvc.SearchForUsersResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		platformkeys.SearchQueryKey: request.Query,
	}, span, s.logger)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	users, err := s.identityDataManager.SearchForUsers(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to search for users")
	}

	x := &identitysvc.SearchForUsersResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      filteringgrpc.PaginationToProto(users.Pagination),
	}

	for _, user := range users.Data {
		x.Results = append(x.Results, converters.ConvertUserToGRPCUser(user))
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserDetails(ctx context.Context, request *identitysvc.UpdateUserDetailsRequest) (*identitysvc.UpdateUserDetailsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCUserDetailsUpdateRequestInputToUserDetailsUpdateRequestInput(request.Input)

	if err = s.identityDataManager.UpdateUserDetails(ctx, sessionContextData.GetUserID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user details")
	}

	x := &identitysvc.UpdateUserDetailsResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserEmailAddress(ctx context.Context, request *identitysvc.UpdateUserEmailAddressRequest) (*identitysvc.UpdateUserEmailAddressResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.identityDataManager.UpdateUserEmailAddress(ctx, sessionContextData.GetUserID(), request.NewEmailAddress); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user email address")
	}

	x := &identitysvc.UpdateUserEmailAddressResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserUsername(ctx context.Context, request *identitysvc.UpdateUserUsernameRequest) (*identitysvc.UpdateUserUsernameResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.identityDataManager.UpdateUserUsername(ctx, sessionContextData.GetUserID(), request.NewUsername); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user username")
	}

	x := &identitysvc.UpdateUserUsernameResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UploadUserAvatar(stream grpc.ClientStreamingServer[uploadedmediasvc.UploadRequest, identitysvc.UploadUserAvatarResponse]) error {
	ctx := stream.Context()
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	userID := sessionContextData.GetUserID()
	logger = logger.WithValue(identitykeys.UserIDKey, userID)

	firstReq, err := stream.Recv()
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive metadata")
	}

	metadata := firstReq.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger, span, codes.InvalidArgument, "first message must contain metadata",
		)
	}

	if metadata.ContentType == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("content_type is required"),
			logger, span, codes.InvalidArgument, "content_type is required",
		)
	}

	mimeType := metadata.ContentType
	if !uploadedmedia.IsValidMimeType(mimeType) {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("unsupported content type: %s", mimeType),
			logger, span, codes.InvalidArgument, "unsupported content type",
		)
	}

	var fileData bytes.Buffer
	totalSize := int64(0)

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
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
		if totalSize+chunkSize > maxAvatarUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxAvatarUploadSize),
				logger, span, codes.InvalidArgument, "file too large",
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
			logger, span, codes.InvalidArgument, "no file data received",
		)
	}

	// The row's id is minted here rather than by the registry, because the
	// storage key is built from it: the bytes have to know where they are going
	// before anything writes them.
	fileID := identifiers.New()

	// The avatar hangs off its user, which is what the registry's belongs-to pair
	// is for. The user_avatars row is still what says which of a user's objects is
	// the current avatar — the subject says an object is one of theirs, not that
	// it is the one on show.
	object := &registry.Object{
		ID:          fileID,
		Scope:       uploadedmedia.Scope(),
		Key:         filepath.Join(userID, fileID, avatarObjectName(mimeType)),
		ContentType: mimeType,
		OwnerID:     userID,
		BelongsTo:   registry.Subject{Type: avatarSubjectType, ID: userID},
	}

	if err = object.ValidateWithContext(ctx); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate uploaded media")
	}

	// Bytes first, then the row, with the size counted as it went past. A failure
	// between the two leaves an object with no row, which is invisible to every
	// read; the other order leaves a row promising bytes that are not there.
	if err = registry.StoreAndRecord(ctx, s.uploadManager, s.registry, object, &fileData); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to store avatar")
	}

	if err = s.identityDataManager.SetUserAvatar(ctx, userID, object.ID); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to set user avatar")
	}

	response := &identitysvc.UploadUserAvatarResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Created:         uploadedmediaconverters.ConvertUploadedMediaToGRPCUploadedMedia(object),
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("avatar uploaded successfully")
	return nil
}
