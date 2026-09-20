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
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformsettings "github.com/primandproper/platform-go/v14/settings"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/database"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/httpclient"
	"github.com/primandproper/primitives-go/v2/identifiers"
	msgconfig "github.com/primandproper/primitives-go/v2/messagequeue/config"
	"github.com/primandproper/primitives-go/v2/messagequeue/redis"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/random"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/testutils/containers/redistest"

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
	directory *platformidentity.Service,
	store platformidentity.Store,
	dbClient database.Client,
	premadeAdminUser *platformidentity.User,
) (*platformidentity.User, error) {
	hasher := authentication.NewArgon2Authenticator(argon2.WithLogger(logger), argon2.WithTracerProvider(tracerProvider))

	actuallyHashedPass, err := hasher.HashPassword(ctx, premadeAdminUser.HashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	premadeAdminUser.HashedPassword = actuallyHashedPass

	if existing, lookupErr := store.GetUserByUsername(ctx, dbClient.Reader(), ddbidentity.Scope(), premadeAdminUser.Username); lookupErr == nil && existing != nil {
		return existing, nil
	}

	// Registered rather than inserted: a user, their account and the membership that puts
	// them in it are one transaction, and the shape that rules out a user with no account
	// is the reason Service ships it.
	registration, err := directory.Register(ctx, ddbidentity.Scope(), premadeAdminUser, &platformidentity.Account{
		Name: premadeAdminUser.Username + "'s account",
	}, []string{authorization.AccountAdminRoleName})
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// The service role is a write of its own, through the operation that exists for it
	// rather than through two statements against a role-assignment table this application
	// no longer owns.
	user, err := directory.SetUserServiceRoles(ctx, ddbidentity.Scope(), registration.User.ID,
		[]string{authorization.ServiceAdminRoleName})
	if err != nil {
		return nil, fmt.Errorf("failed to promote user to service admin: %w", err)
	}

	if _, err = directory.MarkUserTwoFactorSecretVerified(ctx, ddbidentity.Scope(), user.ID); err != nil {
		return nil, fmt.Errorf("failed to mark user as verified: %w", err)
	}

	return user, nil
}

// CreateOAuth2ClientForService registers a client and hands back the secret it was issued.
//
// The credentials are the service's to mint, not the caller's: the plaintext exists on the
// IssuedClient this returns and nowhere else, because the row holds a digest and no read
// reverses it. A caller that wants a predictable credential supplies a generator — see
// oauth2ClientRegistry's option — rather than choosing the value here.
//
// The registration is global and unowned, which is what this deployment's clients are: an
// operator mints one to let an application speak for the service on behalf of whoever signs
// in, and oauth2clients.Client.Admits permits any subject for exactly that arrangement.
func CreateOAuth2ClientForService(
	ctx context.Context,
	pgc database.Client,
	input *platformoauth2clients.CreationInput,
) (*platformoauth2clients.IssuedClient, error) {
	svc, err := oauth2ClientRegistry(pgc)
	if err != nil {
		return nil, err
	}

	issued, err := svc.CreateClient(ctx, tenancy.Global(), "", input)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 client: %w", err)
	}

	return issued, nil
}

// oauth2ClientRegistry builds the registry service over a bare database client.
//
// Undecorated, unlike the one the injector builds: this is a seeding path with no session
// behind it, and an audit entry attributing a localdev client to nobody is noise in a log
// whose whole value is attribution. The API server's registrations go through
// oauth2clientsstore and are recorded.
func oauth2ClientRegistry(pgc database.Client, opts ...platformoauth2clients.ServiceOption) (*platformoauth2clients.Service, error) {
	store, err := platformoauth2clients.NewSQLStore(pgc, platformoauth2clients.WithTablePrefix(oauth.TablePrefix))
	if err != nil {
		return nil, err
	}

	return platformoauth2clients.NewService(pgc, store, opts...)
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
		uploads, err := UploadsRegistry(logger, tracerProvider, dbClient)
		if err != nil {
			return err
		}
		policy, policyErr := authorization.NewDatabaseResolver(dbClient.Reader(), logger, tracerProvider, nil)
		if policyErr != nil {
			return policyErr
		}

		identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditLogRepo, dbClient, nil, uploads, policy)
		return fn(ctx, identityRepo, logger, tracerProvider, dbClient)
	}
}

// WithOAuth2Registry provides the client registry for custom operations.
//
// The generator is the caller's, because the one caller there is seeds a well-known
// credential: localdev's client_id and secret are in checked-in configuration and in the
// web apps' environment, so a minted one would mean nothing could sign in until somebody
// copied it out of the database. platform supplies WithCredentialGenerator for exactly
// this, and the seam is the option rather than a write that bypasses the service.
func WithOAuth2Registry(
	generate platformoauth2clients.CredentialGenerator,
	fn func(ctx context.Context, svc *platformoauth2clients.Service, logger logging.Logger, tracerProvider tracing.Provider) error,
) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, _ *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		svc, err := oauth2ClientRegistry(dbClient,
			platformoauth2clients.WithCredentialGenerator(generate),
			platformoauth2clients.WithServiceLogger(logger),
			platformoauth2clients.WithServiceTracerProvider(tracerProvider),
		)
		if err != nil {
			return err
		}

		return fn(ctx, svc, logger, tracerProvider)
	}
}

// WithPasswordResetTokenStore provides the password reset token store for custom operations.
// The provided function receives a fully configured passwordreset.Store along with logger and tracer.
func WithPasswordResetTokenStore(fn func(ctx context.Context, store passwordreset.Store, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		store, err := authrepo.ProvidePasswordResetTokenStore(logger, tracerProvider, auditLogRepo, dbClient)
		if err != nil {
			return err
		}
		return fn(ctx, store, logger, tracerProvider)
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
		uploads, err := UploadsRegistry(logger, tracerProvider, dbClient)
		if err != nil {
			return err
		}
		identityStore, storeErr := platformidentity.NewSQLStore(dbClient,
			platformidentity.WithTablePrefix(ddbidentity.TablePrefix),
			platformidentity.WithStoreLogger(logger),
			platformidentity.WithStoreTracerProvider(tracerProvider),
		)
		if storeErr != nil {
			return storeErr
		}

		mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditLogRepo, identityStore, dbClient, nil, uploads)
		return fn(ctx, mealPlanningRepo, logger, tracerProvider)
	}
}

// WithSettingsRepository provides a settings store for custom operations.
// The provided function receives a fully configured settings.Store along with logger and tracer.
//
// It also receives the database client, because as of platform-go v14 a store
// write takes the caller's transaction and there is nowhere else for a seed to
// get one. WithIdentityRepository already took it for the same reason.
func WithSettingsRepository(fn func(ctx context.Context, store platformsettings.Store, logger logging.Logger, tracerProvider tracing.Provider, dbClient database.Client) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}

		settingsStore, err := settingsrepo.ProvideSettingsRepository(ctx, logger, tracerProvider, metricsnoop.NewMetricsProvider(), auditLogRepo, dbClient, nil)
		if err != nil {
			return err
		}

		return fn(ctx, settingsStore, logger, tracerProvider, dbClient)
	}
}

// WithWebhooksRepository is gone with the repository it provided.
//
// Nothing called it: it existed so a localdev hook could write webhooks
// directly, and the endpoints are platform's now. A hook that wants one builds
// webhooksstore.RegisterWebhooksStore's dependencies, or asks the API.

// WithNotificationsRepository provides a notifications repository for custom operations.
// The provided function receives a fully configured notifications.Repository along with logger and tracer.
func WithNotificationsRepository(fn func(ctx context.Context, repo notifications.Repository, logger logging.Logger, tracerProvider tracing.Provider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *dbcfg.Config, logger logging.Logger, tracerProvider tracing.Provider) error {
		auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, dbClient)
		if err != nil {
			return err
		}
		notificationsRepo, err := notificationsstore.ProvideAdapter(ctx, logger, tracerProvider, metricsnoop.NewMetricsProvider(), auditLogRepo, nil, dbClient)
		if err != nil {
			return err
		}
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

	res, err := httpClient.Do(req)
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
