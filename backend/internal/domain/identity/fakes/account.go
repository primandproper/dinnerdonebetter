package fakes

import (
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// fakeTimeZone is the zone every faked account is in.
//
// A named zone rather than a generated string, because the type validates one by loading
// it: a random value fails every write, and so does a plausible-looking typo.
const fakeTimeZone = "America/Chicago"

// BuildFakeAccount builds a faked account.
func BuildFakeAccount() *identity.Account {
	account := fake.BuildFakeRecord[identity.Account]()

	account.Scope = ddbidentity.Scope()
	account.TimeZone = fakeTimeZone

	// An account that has not paid, which is the state a new one is in. The processor's
	// customer identifier is empty for the same reason, and the plan and the sync stamp
	// are nil: all four are written by the billing sync rather than at creation.
	account.BillingStatus = identity.BillingUnpaid
	account.PaymentProcessorCustomerID = ""
	account.SubscriptionPlanID = nil
	account.LastPaymentProviderSyncedAt = nil

	// An address that holds together — a city in its own state, a zip in its own city —
	// because the fields are handed to a processor as one address.
	fakeAddress := gofakeit.Address()
	account.BillingAddress = identity.BillingAddress{
		Line1:      fakeAddress.Address,
		City:       fakeAddress.City,
		State:      fakeAddress.State,
		PostalCode: fakeAddress.Zip,
		Country:    fakeAddress.Country,
		Phone:      gofakeit.PhoneFormatted(),
	}

	return account
}

// BuildFakeAccountsList builds a faked page of accounts.
func BuildFakeAccountsList() *filtering.QueryFilteredResult[identity.Account] {
	return fake.BuildFakePage(BuildFakeAccount)
}

// BuildFakeAccountForUser builds a faked account owned by the given user.
func BuildFakeAccountForUser(userID string) *identity.Account {
	account := BuildFakeAccount()
	account.OwnerUserID = userID

	return account
}

// BuildFakeAccountUpdate builds a faked AccountUpdate.
//
// Built from an account rather than from field names, because the three fields here are
// the three an account holder may move and each has to be a value the account type
// accepts — a generated time zone is refused, and a generated address is not an address.
func BuildFakeAccountUpdate() *identity.AccountUpdate {
	account := BuildFakeAccount()
	address := account.BillingAddress

	return &identity.AccountUpdate{
		Name:           &account.Name,
		TimeZone:       &account.TimeZone,
		BillingAddress: &address,
	}
}
