// Package adapters holds the payment provider implementations of
// payments.PaymentProcessor. Both real providers delegate to platform-go's capitalism
// package — Stripe for web checkout, RevenueCat for mobile store purchases — and each
// adapter's whole job is turning the platform-owned event it hands back into one of our
// domain's ParsedWebhookEvents.
package adapters
