package grpcconverters

// The conversions in this file are the ones cmd/tools/codegen/converters does not generate. Each
// is listed in that tool's grpcHandWritten with the reason it is not a set of field assignments;
// everything else on the gRPC axis is in converters_generated.go.

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
)

// ConvertMealPlanToGRPCMealPlanSummary projects a MealPlan onto the wire shape
// the list response uses. The events survive -- their dates are what a list of
// plans is read for -- but each carries no options, which is what keeps a page
// inside the 4 MiB gRPC message bound. See the MealPlanSummary comment in
// mealplanning_messages.proto.
//
// The annotations supply the two things the plan itself cannot, both projections
// of the dropped options: whether the requesting user has voted, and each event's
// chosen meal name. A nil annotations leaves both unset, which is what a caller
// with no session to attribute a vote to should pass.
func ConvertMealPlanToGRPCMealPlanSummary(input *mealplanning.MealPlan, annotations *mealplanning.MealPlanSummaryAnnotations) *mealplanningsvc.MealPlanSummary {
	var (
		events        []*mealplanningsvc.MealPlanEventSummary
		chosenByEvent map[string]string
		hasVoted      bool
	)

	if annotations != nil {
		chosenByEvent = annotations.ChosenMealNamesByEventID
		hasVoted = annotations.VotedOnMealPlanIDs[input.ID]
	}

	for _, event := range input.Events {
		summary := ConvertMealPlanEventToGRPCMealPlanEventSummary(event)
		if chosenMealName, ok := chosenByEvent[event.ID]; ok {
			summary.ChosenMealName = &chosenMealName
		}
		events = append(events, summary)
	}

	return &mealplanningsvc.MealPlanSummary{
		CreatedAt:              converters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:          converters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:             converters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		VotingDeadline:         converters.ConvertTimeToPBTimestamp(input.VotingDeadline),
		ElectionMethod:         ConvertStringToMealPlanElectionMethod(input.ElectionMethod),
		Status:                 ConvertStringToMealPlanStatus(input.Status),
		Notes:                  input.Notes,
		Id:                     input.ID,
		BelongsToAccount:       input.BelongsToAccount,
		CreatedByUser:          input.CreatedByUser,
		Events:                 events,
		GroceryListInitialized: input.GroceryListInitialized,
		TasksCreated:           input.TasksCreated,
		CurrentUserHasVoted:    hasVoted,
	}
}

func ConvertGRPCMealListUpdateRequestInputToMealListUpdateRequestInput(input *mealplanningsvc.MealListUpdateRequestInput) *mealplanning.MealListUpdateRequestInput {
	if input == nil {
		return nil
	}

	var name *string
	if input.Name != nil {
		name = new(input.GetName())
	}

	var desc *string
	if input.Description != nil {
		desc = new(input.GetDescription())
	}

	return &mealplanning.MealListUpdateRequestInput{
		Name:        name,
		Description: desc,
	}
}

func ConvertGRPCMealListItemUpdateRequestInputToMealListItemUpdateRequestInput(input *mealplanningsvc.MealListItemUpdateRequestInput) *mealplanning.MealListItemUpdateRequestInput {
	if input == nil {
		return nil
	}

	var notes *string
	if input.Notes != nil {
		notes = new(input.GetNotes())
	}

	return &mealplanning.MealListItemUpdateRequestInput{
		Notes: notes,
	}
}

func ConvertGRPCMealPlanRecipeOptionSelectionUpdateRequestInputToMealPlanRecipeOptionSelectionUpdateRequestInput(input *mealplanningsvc.MealPlanRecipeOptionSelectionUpdateRequestInput) *mealplanning.MealPlanRecipeOptionSelectionUpdateRequestInput {
	if input == nil {
		return nil
	}

	selectedOptionIndex := uint16(input.SelectedOptionIndex)
	return &mealplanning.MealPlanRecipeOptionSelectionUpdateRequestInput{
		SelectedOptionIndex: &selectedOptionIndex,
	}
}

func ConvertGRPCRecipeListUpdateRequestInputToRecipeListUpdateRequestInput(input *mealplanningsvc.RecipeListUpdateRequestInput) *mealplanning.RecipeListUpdateRequestInput {
	if input == nil {
		return nil
	}

	var name *string
	if input.Name != nil {
		name = new(input.GetName())
	}

	var desc *string
	if input.Description != nil {
		desc = new(input.GetDescription())
	}

	return &mealplanning.RecipeListUpdateRequestInput{
		Name:        name,
		Description: desc,
	}
}

func ConvertGRPCRecipeListItemUpdateRequestInputToRecipeListItemUpdateRequestInput(input *mealplanningsvc.RecipeListItemUpdateRequestInput) *mealplanning.RecipeListItemUpdateRequestInput {
	if input == nil {
		return nil
	}

	var notes *string
	if input.Notes != nil {
		notes = new(input.GetNotes())
	}

	return &mealplanning.RecipeListItemUpdateRequestInput{
		Notes: notes,
	}
}

func ConvertGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(input *mealplanningsvc.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	ingredients := []*mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{}
	for _, ingredient := range input.Ingredients {
		ingredients = append(ingredients, &mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{RecipeStepIngredient: ingredient.RecipeStepIngredient})
	}

	return &mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
		IngredientStateID:   input.IngredientStateId,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}
}

func ConvertGRPCValidIngredientGroupCreationRequestInputToValidIngredientGroupCreationRequestInput(request *mealplanningsvc.ValidIngredientGroupCreationRequestInput) *mealplanning.ValidIngredientGroupCreationRequestInput {
	members := make([]*mealplanning.ValidIngredientGroupMemberCreationRequestInput, len(request.Members))
	for i, member := range request.Members {
		members[i] = &mealplanning.ValidIngredientGroupMemberCreationRequestInput{
			ValidIngredientID: member.ValidIngredientId,
		}
	}

	return &mealplanning.ValidIngredientGroupCreationRequestInput{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		Members:     members,
	}
}

func ConvertValidIngredientGroupCreationRequestInputToGRPCValidIngredientGroupCreationRequestInput(request *mealplanning.ValidIngredientGroupCreationRequestInput) *mealplanningsvc.ValidIngredientGroupCreationRequestInput {
	members := make([]*mealplanningsvc.ValidIngredientGroupMemberCreationRequestInput, len(request.Members))
	for i, member := range request.Members {
		members[i] = &mealplanningsvc.ValidIngredientGroupMemberCreationRequestInput{
			ValidIngredientId: member.ValidIngredientID,
		}
	}

	return &mealplanningsvc.ValidIngredientGroupCreationRequestInput{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		Members:     members,
	}
}

func ConvertValidIngredientGroupToGRPCValidIngredientGroup(x *mealplanning.ValidIngredientGroup) *mealplanningsvc.ValidIngredientGroup {
	members := make([]*mealplanningsvc.ValidIngredientGroupMember, len(x.Members))
	for i, member := range x.Members {
		members[i] = &mealplanningsvc.ValidIngredientGroupMember{
			CreatedAt:       converters.ConvertTimeToPBTimestamp(member.CreatedAt),
			ArchivedAt:      converters.ConvertTimePointerToPBTimestamp(member.ArchivedAt),
			Id:              member.ID,
			BelongsToGroup:  member.BelongsToGroup,
			ValidIngredient: ConvertValidIngredientToGRPCValidIngredient(&member.ValidIngredient),
		}
	}

	return &mealplanningsvc.ValidIngredientGroup{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Id:            x.ID,
		Name:          x.Name,
		Slug:          x.Slug,
		Description:   x.Description,
		Members:       members,
	}
}

func ConvertGRPCValidIngredientGroupToValidIngredientGroup(x *mealplanningsvc.ValidIngredientGroup) *mealplanning.ValidIngredientGroup {
	members := make([]*mealplanning.ValidIngredientGroupMember, len(x.Members))
	for i, member := range x.Members {
		members[i] = &mealplanning.ValidIngredientGroupMember{
			CreatedAt:       converters.ConvertPBTimestampToTime(member.CreatedAt),
			ArchivedAt:      converters.ConvertPBTimestampToTimePointer(member.ArchivedAt),
			ID:              member.Id,
			BelongsToGroup:  member.BelongsToGroup,
			ValidIngredient: *ConvertGRPCValidIngredientToValidIngredient(member.ValidIngredient),
		}
	}

	return &mealplanning.ValidIngredientGroup{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		ID:            x.Id,
		Name:          x.Name,
		Slug:          x.Slug,
		Description:   x.Description,
		Members:       members,
	}
}

func ConvertMeasurementUnitConversionMismatchToGRPCMeasurementUnitConversionMismatch(x *mealplanning.MeasurementUnitConversionMismatch) *mealplanningsvc.MeasurementUnitConversionMismatch {
	if x == nil {
		return nil
	}
	return &mealplanningsvc.MeasurementUnitConversionMismatch{
		Ingredient: ConvertValidIngredientToGRPCValidIngredient(&x.Ingredient),
		FromUnit:   ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.FromUnit),
		ToUnit:     ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.ToUnit),
	}
}

// scaleFactorOrOne is the scale factor a creation input or a read entity carries, with an absent,
// zero or negative one meaning unscaled.
func scaleFactorOrOne(scaleFactor float32) float32 {
	if scaleFactor <= 0 {
		return 1.0
	}

	return scaleFactor
}

// scaleFactorPointerOrOne is the scale factor an update input carries. An absent, zero or negative
// one becomes a pointer to one rather than staying absent, and a present one is copied rather than
// shared with the message.
func scaleFactorPointerOrOne(scaleFactor *float32) *float32 {
	if scaleFactor != nil && *scaleFactor > 0 {
		return new(*scaleFactor)
	}

	return new(float32(1.0))
}
