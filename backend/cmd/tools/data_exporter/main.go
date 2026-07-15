package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	"github.com/primandproper/platform-go/v4/database"
	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	"github.com/primandproper/platform-go/v4/database/postgres"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/observability/logging"
	loggingnoop "github.com/primandproper/platform-go/v4/observability/logging/noop"
	"github.com/primandproper/platform-go/v4/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v4/observability/tracing/noop"

	"github.com/spf13/cobra"
)

// ExportData represents the full export structure.
type ExportData struct {
	ExportedAt   time.Time              `json:"exportedAt"`
	Enumerations ExportedEnumerations   `json:"enumerations"`
	Recipes      []*mealplanning.Recipe `json:"recipes"`
	Meals        []*mealplanning.Meal   `json:"meals"`
}

// ExportedEnumerations holds all valid enumeration data.
type ExportedEnumerations struct {
	ValidIngredients                []*mealplanning.ValidIngredient                `json:"validIngredients"`
	ValidPreparations               []*mealplanning.ValidPreparation               `json:"validPreparations"`
	ValidInstruments                []*mealplanning.ValidInstrument                `json:"validInstruments"`
	ValidVessels                    []*mealplanning.ValidVessel                    `json:"validVessels"`
	ValidMeasurementUnits           []*mealplanning.ValidMeasurementUnit           `json:"validMeasurementUnits"`
	ValidIngredientStates           []*mealplanning.ValidIngredientState           `json:"validIngredientStates"`
	ValidIngredientPreparations     []*mealplanning.ValidIngredientPreparation     `json:"validIngredientPreparations"`
	ValidIngredientMeasurementUnits []*mealplanning.ValidIngredientMeasurementUnit `json:"validIngredientMeasurementUnits"`
	ValidPreparationInstruments     []*mealplanning.ValidPreparationInstrument     `json:"validPreparationInstruments"`
	ValidPreparationVessels         []*mealplanning.ValidPreparationVessel         `json:"validPreparationVessels"`
	ValidIngredientGroups           []*mealplanning.ValidIngredientGroup           `json:"validIngredientGroups"`
	ValidIngredientStateIngredients []*mealplanning.ValidIngredientStateIngredient `json:"validIngredientStateIngredients"`
	ValidMeasurementUnitConversions []*mealplanning.ValidMeasurementUnitConversion `json:"validMeasurementUnitConversions"`
}

func main() {
	var (
		dbHost       string
		dbPort       uint16
		dbUser       string
		dbPassword   string
		dbName       string
		dbSSLDisable bool
		outputFile   string
	)

	root := &cobra.Command{
		Use:   "data_exporter",
		Short: "Export enumeration, recipe, and meal data from the database to JSON",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runExport(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLDisable, outputFile)
		},
	}

	root.Flags().StringVar(&dbHost, "db-host", "", "Postgres host")
	root.Flags().Uint16Var(&dbPort, "db-port", 5432, "Postgres port")
	root.Flags().StringVar(&dbUser, "db-user", "", "Postgres username")
	root.Flags().StringVar(&dbPassword, "db-password", "", "Postgres password")
	root.Flags().StringVar(&dbName, "db-name", "", "Postgres database name")
	root.Flags().BoolVar(&dbSSLDisable, "db-ssl-disable", true, "Disable SSL for DB connection")
	root.Flags().StringVar(&outputFile, "output", "seed_data.json", "Output file path")

	for _, flag := range []string{"db-host", "db-user", "db-password", "db-name"} {
		if err := root.MarkFlagRequired(flag); err != nil {
			log.Fatalln(err)
		}
	}

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func runExport(dbHost string, dbPort uint16, dbUser, dbPassword, dbName string, dbSSLDisable bool, outputFile string) error {
	ctx := context.Background()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	client, err := connectDB(ctx, logger, tracerProvider, dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLDisable)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Println("error closing database:", closeErr)
		}
	}()

	auditRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, client)
	identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditRepo, client)
	repo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditRepo, identityRepo, client)

	export := &ExportData{
		ExportedAt: time.Now().UTC(),
	}

	log.Println("Exporting enumerations...")
	if err = exportEnumerations(ctx, repo, export); err != nil {
		return fmt.Errorf("exporting enumerations: %w", err)
	}

	log.Println("Exporting recipes...")
	if err = exportRecipes(ctx, repo, export); err != nil {
		return fmt.Errorf("exporting recipes: %w", err)
	}

	log.Println("Exporting meals...")
	if err = exportMeals(ctx, repo, export); err != nil {
		return fmt.Errorf("exporting meals: %w", err)
	}

	log.Printf("Writing output to %s...", outputFile)
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling export data: %w", err)
	}

	if err = os.WriteFile(outputFile, data, 0o600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	log.Printf("Export complete: %d ingredients, %d preparations, %d instruments, %d vessels, %d measurement units, %d recipes, %d meals",
		len(export.Enumerations.ValidIngredients),
		len(export.Enumerations.ValidPreparations),
		len(export.Enumerations.ValidInstruments),
		len(export.Enumerations.ValidVessels),
		len(export.Enumerations.ValidMeasurementUnits),
		len(export.Recipes),
		len(export.Meals),
	)

	return nil
}

func connectDB(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, dbHost string, dbPort uint16, dbUser, dbPassword, dbName string, dbSSLDisable bool) (database.Client, error) {
	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return nil, errors.New("database connection requires --db-host, --db-user, --db-password, --db-name")
	}

	connDetails := databasecfg.ConnectionDetails{
		Host:       dbHost,
		Port:       dbPort,
		Username:   dbUser,
		Password:   dbPassword,
		Database:   dbName,
		DisableSSL: dbSSLDisable,
	}

	clientConfig := &exporterClientConfig{connDetails: connDetails}
	return postgres.NewDatabaseClient(ctx, logger, tracerProvider, clientConfig, nil)
}

// fetchAll pages through all results using cursor-based pagination.
func fetchAll[T any](
	ctx context.Context,
	fetchPage func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[T], error),
	getID func(*T) string,
) ([]*T, error) {
	var all []*T
	filter := filtering.DefaultQueryFilter()
	pageSize := uint8(filtering.MaxQueryFilterLimit)
	filter.MaxResponseSize = &pageSize

	var cursor *string
	for {
		filter.Cursor = cursor
		result, err := fetchPage(ctx, filter)
		if err != nil {
			return nil, err
		}
		all = append(all, result.Data...)
		if len(result.Data) == 0 {
			break
		}
		lastID := getID(result.Data[len(result.Data)-1])
		cursor = &lastID
	}
	return all, nil
}

func exportEnumerations(ctx context.Context, repo mealplanning.Repository, export *ExportData) error {
	type enumFetch struct {
		fn   func() error
		name string
	}

	fetches := []enumFetch{
		{
			name: "valid ingredients",
			fn: func() error {
				var err error
				export.Enumerations.ValidIngredients, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
					return repo.GetValidIngredients(ctx, f)
				}, func(v *mealplanning.ValidIngredient) string { return v.ID })
				return err
			}},
		{
			name: "valid preparations",
			fn: func() error {
				var err error
				export.Enumerations.ValidPreparations, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
					return repo.GetValidPreparations(ctx, f)
				}, func(v *mealplanning.ValidPreparation) string { return v.ID })
				return err
			}},
		{
			name: "valid instruments",
			fn: func() error {
				var err error
				export.Enumerations.ValidInstruments, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
					return repo.GetValidInstruments(ctx, f)
				}, func(v *mealplanning.ValidInstrument) string { return v.ID })
				return err
			}},
		{
			name: "valid vessels",
			fn: func() error {
				var err error
				export.Enumerations.ValidVessels, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
					return repo.GetValidVessels(ctx, f)
				}, func(v *mealplanning.ValidVessel) string { return v.ID })
				return err
			}},
		{
			name: "valid measurement units",
			fn: func() error {
				var err error
				export.Enumerations.ValidMeasurementUnits, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
					return repo.GetValidMeasurementUnits(ctx, f)
				}, func(v *mealplanning.ValidMeasurementUnit) string { return v.ID })
				return err
			}},
		{
			name: "valid ingredient states",
			fn: func() error {
				var err error
				export.Enumerations.ValidIngredientStates, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
					return repo.GetValidIngredientStates(ctx, f)
				}, func(v *mealplanning.ValidIngredientState) string { return v.ID })
				return err
			}},
		{
			name: "valid ingredient preparations",
			fn: func() error {
				var err error
				export.Enumerations.ValidIngredientPreparations, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
					return repo.GetValidIngredientPreparations(ctx, f)
				}, func(v *mealplanning.ValidIngredientPreparation) string { return v.ID })
				return err
			}},
		{
			name: "valid ingredient measurement units",
			fn: func() error {
				var err error
				export.Enumerations.ValidIngredientMeasurementUnits, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
					return repo.GetValidIngredientMeasurementUnits(ctx, f)
				}, func(v *mealplanning.ValidIngredientMeasurementUnit) string { return v.ID })
				return err
			}},
		{
			name: "valid preparation instruments",
			fn: func() error {
				var err error
				export.Enumerations.ValidPreparationInstruments, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
					return repo.GetValidPreparationInstruments(ctx, f)
				}, func(v *mealplanning.ValidPreparationInstrument) string { return v.ID })
				return err
			}},
		{
			name: "valid preparation vessels",
			fn: func() error {
				var err error
				export.Enumerations.ValidPreparationVessels, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
					return repo.GetValidPreparationVessels(ctx, f)
				}, func(v *mealplanning.ValidPreparationVessel) string { return v.ID })
				return err
			}},
		{
			name: "valid ingredient groups",
			fn: func() error {
				var err error
				export.Enumerations.ValidIngredientGroups, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
					return repo.GetValidIngredientGroups(ctx, f)
				}, func(v *mealplanning.ValidIngredientGroup) string { return v.ID })
				return err
			}},
		{
			name: "valid ingredient state ingredients",
			fn: func() error {
				var err error
				export.Enumerations.ValidIngredientStateIngredients, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
					return repo.GetValidIngredientStateIngredients(ctx, f)
				}, func(v *mealplanning.ValidIngredientStateIngredient) string { return v.ID })
				return err
			}},
	}

	for _, f := range fetches {
		log.Printf("  Fetching %s...", f.name)
		if err := f.fn(); err != nil {
			return fmt.Errorf("fetching %s: %w", f.name, err)
		}
	}

	// Measurement unit conversions: iterate by unit since there's no generic "get all" method
	log.Println("  Fetching valid measurement unit conversions...")
	seenConversions := make(map[string]bool)
	for _, unit := range export.Enumerations.ValidMeasurementUnits {
		conversions, convErr := fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion], error) {
			return repo.GetValidMeasurementUnitConversionsForUnit(ctx, unit.ID, f)
		}, func(v *mealplanning.ValidMeasurementUnitConversion) string { return v.ID })
		if convErr != nil {
			return fmt.Errorf("fetching conversions for unit %s: %w", unit.ID, convErr)
		}
		for _, c := range conversions {
			if !seenConversions[c.ID] {
				seenConversions[c.ID] = true
				export.Enumerations.ValidMeasurementUnitConversions = append(export.Enumerations.ValidMeasurementUnitConversions, c)
			}
		}
	}
	log.Printf("    %d valid measurement unit conversions", len(export.Enumerations.ValidMeasurementUnitConversions))

	return nil
}

func exportRecipes(ctx context.Context, repo mealplanning.Repository, export *ExportData) error {
	// First get all recipe IDs via pagination
	filter := filtering.DefaultQueryFilter()
	pageSize := uint8(filtering.MaxQueryFilterLimit)
	filter.MaxResponseSize = &pageSize

	var recipeIDs []string
	var cursor *string
	for {
		filter.Cursor = cursor
		result, err := repo.GetRecipes(ctx, mealplanning.RecipeStatusApproved, filter)
		if err != nil {
			return fmt.Errorf("fetching recipes: %w", err)
		}
		for _, r := range result.Data {
			recipeIDs = append(recipeIDs, r.ID)
		}
		if len(result.Data) == 0 {
			break
		}
		lastID := result.Data[len(result.Data)-1].ID
		cursor = &lastID
	}

	// Fetch each recipe fully hydrated
	for i, id := range recipeIDs {
		recipe, err := repo.GetRecipe(ctx, id)
		if err != nil {
			return fmt.Errorf("fetching recipe %s: %w", id, err)
		}
		export.Recipes = append(export.Recipes, recipe)
		if (i+1)%50 == 0 {
			log.Printf("  %d/%d recipes fetched", i+1, len(recipeIDs))
		}
	}
	log.Printf("  %d recipes fetched", len(export.Recipes))

	return nil
}

func exportMeals(ctx context.Context, repo mealplanning.Repository, export *ExportData) error {
	var err error
	export.Meals, err = fetchAll(ctx, func(ctx context.Context, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
		return repo.GetMeals(ctx, f)
	}, func(v *mealplanning.Meal) string { return v.ID })
	if err != nil {
		return fmt.Errorf("fetching meals: %w", err)
	}
	log.Printf("  %d meals fetched", len(export.Meals))
	return nil
}

// exporterClientConfig implements database.ClientConfig.
type exporterClientConfig struct {
	connDetails databasecfg.ConnectionDetails
}

var _ database.ClientConfig = (*exporterClientConfig)(nil)

func (c *exporterClientConfig) GetReadConnectionString() string {
	if c.connDetails.DisableSSL {
		return c.connDetails.URI()
	}
	return c.connDetails.String()
}

func (c *exporterClientConfig) GetWriteConnectionString() string {
	return c.GetReadConnectionString()
}

func (c *exporterClientConfig) GetMaxPingAttempts() uint64 { return 10 }

func (c *exporterClientConfig) GetPingWaitPeriod() time.Duration { return time.Second }

func (c *exporterClientConfig) GetMaxIdleConns() int { return 5 }

func (c *exporterClientConfig) GetMaxOpenConns() int { return 7 }

func (c *exporterClientConfig) GetConnMaxLifetime() time.Duration { return 30 * time.Minute }
