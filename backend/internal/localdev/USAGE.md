# LocalDev AllInOne Usage Guide

The `localdev.AllInOne` function provides a flexible way to set up a complete local development environment with arbitrary database initialization.

## Basic Concept

Instead of hardcoding specific entities to create, `AllInOne` accepts a list of `DatabaseInitFunc` functions that execute during setup. The package provides repository-specific helpers that give you access to fully configured repositories.

## Available Repository Helpers

- `WithIdentityRepository` - User, account, and membership operations
- `WithOAuth2Repository` - OAuth2 client operations
- `WithPasswordResetTokenStore` - Password reset token operations
- `WithMealPlanningRepository` - All meal planning entities (recipes, ingredients, preparations, vessels, instruments, etc.)
- `WithSettingsRepository` - Settings operations
- `WithWebhooksRepository` - Webhook operations
- `WithNotificationsRepository` - Notification operations

## Basic Example

```go
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
        // Create a user
        user, err := repo.CreateUser(ctx, &identity.UserDatabaseCreationInput{
            ID:             "user123",
            Username:       "testuser",
            EmailAddress:   "test@example.com",
            HashedPassword: "hashedpass",
        })
        return err
    }),
)
```

## Creating Admin User & OAuth2 Client (Original Functionality)

```go
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    // Create admin user
    localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
        premadeAdminUser := &identity.User{
            ID:              "admin123",
            TwoFactorSecret: "AAAA...",
            EmailAddress:    "admin@example.com",
            Username:        "admin_user",
            HashedPassword:  "admin_pass",
        }
        _, err := localdev.CreatePremadeAdminUser(ctx, logger, tracerProvider, repo, dbClient, premadeAdminUser)
        return err
    }),
    // Create OAuth2 client
    localdev.WithOAuth2Repository(func(ctx context.Context, repo oauth.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
        _, err := repo.CreateOAuth2Client(ctx, &oauth.OAuth2ClientDatabaseCreationInput{
            ID:           "oauth123",
            Name:         "test_client",
            Description:  "Test OAuth2 client",
            ClientID:     "client_id",
            ClientSecret: "client_secret",
        })
        return err
    }),
)
```

## Creating Test Data with Meal Planning Repository

```go
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    localdev.WithMealPlanningRepository(func(ctx context.Context, repo mealplanning.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
        // Create valid ingredients
        _, err := repo.CreateValidIngredient(ctx, &types.ValidIngredientDatabaseCreationInput{
            ID:          "ingredient1",
            Name:        "Chicken Breast",
            Description: "Boneless skinless chicken breast",
        })
        if err != nil {
            return err
        }
        
        // Create valid preparations
        _, err = repo.CreateValidPreparation(ctx, &types.ValidPreparationDatabaseCreationInput{
            ID:          "prep1",
            Name:        "Diced",
            Description: "Cut into small cubes",
        })
        if err != nil {
            return err
        }
        
        // Create valid instruments
        _, err = repo.CreateValidInstrument(ctx, &types.ValidInstrumentDatabaseCreationInput{
            ID:          "instrument1",
            Name:        "Chef's Knife",
            Description: "A sharp chef's knife",
        })
        return err
    }),
)
```

## Multiple Repository Operations

You can chain multiple repository helpers to set up complex data:

```go
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    // Set up users and accounts
    localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
        // Create admin user
        admin, err := localdev.CreatePremadeAdminUser(ctx, logger, tracerProvider, repo, dbClient, adminUser)
        if err != nil {
            return err
        }
        
        // Create test account
        account, err := repo.CreateAccount(ctx, &identity.AccountDatabaseCreationInput{
            ID:   "account123",
            Name: "Test Account",
        })
        if err != nil {
            return err
        }
        
        // Link user to account
        err = repo.AddUserToAccount(ctx, admin.ID, account.ID)
        return err
    }),
    
    // Set up OAuth2
    localdev.WithOAuth2Repository(func(ctx context.Context, repo oauth.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
        _, err := repo.CreateOAuth2Client(ctx, oauthInput)
        return err
    }),
    
    // Set up meal planning data
    localdev.WithMealPlanningRepository(func(ctx context.Context, repo mealplanning.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
        // Create a bunch of test ingredients, recipes, etc.
        for _, ingredient := range testIngredients {
            if _, err := repo.CreateValidIngredient(ctx, ingredient); err != nil {
                return err
            }
        }
        return nil
    }),
    
    // Set up webhooks
    localdev.WithWebhooksRepository(func(ctx context.Context, repo webhooks.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
        _, err := repo.CreateWebhook(ctx, &webhooks.WebhookDatabaseCreationInput{
            ID:      "webhook1",
            Name:    "Test Webhook",
            URL:     "https://example.com/webhook",
            Events:  []string{"user.created"},
        })
        return err
    }),
)
```

## Using Raw Database Queries

If you need to do something that's not available through a repository, you can use the base `DatabaseInitFunc`:

```go
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    // Custom database initialization
    func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
        // Execute raw SQL
        _, err := dbClient.DB().ExecContext(ctx, `
            INSERT INTO some_table (id, name, value)
            VALUES ($1, $2, $3)
        `, "id1", "test", 42)
        return err
    },
    
    // You can still mix with repository helpers
    localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
        // Use the repo as normal
        return nil
    }),
)
```

## Error Handling

All initialization functions return errors. If any function fails, `AllInOne` will return an error immediately:

```go
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
        user, err := repo.CreateUser(ctx, input)
        if err != nil {
            return fmt.Errorf("failed to create user: %w", err)
        }
        logger.Info("Created user: " + user.ID)
        return nil
    }),
)
if err != nil {
    log.Fatal("Setup failed:", err)
}
```

## Tips

1. **Order matters**: Functions execute sequentially, so if one depends on data from another, order them appropriately.

2. **Use the provided infrastructure**: Logger and TracerProvider are available in all callbacks for proper observability.

3. **Reusable functions**: Extract common initialization patterns into reusable functions:

```go
func createTestUsers(users []*identity.User) localdev.DatabaseInitFunc {
    return localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
        for _, user := range users {
            if _, err := repo.CreateUser(ctx, convertUser(user)); err != nil {
                return err
            }
        }
        return nil
    })
}

// Use it
server, err := localdev.AllInOne(
    ctx,
    apiConfig,
    createTestUsers(myTestUsers),
    createTestOAuth2Client(myOAuthClient),
)
```

1. **Access multiple repositories**: If you need multiple repositories in one function, use nested helpers or use the base `DatabaseInitFunc` and instantiate repositories yourself.
