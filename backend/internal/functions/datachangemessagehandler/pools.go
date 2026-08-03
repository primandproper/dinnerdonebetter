package datachangemessagehandler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/observability"
)

// partialStartDrainTimeout bounds the teardown of pools that did start when a later one failed
// to build. Start is about to return an error and the process is about to exit, so this only
// needs to be long enough for in-flight handlers to finish, not for a graceful drain.
const partialStartDrainTimeout = 15 * time.Second

// poolSpec pairs a topic with the pool that drains it. The topic comes from the queues config
// and the knobs from the pools config, so a topic name is never spelled twice.
type poolSpec struct {
	handler jobs.Handler
	cfg     *jobs.PoolConfig
	topic   string
}

// Start builds a jobs.Pool per topic and begins consuming. Each pool owns its own workers,
// retry policy, and dead-letter path; Close stops them.
//
// This replaces the previous arrangement of six bare messagequeue.Consumers sharing a stop
// channel. The handlers are unchanged — jobs.Handler has the same signature the event handler
// factories already returned — but a failure is now retried with backoff and a message that
// exhausts its attempts is dead-lettered rather than logged and forgotten.
func (a *AsyncDataChangeMessageHandler) Start(ctx context.Context) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if len(a.pools) > 0 {
		return errPoolsAlreadyStarted
	}

	specs := []poolSpec{
		{
			topic:   a.queuesConfig.DataChangesTopicName,
			cfg:     &a.poolsConfig.DataChanges,
			handler: a.DataChangesEventHandler(a.queuesConfig.DataChangesTopicName),
		},
		{
			topic:   a.queuesConfig.OutboundEmailsTopicName,
			cfg:     &a.poolsConfig.OutboundEmails,
			handler: a.OutboundEmailsEventHandler(a.queuesConfig.OutboundEmailsTopicName),
		},
		{
			topic:   a.queuesConfig.SearchIndexRequestsTopicName,
			cfg:     &a.poolsConfig.SearchIndexRequests,
			handler: a.SearchIndexRequestsEventHandler(a.queuesConfig.SearchIndexRequestsTopicName),
		},
		// There is no webhook execution pool. Deliveries are dispatch rows the delivery
		// worker claims, not messages on a topic — which is what lets one endpoint's copy of
		// one event be the unit of retry, rather than the whole fan-out.
		{
			topic:   a.queuesConfig.UserDataAggregationTopicName,
			cfg:     &a.poolsConfig.UserDataAggregation,
			handler: a.UserDataAggregationEventHandler(a.queuesConfig.UserDataAggregationTopicName),
		},
		{
			topic:   a.queuesConfig.MobileNotificationsTopicName,
			cfg:     &a.poolsConfig.MobileNotifications,
			handler: a.MobileNotificationsEventHandler(a.queuesConfig.MobileNotificationsTopicName),
		},
	}

	for i := range specs {
		spec := &specs[i]

		// Topic is not part of the pools config; it lives in the queues config, which is
		// what the publishers on the other side of these topics already read.
		spec.cfg.Topic = spec.topic

		pool, err := jobs.NewPool(
			ctx,
			spec.cfg,
			a.consumerProvider,
			spec.handler,
			jobs.WithPoolDeadLetter(a.deadLetter),
			jobs.WithPoolLogger(a.logger),
			jobs.WithPoolTracerProvider(a.tracerProvider),
			jobs.WithPoolMetricsProvider(a.metricsProvider),
		)
		if err != nil {
			// Close whatever already started, so a failure partway through does not leave
			// half the topics being consumed by a process that is about to exit. The
			// deadline is here because this is a teardown on the error path, not the
			// operator-controlled drain that Close normally performs.
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), partialStartDrainTimeout)
			if closeErr := a.Close(closeCtx); closeErr != nil {
				a.logger.Error("closing partially started job pools", closeErr)
			}
			cancel()

			return observability.PrepareAndLogError(err, a.logger, span, "building job pool for topic %q", spec.topic)
		}

		// Run is started as soon as the pool is built rather than after the whole set is,
		// because Pool.Close waits on the goroutine Run owns: a pool that was constructed
		// but never run has nothing to signal that wait, and closing it would block until
		// its context expired.
		a.pools = append(a.pools, pool)
		a.poolsWG.Add(1)
		go func(p *jobs.Pool) {
			defer a.poolsWG.Done()
			// Run takes no context on purpose: tied to a server context it would stop
			// mid-message the instant that context was canceled. Close is the stop signal.
			p.Run()
		}(pool)
	}

	return nil
}

// Close stops every pool and waits for in-flight handlers to drain. The context bounds the
// wait; when it expires the pools cancel their workers and Close reports why.
func (a *AsyncDataChangeMessageHandler) Close(ctx context.Context) error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, pool := range a.pools {
		wg.Add(1)
		go func(p *jobs.Pool) {
			defer wg.Done()
			if err := p.Close(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(pool)
	}

	wg.Wait()
	a.poolsWG.Wait()
	a.pools = nil

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
