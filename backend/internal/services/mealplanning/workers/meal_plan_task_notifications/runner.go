package mealplantasknotifications

import (
	"context"
	"sync"

	"github.com/primandproper/platform-go/v15/service"

	"github.com/samber/do/v2"
)

// QueueRunner adapts the prep task reminder queue to the service.Runner its lifecycle is joined
// through.
//
// The queue is not a loop — the scheduled job drives it — but it batches enqueues on a goroutine
// of its own, and Close is what writes the last batch out. service.WithRunners is the one seam a
// service.Service offers an application's own components, and an application runner is closed
// first: before the scheduler whose job enqueues into it has drained. A reminder pass still
// running at that moment has its remaining enqueues refused, which costs nothing durable — the
// job finds every task still owed a reminder from the database on each pass, so a refused one is
// picked up by the next — but it is the wrong order, and the right one is a final-flush slot,
// after the loops and before the database, which service gives its own operations queue and
// does not offer an application. See platform-go#1148.
type QueueRunner struct {
	queue *TaskQueue
	stop  chan struct{}
	once  sync.Once
}

var _ service.Runner = (*QueueRunner)(nil)

// NewQueueRunner resolves the reminder queue from i and wraps it.
func NewQueueRunner(i do.Injector) (*QueueRunner, error) {
	queue, err := do.Invoke[*TaskQueue](i)
	if err != nil {
		return nil, err
	}

	return &QueueRunner{queue: queue, stop: make(chan struct{})}, nil
}

// Run blocks until Close.
func (q *QueueRunner) Run() {
	<-q.stop
}

// Close writes out the queue's last batch and stops its goroutine.
func (q *QueueRunner) Close(ctx context.Context) error {
	q.once.Do(func() { close(q.stop) })

	return q.queue.Close(ctx)
}
