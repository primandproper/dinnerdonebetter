package integration

import (
	"context"
	"database/sql"
	"log"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	databasecfg "github.com/primandproper/platform-go/v12/database/config"
	"github.com/primandproper/platform-go/v12/identifiers"
)

const (
	// mcpConfigurationFilepath is the rendered file `ddb serve mcp` reads in this
	// environment. Loading it, rather than building a config in Go the way the API
	// suite does, is a large part of the point: the fields this suite depends on —
	// the oauth2 provider and the table prefix under it — are written by
	// internal/config/environments and can only ever be wrong in the file.
	mcpConfigurationFilepath = "../../../deploy/environments/testing/config_files/mcp_server_config.json"

	// adminUserPassword is the password the premade admin signs in with. The user
	// record stores its argon2 digest; this is the plaintext the login form is given.
	adminUserPassword = "integration-tests-are-cool"
)

var (
	// mcpServiceConfig is what every replica is built from: the rendered file with the
	// database connection repointed at this run's container, and nothing else changed.
	//
	// Replicas take a copy rather than this pointer — building a service fills in the
	// issuer and the resources in place, and a run that let the first replica's base
	// URL stick would silently give every later one the same issuer.
	mcpServiceConfig *config.MCPServiceConfig

	// rawDB is a handle on the database the replicas use, for asserting on rows nothing
	// of ours exposes: the four tables migration 33 creates under the oauth2 prefix.
	rawDB *sql.DB

	// adminUser is the account the login form is driven with. Its HashedPassword field
	// holds the digest by the time CreatePremadeAdminUser returns, which is why the
	// plaintext is a constant beside it rather than read back off this.
	adminUser *identity.User

	// seededIngredient is the row a tool call reads back. A tool returning an empty page
	// proves only that the handler ran; returning this proves it reached Postgres through
	// the container the MCP server built for itself.
	seededIngredient *mealplanning.ValidIngredient

	// fleetBaseURL is the public address every replica in this run advertises: its
	// issuer, the resource its tokens are minted for, and the resource its bearer
	// middleware refuses tokens minted for anywhere else. It is the primary replica's
	// own address, so that a client which follows the discovery document reaches a
	// server.
	fleetBaseURL string

	// primary is the replica bound at fleetBaseURL. Tests that do not need a second
	// process drive this one, and nothing stops it — a test that wants a replica to
	// stop starts its own.
	primary *instance
)

func init() {
	ctx := context.Background()

	cfg, err := config.LoadConfigFromPath[config.MCPServiceConfig](mcpConfigurationFilepath)
	if err != nil {
		log.Fatal(err)
	}

	// The container, and the only edit this suite makes to the rendered config. The
	// hostname in the file names a docker-compose service that is not running here.
	if rawDB, err = buildDatabase(ctx, cfg); err != nil {
		log.Fatal(err)
	}

	mcpServiceConfig = cfg

	if primary, err = startInstance(ctx, "", nil); err != nil {
		log.Fatal(err)
	}
	fleetBaseURL = primary.baseURL
}

// buildDatabase stands up a Postgres container, points cfg at it, and migrates it.
//
// The MCP server does not migrate: its container builds a database client with no
// migrator, because in a deployment `ddb migrate` has already run. So the suite plays
// that part, and the schema the replicas find is the one the real migrator creates —
// which is the only way the table prefix in the config can be checked against the tables
// that actually exist.
func buildDatabase(ctx context.Context, cfg *config.MCPServiceConfig) (*sql.DB, error) {
	_, db, dbCfg, err := pgtesting.BuildDatabaseContainer(ctx, "mcp_integration_testing")
	if err != nil {
		return nil, err
	}

	cfg.Database.ReadConnection = dbCfg.ReadConnection
	cfg.Database.WriteConnection = dbCfg.WriteConnection

	pillars, err := cfg.Observability.NewPillars(ctx)
	if err != nil {
		return nil, err
	}

	migrator, err := repositories.ProvideMigrator(&cfg.Database.Config, pillars.Logger)
	if err != nil {
		return nil, err
	}

	databaseClient, err := databasecfg.NewDatabase(ctx, &cfg.Database.Config, migrator,
		databasecfg.WithLogger(pillars.Logger),
		databasecfg.WithTracerProvider(pillars.TracerProvider),
	)
	if err != nil {
		return nil, err
	}

	auditRepo, err := auditlogentries.ProvideAuditLogRepository(pillars.Logger, pillars.TracerProvider, nil, databaseClient)
	if err != nil {
		return nil, err
	}

	identityRepo := identityrepo.ProvideIdentityRepository(pillars.Logger, pillars.TracerProvider, auditRepo, databaseClient, nil)

	// The MCP login form is admin-only and checks a second factor whenever the account
	// has a verified secret, both of which this helper arranges — so the flow the suite
	// drives is the whole one, not a version of it with the second factor turned off.
	adminUser, err = localdev.CreatePremadeAdminUser(ctx, pillars.Logger, pillars.TracerProvider, identityRepo, databaseClient, &identity.User{
		ID:              identifiers.New(),
		TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		EmailAddress:    "mcp_integration_tests@example.email",
		Username:        "mcp_admin_user",
		HashedPassword:  adminUserPassword,
	})
	if err != nil {
		return nil, err
	}

	mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(pillars.Logger, pillars.TracerProvider, auditRepo, identityRepo, databaseClient, nil)

	seededIngredient, err = mealPlanningRepo.CreateValidIngredient(ctx,
		mealplanningconverters.ConvertValidIngredientToValidIngredientDatabaseCreationInput(mealplanningfakes.BuildFakeValidIngredient()))
	if err != nil {
		return nil, err
	}

	return db, nil
}
