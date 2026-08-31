package scheduler

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestBuildInjector_RegistersAllProviders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &config.SchedulerConfig{}

	i := BuildInjector(ctx, cfg)

	services := i.ListProvidedServices()
	assert.NotEmpty(t, services, "expected providers to be registered")
	assert.Greater(t, len(services), 10, "expected many providers to be registered")
}

// TestBuildInjector_RegistersTheNotificationChain names the providers the meal plan task
// notification job resolves, because a missing one is otherwise invisible until the scheduler
// boots in an environment.
//
// It asserts on declarations rather than resolving them: every link below needs a database.Client,
// so actually invoking the chain means a container, which belongs in the integration suite (see
// TestMealPlanTaskNotifications_Worker) rather than in a unit test of the wiring. What this
// catches is the realistic mistake — a Register line deleted, or a dependency added to one of
// these constructors without being registered here — which used to be a crash loop and is now a
// red test.
//
// The chain is longer than it looks because the job sends its own pushes rather than publishing
// them: the queue it claims from, the fan-out it delivers through, and everything the fan-out
// needs in turn.
func TestBuildInjector_RegistersTheNotificationChain(t *testing.T) {
	t.Parallel()

	i := BuildInjector(context.Background(), &config.SchedulerConfig{})

	declared := map[string]bool{}
	for _, service := range i.ListProvidedServices() {
		declared[service.Service] = true
	}

	for _, name := range []string{
		"*github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications.Worker",
		"*github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications.TaskQueue",
		"*github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push.Fanout",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager.NotificationsDataManager",
		"github.com/primandproper/platform-go/v13/notifications/mobile.PushNotificationSender",
		"*github.com/primandproper/platform-go/v13/workqueue.Config",
	} {
		assert.True(t, declared[name], "the scheduler must provide %s", name)
	}
}
