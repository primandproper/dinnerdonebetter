package mealplantasknotifications

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/workqueue"
	workqueuecfg "github.com/primandproper/platform-go/v13/workqueue/config"

	"github.com/samber/do/v2"
)

// RegisterQueue registers this worker's work queue with the injector.
//
// It is registered separately from the Worker because a Queue owns a goroutine and has to be
// Closed, and the injector will not do it — see the shutdown in cmd/ddb/worker.go.
//
// The name is defaulted here rather than in the environment config, because workqueue.Config has
// no default for it: one table holds every logical queue, partitioned by name, so an unnamed
// queue would share rows with every other unnamed one. A deployment that sets its own name is
// honored — nothing else reads this queue, so there is no second process for a name to drift
// away from.
func RegisterQueue(i do.Injector) {
	do.Provide[*TaskQueue](i, func(i do.Injector) (*TaskQueue, error) {
		cfg := do.MustInvoke[*workqueue.Config](i)
		if cfg.Name == "" {
			cfg.Name = QueueName
		}

		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		queue, err := workqueuecfg.NewQueue[string](
			do.MustInvoke[context.Context](i),
			cfg,
			do.MustInvoke[database.Client](i),
			workqueuecfg.WithPillars(pillars),
		)
		if err != nil {
			return nil, err
		}

		return &TaskQueue{Queue: queue}, nil
	})
}

// RegisterWorker registers the meal plan task notification worker with the injector.
func RegisterWorker(i do.Injector) {
	do.Provide[*Worker](i, func(i do.Injector) (*Worker, error) {
		return NewWorker(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*TaskQueue](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[*push.Fanout](i),
		), nil
	})
}
