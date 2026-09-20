package fakes

import (
	"time"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// fakeInvitationLifetime is how long a faked invitation has left to run.
//
// Far enough out that a test which creates one and answers it several seconds later is
// answering a live invitation, and finite because the type requires an expiry: a link that
// never expires is a bearer credential nobody can retire.
const fakeInvitationLifetime = time.Hour

// BuildFakeInvitation builds a faked Invitation.
func BuildFakeInvitation() *identity.Invitation {
	invitation := fake.BuildFakeRecord[identity.Invitation]()

	invitation.Scope = ddbidentity.Scope()
	invitation.ExpiresAt = time.Now().Add(fakeInvitationLifetime).UTC()

	// The invitation is addressed to an email rather than to a user, because the common
	// case is inviting somebody who has not registered. ToUser is filled in when somebody
	// accepts, so a fresh invitation names nobody.
	invitation.ToEmail = identity.FoldHandle(gofakeit.Email())
	invitation.ToName = gofakeit.Name()
	invitation.ToUser = nil

	// A token the accept path looks up by, and the status every invitation starts in.
	invitation.Token = fake.BuildFakeString()
	invitation.Status = identity.InvitationPending

	// The note explains a status the invitation does not have yet; it is written when
	// somebody answers.
	invitation.StatusNote = ""

	return invitation
}

// BuildFakeInvitationsList builds a faked page of invitations.
func BuildFakeInvitationsList() *filtering.QueryFilteredResult[identity.Invitation] {
	return fake.BuildFakePage(BuildFakeInvitation)
}

// BuildFakeInvitationFromUserToAccount builds a faked Invitation with its two ends named.
func BuildFakeInvitationFromUserToAccount(fromUserID, accountID string) *identity.Invitation {
	invitation := BuildFakeInvitation()
	invitation.FromUser = fromUserID
	invitation.BelongsToAccount = accountID

	return invitation
}
