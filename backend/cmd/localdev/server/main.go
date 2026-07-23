package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/localdev"
	identitygenerated "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/identifiers"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
)

const (
	apiConfigurationFilepath = "deploy/environments/testing/config_files/integration-tests-config.json"
)

func main() {
	ctx := context.Background()

	// create premade admin user
	premadeAdminUser := &identity.User{
		ID:              strings.Repeat("a", 20),
		TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		EmailAddress:    "integration_tests@example.email",
		Username:        "admin_user",
		HashedPassword:  "admin_pass",
	}

	apiConfig, err := config.LoadConfigFromPath[config.APIServiceConfig](apiConfigurationFilepath)
	if err != nil {
		log.Fatal(err)
	}

	var (
		adminUserID    string
		adminAccountID string
	)

	server, err := localdev.AllInOne(
		ctx,
		apiConfig,
		// Create admin user and get account
		localdev.WithIdentityRepository(func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error {
			user, userErr := localdev.CreatePremadeAdminUser(ctx, logger, tracerProvider, repo, dbClient, premadeAdminUser)
			if userErr != nil {
				return userErr
			}
			adminUserID = user.ID

			// Get or create account for admin user
			accounts, accountsErr := repo.GetAccounts(ctx, adminUserID, nil)
			if accountsErr != nil {
				return fmt.Errorf("failed to get accounts for admin user: %w", accountsErr)
			}

			if len(accounts.Data) > 0 {
				// Use first account
				adminAccountID = accounts.Data[0].ID
			} else {
				// Create a new account for the admin user
				accountInput := &identity.AccountCreationRequestInput{
					Name:          "Admin Household",
					BelongsToUser: adminUserID,
				}
				account, accountErr := repo.CreateAccount(ctx, identityconverters.ConvertAccountCreationInputToAccountDatabaseCreationInput(accountInput))
				if accountErr != nil {
					return fmt.Errorf("failed to create account for admin user: %w", accountErr)
				}
				adminAccountID = account.ID
			}

			if adminAccountID == "" {
				return fmt.Errorf("admin account ID not set")
			}

			hasher := authentication.NewArgon2Authenticator(logger, tracerProvider)
			generatedQuerier := identitygenerated.New()

			// Create two member users
			memberUsers := []*struct {
				username  string
				email     string
				password  string
				firstName string
				lastName  string
				userID    string
			}{
				{
					username:  "member_user_1",
					email:     "member1@example.email",
					password:  "member_pass_1",
					firstName: "Member",
					lastName:  "One",
					userID:    strings.Repeat("c", 20),
				},
				{
					username:  "member_user_2",
					email:     "member2@example.email",
					password:  "member_pass_2",
					firstName: "Member",
					lastName:  "Two",
					userID:    strings.Repeat("d", 20),
				},
			}

			for _, memberUser := range memberUsers {
				// Check if user already exists
				existingUser, userExistsErr := repo.GetUserByUsername(ctx, memberUser.username)
				if userExistsErr == nil && existingUser != nil {
					logger.Info(fmt.Sprintf("User %s already exists, skipping creation", memberUser.username))
					// Still add to account if not already a member
					isMember, memberErr := repo.UserIsMemberOfAccount(ctx, existingUser.ID, adminAccountID)
					if memberErr == nil && !isMember {
						membershipID := identifiers.New()
						if err = generatedQuerier.AddUserToAccount(ctx, dbClient.WriteDB(), &identitygenerated.AddUserToAccountParams{
							ID:               membershipID,
							BelongsToUser:    existingUser.ID,
							BelongsToAccount: adminAccountID,
						}); err != nil {
							return fmt.Errorf("failed to add existing user %s to account: %w", memberUser.username, err)
						}
						if err = generatedQuerier.AssignRoleToUser(ctx, dbClient.WriteDB(), &identitygenerated.AssignRoleToUserParams{
							ID:        identifiers.New(),
							UserID:    existingUser.ID,
							RoleID:    authorization.AccountMemberRoleID,
							AccountID: sql.NullString{String: adminAccountID, Valid: true},
						}); err != nil {
							return fmt.Errorf("failed to assign account role to existing user %s: %w", memberUser.username, err)
						}
						logger.Info(fmt.Sprintf("Added existing user %s to admin account", memberUser.username))
					}
					continue
				}

				// Hash password
				hashedPassword, hashErr := hasher.HashPassword(ctx, memberUser.password)
				if hashErr != nil {
					return fmt.Errorf("failed to hash password for user %s: %w", memberUser.username, hashErr)
				}

				// Create user
				userInput := &identity.User{
					ID:              memberUser.userID,
					Username:        memberUser.username,
					EmailAddress:    memberUser.email,
					HashedPassword:  hashedPassword,
					TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
					FirstName:       memberUser.firstName,
					LastName:        memberUser.lastName,
				}

				user, userErr = repo.CreateUser(ctx, identityconverters.ConvertUserToUserDatabaseCreationInput(userInput))
				if userErr != nil {
					return fmt.Errorf("failed to create user %s: %w", memberUser.username, userErr)
				}

				// Mark two-factor secret as verified
				if err = repo.MarkUserTwoFactorSecretAsVerified(ctx, user.ID); err != nil {
					return fmt.Errorf("failed to mark user %s as verified: %w", memberUser.username, err)
				}

				// Add user to admin account as a member
				membershipID := identifiers.New()
				if err = generatedQuerier.AddUserToAccount(ctx, dbClient.WriteDB(), &identitygenerated.AddUserToAccountParams{
					ID:               membershipID,
					BelongsToUser:    user.ID,
					BelongsToAccount: adminAccountID,
				}); err != nil {
					return fmt.Errorf("failed to add user %s to account: %w", memberUser.username, err)
				}

				if err = generatedQuerier.AssignRoleToUser(ctx, dbClient.WriteDB(), &identitygenerated.AssignRoleToUserParams{
					ID:        identifiers.New(),
					UserID:    user.ID,
					RoleID:    authorization.AccountMemberRoleID,
					AccountID: sql.NullString{String: adminAccountID, Valid: true},
				}); err != nil {
					return fmt.Errorf("failed to assign account role to user %s: %w", memberUser.username, err)
				}

				if err = generatedQuerier.MarkAccountUserMembershipAsUserDefault(ctx, dbClient.WriteDB(), &identitygenerated.MarkAccountUserMembershipAsUserDefaultParams{
					BelongsToUser:    user.ID,
					BelongsToAccount: adminAccountID,
				}); err != nil {
					return fmt.Errorf("failed to mark user %s account as default: %w", memberUser.username, err)
				}

				logger.Info(fmt.Sprintf("Created user %s and added to admin account", memberUser.username))
			}

			return nil
		}),
		// Create OAuth2 client
		localdev.WithOAuth2Repository(func(ctx context.Context, repo oauth.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
			// the plaintext secret stays the well-known localdev value; only its digest is stored.
			_, err = repo.CreateOAuth2Client(ctx, &oauth.OAuth2ClientDatabaseCreationInput{
				ID:           strings.Repeat("b", 20),
				Name:         "localdev_admin_client",
				Description:  "localdev admin client",
				ClientID:     strings.Repeat("A", oauth.ClientIDSize),
				ClientSecret: oauth.HashClientSecret(strings.Repeat("A", oauth.ClientSecretSize)),
			})
			return err
		}),
		// Create example service settings
		localdev.WithSettingsRepository(func(ctx context.Context, repo settings.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
			return createExampleServiceSettings(ctx, repo, logger)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("database connection string:", apiConfig.Database.GetReadConnectionString())
	log.Println("starting server")

	if os.Getenv("DRY_RUN") == "true" {
		log.Println("dry run is enabled, skipping server run")
		return
	}
	server.Run(ctx)
}

func createExampleServiceSettings(ctx context.Context, repo settings.Repository, logger logging.Logger) error {
	defaultTheme := "light"
	_, err := repo.CreateServiceSetting(ctx, &settings.ServiceSettingDatabaseCreationInput{
		ID:           identifiers.New(),
		Name:         "user_theme_preference",
		Type:         "user",
		Description:  "User's preferred theme for the application interface",
		Enumeration:  []string{"light", "dark", "auto"},
		DefaultValue: &defaultTheme,
		AdminsOnly:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create theme preference setting: %w", err)
	}
	logger.Debug("Created ServiceSetting: user_theme_preference (enumerated with default)")

	defaultNotificationFreq := "daily"
	_, err = repo.CreateServiceSetting(ctx, &settings.ServiceSettingDatabaseCreationInput{
		ID:           identifiers.New(),
		Name:         "membership_notification_frequency",
		Type:         "membership",
		Description:  "How often to send notifications to membership members",
		Enumeration:  []string{"immediate", "daily", "weekly", "never"},
		DefaultValue: &defaultNotificationFreq,
		AdminsOnly:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create notification frequency setting: %w", err)
	}
	logger.Debug("Created ServiceSetting: membership_notification_frequency (enumerated with default)")

	defaultLanguage := "en"
	_, err = repo.CreateServiceSetting(ctx, &settings.ServiceSettingDatabaseCreationInput{
		ID:           identifiers.New(),
		Name:         "user_language",
		Type:         "user",
		Description:  "User's preferred language for the application",
		Enumeration:  []string{"en", "es", "fr", "de", "it"},
		DefaultValue: &defaultLanguage,
		AdminsOnly:   false,
	})
	if err != nil {
		return fmt.Errorf("failed to create language setting: %w", err)
	}
	logger.Debug("Created ServiceSetting: user_language (enumerated, non-admin)")

	return nil
}
