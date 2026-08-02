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

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/random"
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
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("listener address is not TCP")
	}

	if err = l.Close(); err != nil {
		return 0, err
	}

	return tcpAddr.Port, nil
}

func init() {
	ctx := context.Background()

	cfg, err := config.LoadConfigFromPath[config.APIServiceConfig](apiConfigurationFilepath)
	if err != nil {
		log.Fatal(err)
	}

	// Use random ports to avoid conflicts with other running instances
	httpPort, err := getFreePort()
	if err != nil {
		log.Fatal(err)
	}
	grpcPort, err := getFreePort()
	if err != nil {
		log.Fatal(err)
	}

	cfg.HTTPServer.Port = uint16(httpPort)
	cfg.GRPCServer.Port = uint16(grpcPort)
	httpTestServerAddress = fmt.Sprintf("http://localhost:%d", httpPort)

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
	identityRepo := identityrepo.ProvideIdentityRepository(pillars.Logger, pillars.TracerProvider, auditLogRepo, databaseClient, nil)
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
	})
	if err != nil {
		log.Fatal(err)
	}
	createdClientID, createdClientSecret = createdClient.ClientID, createdClient.ClientSecret

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
