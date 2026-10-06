package catalog

import (
	"fmt"
	"maps"

	"github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v15/authentication/passkeys"
	"github.com/primandproper/platform-go/v15/billing"
	"github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/platform-go/v15/issuereports"
	"github.com/primandproper/platform-go/v15/mediaregistry"
	"github.com/primandproper/platform-go/v15/notifications"
	"github.com/primandproper/platform-go/v15/settings"
	"github.com/primandproper/platform-go/v15/waitlists"
	"github.com/primandproper/platform-go/v15/webhooks"
)

// Catalog returns every event type this application publishes: its own, from the generated
// definitions, and the ones platform publishes on its behalf, from each adopted package's
// EventCatalog fragment. It is the catalog the dispatcher and the emitter are built with.
//
// An event a subscriber may never receive is in the catalog and marked Internal rather than left
// out, so that "why does my webhook never fire" has something to read and the emitter's
// unsubscribable counter counts only constants that fell out of the catalog. Which events those
// are is excluded.go's decision.
//
// A collision between two fragments is a programming error — two packages naming one event — and
// it is caught by this package's tests rather than by a running process; the panic is for the
// process that somehow outran them.
//
// A fresh map each call, because the caller hands it to a dispatcher that retains it: a shared
// map would let one consumer's mutation change what every other consumer considers dispatchable.
func Catalog() webhooks.Catalog {
	merged, err := webhooks.Merge(
		local(),
		waitlists.EventCatalog(),
		settings.EventCatalog(),
		comments.EventCatalog(),
		issuereports.EventCatalog(),
		mediaregistry.EventCatalog(),
		notifications.EventCatalog(),
		billing.EventCatalog(),
		webhooks.EventCatalog(),
		oauth2clients.EventCatalog(),
		passkeys.EventCatalog(),
	)
	if err != nil {
		panic(fmt.Sprintf("composing the webhook catalog: %v", err))
	}

	for eventType := range merged {
		if Excluded(eventType.String()) {
			definition := merged[eventType]
			definition.Internal = true
			merged[eventType] = definition
		}
	}

	return merged
}

// local is this application's own fragment: the generated definitions, copied.
func local() webhooks.Catalog {
	out := make(webhooks.Catalog, len(definitions))
	maps.Copy(out, definitions)

	return out
}

// Known reports whether eventType may be delivered to a webhook: published, and not internal.
//
// The parameter is a plain string rather than a webhooks.EventType because the domains name their
// events with untyped constants and this is where the two vocabularies meet.
func Known(eventType string) bool {
	return Catalog().Subscribable(webhooks.EventType(eventType))
}

// Published reports whether eventType is one this application or platform emits at all,
// deliverable or not.
func Published(eventType string) bool {
	_, ok := Catalog()[webhooks.EventType(eventType)]

	return ok
}
