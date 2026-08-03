package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	postgresmigrations "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"

	"github.com/primandproper/platform-go/v9/database"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	"github.com/primandproper/platform-go/v9/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	var (
		dbHost       string
		dbPort       uint16
		dbUser       string
		dbPassword   string
		dbName       string
		dbSSLDisable bool
	)

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Connects to a Postgres database and applies all pending schema migrations.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMigrate(cmd.Context(), dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLDisable)
		},
	}

	cmd.Flags().StringVar(&dbHost, "db-host", "", "Postgres host (or DB_HOST)")
	cmd.Flags().Uint16Var(&dbPort, "db-port", 5432, "Postgres port (or DB_PORT)")
	cmd.Flags().StringVar(&dbUser, "db-user", "", "Postgres username (or DB_USER)")
	cmd.Flags().StringVar(&dbPassword, "db-password", "", "Postgres password (or DB_PASSWORD)")
	cmd.Flags().StringVar(&dbName, "db-name", "", "Postgres database name (or DB_NAME)")
	cmd.Flags().BoolVar(&dbSSLDisable, "db-ssl-disable", true, "Disable SSL for DB connection (default: true for local/proxy)")

	for _, flag := range []string{"db-host", "db-user", "db-password", "db-name"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			log.Fatalln(err)
		}
	}

	return cmd
}

func runMigrate(ctx context.Context, dbHost string, dbPort uint16, dbUser, dbPassword, dbName string, dbSSLDisable bool) error {
	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return errors.New("database connection requires --db-host, --db-user, --db-password, --db-name")
	}

	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	connDetails := databasecfg.ConnectionDetails{
		Host:       dbHost,
		Port:       dbPort,
		Username:   dbUser,
		Password:   dbPassword,
		Database:   dbName,
		DisableSSL: dbSSLDisable,
	}

	clientConfig := &migrateClientConfig{
		connDetails: connDetails,
	}

	client, err := postgres.NewDatabaseClient(ctx, clientConfig, postgres.WithLogger(logger), postgres.WithTracerProvider(tracerProvider))
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("closing database client: %v", closeErr)
		}
	}()

	// Migrations need the concrete *sql.DB, which lives behind the RawAccess capability
	// rather than the safe Client surface.
	raw, ok := client.(database.RawAccess)
	if !ok {
		return errors.New("database client does not expose raw access required for migrations")
	}

	migrator, err := postgresmigrations.NewMigrator(logger)
	if err != nil {
		return fmt.Errorf("building migrator: %w", err)
	}

	if err = migrator.Migrate(ctx, raw.WriteDB()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	fmt.Println("Migrations completed successfully.")

	return nil
}

type migrateClientConfig struct {
	connDetails databasecfg.ConnectionDetails
}

var _ database.ClientConfig = (*migrateClientConfig)(nil)

func (m *migrateClientConfig) GetReadConnectionString() string {
	if m.connDetails.DisableSSL {
		return m.connDetails.URI()
	}
	return m.connDetails.String()
}

func (m *migrateClientConfig) GetWriteConnectionString() string {
	return m.GetReadConnectionString()
}

func (m *migrateClientConfig) GetMaxPingAttempts() uint64 {
	return 10
}

func (m *migrateClientConfig) GetPingWaitPeriod() time.Duration {
	return time.Second
}

func (m *migrateClientConfig) GetMaxIdleConns() int {
	return 5
}

func (m *migrateClientConfig) GetMaxOpenConns() int {
	return 7
}

func (m *migrateClientConfig) GetConnMaxLifetime() time.Duration {
	return 30 * time.Minute
}
