package datachangemessagehandler

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	commentstargets "github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/queuedmail"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	"github.com/primandproper/dinnerdonebetter/backend/internal/functions/datachangemessagehandler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	identitystore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issue_reports "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooksstore"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v15/service"
	notificationscfg "github.com/primandproper/primitives-go/v2/notifications/mobile/config"

	"github.com/samber/do/v2"
)

// BuildInjector validates cfg and composes the async message handler from it.
//
// service.Register builds what platform owns from cfg.Service — the pillars, the database, the
// broker, the encoding, the outbound HTTP client, analytics and the emailer — and what follows it
// is this application's: the repositories, the indexers, and the handler that routes each message.
// Validation comes first for the reason the scheduler's BuildInjector gives.
func BuildInjector(
	ctx context.Context,
	cfg *config.AsyncMessageHandlerConfig,
) (*do.RootScope, error) {
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating async message handler config: %w", err)
	}

	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	service.Register(i, &cfg.Service)

	RegisterConfigs(i)

	// The push sender is registered by hand, configured outside Service — see
	// config.AsyncMessageHandlerConfig.PushNotifications.
	notificationscfg.RegisterPushSender(i)

	// Domain: mealplanning
	mealplanningregistration.RegisterForDataChangeHandler(i)

	// repos
	auditlogentries.RegisterAuditLogRepository(i)
	// The platform recorder behind it, which the recording spine the mealplanning
	// registration above installs files every write's entry through.
	auditlogentries.RegisterPlatformRecorder(i)
	// No existence checks on the catalog: this process reads and erases comments
	// but never writes one, and the catalog gates writes rather than reads.
	commentstargets.RegisterReadOnlyTargets(i)
	commentsrepo.RegisterCommentsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	// What a role grants, read from the policy tables the migrator seeds. The
	// identity repository resolves a principal's role names through it when it
	// builds a session.
	authorization.RegisterPolicyResolver(i)
	identitystore.RegisterIdentityStore(i)
	// The mailer the identity service hands an invitation to, and sign-in, password
	// reset and waitlists their mail. See internal/build/queuedmail.
	queuedmail.Register(i)
	issue_reports.RegisterIssueReportsRepository(i)
	uploadedmedia.RegisterUploadedMediaRepository(i)
	webhooksstore.RegisterWebhooksStore(i)
	internalopsrepo.RegisterInternalOpsRepository(i)
	notificationsstore.RegisterNotificationsStore(i)

	// The push fan-out over the notifications store's device registry. The scheduler builds
	// the same one for the prep task reminders it now sends itself, so both processes deliver
	// through one component.
	push.RegisterFanout(i)
	settingsrepo.RegisterSettingsRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)

	// Every search index this process drains the topic of: one Registry, built from each
	// domain's indexes. See internal/searchindexes.
	searchindexes.Register(ctx, i, identityindexing.RegisterIndexes, mealplanningindexing.RegisterIndexes)

	// searchers
	RegisterSearchers(i)

	// main handler
	datachangemessagehandler.RegisterAsyncDataChangeMessageHandler(i)

	return i, nil
}

// HandlerRunner joins the data change handler to a service.Service's lifecycle.
//
// The handler is the one loop this process exists for, and it is this application's rather than
// platform's, so it arrives through service.WithRunners. An application runner is closed first,
// which is the right place for a consumer: nothing upstream of it in this process is still
// producing.
//
// Close also flushes the search syncers' stamp buffers, after the handler has drained and before
// the service releases the database client those flushes write through. That is the slot
// service.New gives a *searchsync.Registry; this process registers internal/searchindexes'
// wrapper over one, which service does not resolve, so the flush is made here instead, by
// retiring the wrapper — see searchindexes.Registry.
type HandlerRunner struct {
	handler *datachangemessagehandler.AsyncDataChangeMessageHandler
	i       do.Injector
	stop    chan struct{}
	once    sync.Once
}

var _ service.Runner = (*HandlerRunner)(nil)

// NewHandlerRunner resolves the handler from i and wraps it.
func NewHandlerRunner(i do.Injector) (*HandlerRunner, error) {
	handler, err := do.Invoke[*datachangemessagehandler.AsyncDataChangeMessageHandler](i)
	if err != nil {
		return nil, err
	}

	return &HandlerRunner{handler: handler, i: i, stop: make(chan struct{})}, nil
}

// Start subscribes the handler's pools. It is the one step that can fail, so the caller makes it
// before running the service, and returns its error rather than running a process that drains
// nothing.
func (r *HandlerRunner) Start(ctx context.Context) error {
	return r.handler.Start(ctx)
}

// Run blocks until Close.
func (r *HandlerRunner) Run() {
	<-r.stop
}

// Close stops each pool's consumer, lets the messages already being handled finish, and then
// flushes the stamps those handlers produced.
func (r *HandlerRunner) Close(ctx context.Context) error {
	r.once.Do(func() { close(r.stop) })

	return errors.Join(r.handler.Close(ctx), do.ShutdownWithContext[*searchindexes.Registry](ctx, r.i))
}
