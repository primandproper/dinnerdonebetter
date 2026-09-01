package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/grpc/converters"

	comments "github.com/primandproper/platform-go/v13/comments"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) AddCommentToRecipe(ctx context.Context, request *mealplanningsvc.AddCommentToRecipeRequest) (*mealplanningsvc.AddCommentToRecipeResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.RecipeIDKey: request.GetRecipeId(),
	}, span, s.logger)

	comment, err := s.addComment(ctx, span, logger, request.GetInput(), comments.Target{
		Type: mealplanning.CommentTargetTypeRecipes,
		ID:   request.GetRecipeId(),
	})
	if err != nil {
		return nil, err
	}

	return &mealplanningsvc.AddCommentToRecipeResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: comment,
	}, nil
}

func (s *serviceImpl) AddCommentToMeal(ctx context.Context, request *mealplanningsvc.AddCommentToMealRequest) (*mealplanningsvc.AddCommentToMealResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealIDKey: request.GetMealId(),
	}, span, s.logger)

	comment, err := s.addComment(ctx, span, logger, request.GetInput(), comments.Target{
		Type: mealplanning.CommentTargetTypeMeals,
		ID:   request.GetMealId(),
	})
	if err != nil {
		return nil, err
	}

	return &mealplanningsvc.AddCommentToMealResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: comment,
	}, nil
}

func (s *serviceImpl) AddCommentToMealPlan(ctx context.Context, request *mealplanningsvc.AddCommentToMealPlanRequest) (*mealplanningsvc.AddCommentToMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.GetMealPlanId(),
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	// The meal plan is read as the caller, which is why the comment target catalog
	// registers no existence check for this type: the check platform would run is
	// given only the target ID, and a plan somebody else's household owns is not a
	// plan this caller may comment on. See internal/build/comments.
	if _, err = s.mealPlanningManager.ReadMealPlan(ctx, request.GetMealPlanId(), sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "validating meal plan access")
	}

	comment, err := s.addComment(ctx, span, logger, request.GetInput(), comments.Target{
		Type: mealplanning.CommentTargetTypeMealPlans,
		ID:   request.GetMealPlanId(),
	})
	if err != nil {
		return nil, err
	}

	return &mealplanningsvc.AddCommentToMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Comment: comment,
	}, nil
}

// addComment writes one comment about a thing this service owns.
//
// The target comes from the request path rather than the body, so a client cannot
// file a comment against a recipe by posting to a meal. Whether that recipe is
// there at all is the store's question, answered through the existence check the
// target catalog registers.
func (s *serviceImpl) addComment(
	ctx context.Context,
	span tracing.Span,
	logger logging.Logger,
	input *commentssvc.CommentCreationRequestInput,
	target comments.Target,
) (*commentssvc.Comment, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	comment := converters.ConvertProtoCommentCreationRequestInputToDomain(input, target, sessionContextData.GetUserID())
	if comment == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("input is required"), logger, span, codes.InvalidArgument, "input is required")
	}

	comment.Scope = ddbcomments.Scope()

	if err = s.comments.CreateComment(ctx, comment); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating comment")
	}

	tracing.AttachToSpan(span, commentskeys.CommentIDKey, comment.ID)

	return converters.ConvertCommentToGRPCComment(comment), nil
}
