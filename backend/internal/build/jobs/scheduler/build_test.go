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
		"*github.com/primandproper/platform-go/v14/notifications/push.Fanout",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager.NotificationsDataManager",
		"github.com/primandproper/primitives-go/v2/notifications/mobile.PushNotificationSender",
		"*github.com/primandproper/platform-go/v14/workqueue.Config",
	} {
		assert.True(t, declared[name], "the scheduler must provide %s", name)
	}
}

// TestBuildInjector_RegistersEveryPrivacyCollectorStore names the stores the data privacy
// registry resolves that nothing else in this process touches.
//
// This is the container that fulfills a subject access request, so the registry it builds is
// the one that decides what an export contains. Three of its collectors read a store this
// process has no other use for — passkeys, password reset tokens, registered OAuth2 clients
// — and each arrived here from a different place: the passkey store with the identity store,
// the other two from Register lines added for this and nothing else.
//
// A store the registry wants and this container does not declare is not a compile error,
// because `do` resolves by type at runtime. Without this test the symptom is the privacy
// worker failing on its first request, in an environment, with a message naming a type
// rather than a missing Register line. The API server builds the same registry and has all
// three for its own reasons, so the gap only ever shows up here.
//
// Declarations rather than resolutions, for the reason the notification chain test gives:
// every one of these needs a database.Client.
func TestBuildInjector_RegistersEveryPrivacyCollectorStore(t *testing.T) {
	t.Parallel()

	i := BuildInjector(context.Background(), &config.SchedulerConfig{})

	declared := map[string]bool{}
	for _, service := range i.ListProvidedServices() {
		declared[service.Service] = true
	}

	for _, name := range []string{
		"github.com/primandproper/platform-go/v14/authentication/passkeys.Store",
		"github.com/primandproper/platform-go/v14/authentication/passwordreset.Store",
		"github.com/primandproper/platform-go/v14/authentication/oauth2clients.Store",
	} {
		assert.True(t, declared[name], "the data privacy registry resolves %s and this container does not declare it", name)
	}
}
