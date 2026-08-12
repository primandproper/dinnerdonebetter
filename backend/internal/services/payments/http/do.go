package http

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsHTTP registers the payments HTTP handler with the injector.
func RegisterPaymentsHTTP(i do.Injector) {
	do.Provide[*WebhookHandler](i, func(i do.Injector) (*WebhookHandler, error) {
		return NewWebhookHandler(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[paymentsmanager.PaymentsDataManager](i),
			do.MustInvoke[payments.PaymentProcessorRegistry](i),
		), nil
	})
}
