package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeValidIngredientGroup builds a faked valid ingredient group.
func BuildFakeValidIngredientGroup() *types.ValidIngredientGroup {
	group := fake.BuildFakeRecord[types.ValidIngredientGroup]()

	// Members of this group rather than of three unrelated ones, and at least one of
	// them: the type validates that a group is not empty.
	members := make([]*types.ValidIngredientGroupMember, 0, exampleQuantity)
	for range exampleQuantity {
		member := BuildFakeValidIngredientGroupMember()
		member.BelongsToGroup = group.ID
		members = append(members, member)
	}
	group.Members = members

	return group
}

// BuildFakeValidIngredientGroupMember builds a faked valid ingredient group member.
func BuildFakeValidIngredientGroupMember() *types.ValidIngredientGroupMember {
	member := fake.BuildFakeRecord[types.ValidIngredientGroupMember]()

	// The ingredient the membership is about, built by its own builder so that its
	// storage temperatures are a range rather than two independent numbers.
	member.ValidIngredient = *BuildFakeValidIngredient()

	return member
}

// BuildFakeValidIngredientGroupsList builds a faked ValidIngredientGroupList.
func BuildFakeValidIngredientGroupsList() *filtering.QueryFilteredResult[types.ValidIngredientGroup] {
	return fake.BuildFakePage(BuildFakeValidIngredientGroup)
}

// BuildFakeValidIngredientGroupUpdateRequestInput builds a faked ValidIngredientGroupUpdateRequestInput from a valid ingredient group.
func BuildFakeValidIngredientGroupUpdateRequestInput() *types.ValidIngredientGroupUpdateRequestInput {
	validIngredientGroup := BuildFakeValidIngredientGroup()

	return converters.ConvertValidIngredientGroupToValidIngredientGroupUpdateRequestInput(validIngredientGroup)
}

// BuildFakeValidIngredientGroupCreationRequestInput builds a faked ValidIngredientGroupCreationRequestInput.
func BuildFakeValidIngredientGroupCreationRequestInput() *types.ValidIngredientGroupCreationRequestInput {
	validIngredientGroup := BuildFakeValidIngredientGroup()

	return converters.ConvertValidIngredientGroupToValidIngredientGroupCreationRequestInput(validIngredientGroup)
}
