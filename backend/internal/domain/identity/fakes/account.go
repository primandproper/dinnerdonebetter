package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"

	"github.com/primandproper/platform-go/v8/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeAccount builds a faked account.
func BuildFakeAccount() *types.Account {
	accountID := BuildFakeID()

	var memberships []*types.AccountUserMembershipWithUser
	for range exampleQuantity {
		membership := BuildFakeAccountUserMembershipWithUser()
		membership.BelongsToAccount = accountID
		memberships = append(memberships, membership)
	}

	fakeAddress := fake.Address()
	key := fake.BitcoinPrivateKey()

	return &types.Account{
		ID:                         accountID,
		Name:                       fake.UUID(),
		BillingStatus:              types.UnpaidAccountBillingStatus,
		ContactPhone:               fake.PhoneFormatted(),
		PaymentProcessorCustomerID: fake.UUID(),
		CreatedAt:                  BuildFakeTime(),
		BelongsToUser:              fake.UUID(),
		Members:                    memberships,
		AddressLine1:               fakeAddress.Address,
		AddressLine2:               "",
		City:                       fakeAddress.City,
		State:                      fakeAddress.State,
		ZipCode:                    fakeAddress.Zip,
		Country:                    fakeAddress.Country,
		Latitude:                   new(buildFakeNumber()),
		Longitude:                  new(buildFakeNumber()),
		WebhookEncryptionKey:       key,
	}
}

// BuildFakeAccountsList builds a faked AccountList.
func BuildFakeAccountsList() *filtering.QueryFilteredResult[types.Account] {
	var examples []*types.Account
	for range exampleQuantity {
		examples = append(examples, BuildFakeAccount())
	}

	return &filtering.QueryFilteredResult[types.Account]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

func BuildFakeAccountOwnershipTransferInput() *types.AccountOwnershipTransferInput {
	return &types.AccountOwnershipTransferInput{
		Reason:       fake.Sentence(5),
		CurrentOwner: BuildFakeID(),
		NewOwner:     BuildFakeID(),
	}
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
