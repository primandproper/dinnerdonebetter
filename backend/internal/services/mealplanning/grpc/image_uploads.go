package grpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const maxImageUploadSize = 5 * 1024 * 1024 // 5 MB

// What a piece of media hangs off in the registry's belongs-to pair. The
// vocabulary is this application's — the registry neither knows nor validates
// it — so the words are declared once rather than spelled at each write.
//
// The bridge tables are still what order media within a thing and record who
// attached it. The subject says an object is one of this recipe's, which is what
// makes an orphan sweep and a "what is attached to this" read possible without
// consulting five tables.
const (
	mealSubjectType             = "meal"
	recipeSubjectType           = "recipe"
	recipeStepSubjectType       = "recipe_step"
	validPreparationSubjectType = "valid_preparation"
	validIngredientSubjectType  = "valid_ingredient"
)

// storeAndRegister writes the bytes and registers what was written.
//
// The id is the caller's because the key is built from it: the bytes have to
// know where they are going before anything writes them. The size is not the
// caller's — it is counted as the bytes go past, which is the only number that
// is about what is actually in the bucket.
//
// The order is deliberate and it is the one that fails safe. A failure between
// the two leaves an object with no row, which is invisible to every read; the
// other order leaves a row promising bytes that are not there, which every read
// reports as media the caller may have and every fetch then fails to deliver.
func (s *serviceImpl) storeAndRegister(
	ctx context.Context,
	objectID, key, contentType, ownerID string,
	subject registry.Subject,
	body *bytes.Buffer,
) (*registry.Object, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	object := &registry.Object{
		ID:          objectID,
		Scope:       uploadedmedia.Scope(),
		Key:         key,
		ContentType: contentType,
		OwnerID:     ownerID,
		BelongsTo:   subject,
	}

	if err := object.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	if err := registry.StoreAndRecord(ctx, s.uploadManager, s.registry, object, body); err != nil {
		return nil, err
	}

	return object, nil
}

func (s *serviceImpl) UploadMealImage(stream grpc.ClientStreamingServer[mealplanningsvc.UploadMealMediaRequest, mealplanningsvc.UploadMealImageResponse]) error {
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
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive first message")
	}

	mealID := firstReq.GetMealId()
	if mealID == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("meal_id is required"),
			logger, span, codes.InvalidArgument, "meal_id is required",
		)
	}
	logger = logger.WithValue(mealplanningkeys.MealIDKey, mealID)

	// Verify user owns the meal
	meal, err := s.mealPlanningManager.ReadMeal(ctx, mealID)
	if err != nil || meal == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("meal not found or access denied: %w", err),
			logger, span, codes.PermissionDenied, "meal not found or access denied",
		)
	}
	if meal.CreatedByUser != userID {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("permission denied"),
			logger, span, codes.PermissionDenied, "permission denied",
		)
	}

	upload := firstReq.GetUpload()
	if upload == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain upload"),
			logger, span, codes.InvalidArgument, "first message must contain upload",
		)
	}

	metadata := upload.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger, span, codes.InvalidArgument, "first message must contain metadata",
		)
	}

	if metadata.ObjectName == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("object_name is required"),
			logger, span, codes.InvalidArgument, "object_name is required",
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
	if chunk := upload.GetChunk(); len(chunk) > 0 {
		if _, err = fileData.Write(chunk); err != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to write chunk")
		}
	}

	totalSize := int64(fileData.Len())

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(recvErr, logger, span, codes.Internal, "failed to receive chunk")
		}

		u := req.GetUpload()
		if u == nil {
			continue
		}

		chunk := u.GetChunk()
		if chunk == nil {
			continue
		}

		chunkSize := int64(len(chunk))
		if totalSize+chunkSize > maxImageUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxImageUploadSize),
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

	fileID := identifiers.New()

	created, err := s.storeAndRegister(
		ctx,
		fileID,
		filepath.Join("meals", mealID, fileID, metadata.ObjectName),
		mimeType,
		userID,
		registry.Subject{Type: mealSubjectType, ID: mealID},
		&fileData,
	)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to store uploaded media")
	}

	if err = s.mealPlanningManager.AddMealImage(ctx, mealID, created.ID, userID); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add meal image")
	}

	uploadedMediaID := created.ID
	response := &mealplanningsvc.UploadMealImageResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UploadedMediaId: &uploadedMediaID,
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("meal image uploaded successfully")
	return nil
}

func (s *serviceImpl) UploadRecipeImage(stream grpc.ClientStreamingServer[mealplanningsvc.UploadRecipeMediaRequest, mealplanningsvc.UploadRecipeImageResponse]) error {
	ctx := stream.Context()
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	userID := sessionContextData.GetUserID()

	firstReq, err := stream.Recv()
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive first message")
	}

	recipeID := firstReq.GetRecipeId()
	if recipeID == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("recipe_id is required"),
			logger, span, codes.InvalidArgument, "recipe_id is required",
		)
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)

	// Verify user owns the recipe
	recipe, err := s.mealPlanningManager.ReadRecipe(ctx, recipeID)
	if err != nil || recipe == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("recipe not found or access denied: %w", err),
			logger, span, codes.PermissionDenied, "recipe not found or access denied",
		)
	}
	if recipe.CreatedByUser != userID {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("permission denied"),
			logger, span, codes.PermissionDenied, "permission denied",
		)
	}

	upload := firstReq.GetUpload()
	if upload == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain upload"),
			logger, span, codes.InvalidArgument, "first message must contain upload",
		)
	}

	metadata := upload.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger, span, codes.InvalidArgument, "first message must contain metadata",
		)
	}

	if metadata.ObjectName == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("object_name is required"),
			logger, span, codes.InvalidArgument, "object_name is required",
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
	if chunk := upload.GetChunk(); len(chunk) > 0 {
		if _, err = fileData.Write(chunk); err != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to write chunk")
		}
	}

	totalSize := int64(fileData.Len())

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(recvErr, logger, span, codes.Internal, "failed to receive chunk")
		}

		u := req.GetUpload()
		if u == nil {
			continue
		}

		chunk := u.GetChunk()
		if chunk == nil {
			continue
		}

		chunkSize := int64(len(chunk))
		if totalSize+chunkSize > maxImageUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxImageUploadSize),
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

	fileID := identifiers.New()

	created, err := s.storeAndRegister(
		ctx,
		fileID,
		filepath.Join("recipes", recipeID, fileID, metadata.ObjectName),
		mimeType,
		userID,
		registry.Subject{Type: recipeSubjectType, ID: recipeID},
		&fileData,
	)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to store uploaded media")
	}

	if err = s.mealPlanningManager.AddRecipeImage(ctx, recipeID, created.ID, userID); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add recipe image")
	}

	uploadedMediaID := created.ID
	response := &mealplanningsvc.UploadRecipeImageResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UploadedMediaId: &uploadedMediaID,
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("recipe image uploaded successfully")
	return nil
}

func (s *serviceImpl) UploadPreparationMedia(stream grpc.ClientStreamingServer[mealplanningsvc.UploadPreparationMediaRequest, mealplanningsvc.UploadPreparationMediaResponse]) error {
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
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive first message")
	}

	validPreparationID := firstReq.GetValidPreparationId()
	if validPreparationID == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("valid_preparation_id is required"),
			logger, span, codes.InvalidArgument, "valid_preparation_id is required",
		)
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	// Verify preparation exists
	_, err = s.mealPlanningManager.ReadValidPreparation(ctx, validPreparationID)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("preparation not found: %w", err),
			logger, span, codes.NotFound, "preparation not found",
		)
	}

	upload := firstReq.GetUpload()
	if upload == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain upload"),
			logger, span, codes.InvalidArgument, "first message must contain upload",
		)
	}

	metadata := upload.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger, span, codes.InvalidArgument, "first message must contain metadata",
		)
	}

	if metadata.ObjectName == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("object_name is required"),
			logger, span, codes.InvalidArgument, "object_name is required",
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
	if chunk := upload.GetChunk(); len(chunk) > 0 {
		if _, err = fileData.Write(chunk); err != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to write chunk")
		}
	}

	totalSize := int64(fileData.Len())

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(recvErr, logger, span, codes.Internal, "failed to receive chunk")
		}

		u := req.GetUpload()
		if u == nil {
			continue
		}

		chunk := u.GetChunk()
		if chunk == nil {
			continue
		}

		chunkSize := int64(len(chunk))
		if totalSize+chunkSize > maxImageUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxImageUploadSize),
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

	fileID := identifiers.New()

	created, err := s.storeAndRegister(
		ctx,
		fileID,
		filepath.Join("preparations", validPreparationID, fileID, metadata.ObjectName),
		mimeType,
		userID,
		registry.Subject{Type: validPreparationSubjectType, ID: validPreparationID},
		&fileData,
	)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to store uploaded media")
	}

	var forIngredientID *string
	if v := firstReq.GetForIngredientId(); v != "" {
		forIngredientID = &v
	}

	if err = s.mealPlanningManager.AddPreparationMedia(ctx, validPreparationID, forIngredientID, created.ID, 0); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add preparation media")
	}

	uploadedMediaID := created.ID
	response := &mealplanningsvc.UploadPreparationMediaResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UploadedMediaId: &uploadedMediaID,
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("preparation media uploaded successfully")
	return nil
}

func (s *serviceImpl) UploadIngredientMedia(stream grpc.ClientStreamingServer[mealplanningsvc.UploadIngredientMediaRequest, mealplanningsvc.UploadIngredientMediaResponse]) error {
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
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive first message")
	}

	validIngredientID := firstReq.GetValidIngredientId()
	if validIngredientID == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("valid_ingredient_id is required"),
			logger, span, codes.InvalidArgument, "valid_ingredient_id is required",
		)
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	// Verify ingredient exists
	_, err = s.mealPlanningManager.ReadValidIngredient(ctx, validIngredientID)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("ingredient not found: %w", err),
			logger, span, codes.NotFound, "ingredient not found",
		)
	}

	upload := firstReq.GetUpload()
	if upload == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain upload"),
			logger, span, codes.InvalidArgument, "first message must contain upload",
		)
	}

	metadata := upload.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger, span, codes.InvalidArgument, "first message must contain metadata",
		)
	}

	if metadata.ObjectName == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("object_name is required"),
			logger, span, codes.InvalidArgument, "object_name is required",
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
	if chunk := upload.GetChunk(); len(chunk) > 0 {
		if _, err = fileData.Write(chunk); err != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to write chunk")
		}
	}

	totalSize := int64(fileData.Len())

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(recvErr, logger, span, codes.Internal, "failed to receive chunk")
		}

		u := req.GetUpload()
		if u == nil {
			continue
		}

		chunk := u.GetChunk()
		if chunk == nil {
			continue
		}

		chunkSize := int64(len(chunk))
		if totalSize+chunkSize > maxImageUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxImageUploadSize),
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

	fileID := identifiers.New()

	created, err := s.storeAndRegister(
		ctx,
		fileID,
		filepath.Join("ingredients", validIngredientID, fileID, metadata.ObjectName),
		mimeType,
		userID,
		registry.Subject{Type: validIngredientSubjectType, ID: validIngredientID},
		&fileData,
	)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to store uploaded media")
	}

	if err = s.mealPlanningManager.AddIngredientMedia(ctx, validIngredientID, created.ID, 0); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add ingredient media")
	}

	uploadedMediaID := created.ID
	response := &mealplanningsvc.UploadIngredientMediaResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UploadedMediaId: &uploadedMediaID,
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("ingredient media uploaded successfully")
	return nil
}

func (s *serviceImpl) UploadRecipeStepImage(stream grpc.ClientStreamingServer[mealplanningsvc.UploadRecipeStepImageRequest, mealplanningsvc.UploadRecipeStepImageResponse]) error {
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
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to receive first message")
	}

	recipeID := firstReq.GetRecipeId()
	if recipeID == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("recipe_id is required"),
			logger, span, codes.InvalidArgument, "recipe_id is required",
		)
	}
	recipeStepID := firstReq.GetRecipeStepId()
	if recipeStepID == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("recipe_step_id is required"),
			logger, span, codes.InvalidArgument, "recipe_step_id is required",
		)
	}
	logger = logger.WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})

	// Verify user owns the recipe
	recipe, err := s.mealPlanningManager.ReadRecipe(ctx, recipeID)
	if err != nil || recipe == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("recipe not found or access denied: %w", err),
			logger, span, codes.PermissionDenied, "recipe not found or access denied",
		)
	}
	if recipe.CreatedByUser != userID {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("permission denied"),
			logger, span, codes.PermissionDenied, "permission denied",
		)
	}

	// Verify step exists and belongs to recipe
	_, err = s.mealPlanningManager.ReadRecipeStep(ctx, recipeID, recipeStepID)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("recipe step not found: %w", err),
			logger, span, codes.NotFound, "recipe step not found",
		)
	}

	upload := firstReq.GetUpload()
	if upload == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain upload"),
			logger, span, codes.InvalidArgument, "first message must contain upload",
		)
	}

	metadata := upload.GetMetadata()
	if metadata == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("first message must contain metadata"),
			logger, span, codes.InvalidArgument, "first message must contain metadata",
		)
	}

	if metadata.ObjectName == "" {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("object_name is required"),
			logger, span, codes.InvalidArgument, "object_name is required",
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
	if chunk := upload.GetChunk(); len(chunk) > 0 {
		if _, err = fileData.Write(chunk); err != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to write chunk")
		}
	}

	totalSize := int64(fileData.Len())

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return errorsgrpc.PrepareAndLogGRPCStatus(recvErr, logger, span, codes.Internal, "failed to receive chunk")
		}

		u := req.GetUpload()
		if u == nil {
			continue
		}

		chunk := u.GetChunk()
		if chunk == nil {
			continue
		}

		chunkSize := int64(len(chunk))
		if totalSize+chunkSize > maxImageUploadSize {
			return errorsgrpc.PrepareAndLogGRPCStatus(
				fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxImageUploadSize),
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

	fileID := identifiers.New()

	created, err := s.storeAndRegister(
		ctx,
		fileID,
		filepath.Join("recipes", recipeID, "steps", recipeStepID, fileID, metadata.ObjectName),
		mimeType,
		userID,
		registry.Subject{Type: recipeStepSubjectType, ID: recipeStepID},
		&fileData,
	)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to store uploaded media")
	}

	if err = s.mealPlanningManager.AddRecipeStepImage(ctx, recipeStepID, created.ID, userID); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add recipe step image")
	}

	uploadedMediaID := created.ID
	response := &mealplanningsvc.UploadRecipeStepImageResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UploadedMediaId: &uploadedMediaID,
	}

	if err = stream.SendAndClose(response); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to send response")
	}

	logger.Info("recipe step image uploaded successfully")
	return nil
}
