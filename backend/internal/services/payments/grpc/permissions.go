package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
)

type PaymentsMethodPermissions map[string][]authorization.Permission

func ProvideMethodPermissions() PaymentsMethodPermissions {
	return PaymentsMethodPermissions{
		paymentssvc.PaymentsService_CreateProduct_FullMethodName:               {authorization.CreateProductsPermission},
		paymentssvc.PaymentsService_GetProduct_FullMethodName:                  {authorization.ReadProductsPermission},
		paymentssvc.PaymentsService_GetProducts_FullMethodName:                 {authorization.ReadProductsPermission},
		paymentssvc.PaymentsService_UpdateProduct_FullMethodName:               {authorization.UpdateProductsPermission},
		paymentssvc.PaymentsService_ArchiveProduct_FullMethodName:              {authorization.ArchiveProductsPermission},
		paymentssvc.PaymentsService_CreateSubscription_FullMethodName:          {authorization.CreateSubscriptionsPermission},
		paymentssvc.PaymentsService_GetSubscription_FullMethodName:             {authorization.ReadSubscriptionsPermission},
		paymentssvc.PaymentsService_GetSubscriptionsForAccount_FullMethodName:  {authorization.ReadSubscriptionsPermission},
		paymentssvc.PaymentsService_UpdateSubscription_FullMethodName:          {authorization.UpdateSubscriptionsPermission},
		paymentssvc.PaymentsService_ArchiveSubscription_FullMethodName:         {authorization.ArchiveSubscriptionsPermission},
		paymentssvc.PaymentsService_GetPurchasesForAccount_FullMethodName:      {authorization.ReadPurchasesPermission},
		paymentssvc.PaymentsService_GetPaymentHistoryForAccount_FullMethodName: {authorization.ReadPaymentHistoryPermission},
	}
}
