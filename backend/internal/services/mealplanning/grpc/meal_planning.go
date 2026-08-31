package grpc

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"google.golang.org/grpc/codes"
)

// The selection attributes recorded on more than one span below.
const (
	selectionTypeKey   = "selection_type"
	recipeStepIDKey    = "recipe_step_id"
	ingredientIndexKey = "ingredient_index"
)

// verifyMealPlanAccess fetches the session context and confirms the meal plan belongs to the
// requester's active account. It is the service-layer authorization guard for meal-plan
// sub-resource handlers whose manager methods do not accept an account-scoping argument.
func (s *serviceImpl) verifyMealPlanAccess(ctx context.Context, mealPlanID string, logger logging.Logger, span tracing.Span) error {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if _, err = s.mealPlanningManager.ReadMealPlan(ctx, mealPlanID, sessionContextData.GetActiveAccountID()); err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "meal plan not found or access denied")
	}

	return nil
}

// verifyMealPlanOptionAccess fetches the session context and confirms the meal plan option resolves
// (via its event and meal plan) to the requester's active account. It is the service-layer
// authorization guard for the recipe-option-selection handlers, whose requests carry only an option ID.
func (s *serviceImpl) verifyMealPlanOptionAccess(ctx context.Context, mealPlanOptionID string, logger logging.Logger, span tracing.Span) error {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	belongs, err := s.mealPlanningManager.MealPlanOptionBelongsToAccount(ctx, mealPlanOptionID, sessionContextData.GetActiveAccountID())
	if err != nil {
		return errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to verify meal plan option access")
	}

	if !belongs {
		return errorsgrpc.PrepareAndLogGRPCStatus(platformerrors.New("meal plan option not found for account"), logger, span, codes.NotFound, "meal plan option not found or access denied")
	}

	return nil
}

func (s *serviceImpl) ArchiveMeal(ctx context.Context, request *mealplanningsvc.ArchiveMealRequest) (*mealplanningsvc.ArchiveMealResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealIDKey: request.MealId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.mealPlanningManager.ArchiveMeal(ctx, request.MealId, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive meal")
	}

	// The meal's comments are deliberately left alone; see ArchiveRecipe.

	x := &mealplanningsvc.ArchiveMealResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealPlan(ctx context.Context, request *mealplanningsvc.ArchiveMealPlanRequest) (*mealplanningsvc.ArchiveMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.mealPlanningManager.ArchiveMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive meal plan")
	}

	// The meal plan's comments are deliberately left alone; see ArchiveRecipe.

	x := &mealplanningsvc.ArchiveMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealPlanEvent(ctx context.Context, request *mealplanningsvc.ArchiveMealPlanEventRequest) (*mealplanningsvc.ArchiveMealPlanEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      request.MealPlanId,
		mealplanningkeys.MealPlanEventIDKey: request.MealPlanEventId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveMealPlanEvent(ctx, request.MealPlanId, request.MealPlanEventId); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "failed to archive meal plan event")
	}

	x := &mealplanningsvc.ArchiveMealPlanEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealPlanGroceryListItem(ctx context.Context, request *mealplanningsvc.ArchiveMealPlanGroceryListItemRequest) (*mealplanningsvc.ArchiveMealPlanGroceryListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:                request.MealPlanId,
		mealplanningkeys.MealPlanGroceryListItemIDKey: request.MealPlanGroceryListItemId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveMealPlanGroceryListItem(ctx, request.MealPlanId, request.MealPlanGroceryListItemId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive meal plan grocery list item")
	}

	x := &mealplanningsvc.ArchiveMealPlanGroceryListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealPlanOption(ctx context.Context, request *mealplanningsvc.ArchiveMealPlanOptionRequest) (*mealplanningsvc.ArchiveMealPlanOptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       request.MealPlanId,
		mealplanningkeys.MealPlanEventIDKey:  request.MealPlanEventId,
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveMealPlanOption(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive meal plan option")
	}

	x := &mealplanningsvc.ArchiveMealPlanOptionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealPlanOptionVote(ctx context.Context, request *mealplanningsvc.ArchiveMealPlanOptionVoteRequest) (*mealplanningsvc.ArchiveMealPlanOptionVoteResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanOptionVoteIDKey: request.MealPlanOptionVoteId,
		mealplanningkeys.MealPlanOptionIDKey:     request.MealPlanOptionId,
		mealplanningkeys.MealPlanEventIDKey:      request.MealPlanEventId,
		mealplanningkeys.MealPlanIDKey:           request.MealPlanId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.ArchiveMealPlanOptionVote(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId, request.MealPlanOptionVoteId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive meal plan option vote")
	}

	x := &mealplanningsvc.ArchiveMealPlanOptionVoteResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveUserIngredientPreference(ctx context.Context, request *mealplanningsvc.ArchiveUserIngredientPreferenceRequest) (*mealplanningsvc.ArchiveUserIngredientPreferenceResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: request.UserIngredientPreferenceId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.mealPlanningManager.ArchiveUserIngredientPreference(ctx, sessionContextData.GetUserID(), request.UserIngredientPreferenceId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive user ingredient preference")
	}

	x := &mealplanningsvc.ArchiveUserIngredientPreferenceResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) CreateMeal(ctx context.Context, request *mealplanningsvc.CreateMealRequest) (*mealplanningsvc.CreateMealResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if request.Input == nil {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	input := converters.ConvertGRPCMealCreationRequestInputToMealCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMeal(ctx, sessionContextData.GetUserID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal")
	}

	x := &mealplanningsvc.CreateMealResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealToGRPCMeal(created),
	}

	return x, nil
}

func (s *serviceImpl) GetMealLists(ctx context.Context, request *mealplanningsvc.GetMealListsRequest) (*mealplanningsvc.GetMealListsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	lists, err := s.mealPlanningManager.ListMealLists(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching meal lists")
	}

	x := &mealplanningsvc.GetMealListsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(lists.Pagination),
	}

	for _, l := range lists.Data {
		x.Results = append(x.Results, converters.ConvertMealListToGRPCMealList(l))
	}

	return x, nil
}

func (s *serviceImpl) CreateMealList(ctx context.Context, request *mealplanningsvc.CreateMealListRequest) (*mealplanningsvc.CreateMealListResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCMealListCreationRequestInputToMealListCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMealList(ctx, sessionContextData.GetUserID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating meal list")
	}

	x := &mealplanningsvc.CreateMealListResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealListToGRPCMealList(created),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealList(ctx context.Context, request *mealplanningsvc.UpdateMealListRequest) (*mealplanningsvc.UpdateMealListResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealListIDKey: request.MealListId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCMealListUpdateRequestInputToMealListUpdateRequestInput(request.Input)
	if err = s.mealPlanningManager.UpdateMealList(ctx, request.MealListId, sessionContextData.GetUserID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating meal list")
	}

	x := &mealplanningsvc.UpdateMealListResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealList(ctx context.Context, request *mealplanningsvc.ArchiveMealListRequest) (*mealplanningsvc.ArchiveMealListResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealListIDKey: request.MealListId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if err = s.mealPlanningManager.ArchiveMealList(ctx, request.MealListId, sessionContextData.GetUserID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving meal list")
	}

	x := &mealplanningsvc.ArchiveMealListResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) GetMealListItems(ctx context.Context, request *mealplanningsvc.GetMealListItemsRequest) (*mealplanningsvc.GetMealListItemsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealListIDKey: request.MealListId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	items, err := s.mealPlanningManager.ListMealListItems(ctx, request.MealListId, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching meal list items")
	}

	x := &mealplanningsvc.GetMealListItemsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(items.Pagination),
	}

	for _, item := range items.Data {
		x.Results = append(x.Results, converters.ConvertMealListItemToGRPCMealListItem(item))
	}

	return x, nil
}

func (s *serviceImpl) CreateMealListItem(ctx context.Context, request *mealplanningsvc.CreateMealListItemRequest) (*mealplanningsvc.CreateMealListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if request.Input == nil {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealListIDKey: request.Input.BelongsToMealList,
		mealplanningkeys.MealIDKey:     request.Input.MealId,
	}, span, s.logger)

	input := converters.ConvertGRPCMealListItemCreationRequestInputToMealListItemCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.AddMealToMealList(ctx, request.Input.BelongsToMealList, input.MealID, input.Notes)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating meal list item")
	}

	x := &mealplanningsvc.CreateMealListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealListItemToGRPCMealListItem(created),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealListItem(ctx context.Context, request *mealplanningsvc.UpdateMealListItemRequest) (*mealplanningsvc.UpdateMealListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealListItemIDKey: request.MealListItemId,
	}, span, s.logger)

	input := converters.ConvertGRPCMealListItemUpdateRequestInputToMealListItemUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateMealListItem(ctx, request.MealListItemId, request.Input.GetBelongsToMealList(), request.Input.GetMealId(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating meal list item")
	}

	x := &mealplanningsvc.UpdateMealListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealListItem(ctx context.Context, request *mealplanningsvc.ArchiveMealListItemRequest) (*mealplanningsvc.ArchiveMealListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealListItemIDKey: request.MealListItemId,
		mealplanningkeys.MealListIDKey:     request.MealListId,
	}, span, s.logger)

	if err := s.mealPlanningManager.RemoveMealFromMealList(ctx, request.MealListId, request.MealListItemId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving meal list item")
	}

	x := &mealplanningsvc.ArchiveMealListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) CreateMealPlan(ctx context.Context, request *mealplanningsvc.CreateMealPlanRequest) (*mealplanningsvc.CreateMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCMealPlanCreationRequestInputToMealPlanCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMealPlan(ctx, sessionContextData.GetActiveAccountID(), sessionContextData.GetUserID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal plan")
	}

	x := &mealplanningsvc.CreateMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealPlanToGRPCMealPlan(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateMealPlanEvent(ctx context.Context, request *mealplanningsvc.CreateMealPlanEventRequest) (*mealplanningsvc.CreateMealPlanEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanEventCreationRequestInputToMealPlanEventCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMealPlanEvent(ctx, request.MealPlanId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal plan event")
	}

	x := &mealplanningsvc.CreateMealPlanEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealPlanEventToGRPCMealPlanEvent(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateMealPlanOption(ctx context.Context, request *mealplanningsvc.CreateMealPlanOptionRequest) (*mealplanningsvc.CreateMealPlanOptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanOptionCreationRequestInputToMealPlanOptionCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMealPlanOptionWithEventID(ctx, request.MealPlanEventId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal plan option")
	}

	x := &mealplanningsvc.CreateMealPlanOptionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealPlanOptionToGRPCMealPlanOption(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateMealPlanOptionVote(ctx context.Context, request *mealplanningsvc.CreateMealPlanOptionVoteRequest) (*mealplanningsvc.CreateMealPlanOptionVoteResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if _, err = s.mealPlanningManager.ReadMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "meal plan not found or access denied")
	}

	input := converters.ConvertGRPCMealPlanOptionVoteCreationRequestInputToMealPlanOptionVoteCreationRequestInput(request.Input)
	for i := range input.Votes {
		input.Votes[i].ByUser = sessionContextData.GetUserID()
	}

	created, err := s.mealPlanningManager.CreateMealPlanOptionVotes(ctx, request.MealPlanId, request.MealPlanEventId, sessionContextData.GetUserID(), input)
	if err != nil {
		switch {
		case errors.Is(err, mealplanning.ErrMealPlanEventNotEligibleForVoting):
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.FailedPrecondition, "meal plan event is not eligible for voting")
		case errors.Is(err, mealplanning.ErrMealPlanOptionNotFoundForEvent):
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "meal plan option not found for event")
		default:
			return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal plan option vote")
		}
	}

	x := &mealplanningsvc.CreateMealPlanOptionVoteResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, creation := range created {
		x.Created = append(x.Created, converters.ConvertMealPlanOptionVoteToGRPCMealPlanOptionVote(creation))
	}

	return x, nil
}

func (s *serviceImpl) CreateMealPlanTask(ctx context.Context, request *mealplanningsvc.CreateMealPlanTaskRequest) (*mealplanningsvc.CreateMealPlanTaskResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanTaskCreationRequestInputToMealPlanTaskCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMealPlanTask(ctx, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal plan task")
	}

	x := &mealplanningsvc.CreateMealPlanTaskResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealPlanTaskToGRPCMealPlanTask(created),
	}

	return x, nil
}

func (s *serviceImpl) CreateUserIngredientPreference(ctx context.Context, request *mealplanningsvc.CreateUserIngredientPreferenceRequest) (*mealplanningsvc.CreateUserIngredientPreferenceResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCUserIngredientPreferenceCreationRequestInputToUserIngredientPreferenceCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateUserIngredientPreference(ctx, sessionContextData.GetUserID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create user ingredient preference")
	}

	x := &mealplanningsvc.CreateUserIngredientPreferenceResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, creation := range created {
		x.Created = append(x.Created, converters.ConvertUserIngredientPreferenceToGRPCUserIngredientPreference(creation))
	}

	return x, nil
}

func (s *serviceImpl) FinalizeMealPlan(ctx context.Context, request *mealplanningsvc.FinalizeMealPlanRequest) (*mealplanningsvc.FinalizeMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	finalized, err := s.mealPlanningManager.FinalizeMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to finalize meal plan")
	}

	x := &mealplanningsvc.FinalizeMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Finalized: finalized,
	}

	return x, nil
}

func (s *serviceImpl) GetMermaidDiagramForMeal(ctx context.Context, request *mealplanningsvc.GetMermaidDiagramForMealRequest) (*mealplanningsvc.GetMermaidDiagramForMealResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealIDKey: request.MealId,
	}, span, s.logger)

	meal, err := s.mealPlanningManager.ReadMeal(ctx, request.MealId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal")
	}

	mermaidDiagram, err := s.mealPlanningManager.MealMermaid(ctx, meal)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to generate mermaid diagram")
	}

	x := &mealplanningsvc.GetMermaidDiagramForMealResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Response: mermaidDiagram,
	}

	return x, nil
}

func (s *serviceImpl) GetMeal(ctx context.Context, request *mealplanningsvc.GetMealRequest) (*mealplanningsvc.GetMealResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealIDKey: request.MealId,
	}, span, s.logger)

	meal, err := s.mealPlanningManager.ReadMeal(ctx, request.MealId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal")
	}

	x := &mealplanningsvc.GetMealResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealToGRPCMeal(meal),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlan(ctx context.Context, request *mealplanningsvc.GetMealPlanRequest) (*mealplanningsvc.GetMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	mealPlan, err := s.mealPlanningManager.ReadMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan")
	}

	x := &mealplanningsvc.GetMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealPlanToGRPCMealPlan(mealPlan),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlansForAccount(ctx context.Context, request *mealplanningsvc.GetMealPlansForAccountRequest) (*mealplanningsvc.GetMealPlansForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.InvalidArgument, "invalid query filter")
	}

	logger := observability.ObserveValues(nil, span, s.logger)

	mealPlansResult, err := s.mealPlanningManager.ListMealPlans(ctx, sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to list meal plans")
	}

	x := &mealplanningsvc.GetMealPlansForAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(mealPlansResult.Pagination),
	}

	for _, mealPlan := range mealPlansResult.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanToGRPCMealPlan(mealPlan))
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanEvent(ctx context.Context, request *mealplanningsvc.GetMealPlanEventRequest) (*mealplanningsvc.GetMealPlanEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      request.MealPlanId,
		mealplanningkeys.MealPlanEventIDKey: request.MealPlanEventId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanEvent, err := s.mealPlanningManager.ReadMealPlanEvent(ctx, request.MealPlanId, request.MealPlanEventId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan event")
	}

	x := &mealplanningsvc.GetMealPlanEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealPlanEventToGRPCMealPlanEvent(mealPlanEvent),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanEvents(ctx context.Context, request *mealplanningsvc.GetMealPlanEventsRequest) (*mealplanningsvc.GetMealPlanEventsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	if err = s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanEventsResult, err := s.mealPlanningManager.ListMealPlanEvents(ctx, request.MealPlanId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meal plan events")
	}

	x := &mealplanningsvc.GetMealPlanEventsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(mealPlanEventsResult.Pagination),
	}

	for _, mealPlanEvent := range mealPlanEventsResult.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanEventToGRPCMealPlanEvent(mealPlanEvent))
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanGroceryListItem(ctx context.Context, request *mealplanningsvc.GetMealPlanGroceryListItemRequest) (*mealplanningsvc.GetMealPlanGroceryListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:                request.MealPlanId,
		mealplanningkeys.MealPlanGroceryListItemIDKey: request.MealPlanGroceryListItemId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanGroceryListItem, err := s.mealPlanningManager.ReadMealPlanGroceryListItem(ctx, request.MealPlanId, request.MealPlanGroceryListItemId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan grocery list item")
	}

	x := &mealplanningsvc.GetMealPlanGroceryListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealPlanGroceryListItemToGRPCMealPlanGroceryListItem(mealPlanGroceryListItem),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanGroceryListItemsForMealPlan(ctx context.Context, request *mealplanningsvc.GetMealPlanGroceryListItemsForMealPlanRequest) (*mealplanningsvc.GetMealPlanGroceryListItemsForMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	if err = s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanGroceryListItems, err := s.mealPlanningManager.ListMealPlanGroceryListItemsByMealPlan(ctx, request.MealPlanId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meal plan grocery list items")
	}

	x := &mealplanningsvc.GetMealPlanGroceryListItemsForMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(mealPlanGroceryListItems.Pagination),
	}

	for _, mealPlanGroceryListItem := range mealPlanGroceryListItems.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanGroceryListItemToGRPCMealPlanGroceryListItem(mealPlanGroceryListItem))
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanRecipeOptionSelection(ctx context.Context, request *mealplanningsvc.GetMealPlanRecipeOptionSelectionRequest) (*mealplanningsvc.GetMealPlanRecipeOptionSelectionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
		recipeStepIDKey:                      request.RecipeStepId,
		ingredientIndexKey:                   request.IngredientIndex,
		selectionTypeKey:                     request.SelectionType,
	}, span, s.logger)

	if err := s.verifyMealPlanOptionAccess(ctx, request.MealPlanOptionId, logger, span); err != nil {
		return nil, err
	}

	selection, err := s.mealPlanningManager.GetMealPlanRecipeOptionSelection(ctx, request.MealPlanOptionId, request.RecipeStepId, uint16(request.IngredientIndex), converters.ConvertMealPlanRecipeOptionSelectionTypeToString(request.SelectionType))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan recipe option selection")
	}

	x := &mealplanningsvc.GetMealPlanRecipeOptionSelectionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	if selection != nil {
		x.Result = converters.ConvertMealPlanRecipeOptionSelectionToGRPCMealPlanRecipeOptionSelection(selection)
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx context.Context, request *mealplanningsvc.GetMealPlanRecipeOptionSelectionsForMealPlanOptionRequest) (*mealplanningsvc.GetMealPlanRecipeOptionSelectionsForMealPlanOptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
	}, span, s.logger)

	if err := s.verifyMealPlanOptionAccess(ctx, request.MealPlanOptionId, logger, span); err != nil {
		return nil, err
	}

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	selections, err := s.mealPlanningManager.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, request.MealPlanOptionId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meal plan recipe option selections")
	}

	x := &mealplanningsvc.GetMealPlanRecipeOptionSelectionsForMealPlanOptionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(selections.Pagination),
	}

	for _, selection := range selections.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanRecipeOptionSelectionToGRPCMealPlanRecipeOptionSelection(selection))
	}

	return x, nil
}

func (s *serviceImpl) CreateMealPlanRecipeOptionSelection(ctx context.Context, request *mealplanningsvc.CreateMealPlanRecipeOptionSelectionRequest) (*mealplanningsvc.CreateMealPlanRecipeOptionSelectionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if request.Input == nil {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: request.Input.BelongsToMealPlanOption,
	}, span, s.logger)

	if err := s.verifyMealPlanOptionAccess(ctx, request.MealPlanOptionId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanRecipeOptionSelectionCreationRequestInputToMealPlanRecipeOptionSelectionCreationRequestInput(request.Input)

	created, err := s.mealPlanningManager.CreateMealPlanRecipeOptionSelection(ctx, request.MealPlanOptionId, input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create meal plan recipe option selection")
	}

	x := &mealplanningsvc.CreateMealPlanRecipeOptionSelectionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertMealPlanRecipeOptionSelectionToGRPCMealPlanRecipeOptionSelection(created),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlanRecipeOptionSelection(ctx context.Context, request *mealplanningsvc.UpdateMealPlanRecipeOptionSelectionRequest) (*mealplanningsvc.UpdateMealPlanRecipeOptionSelectionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
		recipeStepIDKey:                      request.RecipeStepId,
		ingredientIndexKey:                   request.IngredientIndex,
		selectionTypeKey:                     request.SelectionType,
	}, span, s.logger)

	if err := s.verifyMealPlanOptionAccess(ctx, request.MealPlanOptionId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanRecipeOptionSelectionUpdateRequestInputToMealPlanRecipeOptionSelectionUpdateRequestInput(request.Input)

	selectionTypeStr := converters.ConvertMealPlanRecipeOptionSelectionTypeToString(request.SelectionType)
	if err := s.mealPlanningManager.UpdateMealPlanRecipeOptionSelection(ctx, request.MealPlanOptionId, request.RecipeStepId, uint16(request.IngredientIndex), selectionTypeStr, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan recipe option selection")
	}

	updated, err := s.mealPlanningManager.GetMealPlanRecipeOptionSelection(ctx, request.MealPlanOptionId, request.RecipeStepId, uint16(request.IngredientIndex), selectionTypeStr)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan recipe option selection")
	}

	x := &mealplanningsvc.UpdateMealPlanRecipeOptionSelectionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanRecipeOptionSelectionToGRPCMealPlanRecipeOptionSelection(updated),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveMealPlanRecipeOptionSelection(ctx context.Context, request *mealplanningsvc.ArchiveMealPlanRecipeOptionSelectionRequest) (*mealplanningsvc.ArchiveMealPlanRecipeOptionSelectionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
		recipeStepIDKey:                      request.RecipeStepId,
		ingredientIndexKey:                   request.IngredientIndex,
		selectionTypeKey:                     request.SelectionType,
	}, span, s.logger)

	if err := s.verifyMealPlanOptionAccess(ctx, request.MealPlanOptionId, logger, span); err != nil {
		return nil, err
	}

	selectionTypeStr := converters.ConvertMealPlanRecipeOptionSelectionTypeToString(request.SelectionType)
	if err := s.mealPlanningManager.ArchiveMealPlanRecipeOptionSelection(ctx, request.MealPlanOptionId, request.RecipeStepId, uint16(request.IngredientIndex), selectionTypeStr); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive meal plan recipe option selection")
	}

	x := &mealplanningsvc.ArchiveMealPlanRecipeOptionSelectionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanOption(ctx context.Context, request *mealplanningsvc.GetMealPlanOptionRequest) (*mealplanningsvc.GetMealPlanOptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       request.MealPlanId,
		mealplanningkeys.MealPlanEventIDKey:  request.MealPlanEventId,
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanOption, err := s.mealPlanningManager.ReadMealPlanOption(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan option")
	}

	x := &mealplanningsvc.GetMealPlanOptionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealPlanOptionToGRPCMealPlanOption(mealPlanOption),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanOptionVote(ctx context.Context, request *mealplanningsvc.GetMealPlanOptionVoteRequest) (*mealplanningsvc.GetMealPlanOptionVoteResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:           request.MealPlanId,
		mealplanningkeys.MealPlanOptionIDKey:     request.MealPlanOptionId,
		mealplanningkeys.MealPlanEventIDKey:      request.MealPlanEventId,
		mealplanningkeys.MealPlanOptionVoteIDKey: request.MealPlanOptionVoteId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanOptionVote, err := s.mealPlanningManager.ReadMealPlanOptionVote(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId, request.MealPlanOptionVoteId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan option vote")
	}

	x := &mealplanningsvc.GetMealPlanOptionVoteResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealPlanOptionVoteToGRPCMealPlanOptionVote(mealPlanOptionVote),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanOptionVotes(ctx context.Context, request *mealplanningsvc.GetMealPlanOptionVotesRequest) (*mealplanningsvc.GetMealPlanOptionVotesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       request.MealPlanId,
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
		mealplanningkeys.MealPlanEventIDKey:  request.MealPlanEventId,
	}, span, s.logger)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	if err = s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanOptionVotesResult, err := s.mealPlanningManager.ListMealPlanOptionVotes(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meal plan option votes")
	}

	x := &mealplanningsvc.GetMealPlanOptionVotesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	for _, mealPlanOptionVote := range mealPlanOptionVotesResult.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanOptionVoteToGRPCMealPlanOptionVote(mealPlanOptionVote))
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanOptions(ctx context.Context, request *mealplanningsvc.GetMealPlanOptionsRequest) (*mealplanningsvc.GetMealPlanOptionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	if err = s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanOptionsResult, err := s.mealPlanningManager.ListMealPlanOptions(ctx, request.MealPlanId, request.MealPlanEventId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meal plan options")
	}

	x := &mealplanningsvc.GetMealPlanOptionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(mealPlanOptionsResult.Pagination),
	}

	for _, mealPlanOption := range mealPlanOptionsResult.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanOptionToGRPCMealPlanOption(mealPlanOption))
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanTask(ctx context.Context, request *mealplanningsvc.GetMealPlanTaskRequest) (*mealplanningsvc.GetMealPlanTaskResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:     request.MealPlanId,
		mealplanningkeys.MealPlanTaskIDKey: request.MealPlanTaskId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanTask, err := s.mealPlanningManager.ReadMealPlanTask(ctx, request.MealPlanId, request.MealPlanTaskId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read meal plan task")
	}

	x := &mealplanningsvc.GetMealPlanTaskResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertMealPlanTaskToGRPCMealPlanTask(mealPlanTask),
	}

	return x, nil
}

func (s *serviceImpl) GetMealPlanTasks(ctx context.Context, request *mealplanningsvc.GetMealPlanTasksRequest) (*mealplanningsvc.GetMealPlanTasksResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	if err = s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	mealPlanTasks, err := s.mealPlanningManager.ListMealPlanTasksByMealPlan(ctx, request.MealPlanId, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meal plan tasks")
	}

	x := &mealplanningsvc.GetMealPlanTasksResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(mealPlanTasks.Pagination),
	}

	for _, mealPlanTask := range mealPlanTasks.Data {
		x.Results = append(x.Results, converters.ConvertMealPlanTaskToGRPCMealPlanTask(mealPlanTask))
	}

	return x, nil
}

func (s *serviceImpl) GetMeals(ctx context.Context, request *mealplanningsvc.GetMealsRequest) (*mealplanningsvc.GetMealsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	mealsResult, err := s.mealPlanningManager.ListMeals(ctx, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of meals")
	}

	x := &mealplanningsvc.GetMealsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(mealsResult.Pagination),
	}

	for _, meal := range mealsResult.Data {
		x.Results = append(x.Results, converters.ConvertMealToGRPCMealSummary(meal))
	}

	return x, nil
}

func (s *serviceImpl) GetUserIngredientPreference(ctx context.Context, request *mealplanningsvc.GetUserIngredientPreferenceRequest) (*mealplanningsvc.GetUserIngredientPreferenceResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userIngredientPreference, err := s.mealPlanningManager.ReadUserIngredientPreference(ctx, sessionContextData.GetUserID(), request.UserIngredientPreferenceId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to read user ingredient preference")
	}

	x := &mealplanningsvc.GetUserIngredientPreferenceResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertUserIngredientPreferenceToGRPCUserIngredientPreference(userIngredientPreference),
	}

	return x, nil
}

func (s *serviceImpl) GetUserIngredientPreferences(ctx context.Context, request *mealplanningsvc.GetUserIngredientPreferencesRequest) (*mealplanningsvc.GetUserIngredientPreferencesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)
	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userIngredientPreferencesResult, err := s.mealPlanningManager.ListUserIngredientPreferences(ctx, sessionContextData.GetUserID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch list of user ingredient preferences")
	}

	x := &mealplanningsvc.GetUserIngredientPreferencesResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(userIngredientPreferencesResult.Pagination),
	}

	for _, userIngredientPreference := range userIngredientPreferencesResult.Data {
		x.Results = append(x.Results, converters.ConvertUserIngredientPreferenceToGRPCUserIngredientPreference(userIngredientPreference))
	}

	return x, nil
}

// RunFinalizeMealPlanWorker starts a finalization saga for every meal plan that needs one, and
// reports how many it started.
//
// Finalizing, creating tasks, and initializing the grocery list used to be three workers with
// three RPCs to run them on demand. They are one saga now, so all three RPCs do this — see
// RunMealPlanGroceryListInitializerWorker. Finalized counts sagas started rather than plans
// finalized: this call no longer waits for the pipeline, it only enters plans into it.
func (s *serviceImpl) RunFinalizeMealPlanWorker(ctx context.Context, _ *mealplanningsvc.RunFinalizeMealPlanWorkerRequest) (*mealplanningsvc.RunFinalizeMealPlanWorkerResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	started, err := s.mealPlanFinalizationStarter.Work(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "starting meal plan finalization sagas")
	}

	x := &mealplanningsvc.RunFinalizeMealPlanWorkerResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Finalized: uint32(started),
	}

	return x, nil
}

// RunMealPlanGroceryListInitializerWorker starts a finalization saga for every meal plan that
// needs one.
//
// It does the same thing RunFinalizeMealPlanWorker does, because there is no longer a separate
// grocery list job to run: a plan whose grocery list is missing is a plan the finalization saga
// owes a step to, and the starter's query picks it up on exactly those terms. The RPC is kept so
// that clients built against the old three-worker shape keep working; it is a candidate for
// removal the next time this service's proto changes.
func (s *serviceImpl) RunMealPlanGroceryListInitializerWorker(ctx context.Context, _ *mealplanningsvc.RunMealPlanGroceryListInitializerWorkerRequest) (*mealplanningsvc.RunMealPlanGroceryListInitializerWorkerResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if _, err := s.mealPlanFinalizationStarter.Work(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "starting meal plan finalization sagas")
	}

	x := &mealplanningsvc.RunMealPlanGroceryListInitializerWorkerResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

// RunMealPlanTaskCreatorWorker starts a finalization saga for every meal plan that needs one.
// See RunMealPlanGroceryListInitializerWorker for why it is the same call.
func (s *serviceImpl) RunMealPlanTaskCreatorWorker(ctx context.Context, _ *mealplanningsvc.RunMealPlanTaskCreatorWorkerRequest) (*mealplanningsvc.RunMealPlanTaskCreatorWorkerResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if _, err := s.mealPlanFinalizationStarter.Work(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "starting meal plan finalization sagas")
	}

	x := &mealplanningsvc.RunMealPlanTaskCreatorWorkerResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) SearchForMeals(ctx context.Context, request *mealplanningsvc.SearchForMealsRequest) (*mealplanningsvc.SearchForMealsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "invalid query filter")
	}

	meals, err := s.mealPlanningManager.SearchMeals(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to search for meals")
	}

	x := &mealplanningsvc.SearchForMealsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(meals.Pagination),
	}

	for _, meal := range meals.Data {
		x.Results = append(x.Results, converters.ConvertMealToGRPCMealSummary(meal))
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlan(ctx context.Context, request *mealplanningsvc.UpdateMealPlanRequest) (*mealplanningsvc.UpdateMealPlanResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: request.MealPlanId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCMealPlanUpdateRequestInputToMealPlanUpdateRequestInput(request.Input)

	if err = s.mealPlanningManager.UpdateMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan")
	}

	updated, err := s.mealPlanningManager.ReadMealPlan(ctx, request.MealPlanId, sessionContextData.GetActiveAccountID())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan")
	}

	x := &mealplanningsvc.UpdateMealPlanResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanToGRPCMealPlan(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlanEvent(ctx context.Context, request *mealplanningsvc.UpdateMealPlanEventRequest) (*mealplanningsvc.UpdateMealPlanEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      request.MealPlanId,
		mealplanningkeys.MealPlanEventIDKey: request.MealPlanEventId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanEventUpdateRequestInputToMealPlanEventUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateMealPlanEvent(ctx, request.MealPlanId, request.MealPlanEventId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan event")
	}

	updated, err := s.mealPlanningManager.ReadMealPlanEvent(ctx, request.MealPlanId, request.MealPlanEventId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan event")
	}

	x := &mealplanningsvc.UpdateMealPlanEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanEventToGRPCMealPlanEvent(updated),
	}

	return x, nil
}

func (s *serviceImpl) SwapMealPlanEvents(ctx context.Context, request *mealplanningsvc.SwapMealPlanEventsRequest) (*mealplanningsvc.SwapMealPlanEventsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      request.MealPlanId,
		mealplanningkeys.MealPlanEventIDKey: request.MealPlanEventIdA + "," + request.MealPlanEventIdB,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	if err := s.mealPlanningManager.SwapMealPlanEvents(ctx, request.MealPlanId, request.MealPlanEventIdA, request.MealPlanEventIdB); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to swap meal plan events")
	}

	x := &mealplanningsvc.SwapMealPlanEventsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlanGroceryListItem(ctx context.Context, request *mealplanningsvc.UpdateMealPlanGroceryListItemRequest) (*mealplanningsvc.UpdateMealPlanGroceryListItemResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:                request.MealPlanId,
		mealplanningkeys.MealPlanGroceryListItemIDKey: request.MealPlanGroceryListItemId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanGroceryListItemUpdateRequestInputToMealPlanGroceryListItemUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateMealPlanGroceryListItem(ctx, request.MealPlanId, request.MealPlanGroceryListItemId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan grocery list item")
	}

	updated, err := s.mealPlanningManager.ReadMealPlanGroceryListItem(ctx, request.MealPlanId, request.MealPlanGroceryListItemId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan grocery list item")
	}

	x := &mealplanningsvc.UpdateMealPlanGroceryListItemResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanGroceryListItemToGRPCMealPlanGroceryListItem(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlanOption(ctx context.Context, request *mealplanningsvc.UpdateMealPlanOptionRequest) (*mealplanningsvc.UpdateMealPlanOptionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       request.MealPlanId,
		mealplanningkeys.MealPlanOptionIDKey: request.MealPlanOptionId,
		mealplanningkeys.MealPlanEventIDKey:  request.MealPlanEventId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanOptionUpdateRequestInputToMealPlanOptionUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateMealPlanOption(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan option")
	}

	updated, err := s.mealPlanningManager.ReadMealPlanOption(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan option")
	}

	x := &mealplanningsvc.UpdateMealPlanOptionResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanOptionToGRPCMealPlanOption(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlanOptionVote(ctx context.Context, request *mealplanningsvc.UpdateMealPlanOptionVoteRequest) (*mealplanningsvc.UpdateMealPlanOptionVoteResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:           request.MealPlanId,
		mealplanningkeys.MealPlanOptionIDKey:     request.MealPlanOptionId,
		mealplanningkeys.MealPlanOptionVoteIDKey: request.MealPlanOptionVoteId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanOptionVoteUpdateRequestInputToMealPlanOptionVoteUpdateRequestInput(request.Input)

	if err := s.mealPlanningManager.UpdateMealPlanOptionVote(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId, request.MealPlanOptionVoteId, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan option vote")
	}

	updated, err := s.mealPlanningManager.ReadMealPlanOptionVote(ctx, request.MealPlanId, request.MealPlanEventId, request.MealPlanOptionId, request.MealPlanOptionVoteId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan option vote")
	}

	x := &mealplanningsvc.UpdateMealPlanOptionVoteResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanOptionVoteToGRPCMealPlanOptionVote(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateMealPlanTaskStatus(ctx context.Context, request *mealplanningsvc.UpdateMealPlanTaskStatusRequest) (*mealplanningsvc.UpdateMealPlanTaskStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:     request.MealPlanId,
		mealplanningkeys.MealPlanTaskIDKey: request.MealPlanTaskId,
	}, span, s.logger)

	if err := s.verifyMealPlanAccess(ctx, request.MealPlanId, logger, span); err != nil {
		return nil, err
	}

	input := converters.ConvertGRPCMealPlanTaskStatusChangeRequestInputToMealPlanTaskStatusChangeRequestInput(request.Input)
	input.MealPlanTaskID = request.MealPlanTaskId

	if err := s.mealPlanningManager.MealPlanTaskStatusChange(ctx, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update meal plan task status")
	}

	updated, err := s.mealPlanningManager.ReadMealPlanTask(ctx, request.MealPlanId, request.MealPlanTaskId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated meal plan task status")
	}

	x := &mealplanningsvc.UpdateMealPlanTaskStatusResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertMealPlanTaskToGRPCMealPlanTask(updated),
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserIngredientPreference(ctx context.Context, request *mealplanningsvc.UpdateUserIngredientPreferenceRequest) (*mealplanningsvc.UpdateUserIngredientPreferenceResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: request.UserIngredientPreferenceId,
	}, span, s.logger)

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	input := converters.ConvertGRPCUserIngredientPreferenceUpdateRequestInputToUserIngredientPreferenceUpdateRequestInput(request.Input)

	if err = s.mealPlanningManager.UpdateUserIngredientPreference(ctx, request.UserIngredientPreferenceId, sessionContextData.GetUserID(), input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user ingredient preference")
	}

	updated, err := s.mealPlanningManager.ReadUserIngredientPreference(ctx, sessionContextData.GetUserID(), request.UserIngredientPreferenceId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch updated user ingredient preference")
	}

	x := &mealplanningsvc.UpdateUserIngredientPreferenceResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Updated: converters.ConvertUserIngredientPreferenceToGRPCUserIngredientPreference(updated),
	}

	return x, nil
}

func (s *serviceImpl) CreateAccountInstrumentOwnership(ctx context.Context, request *mealplanningsvc.CreateAccountInstrumentOwnershipRequest) (*mealplanningsvc.CreateAccountInstrumentOwnershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	logger := observability.ObserveValues(nil, span, s.logger)

	input := converters.ConvertGRPCAccountInstrumentOwnershipCreationRequestInputToAccountInstrumentOwnershipCreationRequestInput(request.Input)
	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate account instrument ownership creation request input")
	}

	created, err := s.mealPlanningManager.CreateAccountInstrumentOwnership(ctx, sessionContextData.GetActiveAccountID(), input)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create account instrument ownership")
	}

	x := &mealplanningsvc.CreateAccountInstrumentOwnershipResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertAccountInstrumentOwnershipToGRPCAccountInstrumentOwnership(created),
	}

	return x, nil
}

func (s *serviceImpl) GetAccountInstrumentOwnership(ctx context.Context, request *mealplanningsvc.GetAccountInstrumentOwnershipRequest) (*mealplanningsvc.GetAccountInstrumentOwnershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.AccountInstrumentOwnershipIDKey: request.AccountInstrumentOwnershipId,
	}, span, s.logger)

	result, err := s.mealPlanningManager.ReadAccountInstrumentOwnership(ctx, sessionContextData.GetActiveAccountID(), request.AccountInstrumentOwnershipId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch account instrument ownership")
	}

	x := &mealplanningsvc.GetAccountInstrumentOwnershipResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertAccountInstrumentOwnershipToGRPCAccountInstrumentOwnership(result),
	}

	return x, nil
}

func (s *serviceImpl) GetAccountInstrumentOwnerships(ctx context.Context, request *mealplanningsvc.GetAccountInstrumentOwnershipsRequest) (*mealplanningsvc.GetAccountInstrumentOwnershipsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.InvalidArgument, "invalid query filter")
	}

	logger := observability.ObserveValues(nil, span, s.logger)

	results, err := s.mealPlanningManager.ListAccountInstrumentOwnerships(ctx, sessionContextData.GetActiveAccountID(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to list account instrument ownerships")
	}

	x := &mealplanningsvc.GetAccountInstrumentOwnershipsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(results.Pagination),
	}

	for _, result := range results.Data {
		x.Results = append(x.Results, converters.ConvertAccountInstrumentOwnershipToGRPCAccountInstrumentOwnership(result))
	}

	return x, nil
}

func (s *serviceImpl) SearchForValidInstrumentsNotOwnedByAccount(ctx context.Context, request *mealplanningsvc.SearchForValidInstrumentsNotOwnedByAccountRequest) (*mealplanningsvc.SearchForValidInstrumentsNotOwnedByAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	filter, err := filteringgrpc.FromProto(request.Filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.InvalidArgument, "invalid query filter")
	}

	logger := observability.ObserveValues(nil, span, s.logger)

	x, err := s.mealPlanningManager.SearchValidInstrumentsNotOwnedByAccount(ctx, sessionContextData.GetActiveAccountID(), request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "searching for valid instruments not owned by account")
	}

	res := &mealplanningsvc.SearchForValidInstrumentsNotOwnedByAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: filteringgrpc.PaginationToProto(x.Pagination),
	}

	for _, y := range x.Data {
		res.Results = append(res.Results, converters.ConvertValidInstrumentToGRPCValidInstrument(y))
	}

	return res, nil
}

func (s *serviceImpl) UpdateAccountInstrumentOwnership(ctx context.Context, request *mealplanningsvc.UpdateAccountInstrumentOwnershipRequest) (*mealplanningsvc.UpdateAccountInstrumentOwnershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.AccountInstrumentOwnershipIDKey: request.AccountInstrumentOwnershipId,
	}, span, s.logger)

	input := converters.ConvertGRPCAccountInstrumentOwnershipUpdateRequestInputToAccountInstrumentOwnershipUpdateRequestInput(request.Input)

	accountInstrumentOwnership, err := s.mealPlanningManager.ReadAccountInstrumentOwnership(ctx, sessionContextData.GetActiveAccountID(), request.AccountInstrumentOwnershipId)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch account instrument ownership")
	}

	if err = s.mealPlanningManager.UpdateAccountInstrumentOwnership(ctx, accountInstrumentOwnership.ID, accountInstrumentOwnership.BelongsToAccount, input); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update account instrument ownership")
	}

	x := &mealplanningsvc.UpdateAccountInstrumentOwnershipResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveAccountInstrumentOwnership(ctx context.Context, request *mealplanningsvc.ArchiveAccountInstrumentOwnershipRequest) (*mealplanningsvc.ArchiveAccountInstrumentOwnershipResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	logger := observability.ObserveValues(map[string]any{
		mealplanningkeys.AccountInstrumentOwnershipIDKey: request.AccountInstrumentOwnershipId,
	}, span, s.logger)

	if err = s.mealPlanningManager.ArchiveAccountInstrumentOwnership(ctx, sessionContextData.GetActiveAccountID(), request.AccountInstrumentOwnershipId); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive account instrument ownership")
	}

	x := &mealplanningsvc.ArchiveAccountInstrumentOwnershipResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
