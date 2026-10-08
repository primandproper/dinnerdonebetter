package scheduler

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config/environments"

	operationscfg "github.com/primandproper/platform-go/v15/operations/config"
	sagacfg "github.com/primandproper/platform-go/v15/saga/config"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schedulerConfig returns the scheduler configuration the integration test environment ships.
// Each call derives a copy of its own, because composing validates and pins it in place.
func schedulerConfig() *config.SchedulerConfig {
	return (&config.EnvironmentConfigSet{RootConfig: environments.BuildIntegrationTestsConfig()}).Derive().Scheduler
}

// buildInjector composes the scheduler from the shipped configuration.
func buildInjector(t *testing.T) *do.RootScope {
	t.Helper()

	i, err := BuildInjector(t.Context(), schedulerConfig())
	require.NoError(t, err)

	return i
}

func TestBuildInjector_RegistersAllProviders(t *testing.T) {
	t.Parallel()

	i := buildInjector(t)

	services := i.ListProvidedServices()
	assert.NotEmpty(t, services, "expected providers to be registered")
	assert.Greater(t, len(services), 10, "expected many providers to be registered")
}

// TestBuildInjector_RegistersThePlatformReapers names the job sets platform's stores register for
// themselves, which service.New hands the scheduler beside this application's own.
//
// Operations' is the one this process went without: recovery re-offers an operation whose worker
// died between its insert and its enqueue, and before this process was composed from a
// service.Config nothing scheduled it, so such an operation sat pending forever. Declared rather
// than resolved for the reason the notification chain test gives; the integration suite's wiring
// test resolves them and asserts the scheduler holds them.
func TestBuildInjector_RegistersThePlatformReapers(t *testing.T) {
	t.Parallel()

	declared := map[string]bool{}
	for _, service := range buildInjector(t).ListProvidedServices() {
		declared[service.Service] = true
	}

	for _, key := range []string{operationscfg.JobsKey, sagacfg.JobsKey} {
		assert.True(t, declared[key], "the scheduler must register %s", key)
	}
}

func TestBuildInjector_RefusesASchedulerWithoutOperations(t *testing.T) {
	t.Parallel()

	cfg := schedulerConfig()
	cfg.Service.Operations = nil

	_, err := BuildInjector(t.Context(), cfg)
	assert.Error(t, err)
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

	i := buildInjector(t)

	declared := map[string]bool{}
	for _, service := range i.ListProvidedServices() {
		declared[service.Service] = true
	}

	for _, name := range []string{
		"*github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications.Worker",
		"*github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications.TaskQueue",
		"*github.com/primandproper/platform-go/v15/notifications/push.Fanout",
		"github.com/primandproper/platform-go/v15/notifications.Inbox",
		"github.com/primandproper/platform-go/v15/notifications.Registry",
		"github.com/primandproper/primitives-go/v2/notifications/mobile.PushNotificationSender",
		"*github.com/primandproper/platform-go/v15/workqueue.Config",
	} {
		assert.True(t, declared[name], "the scheduler must provide %s", name)
	}
}

// TestBuildInjector_RegistersEverySearcherTheReindexJobResolves names the text index clients
// the search index Registry resolves, one per index, because this container went without all
// nine of them.
//
// The reindex job resolves the Registry from inside its closure at tick time, and the Registry
// resolves each domain's index clients as it registers the indexes. Nothing resolved any of
// that at boot, so a scheduler composed without the clients started, reported healthy, and
// failed the reindex job on its first tick — which is what this process did until each domain's
// registration began registering the clients beside the indexes they serve.
//
// Declarations rather than resolutions, for the reason the notification chain test gives.
func TestBuildInjector_RegistersEverySearcherTheReindexJobResolves(t *testing.T) {
	t.Parallel()

	i := buildInjector(t)

	declared := map[string]bool{}
	for _, service := range i.ListProvidedServices() {
		declared[service.Service] = true
	}

	for _, name := range []string{
		"github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing.UserTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.RecipeTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.MealTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.ValidIngredientTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.ValidInstrumentTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.ValidMeasurementUnitTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.ValidPreparationTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.ValidIngredientStateTextSearcher",
		"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex.ValidVesselTextSearcher",
	} {
		assert.True(t, declared[name], "the search index Registry resolves %s and this container does not declare it", name)
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

	i := buildInjector(t)

	declared := map[string]bool{}
	for _, service := range i.ListProvidedServices() {
		declared[service.Service] = true
	}

	for _, name := range []string{
		"github.com/primandproper/platform-go/v15/authentication/passkeys.Store",
		"github.com/primandproper/platform-go/v15/authentication/passwordreset.Store",
		"github.com/primandproper/platform-go/v15/authentication/oauth2clients.Store",
	} {
		assert.True(t, declared[name], "the data privacy registry resolves %s and this container does not declare it", name)
	}
}
