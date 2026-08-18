package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/platform-go/v11/jobs"
)

// poolSpecs describes every topic this process drains: what to consume, with which knobs, and
// which handler. The topic comes from the queues config and the knobs from the pools config, so
// a topic name is never spelled twice.
//
// A spec's Config is copied by the group rather than retained, which is what lets the search
// specs below share one config value without all of them ending up on whichever topic was
// written last.
func (a *AsyncDataChangeMessageHandler) poolSpecs() []jobs.PoolSpec {
	specs := []jobs.PoolSpec{
		{
			Topic:   a.queuesConfig.DataChangesTopicName,
			Config:  &a.poolsConfig.DataChanges,
			Handler: a.DataChangesEventHandler(a.queuesConfig.DataChangesTopicName),
		},
		{
			Topic:   a.queuesConfig.OutboundEmailsTopicName,
			Config:  &a.poolsConfig.OutboundEmails,
			Handler: a.OutboundEmailsEventHandler(a.queuesConfig.OutboundEmailsTopicName),
		},
		// There is no webhook execution pool and no user data aggregation pool. A webhook
		// delivery is a dispatch row the delivery worker claims, and an export is a request
		// row the data privacy worker claims — neither is a message on a topic. That is what
		// lets one endpoint's copy of one event, or one subject's export, be the unit of
		// retry rather than the whole fan-out.
		{
			Topic:   a.queuesConfig.MobileNotificationsTopicName,
			Config:  &a.poolsConfig.MobileNotifications,
			Handler: a.MobileNotificationsEventHandler(a.queuesConfig.MobileNotificationsTopicName),
		},
	}

	// One pool per search index, rather than one pool over one topic that switched on an
	// index-type field. platform-go keys an index event by its topic, so the fan-out that used
	// to happen inside a handler happens here instead — which also means one index's backlog no
	// longer sits behind another's, and a poison message dead-letters for its own index alone.
	for _, syncer := range a.searchSyncers {
		specs = append(specs, jobs.PoolSpec{
			Topic:   syncer.Topic,
			Config:  &a.poolsConfig.SearchIndexRequests,
			Handler: syncer.Handle,
		})
	}

	return specs
}

// newPoolGroup assembles the group that drains every topic above.
//
// It is built at construction rather than at Start so that the reasons a group refuses to exist —
// a pool config that will not validate, two specs resolving to one topic, a handler that is nil —
// are reported by the thing that wires this process together, before anything is consuming. What
// is left for Start is the broker.
func newPoolGroup(ctx context.Context, a *AsyncDataChangeMessageHandler) (*jobs.PoolGroup, error) {
	return jobs.NewPoolGroup(
		ctx,
		a.poolSpecs(),
		a.consumerProvider,
		jobs.WithPoolGroupDeadLetter(a.deadLetter),
		jobs.WithPoolGroupLogger(a.logger),
		jobs.WithPoolGroupTracerProvider(a.tracerProvider),
		jobs.WithPoolGroupMetricsProvider(a.metricsProvider),
	)
}

// Start begins consuming every topic, or none of them.
//
// The handlers are unchanged — jobs.Handler has the same signature the event handler factories
// already returned — but a failure is retried with backoff and a message that exhausts its
// attempts is dead-lettered rather than logged and forgotten.
func (a *AsyncDataChangeMessageHandler) Start(ctx context.Context) error {
	return a.poolGroup.Start(ctx)
}

// Close stops every pool and waits for in-flight handlers to drain. The context bounds the wait;
// when it expires the pools cancel their workers and Close reports why.
func (a *AsyncDataChangeMessageHandler) Close(ctx context.Context) error {
	return a.poolGroup.Close(ctx)
}
