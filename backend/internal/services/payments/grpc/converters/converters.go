package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
)

func ConvertProductCreationRequestInputToDomain(input *paymentssvc.ProductCreationRequestInput) *payments.ProductCreationRequestInput {
	if input == nil {
		return nil
	}
	out := &payments.ProductCreationRequestInput{
		Name:              input.Name,
		Description:       input.Description,
		Kind:              input.Kind,
		AmountCents:       input.AmountCents,
		Currency:          input.Currency,
		ExternalProductID: input.ExternalProductId,
	}
	if input.BillingIntervalMonths != nil {
		out.BillingIntervalMonths = input.BillingIntervalMonths
	}
	return out
}

func ConvertProductCreationRequestInputToGRPC(input *payments.ProductCreationRequestInput) *paymentssvc.ProductCreationRequestInput {
	if input == nil {
		return nil
	}
	out := &paymentssvc.ProductCreationRequestInput{
		Name:              input.Name,
		Description:       input.Description,
		Kind:              input.Kind,
		AmountCents:       input.AmountCents,
		Currency:          input.Currency,
		ExternalProductId: input.ExternalProductID,
	}
	if input.BillingIntervalMonths != nil {
		out.BillingIntervalMonths = input.BillingIntervalMonths
	}
	return out
}

func ConvertSubscriptionCreationRequestInputToGRPC(input *payments.SubscriptionCreationRequestInput) *paymentssvc.SubscriptionCreationRequestInput {
	if input == nil {
		return nil
	}
	out := &paymentssvc.SubscriptionCreationRequestInput{
		BelongsToAccount:       input.BelongsToAccount,
		ProductId:              input.ProductID,
		ExternalSubscriptionId: input.ExternalSubscriptionID,
		Status:                 input.Status,
		CurrentPeriodStart:     grpcconverters.ConvertTimeToPBTimestamp(input.CurrentPeriodStart),
		CurrentPeriodEnd:       grpcconverters.ConvertTimeToPBTimestamp(input.CurrentPeriodEnd),
	}
	return out
}

func ConvertProductUpdateRequestInputToDomain(input *paymentssvc.ProductUpdateRequestInput) *payments.ProductUpdateRequestInput {
	if input == nil {
		return nil
	}
	return &payments.ProductUpdateRequestInput{
		Name:                  input.Name,
		Description:           input.Description,
		Kind:                  input.Kind,
		AmountCents:           input.AmountCents,
		Currency:              input.Currency,
		BillingIntervalMonths: input.BillingIntervalMonths,
		ExternalProductID:     input.ExternalProductId,
	}
}

func ConvertProductToGRPC(p *payments.Product) *paymentssvc.Product {
	if p == nil {
		return nil
	}
	return &paymentssvc.Product{
		Id:                    p.ID,
		Name:                  p.Name,
		Description:           p.Description,
		Kind:                  p.Kind,
		AmountCents:           p.AmountCents,
		Currency:              p.Currency,
		BillingIntervalMonths: p.BillingIntervalMonths,
		ExternalProductId:     p.ExternalProductID,
		CreatedAt:             grpcconverters.ConvertTimeToPBTimestamp(p.CreatedAt),
		LastUpdatedAt:         grpcconverters.ConvertTimePointerToPBTimestamp(p.LastUpdatedAt),
		ArchivedAt:            grpcconverters.ConvertTimePointerToPBTimestamp(p.ArchivedAt),
	}
}

func ConvertGRPCProductToProduct(p *paymentssvc.Product) *payments.Product {
	if p == nil {
		return nil
	}
	return &payments.Product{
		ID:                    p.Id,
		Name:                  p.Name,
		Description:           p.Description,
		Kind:                  p.Kind,
		AmountCents:           p.AmountCents,
		Currency:              p.Currency,
		BillingIntervalMonths: p.BillingIntervalMonths,
		ExternalProductID:     p.ExternalProductId,
		CreatedAt:             grpcconverters.ConvertPBTimestampToTime(p.CreatedAt),
		LastUpdatedAt:         grpcconverters.ConvertPBTimestampToTimePointer(p.LastUpdatedAt),
		ArchivedAt:            grpcconverters.ConvertPBTimestampToTimePointer(p.ArchivedAt),
	}
}

func ConvertGRPCSubscriptionToSubscription(s *paymentssvc.Subscription) *payments.Subscription {
	if s == nil {
		return nil
	}
	return &payments.Subscription{
		ID:                     s.Id,
		BelongsToAccount:       s.BelongsToAccount,
		ProductID:              s.ProductId,
		ExternalSubscriptionID: s.ExternalSubscriptionId,
		Status:                 s.Status,
		CurrentPeriodStart:     grpcconverters.ConvertPBTimestampToTime(s.CurrentPeriodStart),
		CurrentPeriodEnd:       grpcconverters.ConvertPBTimestampToTime(s.CurrentPeriodEnd),
		CreatedAt:              grpcconverters.ConvertPBTimestampToTime(s.CreatedAt),
		LastUpdatedAt:          grpcconverters.ConvertPBTimestampToTimePointer(s.LastUpdatedAt),
		ArchivedAt:             grpcconverters.ConvertPBTimestampToTimePointer(s.ArchivedAt),
	}
}

func ConvertSubscriptionCreationRequestInputToDomain(input *paymentssvc.SubscriptionCreationRequestInput) *payments.SubscriptionCreationRequestInput {
	if input == nil {
		return nil
	}
	return &payments.SubscriptionCreationRequestInput{
		BelongsToAccount:       input.BelongsToAccount,
		ProductID:              input.ProductId,
		ExternalSubscriptionID: input.ExternalSubscriptionId,
		Status:                 input.Status,
		CurrentPeriodStart:     grpcconverters.ConvertPBTimestampToTime(input.CurrentPeriodStart),
		CurrentPeriodEnd:       grpcconverters.ConvertPBTimestampToTime(input.CurrentPeriodEnd),
	}
}

func ConvertSubscriptionUpdateRequestInputToDomain(input *paymentssvc.SubscriptionUpdateRequestInput) *payments.SubscriptionUpdateRequestInput {
	if input == nil {
		return nil
	}
	return &payments.SubscriptionUpdateRequestInput{
		Status:             input.Status,
		CurrentPeriodStart: grpcconverters.ConvertPBTimestampToTimePointer(input.CurrentPeriodStart),
		CurrentPeriodEnd:   grpcconverters.ConvertPBTimestampToTimePointer(input.CurrentPeriodEnd),
	}
}

func ConvertSubscriptionToGRPC(s *payments.Subscription) *paymentssvc.Subscription {
	if s == nil {
		return nil
	}
	return &paymentssvc.Subscription{
		Id:                     s.ID,
		BelongsToAccount:       s.BelongsToAccount,
		ProductId:              s.ProductID,
		ExternalSubscriptionId: s.ExternalSubscriptionID,
		Status:                 s.Status,
		CurrentPeriodStart:     grpcconverters.ConvertTimeToPBTimestamp(s.CurrentPeriodStart),
		CurrentPeriodEnd:       grpcconverters.ConvertTimeToPBTimestamp(s.CurrentPeriodEnd),
		CreatedAt:              grpcconverters.ConvertTimeToPBTimestamp(s.CreatedAt),
		LastUpdatedAt:          grpcconverters.ConvertTimePointerToPBTimestamp(s.LastUpdatedAt),
		ArchivedAt:             grpcconverters.ConvertTimePointerToPBTimestamp(s.ArchivedAt),
	}
}

func ConvertPurchaseToGRPC(p *payments.Purchase) *paymentssvc.Purchase {
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

func ConvertPaymentTransactionToGRPC(t *payments.PaymentTransaction) *paymentssvc.PaymentTransaction {
	if t == nil {
		return nil
	}
	out := &paymentssvc.PaymentTransaction{
		Id:                    t.ID,
		BelongsToAccount:      t.BelongsToAccount,
		ExternalTransactionId: t.ExternalTransactionID,
		AmountCents:           t.AmountCents,
		Currency:              t.Currency,
		Status:                t.Status,
		CreatedAt:             grpcconverters.ConvertTimeToPBTimestamp(t.CreatedAt),
	}
	if t.SubscriptionID != nil {
		out.SubscriptionId = t.SubscriptionID
	}
	if t.PurchaseID != nil {
		out.PurchaseId = t.PurchaseID
	}
	return out
}
