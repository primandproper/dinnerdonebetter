package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"
	converters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	errorsgrpc "github.com/primandproper/platform-go/v6/errors/grpc"
	"github.com/primandproper/platform-go/v6/observability"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) AddCommentToRecipe(ctx context.Context, request *mealplanningsvc.AddCommentToRecipeRequest) (*mealplanningsvc.AddCommentToRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.RecipeId,
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertProtoCommentCreationRequestInputToDomain(
		request.Input,
		mealplanning.CommentTargetTypeRecipes,
		request.RecipeId,
		sessionContextData.GetUserID(),
	)
	if input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment, err := s.commentsManager.CreateComment(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	return &mealplanningsvc.AddCommentToRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: converters.ConvertCommentToGRPCComment(comment),
	}, nil
}

func (s *serviceImpl) AddCommentToMeal(ctx context.Context, request *mealplanningsvc.AddCommentToMealRequest) (*mealplanningsvc.AddCommentToMealResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealIDKey: request.MealId,
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertProtoCommentCreationRequestInputToDomain(
		request.Input,
		mealplanning.CommentTargetTypeMeals,
		request.MealId,
		sessionContextData.GetUserID(),
	)
	if input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment, err := s.commentsManager.CreateComment(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	return &mealplanningsvc.AddCommentToMealResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: converters.ConvertCommentToGRPCComment(comment),
	}, nil
}

func (s *serviceImpl) AddCommentToMealPlan(ctx context.Context, request *mealplanningsvc.AddCommentToMealPlanRequest) (*mealplanningsvc.AddCommentToMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if _, err = s.mealPlanningManager.ReadMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "validating meal plan access")
	}

	input := converters.ConvertProtoCommentCreationRequestInputToDomain(
		request.Input,
		mealplanning.CommentTargetTypeMealPlans,
		request.MealPlanId,
		sessionContextData.GetUserID(),
	)
	if input == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment, err := s.commentsManager.CreateComment(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	return &mealplanningsvc.AddCommentToMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: converters.ConvertCommentToGRPCComment(comment),
	}, nil
}
