package grpc

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	mealplanningdomain "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	errorsgrpc "github.com/primandproper/platform-go/v12/errors/grpc"
	"github.com/primandproper/platform-go/v12/observability"
	platformkeys "github.com/primandproper/platform-go/v12/observability/keys"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) verifyRecipeOwnership(ctx context.Context, recipeID string, logger logging.Logger, span tracing.Span) (string, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return "", errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}
	userID := sessionContextData.GetUserID()

	recipe, err := s.mealPlanningManager.ReadRecipe(ctx, recipeID)
	if err != nil || recipe == nil {
		return "", errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("recipe not found or access denied: %w", err),
			logger, span, codes.PermissionDenied, "recipe not found or access denied",
		)
	}
	if recipe.CreatedByUser != userID {
		return "", errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("permission denied"),
			logger, span, codes.PermissionDenied, "permission denied",
		)
	}

	return userID, nil
}

// verifyRecipeRatingOwnership loads a recipe rating and confirms it was authored by the requester,
// returning a gRPC status error otherwise. Recipe ratings are user-authored, so mutation is limited
// to the original author (mirroring the comments service).
func (s *serviceImpl) verifyRecipeRatingOwnership(ctx context.Context, recipeID, recipeRatingID string, logger logging.Logger, span tracing.Span) error {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	rating, err := s.mealPlanningManager.ReadRecipeRating(ctx, recipeID, recipeRatingID)
	if err != nil || rating == nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			fmt.Errorf("recipe rating not found or access denied: %w", err),
			logger, span, codes.NotFound, "recipe rating not found or access denied",
		)
	}

	if rating.CreatedByUser != sessionContextData.GetUserID() {
		return errorsgrpc.PrepareAndLogGRPCStatus(
			platformerrors.New("recipe rating does not belong to user"),
			logger, span, codes.PermissionDenied, "recipe rating does not belong to user",
		)
	}

	return nil
}

func (s *serviceImpl) ArchiveRecipe(ctx context.Context, request *mealplanning.ArchiveRecipeRequest) (*mealplanning.ArchiveRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	userID, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span)
	if err != nil {
		return nil, err
	}

	if err = s.mealPlanningManager.ArchiveRecipe(ctx, request.RecipeId, userID); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe")
	}

	if err = s.commentsManager.ArchiveCommentsForReference(ctx, mealplanningdomain.CommentTargetTypeRecipes, request.RecipeId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving comments for recipe")
	}

	x := &mealplanning.ArchiveRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipePrepTask(ctx context.Context, request *mealplanning.ArchiveRecipePrepTaskRequest) (*mealplanning.ArchiveRecipePrepTaskResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:         request.RecipeId,
		mealplanningkeys.RecipePrepTaskIDKey: request.RecipePrepTaskId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipePrepTask(ctx, request.RecipeId, request.RecipePrepTaskId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe prep task")
	}

	x := &mealplanning.ArchiveRecipePrepTaskResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeRating(ctx context.Context, request *mealplanning.ArchiveRecipeRatingRequest) (*mealplanning.ArchiveRecipeRatingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:       request.RecipeId,
		mealplanningkeys.RecipeRatingIDKey: request.RecipeRatingId,
	}, span, s.logger)

	if err := s.verifyRecipeRatingOwnership(ctx, request.RecipeId, request.RecipeRatingId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeRating(ctx, request.RecipeId, request.RecipeRatingId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe rating")
	}

	x := &mealplanning.ArchiveRecipeRatingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeStep(ctx context.Context, request *mealplanning.ArchiveRecipeStepRequest) (*mealplanning.ArchiveRecipeStepResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeStep(ctx, request.RecipeId, request.RecipeStepId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe step")
	}

	x := &mealplanning.ArchiveRecipeStepResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeStepCompletionCondition(ctx context.Context, request *mealplanning.ArchiveRecipeStepCompletionConditionRequest) (*mealplanning.ArchiveRecipeStepCompletionConditionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:                        request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:                    request.RecipeStepId,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: request.RecipeStepCompletionConditionId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeStepCompletionCondition(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepCompletionConditionId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe step completion condition")
	}

	x := &mealplanning.ArchiveRecipeStepCompletionConditionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeStepIngredient(ctx context.Context, request *mealplanning.ArchiveRecipeStepIngredientRequest) (*mealplanning.ArchiveRecipeStepIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:           request.RecipeStepId,
		mealplanningkeys.RecipeStepIngredientIDKey: request.RecipeStepIngredientId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeStepIngredient(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepIngredientId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe step ingredient")
	}

	x := &mealplanning.ArchiveRecipeStepIngredientResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeStepInstrument(ctx context.Context, request *mealplanning.ArchiveRecipeStepInstrumentRequest) (*mealplanning.ArchiveRecipeStepInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:           request.RecipeStepId,
		mealplanningkeys.RecipeStepInstrumentIDKey: request.RecipeStepInstrumentId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeStepInstrument(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepInstrumentId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe step instrument")
	}

	x := &mealplanning.ArchiveRecipeStepInstrumentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeStepProduct(ctx context.Context, request *mealplanning.ArchiveRecipeStepProductRequest) (*mealplanning.ArchiveRecipeStepProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:            request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:        request.RecipeStepId,
		mealplanningkeys.RecipeStepProductIDKey: request.RecipeStepProductId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeStepProduct(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepProductId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe step product")
	}

	x := &mealplanning.ArchiveRecipeStepProductResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeStepVessel(ctx context.Context, request *mealplanning.ArchiveRecipeStepVesselRequest) (*mealplanning.ArchiveRecipeStepVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:           request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:       request.RecipeStepId,
		mealplanningkeys.RecipeStepVesselIDKey: request.RecipeStepVesselId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveRecipeStepVessel(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepVesselId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe step vessel")
	}

	x := &mealplanning.ArchiveRecipeStepVesselResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) CloneRecipe(ctx context.Context, request *mealplanning.CloneRecipeRequest) (*mealplanning.CloneRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	cloned, err := s.mealPlanningManager.CloneRecipe(ctx, request.RecipeId, sessionContextData.GetUserID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "cloning recipe")
	}

	x := &mealplanning.CloneRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Cloned: converters.ConvertRecipeToGRPCRecipe(cloned),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipe(ctx context.Context, request *mealplanning.CreateRecipeRequest) (*mealplanning.CreateRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCRecipeCreationRequestInputToRecipeCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipe(ctx, sessionContextData.GetUserID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe")
	}

	x := &mealplanning.CreateRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeToGRPCRecipe(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipePrepTask(ctx context.Context, request *mealplanning.CreateRecipePrepTaskRequest) (*mealplanning.CreateRecipePrepTaskResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipePrepTaskCreationRequestInputToRecipePrepTaskCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipePrepTask(ctx, request.RecipeId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe prep task")
	}

	x := &mealplanning.CreateRecipePrepTaskResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipePrepTaskToGRPCRecipePrepTask(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeRating(ctx context.Context, request *mealplanning.CreateRecipeRatingRequest) (*mealplanning.CreateRecipeRatingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCRecipeRatingCreationRequestInputToRecipeRatingCreationRequestInput(request.Input)
	input.CreatedByUser = sessionContextData.GetUserID()

	created, err := s.mealPlanningManager.CreateRecipeRating(ctx, request.RecipeId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe rating")
	}

	x := &mealplanning.CreateRecipeRatingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeRatingToGRPCRecipeRating(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeStep(ctx context.Context, request *mealplanning.CreateRecipeStepRequest) (*mealplanning.CreateRecipeStepResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	recipeStepInput := converters.ConvertGRPCRecipeStepCreationRequestInputToRecipeStepCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeStep(ctx, request.RecipeId, recipeStepInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe step")
	}

	x := &mealplanning.CreateRecipeStepResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeStepToGRPCRecipeStep(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeStepCompletionCondition(ctx context.Context, request *mealplanning.CreateRecipeStepCompletionConditionRequest) (*mealplanning.CreateRecipeStepCompletionConditionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	creationInput := converters.ConvertGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeStepCompletionCondition(ctx, request.RecipeId, request.RecipeStepId, creationInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe step completion condition")
	}

	x := &mealplanning.CreateRecipeStepCompletionConditionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeStepCompletionConditionToGRPCRecipeStepCompletionCondition(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeStepIngredient(ctx context.Context, request *mealplanning.CreateRecipeStepIngredientRequest) (*mealplanning.CreateRecipeStepIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	creationInput := converters.ConvertGRPCRecipeStepIngredientCreationRequestInputToRecipeStepIngredientCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeStepIngredient(ctx, request.RecipeId, request.RecipeStepId, creationInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe step ingredient")
	}

	x := &mealplanning.CreateRecipeStepIngredientResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeStepIngredientToGRPCRecipeStepIngredient(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeStepInstrument(ctx context.Context, request *mealplanning.CreateRecipeStepInstrumentRequest) (*mealplanning.CreateRecipeStepInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	creationInput := converters.ConvertGRPCRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeStepInstrument(ctx, request.RecipeId, request.RecipeStepId, creationInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe step instrument")
	}

	x := &mealplanning.CreateRecipeStepInstrumentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeStepInstrumentToGRPCRecipeStepInstrument(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeStepProduct(ctx context.Context, request *mealplanning.CreateRecipeStepProductRequest) (*mealplanning.CreateRecipeStepProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	creationInput := converters.ConvertGRPCRecipeStepProductCreationRequestInputToRecipeStepProductCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeStepProduct(ctx, request.RecipeId, request.RecipeStepId, creationInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe step product")
	}

	x := &mealplanning.CreateRecipeStepProductResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeStepProductToGRPCRecipeStepProduct(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeStepVessel(ctx context.Context, request *mealplanning.CreateRecipeStepVesselRequest) (*mealplanning.CreateRecipeStepVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	creationInput := converters.ConvertGRPCRecipeStepVesselCreationRequestInputToRecipeStepVesselCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeStepVessel(ctx, request.RecipeId, request.RecipeStepId, creationInput)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe step vessel")
	}

	x := &mealplanning.CreateRecipeStepVesselResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeStepVesselToGRPCRecipeStepVessel(created),
	}

	return x, nil
}

func (s *serviceImpl) GetMermaidDiagramForRecipe(ctx context.Context, request *mealplanning.GetMermaidDiagramForRecipeRequest) (*mealplanning.GetMermaidDiagramForRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	mermaidDiagram, err := s.mealPlanningManager.RecipeMermaid(ctx, request.RecipeId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to generate mermaid diagram")
	}

	x := &mealplanning.GetMermaidDiagramForRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Response: mermaidDiagram,
	}

	return x, nil
}

func (s *serviceImpl) GetRecipe(ctx context.Context, request *mealplanning.GetRecipeRequest) (*mealplanning.GetRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	recipe, err := s.mealPlanningManager.ReadRecipe(ctx, request.RecipeId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe")
	}

	x := &mealplanning.GetRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeToGRPCRecipe(recipe),
	}

	return x, nil
}

// TODO: very aware that this basically just emits a list of strings, and that that isn't very useful.
func (s *serviceImpl) EstimateRecipePrepTasks(ctx context.Context, request *mealplanning.EstimateRecipePrepTasksRequest) (*mealplanning.EstimateRecipePrepTasksResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	estimatedPrepSteps, err := s.mealPlanningManager.RecipeEstimatedPrepSteps(ctx, request.RecipeId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to estimate prep steps")
	}

	x := &mealplanning.EstimateRecipePrepTasksResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, step := range estimatedPrepSteps {
		x.Results = append(x.Results, converters.ConvertMealPlanTaskDatabaseCreationEstimateToGRPCMealPlanTask(step))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipePrepTask(ctx context.Context, request *mealplanning.GetRecipePrepTaskRequest) (*mealplanning.GetRecipePrepTaskResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:         request.RecipeId,
		mealplanningkeys.RecipePrepTaskIDKey: request.RecipePrepTaskId,
	}, span, s.logger)

	recipePrepTask, err := s.mealPlanningManager.ReadRecipePrepTask(ctx, request.RecipeId, request.RecipePrepTaskId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe prep task")
	}

	x := &mealplanning.GetRecipePrepTaskResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipePrepTaskToGRPCRecipePrepTask(recipePrepTask),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipePrepTasks(ctx context.Context, request *mealplanning.GetRecipePrepTasksRequest) (*mealplanning.GetRecipePrepTasksResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipePrepTasks, err := s.mealPlanningManager.ListRecipePrepTask(ctx, request.RecipeId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe prep tasks")
	}

	x := &mealplanning.GetRecipePrepTasksResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(recipePrepTasks.Pagination, filter),
	}

	for _, recipePrepTask := range recipePrepTasks.Data {
		x.Results = append(x.Results, converters.ConvertRecipePrepTaskToGRPCRecipePrepTask(recipePrepTask))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeRating(ctx context.Context, request *mealplanning.GetRecipeRatingRequest) (*mealplanning.GetRecipeRatingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:       request.RecipeId,
		mealplanningkeys.RecipeRatingIDKey: request.RecipeRatingId,
	}, span, s.logger)

	recipeRating, err := s.mealPlanningManager.ReadRecipeRating(ctx, request.RecipeId, request.RecipeRatingId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe rating")
	}

	x := &mealplanning.GetRecipeRatingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeRatingToGRPCRecipeRating(recipeRating),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeRatingsForRecipe(ctx context.Context, request *mealplanning.GetRecipeRatingsForRecipeRequest) (*mealplanning.GetRecipeRatingsForRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeRatings, err := s.mealPlanningManager.ListRecipeRatings(ctx, request.RecipeId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe ratings")
	}

	x := &mealplanning.GetRecipeRatingsForRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipePrepTask := range recipeRatings.Data {
		x.Results = append(x.Results, converters.ConvertRecipeRatingToGRPCRecipeRating(recipePrepTask))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStep(ctx context.Context, request *mealplanning.GetRecipeStepRequest) (*mealplanning.GetRecipeStepResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	recipeStep, err := s.mealPlanningManager.ReadRecipeStep(ctx, request.RecipeId, request.RecipeStepId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step")
	}

	x := &mealplanning.GetRecipeStepResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeStepToGRPCRecipeStep(recipeStep),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepCompletionCondition(ctx context.Context, request *mealplanning.GetRecipeStepCompletionConditionRequest) (*mealplanning.GetRecipeStepCompletionConditionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:                        request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:                    request.RecipeStepId,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: request.RecipeStepCompletionConditionId,
	}, span, s.logger)

	recipeStepCompletionCondition, err := s.mealPlanningManager.ReadRecipeStepCompletionCondition(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepCompletionConditionId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step completion condition")
	}

	x := &mealplanning.GetRecipeStepCompletionConditionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeStepCompletionConditionToGRPCRecipeStepCompletionCondition(recipeStepCompletionCondition),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepCompletionConditions(ctx context.Context, request *mealplanning.GetRecipeStepCompletionConditionsRequest) (*mealplanning.GetRecipeStepCompletionConditionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeStepCompletionConditions, err := s.mealPlanningManager.ListRecipeStepCompletionConditions(ctx, request.RecipeId, request.RecipeStepId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step completion conditions")
	}

	x := &mealplanning.GetRecipeStepCompletionConditionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipePrepTask := range recipeStepCompletionConditions.Data {
		x.Results = append(x.Results, converters.ConvertRecipeStepCompletionConditionToGRPCRecipeStepCompletionCondition(recipePrepTask))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepIngredient(ctx context.Context, request *mealplanning.GetRecipeStepIngredientRequest) (*mealplanning.GetRecipeStepIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:           request.RecipeStepId,
		mealplanningkeys.RecipeStepIngredientIDKey: request.RecipeStepIngredientId,
	}, span, s.logger)

	recipeStepIngredient, err := s.mealPlanningManager.ReadRecipeStepIngredient(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepIngredientId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step ingredient")
	}

	x := &mealplanning.GetRecipeStepIngredientResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeStepIngredientToGRPCRecipeStepIngredient(recipeStepIngredient),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepIngredients(ctx context.Context, request *mealplanning.GetRecipeStepIngredientsRequest) (*mealplanning.GetRecipeStepIngredientsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeStepIngredients, err := s.mealPlanningManager.ListRecipeStepIngredients(ctx, request.RecipeId, request.RecipeStepId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step ingredients")
	}

	x := &mealplanning.GetRecipeStepIngredientsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipeStepIngredient := range recipeStepIngredients.Data {
		x.Results = append(x.Results, converters.ConvertRecipeStepIngredientToGRPCRecipeStepIngredient(recipeStepIngredient))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepInstrument(ctx context.Context, request *mealplanning.GetRecipeStepInstrumentRequest) (*mealplanning.GetRecipeStepInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:           request.RecipeStepId,
		mealplanningkeys.RecipeStepInstrumentIDKey: request.RecipeStepInstrumentId,
	}, span, s.logger)

	recipeStepInstrument, err := s.mealPlanningManager.ReadRecipeStepInstrument(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepInstrumentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step instrument")
	}

	x := &mealplanning.GetRecipeStepInstrumentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeStepInstrumentToGRPCRecipeStepInstrument(recipeStepInstrument),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepInstruments(ctx context.Context, request *mealplanning.GetRecipeStepInstrumentsRequest) (*mealplanning.GetRecipeStepInstrumentsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeStepInstruments, err := s.mealPlanningManager.ListRecipeStepInstruments(ctx, request.RecipeId, request.RecipeStepId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step instruments")
	}

	x := &mealplanning.GetRecipeStepInstrumentsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipeStepInstrument := range recipeStepInstruments.Data {
		x.Results = append(x.Results, converters.ConvertRecipeStepInstrumentToGRPCRecipeStepInstrument(recipeStepInstrument))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepProduct(ctx context.Context, request *mealplanning.GetRecipeStepProductRequest) (*mealplanning.GetRecipeStepProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:            request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:        request.RecipeStepId,
		mealplanningkeys.RecipeStepProductIDKey: request.RecipeStepProductId,
	}, span, s.logger)

	recipeStepProduct, err := s.mealPlanningManager.ReadRecipeStepProduct(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepProductId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step")
	}

	x := &mealplanning.GetRecipeStepProductResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeStepProductToGRPCRecipeStepProduct(recipeStepProduct),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepProducts(ctx context.Context, request *mealplanning.GetRecipeStepProductsRequest) (*mealplanning.GetRecipeStepProductsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeStepProducts, err := s.mealPlanningManager.ListRecipeStepProducts(ctx, request.RecipeId, request.RecipeStepId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step product")
	}

	x := &mealplanning.GetRecipeStepProductsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipeStepProduct := range recipeStepProducts.Data {
		x.Results = append(x.Results, converters.ConvertRecipeStepProductToGRPCRecipeStepProduct(recipeStepProduct))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepVessel(ctx context.Context, request *mealplanning.GetRecipeStepVesselRequest) (*mealplanning.GetRecipeStepVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:           request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:       request.RecipeStepId,
		mealplanningkeys.RecipeStepVesselIDKey: request.RecipeStepVesselId,
	}, span, s.logger)

	recipeStepVessel, err := s.mealPlanningManager.ReadRecipeStepVessel(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepVesselId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step")
	}

	x := &mealplanning.GetRecipeStepVesselResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertRecipeStepVesselToGRPCRecipeStepVessel(recipeStepVessel),
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeStepVessels(ctx context.Context, request *mealplanning.GetRecipeStepVesselsRequest) (*mealplanning.GetRecipeStepVesselsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeStepVessels, err := s.mealPlanningManager.ListRecipeStepVessels(ctx, request.RecipeId, request.RecipeStepId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step vessels")
	}

	x := &mealplanning.GetRecipeStepVesselsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipeStepVessel := range recipeStepVessels.Data {
		x.Results = append(x.Results, converters.ConvertRecipeStepVesselToGRPCRecipeStepVessel(recipeStepVessel))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeSteps(ctx context.Context, request *mealplanning.GetRecipeStepsRequest) (*mealplanning.GetRecipeStepsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipeSteps, err := s.mealPlanningManager.ListRecipeSteps(ctx, request.RecipeId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get recipe step vessels")
	}

	x := &mealplanning.GetRecipeStepsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipeStep := range recipeSteps.Data {
		x.Results = append(x.Results, converters.ConvertRecipeStepToGRPCRecipeStep(recipeStep))
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeLists(ctx context.Context, request *mealplanning.GetRecipeListsRequest) (*mealplanning.GetRecipeListsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	lists, err := s.mealPlanningManager.ListRecipeLists(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching recipe lists")
	}

	x := &mealplanning.GetRecipeListsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(lists.Pagination, filter),
	}

	for _, l := range lists.Data {
		x.Results = append(x.Results, converters.ConvertRecipeListToGRPCRecipeList(l))
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeList(ctx context.Context, request *mealplanning.CreateRecipeListRequest) (*mealplanning.CreateRecipeListResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCRecipeListCreationRequestInputToRecipeListCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateRecipeList(ctx, sessionContextData.GetUserID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe list")
	}

	x := &mealplanning.CreateRecipeListResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeListToGRPCRecipeList(created),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeList(ctx context.Context, request *mealplanning.UpdateRecipeListRequest) (*mealplanning.UpdateRecipeListResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeListIDKey: request.RecipeListId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCRecipeListUpdateRequestInputToRecipeListUpdateRequestInput(request.Input)
	if err = s.mealPlanningManager.UpdateRecipeList(ctx, request.RecipeListId, sessionContextData.GetUserID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating recipe list")
	}

	x := &mealplanning.UpdateRecipeListResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeList(ctx context.Context, request *mealplanning.ArchiveRecipeListRequest) (*mealplanning.ArchiveRecipeListResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeListIDKey: request.RecipeListId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.mealPlanningManager.ArchiveRecipeList(ctx, request.RecipeListId, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe list")
	}

	x := &mealplanning.ArchiveRecipeListResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) GetRecipeListItems(ctx context.Context, request *mealplanning.GetRecipeListItemsRequest) (*mealplanning.GetRecipeListItemsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeListIDKey: request.RecipeListId,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	items, err := s.mealPlanningManager.ListRecipeListItems(ctx, request.RecipeListId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching recipe list items")
	}

	x := &mealplanning.GetRecipeListItemsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(items.Pagination, filter),
	}

	for _, item := range items.Data {
		x.Results = append(x.Results, converters.ConvertRecipeListItemToGRPCRecipeListItem(item))
	}

	return x, nil
}

func (s *serviceImpl) CreateRecipeListItem(ctx context.Context, request *mealplanning.CreateRecipeListItemRequest) (*mealplanning.CreateRecipeListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeListIDKey: request.Input.BelongsToRecipeList,
		mealplanningkeys.RecipeIDKey:     request.Input.RecipeId,
	}, span, s.logger)

	input := converters.ConvertGRPCRecipeListItemCreationRequestInputToRecipeListItemCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.AddRecipeToRecipeList(ctx, request.Input.BelongsToRecipeList, input.RecipeID, input.Notes)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating recipe list item")
	}

	x := &mealplanning.CreateRecipeListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertRecipeListItemToGRPCRecipeListItem(created),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeListItem(ctx context.Context, request *mealplanning.UpdateRecipeListItemRequest) (*mealplanning.UpdateRecipeListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeListItemIDKey: request.RecipeListItemId,
	}, span, s.logger)

	input := converters.ConvertGRPCRecipeListItemUpdateRequestInputToRecipeListItemUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeListItem(ctx, request.RecipeListItemId, request.Input.GetBelongsToRecipeList(), request.Input.GetRecipeId(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating recipe list item")
	}

	x := &mealplanning.UpdateRecipeListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveRecipeListItem(ctx context.Context, request *mealplanning.ArchiveRecipeListItemRequest) (*mealplanning.ArchiveRecipeListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeListItemIDKey: request.RecipeListItemId,
		mealplanningkeys.RecipeListIDKey:     request.RecipeListId,
	}, span, s.logger)

	if err := s.mealPlanningManager.RemoveRecipeFromRecipeList(ctx, request.RecipeListId, request.RecipeListItemId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving recipe list item")
	}

	x := &mealplanning.ArchiveRecipeListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) GetRecipes(ctx context.Context, request *mealplanning.GetRecipesRequest) (*mealplanning.GetRecipesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipes, err := s.mealPlanningManager.ListRecipes(ctx, request.Status, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching list of recipes")
	}

	x := &mealplanning.GetRecipesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, recipe := range recipes.Data {
		x.Results = append(x.Results, converters.ConvertRecipeToGRPCRecipe(recipe))
	}

	return x, nil
}

func (s *serviceImpl) SearchForRecipes(ctx context.Context, request *mealplanning.SearchForRecipesRequest) (*mealplanning.SearchForRecipesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		platformkeys.SearchQueryKey: request.Query,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipes, err := s.mealPlanningManager.SearchRecipes(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for recipes")
	}

	x := &mealplanning.SearchForRecipesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(recipes.Pagination, filter),
	}

	for _, recipe := range recipes.Data {
		x.Results = append(x.Results, converters.ConvertRecipeToGRPCRecipe(recipe))
	}

	return x, nil
}

func (s *serviceImpl) SearchForMealEligibleRecipes(ctx context.Context, request *mealplanning.SearchForMealEligibleRecipesRequest) (*mealplanning.SearchForMealEligibleRecipesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		platformkeys.SearchQueryKey: request.Query,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipes, err := s.mealPlanningManager.SearchForMealEligibleRecipes(ctx, request.Query, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for recipes")
	}

	x := &mealplanning.SearchForMealEligibleRecipesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(recipes.Pagination, filter),
	}

	for _, recipe := range recipes.Data {
		x.Results = append(x.Results, converters.ConvertRecipeToGRPCRecipe(recipe))
	}

	return x, nil
}

func (s *serviceImpl) SearchForRecipesWithInstrumentOwnership(ctx context.Context, request *mealplanning.SearchForRecipesWithInstrumentOwnershipRequest) (*mealplanning.SearchForRecipesWithInstrumentOwnershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		platformkeys.SearchQueryKey: request.Query,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	tracing.AttachQueryFilterToSpan(span, filter)

	recipes, err := s.mealPlanningManager.SearchRecipesWithInstrumentOwnership(ctx, sessionContextData.GetActiveAccountID(), request.Query, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for recipes with instrument ownership")
	}

	x := &mealplanning.SearchForRecipesWithInstrumentOwnershipResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(recipes.Pagination, filter),
	}

	for _, recipe := range recipes.Data {
		x.Results = append(x.Results, converters.ConvertRecipeToGRPCRecipe(recipe))
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipe(ctx context.Context, request *mealplanning.UpdateRecipeRequest) (*mealplanning.UpdateRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeUpdateRequestInputToRecipeUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipe(ctx, request.RecipeId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe")
	}

	updated, err := s.mealPlanningManager.ReadRecipe(ctx, request.RecipeId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe")
	}

	x := &mealplanning.UpdateRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeToGRPCRecipe(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStatus(ctx context.Context, request *mealplanning.UpdateRecipeStatusRequest) (*mealplanning.UpdateRecipeStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	if err := s.mealPlanningManager.UpdateRecipeStatus(ctx, request.RecipeId, request.NewStatus); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe")
	}

	updated, err := s.mealPlanningManager.ReadRecipe(ctx, request.RecipeId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe")
	}

	x := &mealplanning.UpdateRecipeStatusResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeToGRPCRecipe(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipePrepTask(ctx context.Context, request *mealplanning.UpdateRecipePrepTaskRequest) (*mealplanning.UpdateRecipePrepTaskResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:         request.RecipeId,
		mealplanningkeys.RecipePrepTaskIDKey: request.RecipePrepTaskId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipePrepTaskUpdateRequestInputToRecipePrepTaskUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipePrepTask(ctx, request.RecipeId, request.RecipePrepTaskId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe prep task")
	}

	updated, err := s.mealPlanningManager.ReadRecipePrepTask(ctx, request.RecipeId, request.RecipePrepTaskId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe prep task")
	}

	x := &mealplanning.UpdateRecipePrepTaskResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipePrepTaskToGRPCRecipePrepTask(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeRating(ctx context.Context, request *mealplanning.UpdateRecipeRatingRequest) (*mealplanning.UpdateRecipeRatingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:       request.RecipeId,
		mealplanningkeys.RecipeRatingIDKey: request.RecipeRatingId,
	}, span, s.logger)

	if err := s.verifyRecipeRatingOwnership(ctx, request.RecipeId, request.RecipeRatingId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeRatingUpdateRequestInputToRecipeRatingUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeRating(ctx, request.RecipeId, request.RecipeRatingId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe rating")
	}

	updated, err := s.mealPlanningManager.ReadRecipeRating(ctx, request.RecipeId, request.RecipeRatingId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe rating")
	}

	x := &mealplanning.UpdateRecipeRatingResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeRatingToGRPCRecipeRating(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStep(ctx context.Context, request *mealplanning.UpdateRecipeStepRequest) (*mealplanning.UpdateRecipeStepResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     request.RecipeId,
		mealplanningkeys.RecipeStepIDKey: request.RecipeStepId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeStepUpdateRequestInputToRecipeStepUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeStep(ctx, request.RecipeId, request.RecipeStepId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe step")
	}

	updated, err := s.mealPlanningManager.ReadRecipeStep(ctx, request.RecipeId, request.RecipeStepId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe step")
	}

	x := &mealplanning.UpdateRecipeStepResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeStepToGRPCRecipeStep(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStepCompletionCondition(ctx context.Context, request *mealplanning.UpdateRecipeStepCompletionConditionRequest) (*mealplanning.UpdateRecipeStepCompletionConditionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:                        request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:                    request.RecipeStepId,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: request.RecipeStepCompletionConditionId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeStepCompletionConditionUpdateRequestInputToRecipeStepCompletionConditionUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeStepCompletionCondition(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepCompletionConditionId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe step completion condition")
	}

	updated, err := s.mealPlanningManager.ReadRecipeStepCompletionCondition(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepCompletionConditionId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe step completion condition")
	}

	x := &mealplanning.UpdateRecipeStepCompletionConditionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeStepCompletionConditionToGRPCRecipeStepCompletionCondition(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStepIngredient(ctx context.Context, request *mealplanning.UpdateRecipeStepIngredientRequest) (*mealplanning.UpdateRecipeStepIngredientResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:           request.RecipeStepId,
		mealplanningkeys.RecipeStepIngredientIDKey: request.RecipeStepIngredientId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeStepIngredientUpdateRequestInputToRecipeStepIngredientUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeStepIngredient(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepIngredientId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe step ingredient")
	}

	updated, err := s.mealPlanningManager.ReadRecipeStepIngredient(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepIngredientId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe step ingredient")
	}

	x := &mealplanning.UpdateRecipeStepIngredientResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeStepIngredientToGRPCRecipeStepIngredient(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStepInstrument(ctx context.Context, request *mealplanning.UpdateRecipeStepInstrumentRequest) (*mealplanning.UpdateRecipeStepInstrumentResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:           request.RecipeStepId,
		mealplanningkeys.RecipeStepInstrumentIDKey: request.RecipeStepInstrumentId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeStepInstrumentUpdateRequestInputToRecipeStepInstrumentUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeStepInstrument(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepInstrumentId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe step instrument")
	}

	updated, err := s.mealPlanningManager.ReadRecipeStepInstrument(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepInstrumentId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe step instrument")
	}

	x := &mealplanning.UpdateRecipeStepInstrumentResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeStepInstrumentToGRPCRecipeStepInstrument(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStepProduct(ctx context.Context, request *mealplanning.UpdateRecipeStepProductRequest) (*mealplanning.UpdateRecipeStepProductResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:            request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:        request.RecipeStepId,
		mealplanningkeys.RecipeStepProductIDKey: request.RecipeStepProductId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeStepProductUpdateRequestInputToRecipeStepProductUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeStepProduct(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepProductId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe step product")
	}

	updated, err := s.mealPlanningManager.ReadRecipeStepProduct(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepProductId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe step product")
	}

	x := &mealplanning.UpdateRecipeStepProductResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeStepProductToGRPCRecipeStepProduct(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateRecipeStepVessel(ctx context.Context, request *mealplanning.UpdateRecipeStepVesselRequest) (*mealplanning.UpdateRecipeStepVesselResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey:           request.RecipeId,
		mealplanningkeys.RecipeStepIDKey:       request.RecipeStepId,
		mealplanningkeys.RecipeStepVesselIDKey: request.RecipeStepVesselId,
	}, span, s.logger)

	if _, err := s.verifyRecipeOwnership(ctx, request.RecipeId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCRecipeStepVesselUpdateRequestInputToRecipeStepVesselUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateRecipeStepVessel(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepVesselId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed updating recipe step vessel")
	}

	updated, err := s.mealPlanningManager.ReadRecipeStepVessel(ctx, request.RecipeId, request.RecipeStepId, request.RecipeStepVesselId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed reading recipe step vessel")
	}

	x := &mealplanning.UpdateRecipeStepVesselResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertRecipeStepVesselToGRPCRecipeStepVessel(updated),
	}

	return x, nil
}
