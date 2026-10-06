package integration

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	apiserver "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"

	platformoauth2clients "github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v15/billing"
	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	"github.com/primandproper/platform-go/v15/service"
	"github.com/primandproper/primitives-go/v2/database"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	msgconfig "github.com/primandproper/primitives-go/v2/messagequeue/config"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
)

const (
	apiConfigurationFilepath = "../../../deploy/environments/testing/config_files/integration-tests-config.json"

	// The other two workloads' rendered configurations, loaded rather than derived from the
	// API server's.
	//
	// Derived would be easier and would prove less. The three processes have to agree about
	// several things that live in all three files — the operations queue's name, the data
	// privacy bucket and cipher, the data changes topic — and every one of those agreements is
	// invisible when it breaks: a request submitted to a queue nothing claims, an artifact
	// written under a key nothing can open. Reading what a testing deployment actually renders
	// is what makes those agreements assertions rather than assumptions.
	schedulerConfigurationFilepath           = "../../../deploy/environments/testing/config_files/scheduler_config.json"
	asyncMessageHandlerConfigurationFilepath = "../../../deploy/environments/testing/config_files/async_message_handler_config.json"
)

var (
	dbConnStr                            string
	createdClientID, createdClientSecret string
	databaseClient                       database.Client
	apiServiceConfig                     *config.APIServiceConfig
	notifsInbox                          platformnotifications.Inbox
	notifsRegistry                       platformnotifications.Registry

	// billingStore seeds the rows two RPCs used to write.
	//
	// platform's billing surface has no CreateSubscription or UpdateSubscription, and that
	// is the ruling rather than a gap: a Subscription mirrors what a payment provider says
	// is paid for, so one created over the wire would grant paid features with nothing
	// behind them. A test that needs a subscription to read therefore writes one the way
	// the webhook handler does — through the store.
	billingStore          billing.Store
	httpTestServerAddress string

	// dataPrivacyFulfillment is the scheduler's half of a subject access request, run in this
	// process. See the note beside where it is started.
	dataPrivacyFulfillment *localdev.DataPrivacyFulfillment

	// The other two workloads' configurations, pointed at this suite's containers. They are
	// what the container-resolution tests build their injectors from.
	schedulerConfig *config.SchedulerConfig
)

// getFreePort asks the OS for a free open port that is ready to use.
// reservePort asks the OS for a free port and holds it until the caller releases it.
//
// It returns the listener rather than just the number because closing it here would be a
// time-of-check-to-time-of-use bug, and not a theoretical one: the containers this suite starts
// are mapped to host ports from the same ephemeral range the OS hands out here, so a Redis or
// Postgres container can be given the exact port the server is about to bind. That is a startup
// crash — "bind: address already in use" — presenting as an unrelated test failure.
//
// Holding the listener keeps the port out of that range until the server is ready for it.
func reservePort() (*net.TCPListener, int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, 0, err
	}

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return nil, 0, errors.Join(errors.New("listener address is not TCP"), l.Close())
	}

	return l, tcpAddr.Port, nil
}

func init() {
	ctx := context.Background()

	cfg, err := config.LoadConfigFromPath[config.APIServiceConfig](apiConfigurationFilepath)
	if err != nil {
		log.Fatal(err)
	}

	// Random ports, to avoid conflicts with other running instances. The reservations are held
	// until just before the server binds them — see reservePort.
	httpReservation, httpPort, err := reservePort()
	if err != nil {
		log.Fatal(err)
	}
	grpcReservation, grpcPort, err := reservePort()
	if err != nil {
		log.Fatal(err)
	}

	cfg.Service.HTTPServer.Port = uint16(httpPort)
	cfg.Service.GRPCServer.Port = uint16(grpcPort)
	httpTestServerAddress = fmt.Sprintf("http://localhost:%d", httpPort)

	// The authorization server's identity, which the rendered config cannot know: the port is
	// chosen here, at startup, and every endpoint in the discovery document is derived from the
	// issuer. Resources is the same string because the API server is both the authorization
	// server and the resource server the interceptor checks a token's audience against.
	cfg.Services.Auth.OAuth2.Issuer = httpTestServerAddress
	cfg.Services.Auth.OAuth2.Resources = []string{httpTestServerAddress}

	apiServiceConfig = cfg

	pillars, err := cfg.Service.Observability.NewPillars(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var (
		server *apiserver.Server
		dbCfg  *dbcfg.Config
	)

	server, databaseClient, dbCfg, err = localdev.BuildInProcessServer(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	dbConnStr = dbCfg.ReadConnection.String()

	// create premade admin user
	auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(pillars.Logger, pillars.TracerProvider, nil, databaseClient)
	if err != nil {
		log.Fatal(err)
	}
	identityDirectory, identityStore, err := localdev.IdentityDirectory(pillars.Logger, pillars.TracerProvider, databaseClient)
	if err != nil {
		log.Fatal(err)
	}
	auditRecorder, ok := auditlogentries.RecorderFrom(auditLogRepo)
	if !ok {
		log.Fatal("the audit log repository exposes no platform recorder")
	}
	// The recording spine the seeded stores record through: the same construction the
	// server's injector makes, by hand, because this seeding runs beside the server.
	spine, err := events.New(ctx, databaseClient, auditRecorder, events.WithPillars(pillars))
	if err != nil {
		log.Fatal(err)
	}
	notifsInbox, notifsRegistry, err = notificationsstore.ProvideStores(ctx, pillars.Logger, pillars.TracerProvider,
		metricsnoop.NewMetricsProvider(), spine.Recorder(), databaseClient)
	if err != nil {
		log.Fatal(err)
	}
	billingStore, err = paymentsrepo.ProvidePaymentsRepository(ctx, pillars.Logger, pillars.TracerProvider,
		metricsnoop.NewMetricsProvider(), spine.Recorder(), databaseClient)
	if err != nil {
		log.Fatal(err)
	}
	adminUser, err := localdev.CreatePremadeAdminUser(ctx, pillars.Logger, pillars.TracerProvider, identityDirectory, identityStore, databaseClient, premadeAdminUser)
	if err != nil {
		log.Fatal(err)
	}

	// The credentials are the registry's to mint, and the plaintext secret exists on what
	// this returns and nowhere else — the row holds a digest and no read reverses it.
	issuedClient, err := localdev.CreateOAuth2ClientForService(ctx, databaseClient, &platformoauth2clients.CreationInput{
		Name:        "integration_client",
		Description: "integration test client",
		// Registered, and matched byte for byte at /authorize and again at /token. The suite
		// authorizes against the API server's own address — nothing listens for the redirect,
		// because the code is read off the Location header rather than followed — and that
		// address is only known now, which is why this is not in the rendered config.
		RedirectURIs: []string{httpTestServerAddress},
	})
	if err != nil {
		log.Fatal(err)
	}
	createdClientID, createdClientSecret = issuedClient.Client.ClientID, issuedClient.Secret

	// The scheduler's half of the system. The API only starts sagas; without something
	// advancing them, everything downstream of meal plan finalization would never happen and
	// the tests that assert on it would be asserting on a pipeline that was never run.
	//
	// Never stopped: this process exits when the suite does, and a worker mid-pass at that
	// point has nothing to drain to.
	if _, err = localdev.StartSagaWorker(ctx, pillars.Logger, pillars.TracerProvider, databaseClient); err != nil {
		log.Fatal(err)
	}

	// The other two workloads' configurations, as a testing deployment renders them.
	workerDatabase, workerEvents = cfg.Service.Database, cfg.Service.MessageQueue
	if schedulerConfig, err = loadSchedulerConfig(); err != nil {
		log.Fatal(err)
	}
	// Loaded here only so a file that does not decode fails the suite at once; the wiring test
	// loads its own copy to compose from.
	if _, err = loadAsyncMessageHandlerConfig(); err != nil {
		log.Fatal(err)
	}

	// The other half the API server does not run: data privacy fulfillment. Submitting a
	// subject access request records a row and returns; the gather, the artifact, and the
	// erasure all happen in an operations worker that lives in the scheduler, so without one
	// here every export would stay in progress forever and the tests asserting on one would be
	// asserting on work nothing ever ran.
	//
	// Never stopped, for the same reason the saga worker is not: this process exits when the
	// suite does, and an operation mid-flight at that point has nothing to hand back to.
	if dataPrivacyFulfillment, err = localdev.NewDataPrivacyFulfillment(ctx, schedulerConfig); err != nil {
		log.Fatal(err)
	}

	go func() {
		if runErr := dataPrivacyFulfillment.Worker.Run(ctx); runErr != nil {
			log.Fatal(runErr)
		}
	}()

	// Release the reserved ports immediately before the server takes them. Everything that
	// could have stolen one — every container this suite starts — has already been mapped.
	if err = errors.Join(httpReservation.Close(), grpcReservation.Close()); err != nil {
		log.Fatal(err)
	}

	go func() {
		if runErr := server.Run(ctx); runErr != nil {
			log.Fatal(runErr)
		}
	}()

	fmt.Printf("DB conn str: %s", dbCfg.ReadConnection.String())
	dbConnStr = dbCfg.ReadConnection.String()
	fmt.Println("db conn str: " + dbConnStr)

	// accursed, but nevertheless we ball.
	time.Sleep(1 * time.Second)

	adminClient, err = createClientForUser(ctx, adminUser)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	// workerDatabase and workerEvents are the API server's database and broker, which the two
	// worker processes' configurations are pointed at — see pointAtSuiteContainers.
	workerDatabase *databasecfg.Config
	workerEvents   *msgconfig.Config
)

// loadSchedulerConfig loads the scheduler's configuration as a testing deployment renders it,
// pointed at this suite's containers.
//
// Each call loads a copy of its own. Composing a process validates its config, and validation
// normalizes and defaults the config's blocks in place, so two parallel tests composing from one
// shared config would race on it.
func loadSchedulerConfig() (*config.SchedulerConfig, error) {
	cfg, err := config.LoadConfigFromPath[config.SchedulerConfig](schedulerConfigurationFilepath)
	if err != nil {
		return nil, err
	}

	pointAtSuiteContainers(&cfg.Service)

	return cfg, nil
}

// loadAsyncMessageHandlerConfig is loadSchedulerConfig for the async message handler.
func loadAsyncMessageHandlerConfig() (*config.AsyncMessageHandlerConfig, error) {
	cfg, err := config.LoadConfigFromPath[config.AsyncMessageHandlerConfig](asyncMessageHandlerConfigurationFilepath)
	if err != nil {
		return nil, err
	}

	pointAtSuiteContainers(&cfg.Service)

	return cfg, nil
}

// pointAtSuiteContainers overwrites the two things a rendered config cannot know: which
// containers this suite started. Both are addresses rather than behavior, so everything else in
// those files is the deployment's own.
func pointAtSuiteContainers(workload *service.Config) {
	databaseCfg := *workload.Database
	databaseCfg.WriteConnection = workerDatabase.WriteConnection
	databaseCfg.ReadConnection = workerDatabase.ReadConnection
	// Migrations are the API server's job — see backend/docs/migrations.md — and a worker that
	// ran them here would race the one that already has.
	databaseCfg.RunMigrations = false
	workload.Database = &databaseCfg

	eventsCfg := *workerEvents
	workload.MessageQueue = &eventsCfg
}
