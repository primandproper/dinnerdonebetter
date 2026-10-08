package payments

import (
	"context"

	"github.com/primandproper/platform-go/v15/billing"
	billingsync "github.com/primandproper/platform-go/v15/billing/sync"
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// StripePlace places a Stripe subscription nobody holds yet.
//
// The account is the one holding the delivery's Stripe customer, which the checkout flow
// attached to it. The product is the catalog entry whose external id is the subscription's
// price, because that is what this catalog is keyed on for Stripe: a price is one way of
// paying for a Stripe product, and it is the price that decides what the account pays.
func StripePlace(products billing.ProductStore, accounts identity.DirectoryReader) billingsync.Place {
	return place(products, func(state *capitalism.SubscriptionState) string { return state.PriceID },
		func(ctx context.Context, q database.SQLQueryExecutor, scope tenancy.Scope, customerID string) (*identity.Account, error) {
			return accounts.GetAccountByPaymentProcessorCustomerID(ctx, q, scope, customerID)
		})
}

// RevenueCatPlace places a RevenueCat subscription nobody holds yet.
//
// RevenueCat's customer is the app_user_id, which the mobile app sets to this
// application's account id when it logs in to RevenueCat, so the account is read by that
// id rather than by a stored customer id. It is still read rather than trusted: a purchase
// made before the app logged in is filed under an anonymous RevenueCat id, and that has to
// be a refusal here rather than a foreign key violation three frames further down. The
// product is the catalog entry for the store product, since RevenueCat reports no price.
func RevenueCatPlace(products billing.ProductStore, accounts identity.DirectoryReader) billingsync.Place {
	return place(products, func(state *capitalism.SubscriptionState) string { return state.ProductID }, accounts.GetAccount)
}

// place is a billingsync.Place over this application's catalog and directory. The two
// providers differ only in which of the state's identifiers names the product and how
// their customer becomes an account.
//
// Every refusal here is an error, which billing/http answers with a 500 so the provider
// redelivers. That is the right answer to both: a customer this application does not know
// yet is usually a checkout whose own transaction has not committed, and a product nobody
// sells is an operator who has not added it to the catalog yet.
func place(
	products billing.ProductStore,
	externalProductID func(*capitalism.SubscriptionState) string,
	account func(ctx context.Context, q database.SQLQueryExecutor, scope tenancy.Scope, customerID string) (*identity.Account, error),
) billingsync.Place {
	return func(ctx context.Context, q database.SQLQueryExecutor, scope tenancy.Scope, state *capitalism.SubscriptionState) (*billingsync.Placement, error) {
		owner, err := account(ctx, q, scope, state.CustomerID)
		if err != nil {
			return nil, platformerrors.Wrapf(err, "reading the account for customer %q", state.CustomerID)
		}

		externalID := externalProductID(state)

		product, err := products.GetProductByExternalID(ctx, q, scope, externalID)
		if err != nil {
			return nil, platformerrors.Wrapf(err, "reading the product for %q", externalID)
		}

		return &billingsync.Placement{AccountID: owner.ID, ProductID: product.ID}, nil
	}
}
