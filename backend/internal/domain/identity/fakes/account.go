package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeAccount builds a faked account.
func BuildFakeAccount() *types.Account {
	account := fake.BuildFakeRecord[types.Account]()

	// An account that has not paid, which is the state a new one is in.
	account.BillingStatus = types.UnpaidAccountBillingStatus

	// An address that holds together — a city in its own state, a zip in its own city —
	// because the fields are read as one address by anything that formats or geocodes.
	fakeAddress := gofakeit.Address()
	account.AddressLine1 = fakeAddress.Address
	account.AddressLine2 = ""
	account.City = fakeAddress.City
	account.State = fakeAddress.State
	account.ZipCode = fakeAddress.Zip
	account.Country = fakeAddress.Country
	account.ContactPhone = gofakeit.PhoneFormatted()
	account.Latitude = pointer.To(fake.BuildFakeNumber())
	account.Longitude = pointer.To(fake.BuildFakeNumber())

	// An opaque value that round-trips as a string: nothing here decodes it, and the
	// secret a webhook is actually signed with is minted per webhook rather than here.
	account.WebhookEncryptionKey = fake.BuildFakeString()

	// Members of this account rather than of three unrelated ones.
	members := make([]*types.AccountUserMembershipWithUser, 0, fake.DefaultPageSize)
	for range fake.DefaultPageSize {
		membership := BuildFakeAccountUserMembershipWithUser()
		membership.BelongsToAccount = account.ID
		members = append(members, membership)
	}
	account.Members = members

	return account
}

// BuildFakeAccountsList builds a faked AccountList.
func BuildFakeAccountsList() *filtering.QueryFilteredResult[types.Account] {
	return fake.BuildFakePage(BuildFakeAccount)
}

// BuildFakeAccountOwnershipTransferInput builds a faked AccountOwnershipTransferInput.
func BuildFakeAccountOwnershipTransferInput() *types.AccountOwnershipTransferInput {
	input := fake.BuildFakeRecord[types.AccountOwnershipTransferInput]()

	// Two user IDs whose names say owner rather than user, which is past what a fake
	// built from field names can infer.
	input.CurrentOwner = fake.BuildFakeID()
	input.NewOwner = fake.BuildFakeID()

	return input
}

// BuildFakeAccountUpdateRequestInput builds a faked AccountUpdateRequestInput from an account.
func BuildFakeAccountUpdateRequestInput() *types.AccountUpdateRequestInput {
	account := BuildFakeAccount()

	return converters.ConvertAccountToAccountUpdateRequestInput(account)
}

// BuildFakeAccountCreationRequestInput builds a faked AccountCreationRequestInput.
func BuildFakeAccountCreationRequestInput() *types.AccountCreationRequestInput {
	account := BuildFakeAccount()

	return converters.ConvertAccountToAccountCreationRequestInput(account)
}
