package localdev

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	apiserver "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/api"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	notificationsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications"
	oauthrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	webhooksrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v11/authentication/argon2"
	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	"github.com/primandproper/platform-go/v11/database"
	databasecfg "github.com/primandproper/platform-go/v11/database/config"
	"github.com/primandproper/platform-go/v11/httpclient"
	"github.com/primandproper/platform-go/v11/identifiers"
	msgconfig "github.com/primandproper/platform-go/v11/messagequeue/config"
	"github.com/primandproper/platform-go/v11/messagequeue/redis"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"
	"github.com/primandproper/platform-go/v11/random"
	"github.com/primandproper/platform-go/v11/testutils/containers/redistest"
	webhookscfg "github.com/primandproper/platform-go/v11/webhooks/config"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const redisProtocolPrefix = "redis://"

// buildContainerBackedRedisConfig spins up a Redis testcontainer and returns a
// *redis.Config pointed at it. redistest.Try is the entry point platform provides for
// callers outside a testing.TB: it applies the same image, wait strategy and retry
// policy the test suites get, but enforces neither the RUN_CONTAINER_TESTS gate nor any
// cleanup registration, which is what a long-running localdev process needs.
func buildContainerBackedRedisConfig(ctx context.Context) (*redis.Config, func(context.Context) error, error) {
	redisContainer, shutdown, err := redistest.Try(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build redis container: %w", err)
	}

	redisAddress, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		if shutdownErr := shutdown(ctx); shutdownErr != nil {
			slog.Error("failed to terminate redis container", slog.Any("error", shutdownErr))
		}
		return nil, nil, fmt.Errorf("failed to build redis connection string: %w", err)
	}

	cfg := &redis.Config{
		QueueAddresses: []string{strings.TrimPrefix(redisAddress, redisProtocolPrefix)},
	}

	return cfg, shutdown, nil
}

func CreatePremadeAdminUser(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	identityRepo identity.Repository,
	dbClient database.Client,
	premadeAdminUser *identity.User,
) (*identity.User, error) {
	hasher := authentication.NewArgon2Authenticator(argon2.WithLogger(logger), argon2.WithTracerProvider(tracerProvider))

	actuallyHashedPass, err := hasher.HashPassword(ctx, premadeAdminUser.HashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	premadeAdminUser.HashedPassword = actuallyHashedPass

	var user *identity.User
	if user, err = identityRepo.GetUserByUsername(ctx, premadeAdminUser.Username); err == nil {
		return user, nil
	}

	user, err = identityRepo.CreateUser(ctx, identityconverters.ConvertUserToUserDatabaseCreationInput(premadeAdminUser))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Promote user to service_admin by archiving old service role and assigning new one.
	if _, err = dbClient.Writer().ExecContext(ctx, "UPDATE user_role_assignments SET archived_at = NOW() WHERE user_id = $1 AND account_id IS NULL AND archived_at IS NULL", user.ID); err != nil {
		return nil, fmt.Errorf("failed to archive old service role: %w", err)
	}
	if _, err = dbClient.Writer().ExecContext(ctx, "INSERT INTO user_role_assignments (id, user_id, role_id) VALUES ($1, $2, $3)", identifiers.New(), user.ID, authorization.ServiceAdminRoleID); err != nil {
		return nil, fmt.Errorf("failed to assign service_admin role: %w", err)
	}

	if err = identityRepo.MarkUserTwoFactorSecretAsVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to mark user as verified: %w", err)
	}

	adminUser, err := identityRepo.GetAdminUserByUsername(ctx, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}

	return adminUser, nil
}

func CreateOAuth2ClientForService(ctx context.Context, pgc database.Client, dbCfg *dbcfg.Config, oauth2Input *oauth.OAuth2ClientDatabaseCreationInput) (*oauth.OAuth2Client, error) {
	auditRepo, err := auditlogentries.ProvideAuditLogRepository(nil, nil, nil, pgc)
	if err != nil {
		return nil, err
	}
	oauth2ClientManager := oauthrepo.ProvideOAuthRepository(ctx, nil, nil, auditRepo, dbCfg, pgc)

	// only the digest is persisted; hand the plaintext back to the caller.
	plaintextSecret := oauth2Input.ClientSecret
	oauth2Input.ClientSecret = oauth.HashClientSecret(plaintextSecret)

	createdClient, err := oauth2ClientManager.CreateOAuth2Client(ctx, oauth2Input)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 client: %w", err)
	}
	createdClient.ClientSecret = plaintextSecret

	return createdClient, nil
}

func BuildInProcessServer(ctx context.Context, cfg *config.APIServiceConfig) (server *apiserver.Server, databaseClient database.Client, dbCfg *dbcfg.Config, err error) {
	pillars, err := cfg.Observability.NewPillars(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("setting up observability pillars: %w", err)
	}
	logger := pillars.Logger

	redisConfig, _, err := buildContainerBackedRedisConfig(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connecting to redis: %w", err)
	}
	cfg.Events.Publisher.Provider = msgconfig.ProviderRedis
	cfg.Events.Publisher.Redis = *redisConfig
	cfg.Events.Consumer.Redis = *redisConfig

	// set up a database container, migrate it, and build a connection client
	_, _, dbCfg, err = pgtesting.BuildDatabaseContainer(ctx, "integration_testing")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	cfg.Database.WriteConnection = dbCfg.WriteConnection
	cfg.Database.ReadConnection = dbCfg.ReadConnection

	tracerProvider := tracingnoop.NewTracerProvider()
	migrator, err := repositories.ProvideMigrator(&cfg.Database.Config, logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("building migrator: %w", err)
	}

	databaseClient, err = databasecfg.NewDatabase(ctx, &cfg.Database.Config, migrator,
		databasecfg.WithLogger(logger),
		databasecfg.WithTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("initializing database client: %w", err)
	}

	// create premade admin user
	server, err = apiserver.NewServer(ctx, pillars, cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("building API server: %w", err)
	}

	return server, databaseClient, &cfg.Database, nil
}

// DatabaseInitFunc is a function that performs database initialization operations.
// It receives the database client, config, logger, and tracer to perform arbitrary operations.
type DatabaseInitFunc func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error

// WithIdentityRepository provides an identity repository for custom operations.
// The provided function receives a fully configured identity.Repository along with logger, tracer, and database client.
func WithIdentityRepository(fn func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.Provider, dbClient database.Client) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditLogRepo, dbClient, nil)
		return fn(ctx, identityRepo, logger, tracerProvider, dbClient)
	}
}

// WithOAuth2Repository provides an OAuth2 repository for custom operations.
// The provided function receives a fully configured oauth.Repository along with logger and tracer.
func WithOAuth2Repository(fn func(ctx context.Context, repo oauth.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		oauthRepo := oauthrepo.ProvideOAuthRepository(ctx, logger, tracerProvider, auditLogRepo, dbCfg, dbClient)
		return fn(ctx, oauthRepo, logger, tracerProvider)
	}
}

// WithAuthRepository provides an auth repository for custom operations.
// The provided function receives a fully configured auth.Repository along with logger and tracer.
func WithAuthRepository(fn func(ctx context.Context, repo auth.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		authRepo := authrepo.ProvideAuthRepository(logger, tracerProvider, auditLogRepo, dbClient)
		return fn(ctx, authRepo, logger, tracerProvider)
	}
}

// WithMealPlanningRepository provides a meal planning repository for custom operations.
// The provided function receives a fully configured mealplanning.Repository along with logger and tracer.
// This repository handles all meal planning entities including recipes, ingredients, preparations, vessels, instruments, etc.
func WithMealPlanningRepository(fn func(ctx context.Context, repo mealplanning.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditLogRepo, dbClient, nil)
		mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditLogRepo, identityRepo, dbClient, nil)
		return fn(ctx, mealPlanningRepo, logger, tracerProvider)
	}
}

// WithSettingsRepository provides a settings repository for custom operations.
// The provided function receives a fully configured settings.Repository along with logger and tracer.
func WithSettingsRepository(fn func(ctx context.Context, repo settings.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		settingsRepo := settingsrepo.ProvideSettingsRepository(logger, tracerProvider, auditLogRepo, dbClient, nil)
		return fn(ctx, settingsRepo, logger, tracerProvider)
	}
}

// WithWebhooksRepository provides a webhooks repository for custom operations.
// The provided function receives a fully configured webhooks.Repository along with logger and tracer.
func WithWebhooksRepository(fn func(ctx context.Context, repo webhooks.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}

		// A dispatcher is needed because creating a webhook registers a delivery endpoint,
		// and a nil one refuses rather than silently storing a webhook that never fires.
		store, err := webhookscfg.NewStore(ctx, &webhookscfg.Config{}, dbClient)
		if err != nil {
			return err
		}

		dispatcher, err := webhookscfg.NewDispatcher(
			ctx,
			&webhookscfg.Config{},
			store,
			catalog.Catalog(),
			webhookscfg.WithLogger(logger),
			webhookscfg.WithTracerProvider(tracerProvider),
		)
		if err != nil {
			return err
		}

		webhooksRepo := webhooksrepo.ProvideWebhooksRepository(logger, tracerProvider, auditLogRepo, dbClient, nil, dispatcher, store)
		return fn(ctx, webhooksRepo, logger, tracerProvider)
	}
}

// WithNotificationsRepository provides a notifications repository for custom operations.
// The provided function receives a fully configured notifications.Repository along with logger and tracer.
func WithNotificationsRepository(fn func(ctx context.Context, repo notifications.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		notificationsRepo := notificationsrepo.ProvideNotificationsRepository(logger, tracerProvider, auditLogRepo, &dbCfg.Config, dbClient, nil)
		return fn(ctx, notificationsRepo, logger, tracerProvider)
	}
}

// AllInOne sets up a complete local development environment with a docker-backed server,
// database, and runs the provided database initialization functions.
func AllInOne(ctx context.Context, cfg *config.APIServiceConfig, initFuncs ...DatabaseInitFunc) (*apiserver.Server, error) {
	server, databaseClient, dbCfg, err := BuildInProcessServer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("building in-process server: %w", err)
	}

	log.Printf("%sDATABASE CONNECTION URL: %s%s", strings.Repeat("\n", 10), dbCfg.ReadConnection.URI(), strings.Repeat("\n", 10))

	pillars, err := cfg.Observability.NewPillars(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting o11y pillars: %w", err)
	}

	// Run all database initialization functions
	for i, initFunc := range initFuncs {
		if err = initFunc(ctx, databaseClient, dbCfg, pillars.Logger, pillars.TracerProvider); err != nil {
			return nil, fmt.Errorf("running database init function %d: %w", i, err)
		}
	}

	return server, nil
}

// NewOAuth2ConfigForTestServer builds the OAuth2 config the integration suite and localdev
// drive the authorization code flow with. The redirect URL is the HTTP server's own address:
// nothing is listening for the redirect, because the code is read off the Location header of
// the 302 rather than followed.
func NewOAuth2ConfigForTestServer(clientID, clientSecret, httpServerAddress string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"anything"}, // TODO: This should be nil-able
		RedirectURL:  httpServerAddress,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   httpServerAddress + oauth2server.PathAuthorize,
			TokenURL:  httpServerAddress + oauth2server.PathToken,
		},
	}
}

// NewNonRedirectingHTTPClient returns an HTTP client that hands the caller the 302 instead of
// following it. The authorization code lives on that response's Location header, and the
// redirect target is not a real endpoint.
func NewNonRedirectingHTTPClient() (*http.Client, error) {
	httpClient, err := httpclient.NewHTTPClient(httpclient.WithTracing(true))
	if err != nil {
		return nil, fmt.Errorf("building http client: %w", err)
	}

	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return httpClient, nil
}

// exchangeAuthorizationCodeWithJWT runs the full authorization code flow against the API
// server: POST /authorize authenticated with the caller's JWT, read the code off the redirect,
// then POST /token to exchange it.
//
// POST rather than GET, which is what changed when the API server moved onto the platform's
// authorization server. A GET there renders the login form — the answer for a browser that
// arrived with no session — and only a POST runs the SubjectAuthenticator that reads this
// bearer token. The query string is the same either way: the authorization parameters travel in
// the URL on both methods, so the request that issues the code is the one that was validated.
//
// PKCE is S256, and deliberately not configurable. The `plain` method this used to send is not
// accepted at all any more, and a helper that could still choose it is one that eventually
// would.
func exchangeAuthorizationCodeWithJWT(ctx context.Context, oauth2Config *oauth2.Config, jwt string) (*oauth2.Token, error) {
	state, err := random.GenerateBase64EncodedString(ctx, 32)
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	verifier := oauth2.GenerateVerifier()

	authCodeURL := oauth2Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authCodeURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	// The authorization server parses the form on every request. Without a content type it
	// reads no body at all, which is fine here — every parameter is in the query — but the
	// header keeps the request well-formed rather than accidentally acceptable.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient, err := NewNonRedirectingHTTPClient()
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Do(req) //nolint:gosec // G704: authCodeURL from OAuth config (httpServerAddress), not user-controlled
	if err != nil {
		return nil, fmt.Errorf("fetching OAuth2 code: %w", err)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			log.Println("failed to close oauth2 response body", err)
		}
	}()

	rl, err := res.Location()
	if err != nil {
		return nil, fmt.Errorf("getting location from response: %w", err)
	}

	if returnedState := rl.Query().Get("state"); returnedState != state {
		return nil, fmt.Errorf("state mismatch on oauth2 redirect: sent %q, got %q", state, returnedState)
	}

	code := rl.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("code not returned from oauth2 redirect")
	}

	oauth2Token, err := oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("exchanging OAuth2 code: %w", err)
	}

	return oauth2Token, nil
}

func BuildInsecureOAuthedGRPCClient(
	ctx context.Context,
	createdClientID,
	createdClientSecret,
	httpTestServerAddress,
	grpcServerAddress,
	token string,
) (client.Client, error) {
	oauth2Config := NewOAuth2ConfigForTestServer(createdClientID, createdClientSecret, httpTestServerAddress)

	oauth2Token, err := exchangeAuthorizationCodeWithJWT(ctx, oauth2Config, token)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&insecureOAuth{
			TokenSource: oauth2Config.TokenSource(ctx, oauth2Token),
		}),
	}

	c, err := client.BuildClient(grpcServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("building client: %w", err)
	}

	return c, nil
}

// Custom insecure OAuth2 credentials that skip security checks.
type insecureOAuth struct {
	TokenSource oauth2.TokenSource
}

func (i *insecureOAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	token, err := i.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	return map[string]string{"authorization": token.Type() + " " + token.AccessToken}, nil
}

func (i *insecureOAuth) RequireTransportSecurity() bool {
	return false // Explicitly allow insecure transport
}

func FetchLoginTokenForUser(ctx context.Context, grpcServerAddr string, loginInput *authsvc.UserLoginInput) (string, error) {
	unauthedClient, err := client.BuildUnauthenticatedGRPCClient(grpcServerAddr)
	if err != nil {
		return "", fmt.Errorf("initializing client: %w", err)
	}

	return FetchLoginTokenForUserWithClient(ctx, unauthedClient, loginInput)
}

// FetchLoginTokenForUserWithClient calls LoginForToken using the given client.
// Use this when the client must use TLS (e.g. for api.dinnerdonebetter.com).
func FetchLoginTokenForUserWithClient(ctx context.Context, c client.Client, loginInput *authsvc.UserLoginInput) (string, error) {
	tokenRes, err := c.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
		Input: loginInput,
	})
	if err != nil {
		return "", fmt.Errorf("fetching login token: %w", err)
	}

	return tokenRes.Result.AccessToken, nil
}

// FetchOAuth2TokenForUser performs the OAuth2 authorization code flow using the given JWT
// and returns the OAuth2 access and refresh tokens. Used by integration tests for token revocation.
func FetchOAuth2TokenForUser(
	ctx context.Context,
	httpServerAddress, grpcServerAddress, clientID, clientSecret string,
	loginInput *authsvc.UserLoginInput,
) (*oauth2.Token, error) {
	jwt, err := FetchLoginTokenForUser(ctx, grpcServerAddress, loginInput)
	if err != nil {
		return nil, fmt.Errorf("fetching JWT for OAuth2 exchange: %w", err)
	}

	return exchangeAuthorizationCodeWithJWT(ctx, NewOAuth2ConfigForTestServer(clientID, clientSecret, httpServerAddress), jwt)
}
