package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/notifications/push"
	"github.com/primandproper/primitives-go/v2/analytics"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/email"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAsyncDataChangeMessageHandler registers the async data change message handler with
// the injector, over the outbound notifiers handed in. ctx is the process's own, which the
// handler's publishers and dead-letter topic are opened under.
//
// The notifiers are named by the process wiring itself rather than discovered, for the reason
// searchindexes.Register gives for its registrars: a domain whose notifier is missing has events
// that imply mail nobody sends, and the only symptom is an inbox that stays empty.
func RegisterAsyncDataChangeMessageHandler(ctx context.Context, i do.Injector, notifiers ...OutboundNotifier) {
	do.Provide[*AsyncDataChangeMessageHandler](i, func(i do.Injector) (*AsyncDataChangeMessageHandler, error) {
		handlers, err := outboundNotificationHandlers(i, notifiers)
		if err != nil {
			return nil, err
		}

		return NewAsyncDataChangeMessageHandler(
			ctx,
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*config.AsyncMessageHandlerConfig](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[internalops.InternalOpsDataManager](i),
			do.MustInvoke[messagequeue.ConsumerProvider](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[analytics.EventReporter](i),
			do.MustInvoke[email.Emailer](i),
			do.MustInvoke[metrics.Provider](i),
			searchSyncers(i),
			handlers,
			do.MustInvoke[*push.Fanout](i),
		)
	})
}

// outboundNotificationHandlers resolves each domain's handler, in the order the process listed
// them. A domain whose handler cannot be built fails the handler's construction rather than its
// first event.
func outboundNotificationHandlers(i do.Injector, notifiers []OutboundNotifier) ([]OutboundNotificationHandler, error) {
	handlers := make([]OutboundNotificationHandler, 0, len(notifiers))

	for _, notifier := range notifiers {
		handler, err := notifier(i)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, handler)
	}

	return handlers, nil
}

// searchSyncers is every registered index's Syncer, paired with the topic its events arrive on,
// read off the one Registry every index is built through — see internal/searchindexes. Which
// indexes those are is decided where the Registry is registered, not here.
func searchSyncers(i do.Injector) []SearchSyncer {
	specs := do.MustInvoke[*searchindexes.Registry](i).PoolSpecs()

	syncers := make([]SearchSyncer, 0, len(specs))
	for i := range specs {
		syncers = append(syncers, SearchSyncer{Topic: specs[i].Topic, Handle: specs[i].Handler})
	}

	return syncers
}
