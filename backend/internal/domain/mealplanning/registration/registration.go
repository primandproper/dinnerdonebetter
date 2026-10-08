/*
Package registration is the meal planning domain's contribution to the composition root.

Every process this application runs is composed in internal/build, and each builder takes what
it needs from a domain as an entry in a list it ranges over: the gRPC surface to mount and the
permission table its methods ship with, the comment targets it accepts, the scheduled jobs and
the runners it owns, the search indexes it keeps, the mail its events imply, the tools it offers
over MCP, and the components the process registers for it. This package is where the meal
planning entries come from. The builders name no other mealplanning package, so swapping the
domain is replacing this package and the list entries that name it.

Two of the domain's entries are not functions here, because the list that merges them is not a
builder: the search index rules (searchindex.IndexRules, merged by internal/indexevents) and the
analytics allowlist (mealplanning.AnalyticsEventTypes, merged by internal/domain/analytics)
live beside what they describe, and those two packages carry the marker instead.

Every site outside the three mealplanning roots that names this domain carries a
"// Domain: mealplanning" marker, and TestDomainMarkerCensus holds the composition root to it:
the marker census is the complete edit list. It is slices and a marker rather than a plugin
system or a CRUD kit, on purpose — see docs/adding_a_new_domain.md "Should This Be Generic?".
*/
package registration

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	grocerylistpreparation "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	recipeanalysis "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"
	mealplanningsvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"
	"github.com/primandproper/dinnerdonebetter/backend/internal/recordingspine"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanningmcp "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/mcp"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/outbound"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	platformcomments "github.com/primandproper/platform-go/v15/comments"
	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/platform-go/v15/service"
	"github.com/primandproper/platform-go/v15/workqueue"
	"github.com/primandproper/primitives-go/v2/jobs"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"
	platformgrpc "github.com/primandproper/primitives-go/v2/server/grpc"

	"github.com/samber/do/v2"
	"google.golang.org/grpc"
)

// Job names. These are also the distributed lock keys, so renaming one lets an old replica and
// a new replica both run that job during a rollout.
const (
	jobMealPlanFinalizationStarter = "meal_plan_finalization_starter"
	jobMealPlanTaskNotifications   = "meal_plan_task_notifications"
)

func registerRepository(i do.Injector) {
	// The repository writes its data change events into the outbox as part of the same
	// transaction as the row they describe; see internal/recordingspine. Register is also
	// where the recording spine is assembled — the outbox writer, the webhook emitter, and the
	// recorder over them — for the reason config.SchedulerConfig.OutboxRelay gives.
	recordingspine.Register(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
}

// RegisterForGRPCAPI registers all mealplanning components needed by the gRPC API server.
func RegisterForGRPCAPI(i do.Injector) {
	registerRepository(i)
	mealplanningmgr.RegisterManagers(i)
	mealplanningsvc.RegisterMealPlanningService(i)
	// The API only ever starts a finalization saga — the admin RPC that used to run the three
	// pipeline jobs on demand now runs this. Advancing belongs to the scheduler's saga worker,
	// which is what keeps a durable process from being tied to the lifetime of a request.
	mealplanfinalization.RegisterStarter(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
}

// MountGRPC is the registration the API server mounts the meal planning surface with. It
// resolves the service when called, so a surface that cannot be built fails the server's
// assembly rather than its first request.
func MountGRPC(i do.Injector) platformgrpc.RegistrationFunc {
	impl := do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i)

	return func(server *grpc.Server) {
		mealplanningsvcpb.RegisterMealPlanningServiceServer(server, impl)
	}
}

// GRPCPermissions is the permission table the meal planning surface's methods ship with.
func GRPCPermissions() map[string][]authorization.Permission {
	return mealplanningsvc.ProvideMethodPermissions()
}

// RegisterForDataChangeHandler registers mealplanning components needed by the async message
// handler: the repository, the text index clients behind the indexes it keeps current, and the
// notifier behind this domain's entry in its outbound notification list. The indexes themselves
// are registered into the one Registry every domain's go through, by RegisterIndexes — see
// internal/searchindexes — and the notifier arrives through OutboundNotifications.
func RegisterForDataChangeHandler(i do.Injector) {
	registerRepository(i)
	mealplanningindexing.RegisterSearchers(i)
	outbound.Register(i)
}

// OutboundNotifications is this domain's entry in the data change handler's list of outbound
// notification handlers: the mail its events imply. See internal/services/mealplanning/outbound
// for which events those are.
func OutboundNotifications(i do.Injector) (datachangemessagehandler.OutboundNotificationHandler, error) {
	notifier, err := do.Invoke[*outbound.Notifier](i)
	if err != nil {
		return nil, err
	}

	return notifier.Handle, nil
}

// RegisterForMCP registers mealplanning components needed by the MCP server: the repository
// its tools read through. The tools arrive through MCPTools.
func RegisterForMCP(i do.Injector) {
	registerRepository(i)
}

// MCPTools is this domain's tool surface, the MCP server's one entry for it.
func MCPTools(i do.Injector) (mcptools.Toolset, error) {
	repo, err := do.Invoke[mealplanning.Repository](i)
	if err != nil {
		return nil, err
	}

	return mealplanningmcp.NewTools(repo)
}

// RegisterForSearchIndexInitializer registers mealplanning components needed by a process that
// only rebuilds search indexes: the repository the index sources read from and the text index
// clients they write to. The indexes themselves arrive through RegisterIndexes.
func RegisterForSearchIndexInitializer(i do.Injector) {
	registerRepository(i)
	mealplanningindexing.RegisterSearchers(i)
}

// RegisterForScheduler registers mealplanning components needed by the scheduler: the
// repository and the text index clients the reindex job rebuilds through, the analyzer and the
// grocery list creator the finalization saga's steps run, the saga starter, and the prep task
// reminder queue with the worker that fills and drains it.
func RegisterForScheduler(i do.Injector) {
	registerRepository(i)
	mealplanningindexing.RegisterSearchers(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
	mealplanfinalization.RegisterStarter(i)

	// The one work queue the scheduler runs. It is provided as the bare *workqueue.Config the
	// platform's constructor takes, because there is exactly one — a second would need a name
	// to tell them apart in the container, which is the point at which this stops being a
	// single unnamed provider.
	do.Provide[*workqueue.Config](i, func(i do.Injector) (*workqueue.Config, error) {
		return &do.MustInvoke[*config.ScheduledJobsConfig](i).MealPlanning.MealPlanTaskNotificationQueue, nil
	})
	// The queue owns a goroutine, and is joined to the service's lifecycle through Runners.
	mealplantasknotifications.RegisterQueue(i)
	mealplantasknotifications.RegisterWorker(i)
}

// RegisterIndexes adds this domain's eight search indexes to a Registry. It is a
// searchindexes.Registrar, and the searchers it resolves are registered by the two Register
// functions above for the processes that run indexes.
func RegisterIndexes(i do.Injector, registry *searchsync.Registry) error {
	return mealplanningindexing.RegisterIndexes(i, registry)
}

var _ searchindexes.Registrar = RegisterIndexes

// ScheduledJobs is every periodic job this domain runs in the scheduler, already rendered from
// its config with the disabled ones left out.
//
// Each job resolves what it runs from inside its own closure at tick time, so a dependency that
// stopped being registered is a failed run rather than a failed boot. The integration suite's
// wiring test resolves them for that reason.
func ScheduledJobs(i do.Injector) ([]jobs.Job, error) {
	jobsCfg := &do.MustInvoke[*config.ScheduledJobsConfig](i).MealPlanning

	registrations := []struct {
		run  func(ctx context.Context) error
		cfg  *jobscfg.JobConfig
		name string
	}{
		{
			name: jobMealPlanFinalizationStarter,
			cfg:  &jobsCfg.MealPlanFinalizationStarter,
			// The starter reports how many sagas it began; the scheduler has nowhere to
			// put a count, and the worker already records it as a metric.
			run: func(ctx context.Context) error {
				_, workErr := do.MustInvoke[*mealplanfinalization.Starter](i).Work(ctx)

				return workErr
			},
		},
		{
			name: jobMealPlanTaskNotifications,
			cfg:  &jobsCfg.MealPlanTaskNotifications,
			// One pass enqueues every task still owed a reminder, drains what the
			// queue hands over, and sends under the lease. The count of pushes it
			// sent has nowhere to go here — the queue and the fan-out both record
			// their own counters — so it is dropped the way the finalization
			// starter's is.
			run: func(ctx context.Context) error {
				_, workErr := do.MustInvoke[*mealplantasknotifications.Worker](i).Work(ctx)

				return workErr
			},
		},
	}

	scheduled := make([]jobs.Job, 0, len(registrations))

	for idx := range registrations {
		r := &registrations[idx]

		if r.cfg.Disabled {
			continue
		}

		job, err := r.cfg.Job(r.name, r.run)
		if err != nil {
			return nil, err
		}

		scheduled = append(scheduled, job)
	}

	return scheduled, nil
}

// Runners is every long-lived component of this domain's the scheduler joins to its lifecycle
// through service.WithRunners: the prep task reminder queue, which owns a goroutine and whose
// Close writes its last batch out.
func Runners(i do.Injector) ([]service.Runner, error) {
	queue, err := mealplantasknotifications.NewQueueRunner(i)
	if err != nil {
		return nil, err
	}

	return []service.Runner{queue}, nil
}

// CommentTargets is this domain's entry in the comment catalog, with no existence checks.
//
// It is the entry for a process that reads and erases comments but never writes one — the
// scheduler fulfilling a data privacy request, or the data change handler. The catalog gates
// writes rather than reads, so a hookless one is exactly right there, and it still refuses a
// misspelled type should a write path arrive later.
func CommentTargets() platformcomments.Targets {
	return platformcomments.Targets{
		mealplanning.CommentTargetTypeRecipes:   {Description: "A recipe."},
		mealplanning.CommentTargetTypeMeals:     {Description: "A meal."},
		mealplanning.CommentTargetTypeMealPlans: {Description: "A meal plan."},
	}
}

// CheckedCommentTargets is CommentTargets with an existence check on every type the manager can
// answer "is this there" for from the ID alone, for a process that writes comments.
//
// Meal plans are left unchecked. The check platform runs is handed the comment's scope and
// nothing else, and reading a meal plan takes an owner as well as an ID, so the scope the hook
// receives is not one that read can use; the meal planning service reads its target as the
// caller before it delegates, which is a stronger check than this one would be rather than a
// missing one.
//
// A check narrows the window in which a comment can be written about something that is not
// there; it does not close it. A target deleted between the check and the insert is still a
// comment about nothing.
func CheckedCommentTargets(i do.Injector) platformcomments.Targets {
	manager := do.MustInvoke[mealplanningmgr.MealPlanningManager](i)
	targets := CommentTargets()

	targets[mealplanning.CommentTargetTypeRecipes] = comments.WithExistenceCheck(
		targets[mealplanning.CommentTargetTypeRecipes],
		func(ctx context.Context, targetID string) error {
			_, err := manager.ReadRecipe(ctx, targetID)

			return err
		},
	)

	targets[mealplanning.CommentTargetTypeMeals] = comments.WithExistenceCheck(
		targets[mealplanning.CommentTargetTypeMeals],
		func(ctx context.Context, targetID string) error {
			_, err := manager.ReadMeal(ctx, targetID)

			return err
		},
	)

	return targets
}
