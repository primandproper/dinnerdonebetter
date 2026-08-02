package http

import (
	paymentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/manager"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsHTTP registers the payments HTTP handler with the injector.
func RegisterPaymentsHTTP(i do.Injector) {
	do.ProvideValue(i, WebhookSignatureHeader(StripeSignatureHeader))

	do.Provide[*WebhookHandler](i, func(i do.Injector) (*WebhookHandler, error) {
		return NewWebhookHandler(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[paymentsmanager.PaymentsDataManager](i),
			do.MustInvoke[WebhookSignatureHeader](i),
		), nil
	})
}
