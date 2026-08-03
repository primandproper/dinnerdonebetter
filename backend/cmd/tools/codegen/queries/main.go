package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/pflag"
)

var (
	checkOnlyFlag = pflag.Bool("check", false, "only check if files match")
	databaseFlag  = pflag.String("database", postgres, "what database to use")
)

const (
	postgres = "postgres"
)

func main() {
	pflag.Parse()

	runErrors := &multierror.Error{}
	databaseToUse := *databaseFlag

	queryOutput := map[string][]*Query{
		"internalops/sqlc_queries/internalops":                                   buildMaintenanceQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_ingredients":                            buildValidIngredientsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_instruments":                            buildValidInstrumentsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_preparations":                           buildValidPreparationsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_measurement_units":                      buildValidMeasurementUnitsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_ingredient_states":                      buildValidIngredientStatesQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_vessels":                                buildValidVesselsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_ingredient_groups":                      buildValidIngredientGroupsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_ingredient_preparations":                buildValidIngredientPreparationsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_prep_task_configs":                      buildValidPrepTaskConfigsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_preparation_vessels":                    buildValidPreparationVesselsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_ingredient_measurement_units":           buildValidIngredientMeasurementUnitsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_measurement_unit_conversions":           buildValidMeasurementUnitConversionsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_ingredient_state_ingredients":           buildValidIngredientStateIngredientsQueries(databaseToUse),
		"mealplanning/sqlc_queries/valid_preparation_instruments":                buildValidPreparationInstrumentsQueries(databaseToUse),
		"mealplanning/sqlc_queries/account_instrument_ownerships":                buildAccountInstrumentOwnershipQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_components":                              buildMealComponentsQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plan_events":                             buildMealPlanEventsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_media":                                 buildRecipeMediaQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_prep_task_steps":                       buildRecipePrepTaskStepsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_ratings":                               buildRecipeRatingsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_step_completion_condition_ingredients": buildRecipeStepCompletionConditionIngredientsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_prep_tasks":                            buildRecipePrepTasksQueries(databaseToUse),
		"mealplanning/sqlc_queries/meals":                                        buildMealsQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plans":                                   buildMealPlansQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_step_completion_conditions":            buildRecipeStepCompletionConditionQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plan_option_votes":                       buildMealPlanOptionVotesQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plan_options":                            buildMealPlanOptionsQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plan_tasks":                              buildMealPlanTasksQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipes":                                      buildRecipesQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_lists":                                 buildRecipeListsQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_lists":                                   buildMealListsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_step_ingredients":                      buildRecipeStepIngredientsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_step_instruments":                      buildRecipeStepInstrumentsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_step_products":                         buildRecipeStepProductsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_steps":                                 buildRecipeStepsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_step_vessels":                          buildRecipeStepVesselsQueries(databaseToUse),
		"mealplanning/sqlc_queries/user_ingredient_preferences":                  buildUserIngredientPreferencesQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plan_grocery_list_items":                 buildMealPlanGroceryListItemsQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_plan_recipe_option_selections":           buildMealPlanRecipeOptionSelectionsQueries(databaseToUse),
		"mealplanning/sqlc_queries/meal_list_items":                              buildMealListItemsQueries(databaseToUse),
		"mealplanning/sqlc_queries/recipe_list_items":                            buildRecipeListItemsQueries(databaseToUse),
		"oauth/sqlc_queries/oauth2_client_tokens":                                buildOAuth2ClientTokensQueries(databaseToUse),
		"oauth/sqlc_queries/oauth2_clients":                                      buildOAuth2ClientsQueries(databaseToUse),
		"identity/sqlc_queries/account_invitations":                              buildAccountInvitationsQueries(databaseToUse),
		"identity/sqlc_queries/account_user_memberships":                         buildAccountUserMembershipsQueries(databaseToUse),
		"identity/sqlc_queries/accounts":                                         buildAccountsQueries(databaseToUse),
		"identity/sqlc_queries/admin":                                            buildAdminQueries(databaseToUse),
		"auth/sqlc_queries/password_reset_tokens":                                buildPasswordResetTokensQueries(databaseToUse),
		"auth/sqlc_queries/user_sessions":                                        buildUserSessionsQueries(databaseToUse),
		"identity/sqlc_queries/users":                                            buildUsersQueries(databaseToUse),
		"settings/sqlc_queries/service_settings":                                 buildServiceSettingQueries(databaseToUse),
		"settings/sqlc_queries/service_setting_configurations":                   buildServiceSettingConfigurationQueries(databaseToUse),
		"webhooks/sqlc_queries/webhooks":                                         buildWebhooksQueries(databaseToUse),
		"webhooks/sqlc_queries/webhook_trigger_configs":                          buildWebhookTriggerConfigsQueries(databaseToUse),
		"notifications/sqlc_queries/user_notifications":                          buildUserNotificationQueries(databaseToUse),
		"waitlists/sqlc_queries/waitlists":                                       buildWaitlistsQueries(databaseToUse),
		"waitlists/sqlc_queries/waitlist_signups":                                buildWaitlistSignupsQueries(databaseToUse),
		"issuereports/sqlc_queries/issue_reports":                                buildIssueReportsQueries(databaseToUse),
		"uploadedmedia/sqlc_queries/uploaded_media":                              buildUploadedMediaQueries(databaseToUse),
		"payments/sqlc_queries/products":                                         buildPaymentsProductsQueries(databaseToUse),
		"payments/sqlc_queries/subscriptions":                                    buildPaymentsSubscriptionsQueries(databaseToUse),
		"payments/sqlc_queries/purchases":                                        buildPaymentsPurchasesQueries(databaseToUse),
		"payments/sqlc_queries/payment_transactions":                             buildPaymentsTransactionsQueries(databaseToUse),
		"comments/sqlc_queries/comments":                                         buildCommentsQueries(databaseToUse),
		"identity/sqlc_queries/user_roles":                                       buildUserRolesQueries(databaseToUse),
		"identity/sqlc_queries/permissions":                                      buildPermissionsQueries(databaseToUse),
		"identity/sqlc_queries/user_role_permissions":                            buildUserRolePermissionsQueries(databaseToUse),
		"identity/sqlc_queries/user_role_assignments":                            buildUserRoleAssignmentsQueries(databaseToUse),
		"identity/sqlc_queries/user_role_hierarchy":                              buildUserRoleHierarchyQueries(databaseToUse),
	}

	checkOnly := *checkOnlyFlag

	for filePath, queries := range queryOutput {
		fp := fmt.Sprintf("internal/repositories/postgres/%s.generated.sql", filePath)
		if err := os.MkdirAll(filepath.Dir(fp), 0o750); err != nil {
			log.Fatal(fmt.Errorf("creating directory: %w", err))
		}

		existingFile, err := os.ReadFile(fp)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			log.Println("file does not exist")
		}

		var fileContent strings.Builder
		for i, query := range queries {
			if i != 0 {
				fileContent.WriteString("\n")
			}
			fileContent.WriteString(query.Render())
		}

		var fileOutput strings.Builder
		for line := range strings.SplitSeq(strings.TrimSpace(fileContent.String()), "\n") {
			fileOutput.WriteString(strings.TrimSuffix(line, " ") + "\n")
		}

		if string(existingFile) != fileOutput.String() && checkOnly {
			runErrors = multierror.Append(runErrors, fmt.Errorf("files don't match: %s", filePath))
		}

		if !checkOnly {
			if err = os.WriteFile(fp, []byte(fileOutput.String()), 0o600); err != nil {
				runErrors = multierror.Append(runErrors, err)
			}
		}
	}

	if runErrors.ErrorOrNil() != nil {
		log.Fatal(runErrors)
	}
}
