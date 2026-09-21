package authorization

import (
	billinggrpc "github.com/primandproper/platform-go/v14/billing/grpc"
)

// The billing permissions are platform's, re-exported under the names this
// application's policy already spells. See comments_permissions.go.
//
// The three list_all grants are new. They gate the cross-account reads an
// operator makes, which the deleted service expressed as an IsServiceAdmin()
// check inside a handler; as grants they are a line in the role grid instead.
//
// Two grants are gone with the RPCs they guarded. CreateSubscriptions and
// UpdateSubscriptions gated a client opening or moving a subscription, and a
// subscription is what a payment provider reports — the RevenueCat webhook is
// the door, and it is not on this surface. Nothing grants them because nothing
// asks.
const (
	// CreateProductsPermission is a permission.
	CreateProductsPermission = billinggrpc.PermissionCreateProducts
	// ReadProductsPermission is a permission.
	ReadProductsPermission = billinggrpc.PermissionReadProducts
	// UpdateProductsPermission is a permission.
	UpdateProductsPermission = billinggrpc.PermissionUpdateProducts
	// ArchiveProductsPermission is a permission.
	ArchiveProductsPermission = billinggrpc.PermissionArchiveProducts
	// ReadSubscriptionsPermission is a permission.
	ReadSubscriptionsPermission = billinggrpc.PermissionReadSubscriptions
	// ListAllSubscriptionsPermission gates the cross-account read.
	ListAllSubscriptionsPermission = billinggrpc.PermissionListAllSubscriptions
	// ArchiveSubscriptionsPermission is a permission.
	ArchiveSubscriptionsPermission = billinggrpc.PermissionArchiveSubscriptions
	// ReadPurchasesPermission is a permission.
	ReadPurchasesPermission = billinggrpc.PermissionReadPurchases
	// ListAllPurchasesPermission gates the cross-account read.
	ListAllPurchasesPermission = billinggrpc.PermissionListAllPurchases
	// ArchivePurchasesPermission is a permission.
	ArchivePurchasesPermission = billinggrpc.PermissionArchivePurchases
	// ReadTransactionsPermission is a permission.
	ReadTransactionsPermission = billinggrpc.PermissionReadTransactions
	// ListAllTransactionsPermission gates the cross-account read.
	ListAllTransactionsPermission = billinggrpc.PermissionListAllTransactions
	// ArchiveTransactionsPermission is a permission.
	ArchiveTransactionsPermission = billinggrpc.PermissionArchiveTransactions
)

// The checkout grants are this application's and stay so.
//
// They gate the capitalism handlers — the RevenueCat checkout and cancellation
// paths — which are HTTP and outside platform's billing surface either way.
// platform has no opinion about them because the provider integration is the
// consumer's.
const (
	// CreateCheckoutSessionPermission gates opening a checkout with the provider.
	CreateCheckoutSessionPermission Permission = "create.checkout_sessions"
	// CancelSubscriptionPermission gates asking the provider to cancel.
	CancelSubscriptionPermission Permission = "cancel.subscriptions"
)

// ReadPaymentHistoryPermission is the transaction read under the name this
// application's policy spells.
//
// The deleted GetPaymentHistoryForAccount is platform's
// ListTransactionsForAccount, and a payment history is a page of transactions,
// so the grant is the same one.
const ReadPaymentHistoryPermission = billinggrpc.PermissionReadTransactions
