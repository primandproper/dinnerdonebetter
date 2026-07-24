package integration

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	apiserver "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/services/api"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/localdev"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	notificationsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/notifications"

	"github.com/primandproper/platform-go/v6/database"
	databasecfg "github.com/primandproper/platform-go/v6/database/config"
	"github.com/primandproper/platform-go/v6/identifiers"
	"github.com/primandproper/platform-go/v6/random"
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
		dbCfg  *databasecfg.Config
	)

	server, databaseClient, dbCfg, err = localdev.BuildInProcessServer(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	dbConnStr = dbCfg.ReadConnection.String()

	// create premade admin user
	auditLogRepo := auditlogentries.ProvideAuditLogRepository(pillars.Logger, pillars.TracerProvider, databaseClient)
	identityRepo := identityrepo.ProvideIdentityRepository(pillars.Logger, pillars.TracerProvider, auditLogRepo, databaseClient)
	notifsRepo = notificationsrepo.ProvideNotificationsRepository(nil, nil, auditLogRepo, dbCfg, databaseClient)
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
