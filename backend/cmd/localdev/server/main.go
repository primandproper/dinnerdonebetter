package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformsettings "github.com/primandproper/platform-go/v14/settings"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/pointer"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	apiConfigurationFilepath = "deploy/environments/testing/config_files/integration-tests-config.json"
)

func main() {
	ctx := context.Background()

	// create premade admin user
	premadeAdminUser := &platformidentity.User{
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
		localdev.WithIdentityDirectory(func(ctx context.Context, directory *platformidentity.Service, store platformidentity.Store, logger logging.Logger, tracerProvider tracing.Provider, dbClient database.Client) error {
			user, userErr := localdev.CreatePremadeAdminUser(ctx, logger, tracerProvider, directory, store, dbClient, premadeAdminUser)
			if userErr != nil {
				return userErr
			}
			adminUserID = user.ID

			// The account the registration minted. Registering is what creates it, so
			// there is no branch here for a user who has one and a user who does not —
			// every user in this directory owns exactly one account from the moment
			// they exist, and this read finds it whether this run made it or a previous
			// one did.
			accounts, accountsErr := store.ListAccountsForUser(ctx, dbClient.Reader(), ddbidentity.Scope(), adminUserID, nil)
			if accountsErr != nil {
				return fmt.Errorf("failed to get accounts for admin user: %w", accountsErr)
			}

			if len(accounts.Data) == 0 {
				return fmt.Errorf("admin user %s has no account", adminUserID)
			}

			adminAccountID = accounts.Data[0].ID

			hasher := authentication.NewArgon2Authenticator(argon2.WithLogger(logger), argon2.WithTracerProvider(tracerProvider))

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
				existingUser, userExistsErr := store.GetUserByUsername(ctx, dbClient.Reader(), ddbidentity.Scope(), memberUser.username)
				if userExistsErr == nil && existingUser != nil {
					logger.Info(fmt.Sprintf("User %s already exists, skipping creation", memberUser.username))

					if err = joinAdminAccount(ctx, directory, store, dbClient, existingUser.ID, adminAccountID); err != nil {
						return fmt.Errorf("failed to add existing user %s to account: %w", memberUser.username, err)
					}

					continue
				}

				hashedPassword, hashErr := hasher.HashPassword(ctx, memberUser.password)
				if hashErr != nil {
					return fmt.Errorf("failed to hash password for user %s: %w", memberUser.username, hashErr)
				}

				// Registered rather than inserted, for the reason CreatePremadeAdminUser
				// gives: a user, an account and the membership between them are one
				// transaction, and a user with no account is a state nothing here can
				// represent.
				registration, registerErr := directory.Register(ctx, ddbidentity.Scope(), &platformidentity.User{
					ID:              memberUser.userID,
					Username:        memberUser.username,
					EmailAddress:    memberUser.email,
					HashedPassword:  hashedPassword,
					TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
					FirstName:       memberUser.firstName,
					LastName:        memberUser.lastName,
					ServiceRoles:    []string{authorization.ServiceUserRoleName},
				}, &platformidentity.Account{
					Name: memberUser.username + "'s account",
				}, []string{authorization.AccountAdminRoleName})
				if registerErr != nil {
					return fmt.Errorf("failed to create user %s: %w", memberUser.username, registerErr)
				}

				if _, err = directory.MarkUserTwoFactorSecretVerified(ctx, ddbidentity.Scope(), registration.User.ID); err != nil {
					return fmt.Errorf("failed to mark user %s as verified: %w", memberUser.username, err)
				}

				if err = joinAdminAccount(ctx, directory, store, dbClient, registration.User.ID, adminAccountID); err != nil {
					return fmt.Errorf("failed to add user %s to account: %w", memberUser.username, err)
				}

				logger.Info(fmt.Sprintf("Created user %s and added to admin account", memberUser.username))
			}

			return nil
		}),
		// Create OAuth2 client.
		//
		// The credential is well known and stays that way: it is in checked-in config and
		// in both web apps' environments, so a minted one would mean nothing could sign in
		// to localdev until somebody copied it out of the database. platform mints through
		// a CredentialGenerator precisely so a deployment that needs a predictable one can
		// say so, rather than writing the row behind the service's back.
		localdev.WithOAuth2Registry(
			func() (clientID, secret string, err error) {
				return strings.Repeat("A", oauth.ClientIDSize), strings.Repeat("A", oauth.ClientSecretSize), nil
			},
			func(ctx context.Context, svc *platformoauth2clients.Service, logger logging.Logger, tracerProvider tracing.Provider) error {
				_, err = svc.CreateClient(ctx, tenancy.Global(), "", &platformoauth2clients.CreationInput{
					Name:        "localdev_admin_client",
					Description: "localdev admin client",
					// Matched byte for byte, so this is every address a localdev client
					// actually redirects to: the API server's own (which is what the CLI
					// helpers authorize against, reading the code off the Location header
					// rather than following it) and the two web apps' callbacks.
					RedirectURIs: []string{
						"http://localhost:9000",
						branding.LocalDevConsumerWebAppURL + "/auth/callback",
						branding.LocalDevAdminWebAppURL + "/auth/callback",
					},
				})

				return err
			}),
		// Create the example settings catalog
		localdev.WithSettingsRepository(func(ctx context.Context, store platformsettings.Store, logger logging.Logger, tracerProvider tracing.Provider, dbClient database.Client) error {
			return createExampleSettingDefinitions(ctx, store, logger, dbClient)
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
	if err = server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

// createExampleSettingDefinitions seeds the catalog a locally-run instance
// starts with.
//
// Every one of them is a text setting that enumerates its values, which is the
// shape a preferences page can render: a select box needs to know what to offer,
// and a setting with no enumeration is a free-text field nothing validates.
func createExampleSettingDefinitions(ctx context.Context, store platformsettings.Store, logger logging.Logger, dbClient database.Client) error {
	for _, definition := range []*platformsettings.Definition{
		{
			Name:        "user_theme_preference",
			Description: "User's preferred theme for the application interface",
			Kind:        platformsettings.KindString,
			Enumeration: []string{"light", "dark", "auto"},
			Default:     pointer.To("light"),
			AdminOnly:   true,
		},
		{
			Name:        "user_notification_frequency",
			Description: "How often to send notifications",
			Kind:        platformsettings.KindString,
			Enumeration: []string{"immediate", "daily", "weekly", "never"},
			Default:     pointer.To("daily"),
			AdminOnly:   true,
		},
		{
			Name:        "user_language",
			Description: "User's preferred language for the application",
			Kind:        platformsettings.KindString,
			Enumeration: []string{"en", "es", "fr", "de", "it"},
			Default:     pointer.To("en"),
			AdminOnly:   false,
		},
	} {
		if err := dbClient.WithTransaction(ctx, func(tx database.Tx) error {
			_, createErr := store.CreateDefinition(ctx, tx, settings.Scope(), definition)

			return createErr
		}); err != nil {
			return fmt.Errorf("failed to create the %s setting: %w", definition.Name, err)
		}

		logger.Debug("created setting definition: " + definition.Name)
	}

	return nil
}

// joinAdminAccount puts a user in the admin's household and makes it where they land.
//
// The membership is written through the store on a transaction of its own, because the
// directory service has no "add somebody to an account" of its own: joining is answering
// an invitation, and a seed has nobody to send one to. The default is moved afterwards
// through the service, so the one-default-per-user invariant is the store's to keep rather
// than this function's to remember.
//
// It is idempotent: a user already in the account keeps the membership they have.
func joinAdminAccount(
	ctx context.Context,
	directory *platformidentity.Service,
	store platformidentity.Store,
	dbClient database.Client,
	userID, accountID string,
) error {
	existing, err := store.GetMembership(ctx, dbClient.Reader(), ddbidentity.Scope(), userID, accountID)
	if err == nil && existing != nil {
		return nil
	}

	if err = dbClient.WithTransaction(ctx, func(tx database.Tx) error {
		_, createErr := store.CreateMembership(ctx, tx, ddbidentity.Scope(), &platformidentity.Membership{
			Scope:            ddbidentity.Scope(),
			BelongsToUser:    userID,
			BelongsToAccount: accountID,
			Roles:            []string{authorization.AccountMemberRoleName},
		})

		return createErr
	}); err != nil {
		return err
	}

	_, err = directory.SetDefaultAccount(ctx, ddbidentity.Scope(), userID, accountID)

	return err
}
