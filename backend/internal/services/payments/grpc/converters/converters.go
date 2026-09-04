// Package converters renders the platform's billing types on the wire and reads
// the wire's inputs back into them.
package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/capitalism"
)

// ConvertProductToGRPC renders a product.
func ConvertProductToGRPC(p *billing.Product) *paymentssvc.Product {
	if p == nil {
		return nil
	}

	return &paymentssvc.Product{
		Id:                    p.ID,
		Name:                  p.Name,
		Description:           p.Description,
		Kind:                  p.Kind.String(),
		AmountCents:           p.AmountCents,
		Currency:              p.Currency,
		BillingIntervalMonths: p.BillingIntervalMonths,
		ExternalProductId:     p.ExternalProductID,
		CreatedAt:             grpcconverters.ConvertTimeToPBTimestamp(p.CreatedAt),
		LastUpdatedAt:         grpcconverters.ConvertTimePointerToPBTimestamp(p.LastUpdatedAt),
		ArchivedAt:            grpcconverters.ConvertTimePointerToPBTimestamp(p.ArchivedAt),
	}
}

// ConvertGRPCProductToProduct reads a product back off the wire.
//
// The scope is this application's, because the wire does not carry one: every
// product a client can see is in the global catalog. See payments.Scope.
func ConvertGRPCProductToProduct(p *paymentssvc.Product) *billing.Product {
	if p == nil {
		return nil
	}

	return &billing.Product{
		ID:                    p.GetId(),
		Name:                  p.GetName(),
		Description:           p.GetDescription(),
		Kind:                  billing.Kind(p.GetKind()),
		AmountCents:           p.GetAmountCents(),
		Currency:              p.GetCurrency(),
		BillingIntervalMonths: p.GetBillingIntervalMonths(),
		ExternalProductID:     p.GetExternalProductId(),
		Scope:                 payments.Scope(),
		CreatedAt:             grpcconverters.ConvertPBTimestampToTime(p.GetCreatedAt()),
		LastUpdatedAt:         grpcconverters.ConvertPBTimestampToTimePointer(p.GetLastUpdatedAt()),
		ArchivedAt:            grpcconverters.ConvertPBTimestampToTimePointer(p.GetArchivedAt()),
	}
}

// ConvertGRPCProductCreationRequestInputToProduct reads a creation request as the
// product it asks for. The store mints the id and stamps the creation time.
func ConvertGRPCProductCreationRequestInputToProduct(input *paymentssvc.ProductCreationRequestInput) *billing.Product {
	if input == nil {
		return nil
	}

	return &billing.Product{
		Name:                  input.GetName(),
		Description:           input.GetDescription(),
		Kind:                  billing.Kind(input.GetKind()),
		AmountCents:           input.GetAmountCents(),
		Currency:              input.GetCurrency(),
		BillingIntervalMonths: input.GetBillingIntervalMonths(),
		ExternalProductID:     input.GetExternalProductId(),
		Scope:                 payments.Scope(),
	}
}

// ConvertProductToGRPCProductCreationRequestInput renders a product as the request
// that would create it, which is what a test that builds a fake and sends it needs.
func ConvertProductToGRPCProductCreationRequestInput(p *billing.Product) *paymentssvc.ProductCreationRequestInput {
	if p == nil {
		return nil
	}

	return &paymentssvc.ProductCreationRequestInput{
		Name:                  p.Name,
		Description:           p.Description,
		Kind:                  p.Kind.String(),
		AmountCents:           p.AmountCents,
		Currency:              p.Currency,
		BillingIntervalMonths: p.BillingIntervalMonths,
		ExternalProductId:     p.ExternalProductID,
	}
}

// ApplyGRPCProductUpdateRequestInput applies a patch to a product: every field the
// request set is written, and every field it left unset is kept.
func ApplyGRPCProductUpdateRequestInput(product *billing.Product, input *paymentssvc.ProductUpdateRequestInput) {
	if product == nil || input == nil {
		return
	}

	if input.Name != nil {
		product.Name = input.GetName()
	}

	if input.Description != nil {
		product.Description = input.GetDescription()
	}

	if input.Kind != nil {
		product.Kind = billing.Kind(input.GetKind())
	}

	if input.AmountCents != nil {
		product.AmountCents = input.GetAmountCents()
	}

	if input.Currency != nil {
		product.Currency = input.GetCurrency()
	}

	if input.BillingIntervalMonths != nil {
		product.BillingIntervalMonths = input.GetBillingIntervalMonths()
	}

	if input.ExternalProductId != nil {
		product.ExternalProductID = input.GetExternalProductId()
	}
}

// ConvertSubscriptionToGRPC renders a subscription.
func ConvertSubscriptionToGRPC(s *billing.Subscription) *paymentssvc.Subscription {
	if s == nil {
		return nil
	}

	return &paymentssvc.Subscription{
		Id:                     s.ID,
		BelongsToAccount:       s.BelongsToAccount,
		ProductId:              s.ProductID,
		ExternalSubscriptionId: s.ExternalSubscriptionID,
		Status:                 string(s.Status),
		CurrentPeriodStart:     grpcconverters.ConvertTimeToPBTimestamp(s.CurrentPeriodStart),
		CurrentPeriodEnd:       grpcconverters.ConvertTimeToPBTimestamp(s.CurrentPeriodEnd),
		CreatedAt:              grpcconverters.ConvertTimeToPBTimestamp(s.CreatedAt),
		LastUpdatedAt:          grpcconverters.ConvertTimePointerToPBTimestamp(s.LastUpdatedAt),
		ArchivedAt:             grpcconverters.ConvertTimePointerToPBTimestamp(s.ArchivedAt),
	}
}

// ConvertGRPCSubscriptionToSubscription reads a subscription back off the wire.
func ConvertGRPCSubscriptionToSubscription(s *paymentssvc.Subscription) *billing.Subscription {
	if s == nil {
		return nil
	}

	return &billing.Subscription{
		ID:                     s.GetId(),
		BelongsToAccount:       s.GetBelongsToAccount(),
		ProductID:              s.GetProductId(),
		ExternalSubscriptionID: s.GetExternalSubscriptionId(),
		Status:                 capitalism.SubscriptionStatus(s.GetStatus()),
		Scope:                  payments.Scope(),
		CurrentPeriodStart:     grpcconverters.ConvertPBTimestampToTime(s.GetCurrentPeriodStart()),
		CurrentPeriodEnd:       grpcconverters.ConvertPBTimestampToTime(s.GetCurrentPeriodEnd()),
		CreatedAt:              grpcconverters.ConvertPBTimestampToTime(s.GetCreatedAt()),
		LastUpdatedAt:          grpcconverters.ConvertPBTimestampToTimePointer(s.GetLastUpdatedAt()),
		ArchivedAt:             grpcconverters.ConvertPBTimestampToTimePointer(s.GetArchivedAt()),
	}
}

// ConvertGRPCSubscriptionCreationRequestInputToSubscription reads a creation
// request as the subscription it asks for.
func ConvertGRPCSubscriptionCreationRequestInputToSubscription(input *paymentssvc.SubscriptionCreationRequestInput) *billing.Subscription {
	if input == nil {
		return nil
	}

	return &billing.Subscription{
		BelongsToAccount:       input.GetBelongsToAccount(),
		ProductID:              input.GetProductId(),
		ExternalSubscriptionID: input.GetExternalSubscriptionId(),
		Status:                 capitalism.SubscriptionStatus(input.GetStatus()),
		Scope:                  payments.Scope(),
		CurrentPeriodStart:     grpcconverters.ConvertPBTimestampToTime(input.GetCurrentPeriodStart()),
		CurrentPeriodEnd:       grpcconverters.ConvertPBTimestampToTime(input.GetCurrentPeriodEnd()),
	}
}

// ConvertSubscriptionToGRPCSubscriptionCreationRequestInput renders a subscription
// as the request that would create it.
func ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(s *billing.Subscription) *paymentssvc.SubscriptionCreationRequestInput {
	if s == nil {
		return nil
	}

	return &paymentssvc.SubscriptionCreationRequestInput{
		BelongsToAccount:       s.BelongsToAccount,
		ProductId:              s.ProductID,
		ExternalSubscriptionId: s.ExternalSubscriptionID,
		Status:                 string(s.Status),
		CurrentPeriodStart:     grpcconverters.ConvertTimeToPBTimestamp(s.CurrentPeriodStart),
		CurrentPeriodEnd:       grpcconverters.ConvertTimeToPBTimestamp(s.CurrentPeriodEnd),
	}
}

// ApplyGRPCSubscriptionUpdateRequestInput applies a patch to a subscription.
//
// The account is not among the fields it can touch, and the store would refuse
// it anyway: moving a subscription between accounts is not an edit.
func ApplyGRPCSubscriptionUpdateRequestInput(subscription *billing.Subscription, input *paymentssvc.SubscriptionUpdateRequestInput) {
	if subscription == nil || input == nil {
		return
	}

	if input.Status != nil {
		subscription.Status = capitalism.SubscriptionStatus(input.GetStatus())
	}

	if input.ProductId != nil {
		subscription.ProductID = input.GetProductId()
	}

	if input.CurrentPeriodStart != nil {
		subscription.CurrentPeriodStart = grpcconverters.ConvertPBTimestampToTime(input.GetCurrentPeriodStart())
	}

	if input.CurrentPeriodEnd != nil {
		subscription.CurrentPeriodEnd = grpcconverters.ConvertPBTimestampToTime(input.GetCurrentPeriodEnd())
	}
}

// ConvertPurchaseToGRPC renders a purchase.
func ConvertPurchaseToGRPC(p *billing.Purchase) *paymentssvc.Purchase {
	if p == nil {
		return nil
	}

	return &paymentssvc.Purchase{
		Id:                    p.ID,
		BelongsToAccount:      p.BelongsToAccount,
		ProductId:             p.ProductID,
		AmountCents:           p.AmountCents,
		Currency:              p.Currency,
		CompletedAt:           grpcconverters.ConvertTimePointerToPBTimestamp(p.CompletedAt),
		ExternalTransactionId: p.ExternalTransactionID,
		CreatedAt:             grpcconverters.ConvertTimeToPBTimestamp(p.CreatedAt),
		LastUpdatedAt:         grpcconverters.ConvertTimePointerToPBTimestamp(p.LastUpdatedAt),
		ArchivedAt:            grpcconverters.ConvertTimePointerToPBTimestamp(p.ArchivedAt),
	}
}

// ConvertTransactionToGRPC renders a ledger row.
func ConvertTransactionToGRPC(t *billing.Transaction) *paymentssvc.PaymentTransaction {
	if t == nil {
		return nil
	}

	return &paymentssvc.PaymentTransaction{
		Id:                    t.ID,
		BelongsToAccount:      t.BelongsToAccount,
		SubscriptionId:        t.SubscriptionID,
		PurchaseId:            t.PurchaseID,
		ExternalTransactionId: t.ExternalTransactionID,
		AmountCents:           t.AmountCents,
		Currency:              t.Currency,
		Status:                t.Status.String(),
		CreatedAt:             grpcconverters.ConvertTimeToPBTimestamp(t.CreatedAt),
		LastUpdatedAt:         grpcconverters.ConvertTimePointerToPBTimestamp(t.LastUpdatedAt),
		ArchivedAt:            grpcconverters.ConvertTimePointerToPBTimestamp(t.ArchivedAt),
	}
}
