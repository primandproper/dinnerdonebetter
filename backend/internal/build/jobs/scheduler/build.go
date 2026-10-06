package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	commentstargets "github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	dataprivacybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/queuedmail"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/recordingspine"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	identitystore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"
	oauth2clientsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth2clientsstore"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	signindevicesrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/signindevices"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	"github.com/primandproper/platform-go/v15/service"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	notificationscfg "github.com/primandproper/primitives-go/v2/notifications/mobile/config"

	"github.com/samber/do/v2"
)

// BuildInjector validates cfg and composes the scheduler process from it.
//
// The container is the union of what the six periodic jobs used to build separately, one
// short-lived process each. Consolidating them means the connection pools, the tracer, and the
// repositories are constructed once at startup rather than once per tick — which is most of the
// cost of a job that runs every minute.
//
// service.Register builds the half of it platform owns, from cfg.Service: the database, the
// broker, the pillars, the scheduler and its lock, and every platform loop this process runs. What
// follows it is what platform cannot know — the repositories, the domain, the registries the
// loops dispatch into, and the jobs. Validation comes first and is not optional: it is what
// releases the blocks env parsing allocated and nobody configured, and Register reads presence as
// the decision to build.
func BuildInjector(
	ctx context.Context,
	cfg *config.SchedulerConfig,
) (*do.RootScope, error) {
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating scheduler config: %w", err)
	}

	// The prefix and dialect of platform's data privacy tables are this application's, not
	// the deployment's — see dataprivacycfg.Pin.
	dataprivacycfg.Pin(cfg.Service.DataPrivacy)

	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	service.Register(i, &cfg.Service)

	RegisterConfigs(i)

	// repositories
	//
	// This process holds every domain repository, which it did not before, because the data
	// privacy fulfillment worker runs here and one export is a fan-out over all of them. That
	// is the cost of moving the gather off the message queue and onto a claimed row: the
	// process that claims has to be able to answer. It is paid once at startup — the pools and
	// the tracer are constructed here regardless — rather than per request.
	auditlogentries.RegisterAuditLogRepository(i)
	auditlogentries.RegisterPlatformReader(i)
	// And the recorder, which the recording spine registered below files every write's
	// entry through.
	auditlogentries.RegisterPlatformRecorder(i)
	// No existence checks on the catalog: this process reads and erases comments
	// but never writes one, and the catalog gates writes rather than reads.
	commentstargets.RegisterReadOnlyTargets(i)
	commentsrepo.RegisterCommentsRepository(i)
	// What a role grants, read from the policy tables the migrator seeds. The
	// identity repository resolves a principal's role names through it when it
	// builds a session.
	authorization.RegisterPolicyResolver(i)
	identitystore.RegisterIdentityStore(i)
	// The mailer the identity service hands an invitation to, and sign-in, password
	// reset and waitlists their mail. See internal/build/queuedmail.
	queuedmail.Register(i)
	internalopsrepo.RegisterInternalOpsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	settingsrepo.RegisterSettingsRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)

	// The webhook store, dispatcher and delivery worker are platform's, from the Webhooks
	// block. The catalog is this application's: what an event means is an application
	// opinion, and generated Go rather than configuration. The dispatcher is needed in both
	// directions here — dispatch happens inside the transaction that causes the event, and
	// the meal plan finalizer emits events like any request does.
	do.ProvideValue[platformwebhooks.Catalog](i, catalog.Catalog())

	// The notifications store and the push fan-out over its device registry. The fan-out is
	// the part of mobile notifications that has nothing to do with why one is owed: device
	// tokens in, pushes out, dead tokens retired — and the async message handler builds the
	// same one. This process delivers prep task reminders itself rather than handing them to
	// the async message handler over a topic — see the meal plan task notification worker for
	// why the send has to happen under the queue lease that claimed the task — so it holds the
	// push sender too, configured outside Service for the reason the config gives.
	notificationscfg.RegisterPushSender(i)
	notificationsstore.RegisterNotificationsStore(i)
	push.RegisterFanout(i)

	// The two credential stores the privacy registry collects from and nothing else in
	// this process touches. The passkey store arrives with the identity store above.
	//
	// They are here because this is the process that fulfills a subject access request, and
	// an export owes the subject every domain that holds them — including which devices can
	// sign in as them, which reset links are outstanding, and what holds API access on their
	// behalf. A registry built in a container missing one of these fails when the worker
	// starts; a registry that skipped it instead would deliver an export that looks complete.
	authrepo.RegisterAuthRepository(i)
	// The devices behind a person's sign-ins, which a subject access request exports.
	signindevicesrepo.RegisterSignInDevicesRepository(i)
	oauth2clientsstore.RegisterOAuth2ClientsStore(i)

	// The data privacy machinery is platform's, from the DataPrivacy block: the request store,
	// the artifact storage — a bucket and a keyring of its own — the fulfiller, and the sweep
	// that expires what it wrote. What it cannot build from configuration is registered here:
	// the registry of who holds data about a person, the operations registry the fulfiller
	// files its kinds into, the key the artifact keyring is built over, and the codec.
	dataprivacycfg.RegisterKeyset(i, cfg.DataPrivacyArtifactEncryptionKey)
	dataprivacycfg.RegisterCompressor(i)
	dataprivacybuild.RegisterRegistry(i)
	dataprivacybuild.RegisterOperationsRegistry(i)
	dataprivacybuild.RegisterCompletionNotifier(i)

	// Domain: mealplanning
	//
	// recordingspine.Register is also where the recording spine is assembled — the outbox
	// writer, the webhook emitter, and the recorder over them — for the reason
	// config.SchedulerConfig.OutboxRelay gives.
	recordingspine.Register(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)

	// the periodic jobs themselves
	mealplanfinalization.RegisterStarter(i)
	queuetest.RegisterQueueTest(i)

	do.Provide[*queuetest.JobParams](i, func(i do.Injector) (*queuetest.JobParams, error) {
		return &queuetest.JobParams{Queues: *do.MustInvoke[*queuescfg.Config](i)}, nil
	})

	// The prep task reminder queue and the worker that fills and drains it. The queue owns a
	// goroutine, and is joined to the service's lifecycle by cmd/ddb — see NotificationQueue.
	mealplantasknotifications.RegisterQueue(i)
	mealplantasknotifications.RegisterWorker(i)

	// The Reindexers this process drives on a schedule. They come from the same Registry as
	// the Syncers the consumer runs, because the two are halves of keeping one index right.
	searchindexes.Register(ctx, i, identityindexing.RegisterIndexes, mealplanningindexing.RegisterIndexes)

	// The saga definitions, and the publisher and runners over them. The worker that advances
	// every definition in the process is platform's, from the Saga block, along with its store
	// and the job that prunes finished instances.
	//
	// It is built without an idempotency manager, and that is a decision rather than an
	// omission. The manager suppresses a step whose result was recorded but whose instance row
	// did not catch up, and it does so from a store that commits separately from the step — so
	// for a step that writes to this database it is a weaker guarantee than the step already
	// has. Meal plan finalization's steps each write their work and the flag saying they did it
	// in one transaction, and re-read that flag before doing anything; a step that reached out
	// to something that cannot join a transaction would need the manager, and there is not one
	// yet. platform's worker takes one only if the container holds one, and this one does not.
	sagas.RegisterSagas(i)

	RegisterMetering(i)
	RegisterRetentionPolicies(i)
	RegisterOutboxRelay(i)
	RegisterJobs(i)

	return i, nil
}

// NotificationQueue adapts the prep task reminder queue to the service.Runner its lifecycle is
// joined through.
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
type NotificationQueue struct {
	queue *mealplantasknotifications.TaskQueue
	stop  chan struct{}
	once  sync.Once
}

var _ service.Runner = (*NotificationQueue)(nil)

// NewNotificationQueue resolves the reminder queue from i and wraps it.
func NewNotificationQueue(i do.Injector) (*NotificationQueue, error) {
	queue, err := do.Invoke[*mealplantasknotifications.TaskQueue](i)
	if err != nil {
		return nil, err
	}

	return &NotificationQueue{queue: queue, stop: make(chan struct{})}, nil
}

// Run blocks until Close.
func (q *NotificationQueue) Run() {
	<-q.stop
}

// Close writes out the queue's last batch and stops its goroutine.
func (q *NotificationQueue) Close(ctx context.Context) error {
	q.once.Do(func() { close(q.stop) })

	return q.queue.Close(ctx)
}
