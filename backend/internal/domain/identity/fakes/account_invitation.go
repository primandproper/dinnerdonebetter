package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"

	"github.com/primandproper/platform-go/v6/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeAccountInvitation builds a faked AccountInvitation.
func BuildFakeAccountInvitation() *types.AccountInvitation {
	return &types.AccountInvitation{
		FromUser:           *BuildFakeUser(),
		ToEmail:            fake.Email(),
		ToName:             buildUniqueString(),
		ToUser:             func(s string) *string { return &s }(buildUniqueString()),
		Note:               buildUniqueString(),
		StatusNote:         buildUniqueString(),
		Token:              fake.UUID(),
		DestinationAccount: *BuildFakeAccount(),
		ID:                 BuildFakeID(),
		ExpiresAt:          BuildFakeTime(),
		Status:             string(types.PendingAccountInvitationStatus),
		CreatedAt:          BuildFakeTime(),
	}
}

// BuildFakeAccountInvitationsList builds a faked AccountInvitationList.
func BuildFakeAccountInvitationsList() *filtering.QueryFilteredResult[types.AccountInvitation] {
	var examples []*types.AccountInvitation
	for range exampleQuantity {
		examples = append(examples, BuildFakeAccountInvitation())
	}

	return &filtering.QueryFilteredResult[types.AccountInvitation]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

func BuildFakeAccountInvitationUpdateRequestInput() *types.AccountInvitationUpdateRequestInput {
	return &types.AccountInvitationUpdateRequestInput{
		Token: BuildFakeID(),
		Note:  fake.Sentence(3),
	}
}

// BuildFakeAccountInvitationCreationRequestInput builds a faked AccountInvitationCreationRequestInput from a webhook.
func BuildFakeAccountInvitationCreationRequestInput() *types.AccountInvitationCreationRequestInput {
	invitation := BuildFakeAccountInvitation()
	return converters.ConvertAccountInvitationToAccountInvitationCreationInput(invitation)
}
