// Package adapters holds the payment provider implementations of
// payments.PaymentProcessor. Stripe delegates to platform-go's capitalism package;
// RevenueCat, which capitalism does not model, is implemented here.
package adapters

import (
	platformerrors "github.com/primandproper/platform-go/v13/errors"
)

// maxWebhookBodyBytes bounds how much of a webhook request body an adapter reads. Provider
// event payloads are well under this, and it stops a hostile client from forcing an unbounded
// allocation on a public, unauthenticated endpoint. It matches the cap capitalism applies on
// the Stripe path.
const maxWebhookBodyBytes = 64 << 10 // 64 KiB

// ErrInvalidWebhookSignature indicates a webhook request failed signature verification.
var ErrInvalidWebhookSignature = platformerrors.New("invalid webhook signature")
