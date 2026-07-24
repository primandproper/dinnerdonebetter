package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	paymentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
)

// ConvertUserDataCollectionToGRPCDataCollection converts a domain payments UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *payments.UserDataCollection) *paymentssvc.DataCollection {
	result := &paymentssvc.DataCollection{}

	for i := range input.Subscriptions {
		result.Subscriptions = append(result.Subscriptions, ConvertSubscriptionToGRPC(&input.Subscriptions[i]))
	}

	for i := range input.Purchases {
		result.Purchases = append(result.Purchases, ConvertPurchaseToGRPC(&input.Purchases[i]))
	}

	for i := range input.PaymentTransactions {
		result.PaymentTransactions = append(result.PaymentTransactions, ConvertPaymentTransactionToGRPC(&input.PaymentTransactions[i]))
	}

	return result
}
