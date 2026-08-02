package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	"github.com/primandproper/platform-go/v9/database/postgres"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability/logging"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
	"github.com/primandproper/platform-go/v9/search/text/algolia"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"

	"github.com/spf13/cobra"
)

const (
	defaultBatchSize = 50
)

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
	cmd.Flags().IntVar(&batchSize, "batch-size", defaultBatchSize, "Page size for cursor pagination")

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

	indices := strings.Split(strings.TrimSpace(indicesStr), ",")
	var trimmed []string
	for _, idx := range indices {
		if s := strings.TrimSpace(idx); s != "" {
			trimmed = append(trimmed, s)
		}
	}
	indices = trimmed
	if len(indices) == 0 {
		return fmt.Errorf("at least one index is required in --indices")
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

	auditRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, client)
	identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditRepo, client, nil)
	mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditRepo, identityRepo, client, nil)

	searchCfg := &textsearchcfg.Config{
		Provider: searchProvider,
		Algolia: &algolia.Config{
			AppID:  algoliaAppID,
			APIKey: algoliaAPIKey,
		},
	}

	if batchSize < 1 {
		batchSize = defaultBatchSize
	}

	mealPlanningIndexer, userIndexer, err := buildIndexers(ctx, logger, tracerProvider, metricsProvider, searchCfg, mealPlanningRepo, identityRepo)
	if err != nil {
		return fmt.Errorf("building indexers: %w", err)
	}

	for _, indexType := range indices {
		if err = runIndex(ctx, logger, indexType, mealPlanningRepo, identityRepo, mealPlanningIndexer, userIndexer, searchCfg, wipe, batchSize); err != nil {
			return fmt.Errorf("indexing %s: %w", indexType, err)
		}
	}

	return nil
}

func buildIndexers(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	searchCfg *textsearchcfg.Config,
	mealPlanningRepo mealplanning.Repository,
	identityRepo identity.Repository,
) (*mealplanningindexing.MealPlanningDataIndexer, *identityindexing.UserDataIndexer, error) {
	recipeIdx, err := textsearchcfg.NewIndex[mealplanningindexing.RecipeSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeRecipes, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	mealIdx, err := textsearchcfg.NewIndex[mealplanningindexing.MealSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeMeals, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	validIngredientIdx, err := textsearchcfg.NewIndex[mealplanningindexing.ValidIngredientSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeValidIngredients, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	validInstrumentIdx, err := textsearchcfg.NewIndex[mealplanningindexing.ValidInstrumentSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeValidInstruments, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	validMeasurementUnitIdx, err := textsearchcfg.NewIndex[mealplanningindexing.ValidMeasurementUnitSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeValidMeasurementUnits, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	validPreparationIdx, err := textsearchcfg.NewIndex[mealplanningindexing.ValidPreparationSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeValidPreparations, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	validIngredientStateIdx, err := textsearchcfg.NewIndex[mealplanningindexing.ValidIngredientStateSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeValidIngredientStates, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	validVesselIdx, err := textsearchcfg.NewIndex[mealplanningindexing.ValidVesselSearchSubset](ctx, searchCfg, mealplanningindexing.IndexTypeValidVessels, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}
	userIdx, err := textsearchcfg.NewIndex[identityindexing.UserSearchSubset](ctx, searchCfg, identityindexing.IndexTypeUsers, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, err
	}

	mpIndexer := mealplanningindexing.NewMealPlanningDataIndexer(
		logger, tracerProvider, mealPlanningRepo,
		recipeIdx, mealIdx, validIngredientIdx, validInstrumentIdx,
		validMeasurementUnitIdx, validPreparationIdx, validIngredientStateIdx, validVesselIdx,
	)
	userIndexer := identityindexing.NewCoreDataIndexer(logger, tracerProvider, identityRepo, userIdx)

	return mpIndexer, userIndexer, nil
}

func runIndex(
	ctx context.Context,
	logger logging.Logger,
	indexType string,
	mealPlanningRepo mealplanning.Repository,
	identityRepo identity.Repository,
	mpIndexer *mealplanningindexing.MealPlanningDataIndexer,
	userIndexer *identityindexing.UserDataIndexer,
	searchCfg *textsearchcfg.Config,
	wipe bool,
	batchSize int,
) error {
	log.Printf("Starting index: %s", indexType)

	im, err := getIndexManager(ctx, logger, indexType, searchCfg)
	if err != nil {
		return err
	}

	if wipe && im != nil {
		log.Printf("Wiping index: %s", indexType)
		if err = im.Wipe(ctx); err != nil {
			return fmt.Errorf("wiping index: %w", err)
		}
		log.Printf("Wiped index: %s", indexType)
	}

	filter := filtering.DefaultQueryFilter()
	pageSize := min(uint16(batchSize), filtering.MaxQueryFilterLimit)
	filter.MaxResponseSize = &pageSize

	var cursor *string
	pageNum := 0
	totalIndexed := 0

	for {
		pageNum++
		filter.Cursor = cursor

		var ids []string
		switch indexType {
		case mealplanningindexing.IndexTypeRecipes:
			result, fetchErr := mealPlanningRepo.GetRecipes(ctx, mealplanning.RecipeStatusApproved, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting recipes: %w", fetchErr)
			}
			for _, r := range result.Data {
				ids = append(ids, r.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeMeals:
			result, fetchErr := mealPlanningRepo.GetMeals(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting meals: %w", fetchErr)
			}
			for _, m := range result.Data {
				ids = append(ids, m.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeValidIngredients:
			result, fetchErr := mealPlanningRepo.GetValidIngredients(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting valid ingredients: %w", fetchErr)
			}
			for _, v := range result.Data {
				ids = append(ids, v.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeValidInstruments:
			result, fetchErr := mealPlanningRepo.GetValidInstruments(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting valid instruments: %w", fetchErr)
			}
			for _, v := range result.Data {
				ids = append(ids, v.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeValidMeasurementUnits:
			result, fetchErr := mealPlanningRepo.GetValidMeasurementUnits(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting valid measurement units: %w", fetchErr)
			}
			for _, v := range result.Data {
				ids = append(ids, v.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeValidPreparations:
			result, fetchErr := mealPlanningRepo.GetValidPreparations(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting valid preparations: %w", fetchErr)
			}
			for _, v := range result.Data {
				ids = append(ids, v.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeValidIngredientStates:
			result, fetchErr := mealPlanningRepo.GetValidIngredientStates(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting valid ingredient states: %w", fetchErr)
			}
			for _, v := range result.Data {
				ids = append(ids, v.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case mealplanningindexing.IndexTypeValidVessels:
			result, fetchErr := mealPlanningRepo.GetValidVessels(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting valid vessels: %w", fetchErr)
			}
			for _, v := range result.Data {
				ids = append(ids, v.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		case identityindexing.IndexTypeUsers:
			result, fetchErr := identityRepo.GetUsers(ctx, filter)
			if fetchErr != nil {
				return fmt.Errorf("getting users: %w", fetchErr)
			}
			for _, u := range result.Data {
				ids = append(ids, u.ID)
			}
			if len(result.Data) > 0 {
				cursor = &result.Data[len(result.Data)-1].ID
			} else {
				cursor = nil
			}
		default:
			return fmt.Errorf("unknown index type: %s", indexType)
		}

		if len(ids) == 0 {
			break
		}

		for _, id := range ids {
			req := &textsearch.IndexRequest{RowID: id, IndexType: indexType}
			switch indexType {
			case mealplanningindexing.IndexTypeRecipes,
				mealplanningindexing.IndexTypeMeals,
				mealplanningindexing.IndexTypeValidIngredients,
				mealplanningindexing.IndexTypeValidInstruments,
				mealplanningindexing.IndexTypeValidMeasurementUnits,
				mealplanningindexing.IndexTypeValidPreparations,
				mealplanningindexing.IndexTypeValidIngredientStates,
				mealplanningindexing.IndexTypeValidVessels:
				if err = mpIndexer.HandleIndexRequest(ctx, req); err != nil {
					return fmt.Errorf("indexing %s %s: %w", indexType, id, err)
				}
			case identityindexing.IndexTypeUsers:
				if err = userIndexer.HandleIndexRequest(ctx, req); err != nil {
					return fmt.Errorf("indexing %s %s: %w", indexType, id, err)
				}
			}
			totalIndexed++
		}

		log.Printf("Indexed page: %s page=%d items=%d total=%d", indexType, pageNum, len(ids), totalIndexed)

		if len(ids) < batchSize {
			break
		}
	}

	log.Printf("Finished index: %s total_indexed=%d", indexType, totalIndexed)
	return nil
}

func getIndexManager(
	ctx context.Context,
	logger logging.Logger,
	indexType string,
	searchCfg *textsearchcfg.Config,
) (textsearch.IndexManager, error) {
	tracerProvider := tracingnoop.NewTracerProvider()
	metricsProvider := metricsnoop.NewMetricsProvider()

	switch indexType {
	case mealplanningindexing.IndexTypeRecipes:
		return textsearchcfg.NewIndex[mealplanningindexing.RecipeSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeMeals:
		return textsearchcfg.NewIndex[mealplanningindexing.MealSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeValidIngredients:
		return textsearchcfg.NewIndex[mealplanningindexing.ValidIngredientSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeValidInstruments:
		return textsearchcfg.NewIndex[mealplanningindexing.ValidInstrumentSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeValidMeasurementUnits:
		return textsearchcfg.NewIndex[mealplanningindexing.ValidMeasurementUnitSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeValidPreparations:
		return textsearchcfg.NewIndex[mealplanningindexing.ValidPreparationSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeValidIngredientStates:
		return textsearchcfg.NewIndex[mealplanningindexing.ValidIngredientStateSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case mealplanningindexing.IndexTypeValidVessels:
		return textsearchcfg.NewIndex[mealplanningindexing.ValidVesselSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	case identityindexing.IndexTypeUsers:
		return textsearchcfg.NewIndex[identityindexing.UserSearchSubset](ctx, searchCfg, indexType, textsearchcfg.WithLogger(logger), textsearchcfg.WithTracerProvider(tracerProvider), textsearchcfg.WithMetricsProvider(metricsProvider))
	default:
		return nil, fmt.Errorf("unknown index type: %s", indexType)
	}
}
