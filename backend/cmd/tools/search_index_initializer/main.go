/*
Command search-index-initializer loads every search index from the database it is derived from.

It is the scheduler's rebuild, run once by hand: it registers the same searchsync.Registry the
scheduler and the async handler build, through the same registrars, and walks every index in it
with ReindexAll. That is the point of it being thin: an index this tool built and an index the
scheduler rebuilt should be byte-for-byte the same, and the only way to be sure of that is for
both to run the same code over the same Source. An index added to a domain is rebuilt by this
tool without this tool changing.

What it was before platform-go v10 was 400 lines of its own pagination, batching and
row-to-document conversion, one branch per index — a second implementation of the indexing
pipeline that could drift from the real one. What it was until #1469 was a switch over the
eight meal planning index names, one arm per generic instantiation — a third list of the
domain's indexes, beside the registrar's and the rule table's, that the domain could grow past.

Two flags it used to take are gone with the switch, because platform's Registry offers neither:
--indices, which rebuilt a subset by name, and --batch-size, which sized the walk. Every index is
rebuilt, at searchsync.DefaultReindexBatchSize. platform-go#1166 asks for both on the Registry.
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/search/text/algolia"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// shutdownTimeout bounds releasing the container: the stamp buffers' last flush and the
// database pool.
const shutdownTimeout = 30 * time.Second

func main() {
	var (
		databaseURL    string
		searchProvider string
		algoliaAppID   string
		algoliaAPIKey  string
	)

	root := &cobra.Command{
		Use:   "search-index-initializer",
		Short: "Initialize search indices from database (for use with proxied production DB)",
	}

	root.PersistentFlags().StringVar(&databaseURL, "database-url", "", "Postgres connection URL (or set DATABASE_URL)")
	root.PersistentFlags().StringVar(&searchProvider, "search-provider", textsearchcfg.AlgoliaProvider, "Search provider: algolia or elasticsearch")
	root.PersistentFlags().StringVar(&algoliaAppID, "algolia-app-id", "", "Algolia app ID (or set ALGOLIA_APP_ID)")
	root.PersistentFlags().StringVar(&algoliaAPIKey, "algolia-api-key", "", "Algolia API key (or set ALGOLIA_API_KEY)")

	root.AddCommand(initCmd(&databaseURL, &searchProvider, &algoliaAppID, &algoliaAPIKey))

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCmd(databaseURL, searchProvider, algoliaAppID, algoliaAPIKey *string) *cobra.Command {
	var wipe bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Load all data from database into every search index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd.Context(), *databaseURL, *searchProvider, *algoliaAppID, *algoliaAPIKey, wipe)
		},
	}

	cmd.Flags().BoolVar(&wipe, "wipe", false, "Wipe every index before reindexing")

	return cmd
}

func runInit(ctx context.Context, databaseURL, searchProvider, algoliaAppID, algoliaAPIKey string, wipe bool) error {
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		return fmt.Errorf("--database-url or DATABASE_URL is required")
	}

	if algoliaAppID == "" {
		algoliaAppID = os.Getenv("ALGOLIA_APP_ID")
	}
	if algoliaAPIKey == "" {
		algoliaAPIKey = os.Getenv("ALGOLIA_API_KEY")
	}
	if searchProvider == textsearchcfg.AlgoliaProvider && (algoliaAppID == "" || algoliaAPIKey == "") {
		return fmt.Errorf("--algolia-app-id and --algolia-api-key (or env vars) are required for Algolia")
	}

	dbConfig := &databasecfg.Config{
		Provider:        databasecfg.ProviderPostgres,
		MaxPingAttempts: 10,
		PingWaitPeriod:  time.Second,
	}
	if err := dbConfig.LoadConnectionDetailsFromURL(databaseURL); err != nil {
		return fmt.Errorf("loading database config: %w", err)
	}
	dbConfig.WriteConnection = dbConfig.ReadConnection

	searchCfg := &textsearchcfg.Config{
		Provider: searchProvider,
		Algolia: &algolia.Config{
			AppID:  algoliaAppID,
			APIKey: algoliaAPIKey,
		},
	}

	i := buildInjector(ctx, dbConfig, searchCfg)

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()

		if report := i.ShutdownWithContext(shutdownCtx); report != nil && !report.Succeed {
			log.Printf("releasing the container: %v", *report)
		}
	}()

	// Resolved rather than must-resolved: a registrar that cannot build its index — a
	// repository it needs missing, a search backend it cannot reach — is the error worth
	// reporting, and this is the one place it surfaces.
	registry, err := do.Invoke[*searchindexes.Registry](i)
	if err != nil {
		return fmt.Errorf("building the search index registry: %w", err)
	}

	if wipe {
		// Before the walk, not after: a wipe that ran afterwards would empty the index it
		// had just filled, and one that failed halfway leaves the reindex to refill it.
		if err = wipeAll(ctx, i, searchCfg, registry.Names()); err != nil {
			return err
		}
	}

	results, err := registry.ReindexAll(ctx)
	for name, result := range results {
		log.Printf("%s: scanned %d, upserted %d, in %d batches", name, result.Scanned, result.Upserted, result.Batches)
	}

	if err != nil {
		return fmt.Errorf("reindexing: %w", err)
	}

	return nil
}

// buildInjector composes what the registrars resolve: the database, the recording spine the
// repositories are built over, the stores the index sources read from, and the text index
// clients they write to.
//
// It is the same registration order the scheduler's BuildInjector runs, over hand-built pillars
// rather than a config file, because this tool runs against a database no config file names.
func buildInjector(ctx context.Context, dbConfig *databasecfg.Config, searchCfg *textsearchcfg.Config) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue[logging.Logger](i, loggingnoop.NewLogger())
	do.ProvideValue[tracing.Provider](i, tracingnoop.NewTracerProvider())
	do.ProvideValue[metrics.Provider](i, metricsnoop.NewMetricsProvider())
	do.ProvideValue(i, dbConfig)
	do.ProvideValue(i, searchCfg)

	postgres.RegisterDatabaseClient(i)

	// The repositories record through the spine even though nothing here writes a row: a
	// repository is built over its emitter and recorder whether or not a given process
	// exercises them. The spine itself is assembled by the domain registration below, as it
	// is in every other process; what it needs from here is the audit recorder behind it.
	auditlogentries.RegisterAuditLog(i)
	auditlogentries.RegisterPlatformRecorder(i)
	uploadedmedia.RegisterUploadedMediaRepository(i)
	identitystore.RegisterIdentityStore(i)

	identityindexing.RegisterSearchers(i)
	// Domain: mealplanning
	mealplanningregistration.RegisterForSearchIndexInitializer(i)

	searchindexes.Register(ctx, i,
		identityindexing.RegisterIndexes,
		// Domain: mealplanning
		mealplanningregistration.RegisterIndexes,
	)

	return i
}

// wipeAll empties every index the registry names.
//
// The index is opened by name with no document type, because wiping is the one operation on an
// index that does not read or write a document: the type parameter only decides how a search
// result decodes, and nothing here searches. The clients the registrars built are typed per
// index and resolvable only by that type, which a loop over names cannot spell.
func wipeAll(ctx context.Context, i do.Injector, searchCfg *textsearchcfg.Config, names []string) error {
	for _, name := range names {
		index, err := textsearchcfg.NewIndex[struct{}](ctx, searchCfg, name,
			textsearchcfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			textsearchcfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			textsearchcfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
		if err != nil {
			return fmt.Errorf("opening %s index to wipe it: %w", name, err)
		}

		if err = index.Wipe(ctx); err != nil {
			return fmt.Errorf("wiping %s index: %w", name, err)
		}

		log.Printf("%s: wiped", name)
	}

	return nil
}
