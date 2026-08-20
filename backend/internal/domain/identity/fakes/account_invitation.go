package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeAccountInvitation builds a faked AccountInvitation.
func BuildFakeAccountInvitation() *types.AccountInvitation {
	invitation := fake.BuildFakeRecord[types.AccountInvitation]()

	// The invitation is delivered by email, and it is addressed to someone who may not
	// have an account yet — so the address is validated as one and the user is optional.
	invitation.ToEmail = gofakeit.Email()
	invitation.ToUser = pointer.To(fake.BuildFakeID())

	// A token the accept path looks up by, and the status every invitation starts in.
	invitation.Token = fake.BuildFakeString()
	invitation.Status = string(types.PendingAccountInvitationStatus)

	// The note explains a status the invitation does not have yet, and the creation input
	// carries neither — both are written when someone accepts or rejects.
	invitation.StatusNote = ""

	// The two whole records an invitation carries: who sent it, and what they are
	// inviting someone into. BuildFakeRecord fills a nested struct too, but with one
	// whose own constrained fields are random.
	invitation.FromUser = *BuildFakeUser()
	invitation.DestinationAccount = *BuildFakeAccount()

	return invitation
}

// BuildFakeAccountInvitationsList builds a faked AccountInvitationList.
func BuildFakeAccountInvitationsList() *filtering.QueryFilteredResult[types.AccountInvitation] {
	return fake.BuildFakePage(BuildFakeAccountInvitation)
}

// BuildFakeAccountInvitationUpdateRequestInput builds a faked AccountInvitationUpdateRequestInput.
func BuildFakeAccountInvitationUpdateRequestInput() *types.AccountInvitationUpdateRequestInput {
	input := fake.BuildFakeRecord[types.AccountInvitationUpdateRequestInput]()
	input.Token = fake.BuildFakeID()

	return input
}

// BuildFakeAccountInvitationCreationRequestInput builds a faked AccountInvitationCreationRequestInput.
func BuildFakeAccountInvitationCreationRequestInput() *types.AccountInvitationCreationRequestInput {
	invitation := BuildFakeAccountInvitation()

	return converters.ConvertAccountInvitationToAccountInvitationCreationRequestInput(invitation)
}
