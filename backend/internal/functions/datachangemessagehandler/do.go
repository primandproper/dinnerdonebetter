package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
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

// RegisterAsyncDataChangeMessageHandler registers the async data change message handler with the injector.
func RegisterAsyncDataChangeMessageHandler(i do.Injector) {
	do.Provide[*AsyncDataChangeMessageHandler](i, func(i do.Injector) (*AsyncDataChangeMessageHandler, error) {
		return NewAsyncDataChangeMessageHandler(
			do.MustInvoke[context.Context](i),
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
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[*push.Fanout](i),
		)
	})
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
