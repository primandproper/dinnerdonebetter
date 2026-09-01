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
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	notificationsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/random"
)

const (
	apiConfigurationFilepath = "../../../deploy/environments/testing/config_files/integration-tests-config.json"
)

var (
	dbConnStr                            string
	createdClientID, createdClientSecret string
	databaseClient                       database.Client
	apiServiceConfig                     *config.APIServiceConfig
	notifsRepo                           notifications.Repository
	httpTestServerAddress                string
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

	cfg.HTTPServer.Port = uint16(httpPort)
	cfg.GRPCServer.Port = uint16(grpcPort)
	httpTestServerAddress = fmt.Sprintf("http://localhost:%d", httpPort)

	// The authorization server's identity, which the rendered config cannot know: the port is
	// chosen here, at startup, and every endpoint in the discovery document is derived from the
	// issuer. Resources is the same string because the API server is both the authorization
	// server and the resource server the interceptor checks a token's audience against.
	cfg.Services.Auth.OAuth2.Issuer = httpTestServerAddress
	cfg.Services.Auth.OAuth2.Resources = []string{httpTestServerAddress}

	apiServiceConfig = cfg

	pillars, err := cfg.Observability.NewPillars(ctx)
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
	uploadsRegistryStore, err := localdev.UploadsRegistry(pillars.Logger, pillars.TracerProvider, databaseClient)
	if err != nil {
		log.Fatal(err)
	}
	identityRepo := identityrepo.ProvideIdentityRepository(pillars.Logger, pillars.TracerProvider, auditLogRepo, databaseClient, nil, uploadsRegistryStore)
	notifsRepo = notificationsrepo.ProvideNotificationsRepository(nil, nil, auditLogRepo, &dbCfg.Config, databaseClient, nil)
	adminUser, err := localdev.CreatePremadeAdminUser(ctx, pillars.Logger, pillars.TracerProvider, identityRepo, databaseClient, premadeAdminUser)
	if err != nil {
		log.Fatal(err)
	}

	createdClient, err := localdev.CreateOAuth2ClientForService(ctx, databaseClient, dbCfg, &oauth.OAuth2ClientDatabaseCreationInput{
		ID:           identifiers.New(),
		Name:         "integration_client",
		Description:  "integration test client",
		ClientID:     random.MustGenerateHexEncodedString(ctx, oauth.ClientIDSize),
		ClientSecret: random.MustGenerateHexEncodedString(ctx, oauth.ClientSecretSize),
		// Registered, and matched byte for byte at /authorize and again at /token. The suite
		// authorizes against the API server's own address — nothing listens for the redirect,
		// because the code is read off the Location header rather than followed — and that
		// address is only known now, which is why this is not in the rendered config.
		RedirectURIs: []string{httpTestServerAddress},
	})
	if err != nil {
		log.Fatal(err)
	}
	createdClientID, createdClientSecret = createdClient.ClientID, createdClient.ClientSecret

	// The scheduler's half of the system. The API only starts sagas; without something
	// advancing them, everything downstream of meal plan finalization would never happen and
	// the tests that assert on it would be asserting on a pipeline that was never run.
	//
	// Never stopped: this process exits when the suite does, and a worker mid-pass at that
	// point has nothing to drain to.
	if _, err = localdev.StartSagaWorker(ctx, pillars.Logger, pillars.TracerProvider, databaseClient); err != nil {
		log.Fatal(err)
	}

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
