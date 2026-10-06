package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeWebhookEventType returns an event type from the real catalog.
//
// Not a random string: an event type outside the catalog is rejected at every boundary that
// takes one, so a fake that invented them would only ever exercise the rejection path.
func BuildFakeWebhookEventType() string {
	eventTypes := catalog.Catalog().EventTypes()

	return eventTypes[gofakeit.Number(0, len(eventTypes)-1)].String()
}
