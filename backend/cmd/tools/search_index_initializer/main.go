/*
Command search-index-initializer loads a search index from the database it is derived from.

It is a thin driver over searchsync.Reindexer, which is the same walk the scheduler runs on a
timer. That is the point of it being thin: an index this tool built and an index the scheduler
rebuilt should be byte-for-byte the same, and the only way to be sure of that is for both to run
the same code over the same Scanner.

What it was before platform-go v10 was 400 lines of its own pagination, batching and
row-to-document conversion, one branch per index — a second implementation of the indexing
pipeline that could drift from the real one, and did not share its ordering guarantees.
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/database/postgres"
	"github.com/primandproper/platform-go/v13/observability/logging"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	searchsync "github.com/primandproper/platform-go/v13/search/sync"
	syncsource "github.com/primandproper/platform-go/v13/search/sync/source"
	"github.com/primandproper/platform-go/v13/search/text/algolia"
	textsearchcfg "github.com/primandproper/platform-go/v13/search/text/config"
	"github.com/primandproper/platform-go/v13/uploads/registry"

	"github.com/spf13/cobra"
)

const defaultBatchSize = 50

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
	var indices string
	var wipe bool
	var batchSize int

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Load all data from database into search indices",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInit(*databaseURL, *searchProvider, *algoliaAppID, *algoliaAPIKey, indices, wipe, batchSize)
		},
	}

	cmd.Flags().StringVar(&indices, "indices", "", "Comma-separated indices to initialize (e.g. recipes,meals,users)")
	cmd.Flags().BoolVar(&wipe, "wipe", false, "Wipe index before reindexing")
	cmd.Flags().IntVar(&batchSize, "batch-size", defaultBatchSize, "Page size for the keyset walk")

	if err := cmd.MarkFlagRequired("indices"); err != nil {
		log.Fatal(err)
	}

	return cmd
}

func runInit(databaseURL, searchProvider, algoliaAppID, algoliaAPIKey, indicesStr string, wipe bool, batchSize int) error {
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

	var requested []string
	for idx := range strings.SplitSeq(strings.TrimSpace(indicesStr), ",") {
		if s := strings.TrimSpace(idx); s != "" {
			requested = append(requested, s)
		}
	}
	if len(requested) == 0 {
		return fmt.Errorf("at least one index is required in --indices")
	}

	if batchSize < 1 {
		batchSize = defaultBatchSize
	}

	ctx := context.Background()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()
	metricsProvider := metricsnoop.NewMetricsProvider()

	dbConfig := &databasecfg.Config{
		Provider:        databasecfg.ProviderPostgres,
		MaxPingAttempts: 10,
		PingWaitPeriod:  time.Second,
	}
	if err := dbConfig.LoadConnectionDetailsFromURL(databaseURL); err != nil {
		return fmt.Errorf("loading database config: %w", err)
	}
	dbConfig.WriteConnection = dbConfig.ReadConnection

	client, err := postgres.NewDatabaseClient(ctx, dbConfig, postgres.WithLogger(logger), postgres.WithTracerProvider(tracerProvider))
	if err != nil {
		return fmt.Errorf("initializing database client: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			logger.Error("closing database client", closeErr)
		}
	}()

	auditRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, nil, client)
	if err != nil {
		return fmt.Errorf("building audit log repository: %w", err)
	}

	// A real registry store rather than nil: both repositories hydrate media
	// through it. It needs no emitter or metrics — nothing here writes an object.
	uploadsRegistry, err := registry.NewSQLStore(
		client,
		registry.WithTablePrefix(uploadedmedia.TablePrefix),
		registry.WithStoreLogger(logger),
		registry.WithStoreTracerProvider(tracerProvider),
	)
	if err != nil {
		return fmt.Errorf("building upload registry store: %w", err)
	}

	policy, err := authorization.NewDatabaseResolver(client.Reader(), logger, tracerProvider, nil)
	if err != nil {
		return fmt.Errorf("building authorization policy resolver: %w", err)
	}

	identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditRepo, client, nil, uploadsRegistry, policy)
	mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditRepo, identityRepo, client, nil, uploadsRegistry)

	searchCfg := &textsearchcfg.Config{
		Provider: searchProvider,
		Algolia: &algolia.Config{
			AppID:  algoliaAppID,
			APIKey: algoliaAPIKey,
		},
	}

	o11y := observability{logger: logger, tracerProvider: tracerProvider, metricsProvider: metricsProvider}

	for _, indexType := range requested {
		if err = reindexOne(ctx, indexType, searchCfg, identityRepo, mealPlanningRepo, o11y, wipe, batchSize); err != nil {
			return fmt.Errorf("indexing %s: %w", indexType, err)
		}
	}

	return nil
}

// observability bundles the three providers every index build needs, so the per-index
// constructors below take one argument for them rather than three.
type observability struct {
	logger          logging.Logger
	tracerProvider  tracing.Provider
	metricsProvider metrics.Provider
}

// reindexOne builds the index named by indexType and walks its source into it.
//
// The switch is over index names rather than a map because each arm resolves a different
// generic instantiation, and a map would need them all erased to a common type before the
// compiler could check any of them against the index it writes to.
func reindexOne(
	ctx context.Context,
	indexType string,
	searchCfg *textsearchcfg.Config,
	identityRepo identity.Repository,
	mealPlanningRepo mealplanning.Repository,
	o11y observability,
	wipe bool,
	batchSize int,
) error {
	// Each arm names its constructor and then builds, rather than nesting the two, because
	// syncsource.New returns an error and Go will only expand a multi-valued call into a call
	// that takes nothing else.
	switch indexType {
	case identityindexing.IndexTypeUsers:
		source, err := identityindexing.UserSource(identityRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeMeals:
		source, err := mealplanningindexing.NewMealSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeRecipes:
		source, err := mealplanningindexing.NewRecipeSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeValidIngredients:
		source, err := mealplanningindexing.NewValidIngredientSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeValidInstruments:
		source, err := mealplanningindexing.NewValidInstrumentSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeValidMeasurementUnits:
		source, err := mealplanningindexing.NewValidMeasurementUnitSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeValidPreparations:
		source, err := mealplanningindexing.NewValidPreparationSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeValidIngredientStates:
		source, err := mealplanningindexing.NewValidIngredientStateSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	case mealplanningindexing.IndexTypeValidVessels:
		source, err := mealplanningindexing.NewValidVesselSource(mealPlanningRepo)
		if err != nil {
			return err
		}

		return build(ctx, searchCfg, source, o11y, wipe, batchSize)
	default:
		return fmt.Errorf("unknown index type %q, expected one of %s", indexType, strings.Join(knownIndexTypes(), ", "))
	}
}

// build opens the index this Source feeds and walks the whole source into it.
func build[E, T any](
	ctx context.Context,
	searchCfg *textsearchcfg.Config,
	source *syncsource.Source[E, T],
	o11y observability,
	wipe bool,
	batchSize int,
) error {
	index, err := textsearchcfg.NewIndex[T](
		ctx,
		searchCfg,
		source.Name(),
		textsearchcfg.WithLogger(o11y.logger),
		textsearchcfg.WithTracerProvider(o11y.tracerProvider),
		textsearchcfg.WithMetricsProvider(o11y.metricsProvider),
	)
	if err != nil {
		return fmt.Errorf("building index: %w", err)
	}

	if wipe {
		// Before the walk, not after: a wipe that ran afterwards would empty the index it
		// had just filled, and one that failed halfway leaves the reindex to refill it.
		if err = index.Wipe(ctx); err != nil {
			return fmt.Errorf("wiping index: %w", err)
		}
	}

	reindexer, err := syncsource.NewReindexer(source, index,
		syncsource.WithLogger(o11y.logger),
		syncsource.WithTracerProvider(o11y.tracerProvider),
		syncsource.WithMetricsProvider(o11y.metricsProvider),
		syncsource.WithReindexOptions(searchsync.WithReindexBatchSize(batchSize)),
	)
	if err != nil {
		return err
	}

	result, err := reindexer.Reindex(ctx)
	if err != nil {
		return fmt.Errorf("reindexing: %w", err)
	}

	log.Printf("%s: scanned %d, upserted %d, in %d batches", source.Name(), result.Scanned, result.Upserted, result.Batches)

	return nil
}

func knownIndexTypes() []string {
	types := []string{
		identityindexing.IndexTypeUsers,
		mealplanningindexing.IndexTypeMeals,
		mealplanningindexing.IndexTypeRecipes,
		mealplanningindexing.IndexTypeValidIngredients,
		mealplanningindexing.IndexTypeValidInstruments,
		mealplanningindexing.IndexTypeValidMeasurementUnits,
		mealplanningindexing.IndexTypeValidPreparations,
		mealplanningindexing.IndexTypeValidIngredientStates,
		mealplanningindexing.IndexTypeValidVessels,
	}
	slices.Sort(types)

	return types
}
