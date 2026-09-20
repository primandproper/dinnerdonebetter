package authorization

import (
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
)

type (
	role int

	// Permission names an action a principal may be authorized to perform.
	//
	// It is an alias for the platform's type rather than a defined type of its
	// own, which is the adoption platform documents and the thing that lets this
	// application's permission maps be composed with the ones platform's own gRPC
	// surfaces ship — see internal/build/services/api/grpc.AggregateMethodPermissions,
	// which now merges commentsgrpc.Permissions() alongside this repo's own. A
	// defined type would have made that one conversion per domain, for thirteen
	// domains.
	Permission = platformauthz.Permission
)

var (
	// ServiceAdminPermissions is every service admin permission.
	ServiceAdminPermissions = []Permission{
		ReadUserDataPermission,
		UpdateUserStatusPermission,
		ReadUserPermission,
		SearchUserPermission,
		ArchiveUserPermission,
		CreateOAuth2ClientsPermission,
		ArchiveOAuth2ClientsPermission,
		ArchiveSettingDefinitionsPermission,
		ImpersonateUserPermission,
		ManageUserSessionsPermission,
		PublishArbitraryQueueMessagePermission,
		RunMealPlanWorkersPermission,
		UpdateRecipesStatusPermission,
		// only admins can arbitrarily create these via the API, this is exclusively for integration test purposes.
		CreateSettingDefinitionsPermission,
		UpdateSettingDefinitionsPermission,
		CreateMealPlanTasksPermission,
		CreateMealPlanGroceryListItemsPermission,
		CreateWaitlistsPermission,
		UpdateWaitlistsPermission,
		ArchiveWaitlistsPermission,
		CreateProductsPermission,
		ReadProductsPermission,
		UpdateProductsPermission,
		ArchiveProductsPermission,
		ReadSubscriptionsPermission,
		ArchiveSubscriptionsPermission,

		// The rest of the adopted surfaces' administrative halves. Each of these is
		// a method platform's server mounts, and a permission no role held until
		// now — which made the method callable by nobody, refused exactly as it
		// would be for a caller who genuinely lacked it. See
		// internal/build/services/api/grpc.TestMethodTableIsCoveredByThePolicy,
		// which is the check that says so.
		VerifyAuditChainPermission,
		ModerateCommentsPermission,
		ReadAllSettingValuesPermission,
		WriteAdminSettingValuesPermission,
		InviteWaitlistSignupsPermission,
		ConvertWaitlistSignupsPermission,
		EraseWaitlistSignupsPermission,

		// Reading signups is a service admin's and nobody else's, which costs a member
		// the ability to see their own place in a queue.
		//
		// platform puts four reads behind this one grant — a signup by id, one by the
		// address it was made with, a list's page, and one subject's signups — and warns
		// that it is the grant to think hardest about, because a holder can ask whether
		// any address they can type is on any list. Granting it to a member would hand
		// every signed-in user that question about every other user's address.
		//
		// The obvious narrower grant is not available either: ListSignupsForSubject takes
		// the subject from the request, and unlike settings there is no authorizer to
		// refuse a subject that is not the caller's own. So a member holding it could read
		// anybody's signups by naming them. Filed upstream; when a subject authorizer or a
		// split grant lands, the own-signup half comes back to the member.
		ReadWaitlistSignupsPermission,

		// The fleet-wide ledger reads, which are separate permissions from the
		// account-scoped ones an account admin holds precisely so that reading one
		// account's money is not reading everybody's.
		ListAllSubscriptionsPermission,
		ListAllPurchasesPermission,
		ListAllTransactionsPermission,
		ReadTransactionsPermission,
		ArchivePurchasesPermission,
		ArchiveTransactionsPermission,
	}

	// ServiceDataAdminPermissions is every service data admin permission.
	ServiceDataAdminPermissions = []Permission{
		CreateValidInstrumentsPermission,
		UpdateValidInstrumentsPermission,
		ArchiveValidInstrumentsPermission,
		CreateValidVesselsPermission,
		UpdateValidVesselsPermission,
		ArchiveValidVesselsPermission,
		CreateValidIngredientsPermission,
		UpdateValidIngredientsPermission,
		ArchiveValidIngredientsPermission,
		CreateValidIngredientGroupsPermission,
		UpdateValidIngredientGroupsPermission,
		ArchiveValidIngredientGroupsPermission,
		CreateValidPreparationsPermission,
		UpdateValidPreparationsPermission,
		ArchiveValidPreparationsPermission,
		CreateValidMeasurementUnitsPermission,
		UpdateValidMeasurementUnitsPermission,
		ArchiveValidMeasurementUnitsPermission,
		CreateValidMeasurementUnitConversionsPermission,
		UpdateValidMeasurementUnitConversionsPermission,
		ArchiveValidMeasurementUnitConversionsPermission,
		CreateValidIngredientPreparationsPermission,
		UpdateValidIngredientPreparationsPermission,
		ArchiveValidIngredientPreparationsPermission,
		CreateValidPrepTaskConfigsPermission,
		UpdateValidPrepTaskConfigsPermission,
		ArchiveValidPrepTaskConfigsPermission,
		CreateValidIngredientStateIngredientsPermission,
		UpdateValidIngredientStateIngredientsPermission,
		ArchiveValidIngredientStateIngredientsPermission,
		CreateValidPreparationInstrumentsPermission,
		UpdateValidPreparationInstrumentsPermission,
		ArchiveValidPreparationInstrumentsPermission,
		CreateValidPreparationVesselsPermission,
		UpdateValidPreparationVesselsPermission,
		ArchiveValidPreparationVesselsPermission,
		CreateValidIngredientMeasurementUnitsPermission,
		UpdateValidIngredientMeasurementUnitsPermission,
		ArchiveValidIngredientMeasurementUnitsPermission,
		CreateValidIngredientStatesPermission,
		UpdateValidIngredientStatesPermission,
		ArchiveValidIngredientStatesPermission,
	}

	// AccountAdminPermissions is every account admin permission.
	AccountAdminPermissions = []Permission{
		UpdateAccountPermission,
		ArchiveAccountPermission,
		TransferAccountPermission,
		InviteUserToAccountPermission,
		ModifyMemberPermissionsForAccountPermission,
		RemoveMemberAccountPermission,
		CreateIssueReportsPermission,
		UpdateIssueReportsPermission,
		ArchiveIssueReportsPermission,
		// Working the queue is an account admin's, because the queue is the account's:
		// a report is filed in the account it is about and no read crosses that line.
		// These are the two halves of what "update.issue_reports" gated before the
		// adoption split it — paging the queue by status or subject, and moving a
		// report through it.
		TriageIssueReportsPermission,
		TransitionIssueReportsPermission,
		CreateMealPlansPermission,
		UpdateMealPlansPermission,
		ArchiveMealPlansPermission,
		CreateMealPlanEventsPermission,
		UpdateMealPlanEventsPermission,
		ArchiveMealPlanEventsPermission,
		CreateMealPlanOptionsPermission,
		UpdateMealPlanOptionsPermission,
		ArchiveMealPlanOptionsPermission,
		CreateAccountInstrumentOwnershipsPermission,
		UpdateAccountInstrumentOwnershipsPermission,
		ArchiveAccountInstrumentOwnershipsPermission,
		CreateMealListsPermission,
		ReadMealListsPermission,
		UpdateMealListsPermission,
		ArchiveMealListsPermission,
		CreateRecipeListsPermission,
		ReadRecipeListsPermission,
		UpdateRecipeListsPermission,
		ArchiveRecipeListsPermission,
		CreateCheckoutSessionPermission,
		CancelSubscriptionPermission,
		ReadPurchasesPermission,
		ReadPaymentHistoryPermission,
		ReadSubscriptionsPermission,

		// Platform's webhook writes, which are account-scoped in this application:
		// an account's admin manages that account's endpoints and what they hear
		// about. They are finer than the four this repository's own webhooks
		// service used — an endpoint, a subscription and a secret rotation are
		// separately grantable now, where "update.webhooks" covered all three.
		// The reads are a member's, below, as "read.webhooks" was.
		SaveWebhookEndpointsPermission,
		ArchiveWebhookEndpointsPermission,
		RotateWebhookSecretPermission,
		AddWebhookSubscriptionsPermission,
		ArchiveWebhookSubscriptionsPermission,
	}

	// AccountMemberPermissions is every account member permission.
	AccountMemberPermissions = []Permission{
		ReportAnalyticsEventsPermission,
		ReadIssueReportsPermission,
		ReadAuditLogEntriesPermission,
		ReadOAuth2ClientsPermission,
		ReadSettingDefinitionsPermission,
		CreateUploadedMediaPermission,
		ReadUploadedMediaPermission,
		ArchiveUploadedMediaPermission,
		CreateMealsPermission,
		ReadMealsPermission,
		UpdateMealsPermission,
		ArchiveMealsPermission,
		CreateRecipesPermission,
		CloneRecipesPermission,
		ReadRecipesPermission,
		SearchRecipesPermission,
		UpdateRecipesPermission,
		ArchiveRecipesPermission,
		CreateRecipeStepsPermission,
		ReadRecipeStepsPermission,
		SearchRecipeStepsPermission,
		UpdateRecipeStepsPermission,
		ArchiveRecipeStepsPermission,
		CreateRecipePrepTasksPermission,
		ReadRecipePrepTasksPermission,
		UpdateRecipePrepTasksPermission,
		ArchiveRecipePrepTasksPermission,
		CreateRecipeStepInstrumentsPermission,
		ReadRecipeStepInstrumentsPermission,
		SearchRecipeStepInstrumentsPermission,
		UpdateRecipeStepInstrumentsPermission,
		ArchiveRecipeStepInstrumentsPermission,
		CreateRecipeStepVesselsPermission,
		ReadRecipeStepVesselsPermission,
		SearchRecipeStepVesselsPermission,
		UpdateRecipeStepVesselsPermission,
		ArchiveRecipeStepVesselsPermission,
		CreateRecipeStepIngredientsPermission,
		ReadRecipeStepIngredientsPermission,
		SearchRecipeStepIngredientsPermission,
		UpdateRecipeStepIngredientsPermission,
		ArchiveRecipeStepIngredientsPermission,
		CreateRecipeStepCompletionConditionsPermission,
		ReadRecipeStepCompletionConditionsPermission,
		SearchRecipeStepCompletionConditionsPermission,
		UpdateRecipeStepCompletionConditionsPermission,
		ArchiveRecipeStepCompletionConditionsPermission,
		CreateRecipeStepProductsPermission,
		ReadRecipeStepProductsPermission,
		SearchRecipeStepProductsPermission,
		UpdateRecipeStepProductsPermission,
		ArchiveRecipeStepProductsPermission,
		ReadValidInstrumentsPermission,
		SearchValidInstrumentsPermission,
		ReadValidVesselsPermission,
		SearchValidVesselsPermission,
		ReadValidIngredientsPermission,
		SearchValidIngredientsPermission,
		ReadValidIngredientGroupsPermission,
		SearchValidIngredientGroupsPermission,
		ReadValidPreparationsPermission,
		SearchValidPreparationsPermission,
		ReadValidMeasurementUnitsPermission,
		SearchValidMeasurementUnitsPermission,
		ReadValidMeasurementUnitConversionsPermission,
		ReadValidIngredientPreparationsPermission,
		SearchValidIngredientPreparationsPermission,
		ReadValidIngredientStateIngredientsPermission,
		SearchValidIngredientStateIngredientsPermission,
		ReadValidPreparationInstrumentsPermission,
		SearchValidPreparationInstrumentsPermission,
		ReadValidPreparationVesselsPermission,
		SearchValidPreparationVesselsPermission,
		ReadValidIngredientMeasurementUnitsPermission,
		SearchValidIngredientMeasurementUnitsPermission,
		ReadMealPlansPermission,
		SearchMealPlansPermission,
		ReadMealPlanEventsPermission,
		ReadMealPlanOptionsPermission,
		SearchMealPlanOptionsPermission,
		ReadValidIngredientStatesPermission,
		ReadMealPlanGroceryListItemsPermission,
		UpdateMealPlanGroceryListItemsPermission,
		ArchiveMealPlanGroceryListItemsPermission,
		CreateMealPlanOptionVotesPermission,
		ReadMealPlanOptionVotesPermission,
		SearchMealPlanOptionVotesPermission,
		UpdateMealPlanOptionVotesPermission,
		ArchiveMealPlanOptionVotesPermission,
		CreateMealPlanRecipeOptionSelectionsPermission,
		ReadMealPlanRecipeOptionSelectionsPermission,
		UpdateMealPlanRecipeOptionSelectionsPermission,
		ArchiveMealPlanRecipeOptionSelectionsPermission,
		WriteSettingValuesPermission,
		ReadSettingValuesPermission,
		ReadMealPlanTasksPermission,
		UpdateMealPlanTasksPermission,
		CreateUserIngredientPreferencesPermission,
		ReadUserIngredientPreferencesPermission,
		UpdateUserIngredientPreferencesPermission,
		ArchiveUserIngredientPreferencesPermission,
		ReadAccountInstrumentOwnershipsPermission,
		CreateRecipeRatingsPermission,
		ReadRecipeRatingsPermission,
		CreateCommentsPermission,
		ReadCommentsPermission,
		UpdateCommentsPermission,
		ArchiveCommentsPermission,
		UpdateRecipeRatingsPermission,
		ArchiveRecipeRatingsPermission,
		ReadUserNotificationsPermission,
		MarkUserNotificationsReadPermission,
		ArchiveUserNotificationsPermission,
		CreateUserDeviceTokensPermission,
		ReadUserDeviceTokensPermission,
		ArchiveUserDeviceTokensPermission,
		JoinWaitlistsPermission,
		UpdateWaitlistSignupsPermission,
		ArchiveWaitlistSignupsPermission,
		ReadWaitlistsPermission,
		ReadValidPrepTaskConfigsPermission,
		CreateCheckoutSessionPermission,
		CancelSubscriptionPermission,
		ReadPurchasesPermission,
		ReadPaymentHistoryPermission,
		ReadSubscriptionsPermission,
		CreateUserDataReportsPermission,
		ReadUserDataReportsPermission,
		DestroyUserDataPermission,

		// The webhook reads, which a member held as "read.webhooks" before the
		// adoption split it. A member seeing an endpoint learns nothing they
		// should not: platform's converter never renders the signing secret, for
		// exactly the reason a read permission would otherwise be a write one.
		ReadWebhookEndpointsPermission,
		ReadWebhookSubscriptionsPermission,
		ReadWebhookAttemptsPermission,
		ReadWebhookEventTypesPermission,
	}
)
