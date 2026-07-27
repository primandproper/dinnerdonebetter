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
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	"github.com/primandproper/platform-go/v7/database"
	databasecfg "github.com/primandproper/platform-go/v7/database/config"
	"github.com/primandproper/platform-go/v7/database/postgres"
	"github.com/primandproper/platform-go/v7/observability/logging"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/spf13/cobra"
)

func main() {
	var (
		dbHost       string
		dbPort       uint16
		dbUser       string
		dbPassword   string
		dbName       string
		dbSSLDisable bool
		inputFile    string
	)

	root := &cobra.Command{
		Use:   "data_importer",
		Short: "Import enumeration, recipe, and meal data from JSON into the database",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runImport(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLDisable, inputFile)
		},
	}

	root.Flags().StringVar(&dbHost, "db-host", "", "Postgres host")
	root.Flags().Uint16Var(&dbPort, "db-port", 5432, "Postgres port")
	root.Flags().StringVar(&dbUser, "db-user", "", "Postgres username")
	root.Flags().StringVar(&dbPassword, "db-password", "", "Postgres password")
	root.Flags().StringVar(&dbName, "db-name", "", "Postgres database name")
	root.Flags().BoolVar(&dbSSLDisable, "db-ssl-disable", true, "Disable SSL for DB connection")
	root.Flags().StringVar(&inputFile, "input", "seed_data.json", "Input JSON file path")

	for _, flag := range []string{"db-host", "db-user", "db-password", "db-name"} {
		if err := root.MarkFlagRequired(flag); err != nil {
			log.Fatalln(err)
		}
	}

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}

// ExportData mirrors the exporter's output structure.
type ExportData struct {
	ExportedAt   time.Time              `json:"exportedAt"`
	Enumerations ExportedEnumerations   `json:"enumerations"`
	Recipes      []*mealplanning.Recipe `json:"recipes"`
	Meals        []*mealplanning.Meal   `json:"meals"`
}

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

func runImport(dbHost string, dbPort uint16, dbUser, dbPassword, dbName string, dbSSLDisable bool, inputFile string) error {
	ctx := context.Background()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	log.Printf("Reading input file: %s", inputFile)
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	var export ExportData
	if err = json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("unmarshaling input data: %w", err)
	}
	log.Printf("Loaded export from %s", export.ExportedAt.Format(time.RFC3339))

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

	log.Println("Importing base enumerations...")
	if err = importBaseEnumerations(ctx, repo, &export.Enumerations); err != nil {
		return fmt.Errorf("importing base enumerations: %w", err)
	}

	log.Println("Importing bridge types...")
	if err = importBridgeTypes(ctx, repo, &export.Enumerations); err != nil {
		return fmt.Errorf("importing bridge types: %w", err)
	}

	log.Println("Importing recipes...")
	if err = importRecipes(ctx, repo, export.Recipes); err != nil {
		return fmt.Errorf("importing recipes: %w", err)
	}

	log.Println("Importing meals...")
	if err = importMeals(ctx, repo, export.Meals); err != nil {
		return fmt.Errorf("importing meals: %w", err)
	}

	log.Println("Import complete!")
	return nil
}

func importBaseEnumerations(ctx context.Context, repo mealplanning.Repository, enums *ExportedEnumerations) error {
	for i, v := range enums.ValidIngredients {
		exists, existsErr := repo.ValidIngredientExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid ingredient %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidIngredient(ctx, converters.ConvertValidIngredientToValidIngredientDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid ingredient %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid ingredients processed", len(enums.ValidIngredients))

	for i, v := range enums.ValidPreparations {
		exists, existsErr := repo.ValidPreparationExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid preparation %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidPreparation(ctx, converters.ConvertValidPreparationToValidPreparationDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid preparation %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid preparations processed", len(enums.ValidPreparations))

	for i, v := range enums.ValidInstruments {
		exists, existsErr := repo.ValidInstrumentExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid instrument %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidInstrument(ctx, converters.ConvertValidInstrumentToValidInstrumentDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid instrument %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid instruments processed", len(enums.ValidInstruments))

	for i, v := range enums.ValidVessels {
		exists, existsErr := repo.ValidVesselExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid vessel %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidVessel(ctx, converters.ConvertValidVesselToValidVesselDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid vessel %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid vessels processed", len(enums.ValidVessels))

	for i, v := range enums.ValidMeasurementUnits {
		exists, existsErr := repo.ValidMeasurementUnitExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid measurement unit %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidMeasurementUnit(ctx, converters.ConvertValidMeasurementUnitToValidMeasurementUnitDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid measurement unit %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid measurement units processed", len(enums.ValidMeasurementUnits))

	for i, v := range enums.ValidIngredientStates {
		exists, existsErr := repo.ValidIngredientStateExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid ingredient state %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidIngredientState(ctx, converters.ConvertValidIngredientStateToValidIngredientStateDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid ingredient state %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid ingredient states processed", len(enums.ValidIngredientStates))

	return nil
}

func importBridgeTypes(ctx context.Context, repo mealplanning.Repository, enums *ExportedEnumerations) error {
	for i, v := range enums.ValidIngredientPreparations {
		exists, existsErr := repo.ValidIngredientPreparationExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid ingredient preparation %d: %w", i, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidIngredientPreparation(ctx, converters.ConvertValidIngredientPreparationToValidIngredientPreparationDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid ingredient preparation %d: %w", i, err)
		}
	}
	log.Printf("  %d valid ingredient preparations processed", len(enums.ValidIngredientPreparations))

	for i, v := range enums.ValidIngredientMeasurementUnits {
		exists, existsErr := repo.ValidIngredientMeasurementUnitExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid ingredient measurement unit %d: %w", i, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidIngredientMeasurementUnit(ctx, converters.ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid ingredient measurement unit %d: %w", i, err)
		}
	}
	log.Printf("  %d valid ingredient measurement units processed", len(enums.ValidIngredientMeasurementUnits))

	for i, v := range enums.ValidPreparationInstruments {
		exists, existsErr := repo.ValidPreparationInstrumentExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid preparation instrument %d: %w", i, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidPreparationInstrument(ctx, converters.ConvertValidPreparationInstrumentToValidPreparationInstrumentDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid preparation instrument %d: %w", i, err)
		}
	}
	log.Printf("  %d valid preparation instruments processed", len(enums.ValidPreparationInstruments))

	for i, v := range enums.ValidPreparationVessels {
		exists, existsErr := repo.ValidPreparationVesselExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid preparation vessel %d: %w", i, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidPreparationVessel(ctx, converters.ConvertValidPreparationVesselToValidPreparationVesselDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid preparation vessel %d: %w", i, err)
		}
	}
	log.Printf("  %d valid preparation vessels processed", len(enums.ValidPreparationVessels))

	for i, v := range enums.ValidIngredientGroups {
		exists, existsErr := repo.ValidIngredientGroupExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid ingredient group %d (%s): %w", i, v.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidIngredientGroup(ctx, converters.ConvertValidIngredientGroupToValidIngredientGroupDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid ingredient group %d (%s): %w", i, v.Name, err)
		}
	}
	log.Printf("  %d valid ingredient groups processed", len(enums.ValidIngredientGroups))

	for i, v := range enums.ValidIngredientStateIngredients {
		exists, existsErr := repo.ValidIngredientStateIngredientExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid ingredient state ingredient %d: %w", i, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidIngredientStateIngredient(ctx, converters.ConvertValidIngredientStateIngredientToValidIngredientStateIngredientDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid ingredient state ingredient %d: %w", i, err)
		}
	}
	log.Printf("  %d valid ingredient state ingredients processed", len(enums.ValidIngredientStateIngredients))

	for i, v := range enums.ValidMeasurementUnitConversions {
		exists, existsErr := repo.ValidMeasurementUnitConversionExists(ctx, v.ID)
		if existsErr != nil {
			return fmt.Errorf("checking valid measurement unit conversion %d: %w", i, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateValidMeasurementUnitConversion(ctx, converters.ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionDatabaseCreationInput(v)); err != nil {
			return fmt.Errorf("creating valid measurement unit conversion %d: %w", i, err)
		}
	}
	log.Printf("  %d valid measurement unit conversions processed", len(enums.ValidMeasurementUnitConversions))

	return nil
}

func importRecipes(ctx context.Context, repo mealplanning.Repository, recipes []*mealplanning.Recipe) error {
	for i, r := range recipes {
		exists, existsErr := repo.RecipeExists(ctx, r.ID)
		if existsErr != nil {
			return fmt.Errorf("checking recipe %d (%s): %w", i, r.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateRecipe(ctx, converters.ConvertRecipeToRecipeDatabaseCreationInput(r)); err != nil {
			return fmt.Errorf("creating recipe %d (%s): %w", i, r.Name, err)
		}
		if (i+1)%50 == 0 {
			log.Printf("  %d/%d recipes processed", i+1, len(recipes))
		}
	}
	log.Printf("  %d recipes processed", len(recipes))
	return nil
}

func importMeals(ctx context.Context, repo mealplanning.Repository, meals []*mealplanning.Meal) error {
	for i, m := range meals {
		exists, existsErr := repo.MealExists(ctx, m.ID)
		if existsErr != nil {
			return fmt.Errorf("checking meal %d (%s): %w", i, m.Name, existsErr)
		}
		if exists {
			continue
		}
		if _, err := repo.CreateMeal(ctx, converters.ConvertMealToMealDatabaseCreationInput(m)); err != nil {
			return fmt.Errorf("creating meal %d (%s): %w", i, m.Name, err)
		}
	}
	log.Printf("  %d meals processed", len(meals))
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

	clientConfig := &importerClientConfig{connDetails: connDetails}
	return postgres.NewDatabaseClient(ctx, logger, tracerProvider, clientConfig, nil)
}

type importerClientConfig struct {
	connDetails databasecfg.ConnectionDetails
}

var _ database.ClientConfig = (*importerClientConfig)(nil)

func (c *importerClientConfig) GetReadConnectionString() string {
	if c.connDetails.DisableSSL {
		return c.connDetails.URI()
	}
	return c.connDetails.String()
}

func (c *importerClientConfig) GetWriteConnectionString() string  { return c.GetReadConnectionString() }
func (c *importerClientConfig) GetMaxPingAttempts() uint64        { return 10 }
func (c *importerClientConfig) GetPingWaitPeriod() time.Duration  { return time.Second }
func (c *importerClientConfig) GetMaxIdleConns() int              { return 5 }
func (c *importerClientConfig) GetMaxOpenConns() int              { return 7 }
func (c *importerClientConfig) GetConnMaxLifetime() time.Duration { return 30 * time.Minute }
