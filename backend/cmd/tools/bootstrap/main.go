package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"

	platformidentity "github.com/primandproper/platform-go/v14/identity"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/database"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/secrets/kubernetes"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Placeholder TOTP secret for bootstrap admin (2FA is marked verified without real TOTP).
	twoFactorSecretPlaceholder = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	// Prod Kubernetes secret coordinates.
	prodNamespace  = "prod"
	prodSecretName = "api-service-config"
	prodDBUser     = "api_db_user"
	/* #nosec G101 */
	prodDBPassKey = "DATABASE_API_PASSWORD"

	// prodAPIServerURL is where the production API server answers, and so the redirect URI
	// every first-party OAuth2 client sends. Overridable with --api-server-url, because a
	// bootstrap run against anything other than prod is registering a different address.
	prodAPIServerURL = "https://http-api." + branding.PublicDomain
)

// dbFlags holds database connection flags shared across subcommands.
type dbFlags struct {
	host       string
	user       string
	password   string
	name       string
	kubeconfig string
	port       uint16
	sslDisable bool
	prod       bool
}

func main() {
	var db dbFlags

	root := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap tooling for database initialization",
		Long: `Bootstrap tooling for initializing an empty database.

The --prod flag fetches DB connection details (host, password, OAuth2 token
encryption key) from the "api-service-config" Kubernetes secret in the "prod"
namespace. It uses your current kubeconfig context by default. NOTE: You MUST
be proxied into production via 'make proxy_db' for this to work in production.

The remaining DB defaults (host=127.0.0.1, port=5434, user=api_db_user,
dbname=dinner-done-better, sslmode=disable) assume a local Cloud SQL Auth
Proxy is running. Override any of them with the corresponding --db-* flag.

Quick start (prod, via Cloud SQL Auth Proxy):

  bootstrap --prod init --username=you --password="hunter2" --email="you@example.com"

Local dev (explicit credentials):

  bootstrap --db-password=localpass init --username=admin --password=admin123`,
	}

	root.PersistentFlags().StringVar(&db.host, "db-host", "127.0.0.1", "Postgres host")
	root.PersistentFlags().Uint16Var(&db.port, "db-port", 5434, "Postgres port")
	root.PersistentFlags().StringVar(&db.user, "db-user", prodDBUser, "Postgres username")
	root.PersistentFlags().StringVar(&db.password, "db-password", "", "Postgres password")
	root.PersistentFlags().StringVar(&db.name, "db-name", branding.CompanySlug, "Postgres database name")
	root.PersistentFlags().BoolVar(&db.sslDisable, "db-ssl-disable", true, "Disable SSL for DB connection (default: true for local/proxy)")
	root.PersistentFlags().BoolVar(&db.prod, "prod", false, "Fetch DB connection details from prod Kubernetes secrets")
	root.PersistentFlags().StringVar(&db.kubeconfig, "kubeconfig", "", "Path to kubeconfig file (defaults to ~/.kube/config)")

	// init subcommand
	var (
		adminUsername string
		adminPassword string
		adminEmail    string
		apiServerURL  string
	)

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create admin user and OAuth2 clients",
		Long: `Creates an admin user (with 2FA pre-verified) and OAuth2 clients for
Admin Webapp, Consumer Webapp, iOS App, and MCP Server.

Idempotent: safe to run multiple times. Existing users, roles, and OAuth2
clients are detected by name and skipped.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := resolveDBFlags(cmd, &db); err != nil {
				return err
			}
			return runInit(&db, adminUsername, adminPassword, adminEmail, apiServerURL)
		},
	}

	initCmd.Flags().StringVar(&adminUsername, "username", "", "Admin username to create")
	initCmd.Flags().StringVar(&adminPassword, "password", "", "Admin password (will be hashed with Argon2)")
	initCmd.Flags().StringVar(&adminEmail, "email", "", "Admin email (defaults to <username>@bootstrap.local)")
	initCmd.Flags().StringVar(&apiServerURL, "api-server-url", prodAPIServerURL, "Public URL of the API server, registered as every bootstrapped OAuth2 client's redirect URI")

	for _, flag := range []string{"username", "password"} {
		if err := initCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalln(err)
		}
	}

	root.AddCommand(initCmd)

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}

// resolveDBFlags applies --prod defaults and validates that all DB fields are set.
func resolveDBFlags(cmd *cobra.Command, db *dbFlags) error {
	if db.prod {
		kubecfgPath := db.kubeconfig
		if kubecfgPath == "" {
			kubecfgPath = clientcmd.RecommendedHomeFile
		}

		secrets, err := fetchProdSecrets(cmd.Context(), kubecfgPath)
		if err != nil {
			return fmt.Errorf("fetching prod secrets: %w", err)
		}

		if !cmd.Flags().Changed("db-password") {
			db.password = secrets.dbPassword
		}
	}

	if db.host == "" || db.user == "" || db.password == "" || db.name == "" {
		return errors.New("database connection requires --db-host, --db-user, --db-password, --db-name (or use --prod)")
	}

	return nil
}

func runInit(db *dbFlags, adminUsername, adminPassword, adminEmail, apiServerURL string) error {
	if adminEmail == "" {
		adminEmail = adminUsername + "@bootstrap.local"
	}

	ctx := context.Background()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	connDetails := databasecfg.ConnectionDetails{
		Host:       db.host,
		Port:       db.port,
		Username:   db.user,
		Password:   db.password,
		Database:   db.name,
		DisableSSL: db.sslDisable,
	}

	clientConfig := &bootstrapClientConfig{connDetails: connDetails}
	client, err := postgres.NewDatabaseClient(ctx, clientConfig, postgres.WithLogger(logger), postgres.WithTracerProvider(tracerProvider))
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Error("closing database client", closeErr)
		}
	}()

	// Pinging is a driver feature off the executor seam, so it needs the concrete pool behind
	// the RawAccess capability rather than the safe Client surface. platform-go v10 has the
	// provider constructor return its own concrete type, so this is a static conversion now
	// rather than a checked assertion — the compiler enforces what the check used to.
	var raw database.RawAccess = client
	if err = raw.ReadDB().PingContext(ctx); err != nil {
		return fmt.Errorf("pinging database client: %w", err)
	}

	// The directory, without hooks: bootstrap runs before there is anybody to attribute
	// a registration to, and an audit entry naming nobody is noise in a log whose value
	// is attribution. The API server's registrations are recorded.
	directory, identityStore, err := localdev.IdentityDirectory(logger, tracerProvider, client)
	if err != nil {
		return fmt.Errorf("building identity directory: %w", err)
	}
	oauthStore, err := platformoauth2clients.NewSQLStore(client, platformoauth2clients.WithTablePrefix(oauth.TablePrefix))
	if err != nil {
		return fmt.Errorf("building OAuth2 client store: %w", err)
	}

	oauthRegistry, err := platformoauth2clients.NewService(client, oauthStore)
	if err != nil {
		return fmt.Errorf("building OAuth2 client registry: %w", err)
	}

	// --- Admin user (idempotent) ---
	user, err := identityStore.GetUserByUsername(ctx, client.Reader(), ddbidentity.Scope(), adminUsername)
	if err != nil {
		hasher := authentication.NewArgon2Authenticator(argon2.WithLogger(logger), argon2.WithTracerProvider(tracerProvider))
		hashedPassword, hashErr := hasher.HashPassword(ctx, adminPassword)
		if hashErr != nil {
			return fmt.Errorf("hashing password: %w", hashErr)
		}

		// Registered rather than inserted: the user, the account they own and the
		// membership between them are one transaction, which is what makes a half-made
		// administrator unrepresentable rather than merely unlikely.
		registration, registerErr := directory.Register(ctx, ddbidentity.Scope(), &platformidentity.User{
			ID:              identifiers.New(),
			Username:        strings.TrimSpace(adminUsername),
			EmailAddress:    strings.TrimSpace(strings.ToLower(adminEmail)),
			FirstName:       "Admin",
			HashedPassword:  hashedPassword,
			TwoFactorSecret: twoFactorSecretPlaceholder,
			ServiceRoles:    []string{authorization.ServiceUserRoleName},
		}, &platformidentity.Account{
			Name: "Bootstrap account",
		}, []string{authorization.AccountAdminRoleName})
		if registerErr != nil {
			return fmt.Errorf("creating user: %w", registerErr)
		}

		user = registration.User
		fmt.Printf("Admin user %q created.\n", adminUsername)
	} else {
		fmt.Printf("Admin user %q already exists, skipping creation.\n", adminUsername)
	}

	// --- Service admin role (idempotent) ---
	//
	// Through the operation that exists for it rather than through two statements against
	// a role-assignment table this application no longer owns. It replaces rather than
	// merges, which is why the archival of the old row has gone with the insert: setting
	// the set is one write.
	if !slices.Contains(user.ServiceRoles, authorization.ServiceAdminRoleName) {
		if user, err = directory.SetUserServiceRoles(ctx, ddbidentity.Scope(), user.ID,
			[]string{authorization.ServiceAdminRoleName}); err != nil {
			return fmt.Errorf("promoting user to admin: %w", err)
		}
		fmt.Println("Promoted user to service_admin.")
	} else {
		fmt.Println("User already has service_admin role, skipping promotion.")
	}

	// --- 2FA verification (idempotent) ---
	if user.TwoFactorSecretVerifiedAt == nil {
		if _, err = directory.MarkUserTwoFactorSecretVerified(ctx, ddbidentity.Scope(), user.ID); err != nil {
			return fmt.Errorf("marking 2FA as verified: %w", err)
		}
		fmt.Println("Marked 2FA as verified.")
	} else {
		fmt.Println("2FA already verified, skipping.")
	}

	// --- OAuth2 clients (idempotent) ---
	// Every first-party client sends the API server's own address as its redirect_uri and reads
	// the authorization code off the Location header rather than following it — there is no
	// callback endpoint, and never was. Registering that address is therefore registering what
	// the clients actually send.
	//
	// It matters that this is exact. The authorization server compares redirect_uri byte for
	// byte, at /authorize and again at /token against the URI the code was issued for, so a
	// trailing slash or a missing port here is a client that authenticates at the token
	// endpoint and then fails every authorization request.
	redirectURIs := []string{apiServerURL}

	wantClients := []*struct {
		name         string
		desc         string
		redirectURIs []string
	}{
		{"Admin Webapp", "Admin web application OAuth2 client", redirectURIs},
		{"Consumer Webapp", "Consumer web application OAuth2 client", redirectURIs},
		{"iOS App", "iOS mobile application OAuth2 client", redirectURIs},
		{"MCP Server", "MCP server OAuth2 client", redirectURIs},
	}

	existingClients, err := oauthStore.ListClients(ctx, client.Reader(), tenancy.Global(), nil)
	if err != nil {
		return fmt.Errorf("listing existing OAuth2 clients: %w", err)
	}

	existingByName := make(map[string]*platformoauth2clients.Client)
	for _, c := range existingClients.Data {
		existingByName[c.Name] = c
	}

	fmt.Println()
	fmt.Println("OAuth2 clients:")
	for _, want := range wantClients {
		if existing, ok := existingByName[want.name]; ok {
			fmt.Printf("  %s: already exists (client_id=%s)\n", want.name, existing.ClientID)
			continue
		}

		// Global and unowned: an operator mints these to let four applications speak for
		// the service on behalf of whoever signs in, which is the registry
		// oauth2clients.Client.Admits lets any subject through.
		issued, creationErr := oauthRegistry.CreateClient(ctx, tenancy.Global(), "", &platformoauth2clients.CreationInput{
			Name:         want.name,
			Description:  want.desc,
			RedirectURIs: want.redirectURIs,
		})
		if creationErr != nil {
			return fmt.Errorf("creating OAuth2 client %s: %w", want.name, creationErr)
		}
		// print the plaintext secret: this is the only time it is recoverable.
		fmt.Printf("  %s: created (client_id=%s client_secret=%s)\n", want.name, issued.Client.ClientID, issued.Secret)
	}

	fmt.Println()
	fmt.Println("Bootstrap init complete.")

	return nil
}

// bootstrapClientConfig implements database.ClientConfig for bootstrap.
type bootstrapClientConfig struct {
	connDetails databasecfg.ConnectionDetails
}

var _ database.ClientConfig = (*bootstrapClientConfig)(nil)

func (b *bootstrapClientConfig) GetReadConnectionString() string {
	s := b.connDetails.String()
	if b.connDetails.DisableSSL {
		s += " sslmode=disable"
	}
	return s
}

func (b *bootstrapClientConfig) GetWriteConnectionString() string {
	return b.GetReadConnectionString()
}

func (b *bootstrapClientConfig) GetMaxPingAttempts() uint64 {
	return 10
}

func (b *bootstrapClientConfig) GetPingWaitPeriod() time.Duration {
	return time.Second
}

func (b *bootstrapClientConfig) GetMaxIdleConns() int {
	return 5
}

func (b *bootstrapClientConfig) GetMaxOpenConns() int {
	return 7
}

func (b *bootstrapClientConfig) GetConnMaxLifetime() time.Duration {
	return 30 * time.Minute
}

type prodSecrets struct {
	dbPassword string
}

func fetchProdSecrets(ctx context.Context, kubeconfigPath string) (*prodSecrets, error) {
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()
	metricsProvider := metricsnoop.NewMetricsProvider()

	cfg := &kubernetes.Config{
		Namespace:  prodNamespace,
		Kubeconfig: kubeconfigPath,
	}

	secretSource, err := kubernetes.NewSecretSource(ctx, cfg, nil,
		kubernetes.WithLogger(logger),
		kubernetes.WithTracerProvider(tracerProvider),
		kubernetes.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes secret source: %w", err)
	}
	defer func() {
		if err = secretSource.Close(); err != nil {
			logger.Error("closing secret source", err)
		}
	}()

	var s prodSecrets

	s.dbPassword, err = secretSource.GetSecret(ctx, prodSecretName+"/"+prodDBPassKey)
	if err != nil {
		return nil, fmt.Errorf("fetching %s/%s: %w", prodSecretName, prodDBPassKey, err)
	}

	return &s, nil
}
