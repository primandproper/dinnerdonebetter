/**
 * Admin gRPC client factory. Uses AdminLoginForToken (no OAuth exchange) and exposes
 * all admin-required methods with token-first signature.
 */

import * as grpc from '@grpc/grpc-js';
import type { Metadata } from '@grpc/grpc-js';
import { AuthServiceClient } from './auth/auth_service.js';
import type {
  AdminLoginForTokenRequest,
  AdminListSessionsForUserRequest,
  AdminRevokeUserSessionRequest,
  AdminRevokeAllUserSessionsRequest,
} from './auth/auth_service_types.js';
import type {
  LoginForTokenResponse,
  ListActiveSessionsResponse,
  RevokeSessionResponse,
  RevokeAllOtherSessionsResponse,
  RevokeCurrentSessionRequest,
  RevokeCurrentSessionResponse,
} from './auth/auth_service_types.js';
import { IdentityServiceClient } from './identity/identity_service.js';
import { OAuthServiceClient } from './oauth/oauth_service.js';
import { PaymentsServiceClient } from './payments/payments_service.js';
import { SettingsServiceClient } from './settings/settings_service.js';
import { WaitlistsServiceClient } from './waitlists/waitlists_service.js';
import { IssueReportsServiceClient } from './issue_reports/issue_reports_service.js';
import { InternalOperationsClient } from './internal_ops/internal_ops_service.js';
import { AuditServiceClient } from './audit/audit_service.js';
import { AnalyticsServiceClient } from './analytics/analytics_service.js';
import { MealPlanningServiceClient } from './mealplanning/mealplanning_service.js';
import type { CreateRecipeRequest, CreateRecipeResponse } from './mealplanning/mealplanning_service_types.js';
import type { GrpcClientConfig } from './create-clients.js';

function promisifyUnary<TRequest, TResponse>(
  call: (
    req: TRequest,
    metadata: Metadata,
    callback: (err: grpc.ServiceError | null, res: TResponse) => void,
  ) => grpc.ClientUnaryCall,
): (req: TRequest, metadata: Metadata) => Promise<TResponse> {
  return (req, metadata) =>
    new Promise((resolve, reject) => {
      call(req, metadata, (err, res) => {
        if (err) reject(err);
        else if (res) resolve(res);
        else reject(new Error('No response'));
      });
    });
}

export function createAdminGrpcClients(config: GrpcClientConfig) {
  const credentials = config.insecure ? grpc.credentials.createInsecure() : grpc.credentials.createSsl();
  const serverUrl = config.serverUrl;

  let authClient: AuthServiceClient | null = null;
  let identityClient: IdentityServiceClient | null = null;
  let oauthClient: OAuthServiceClient | null = null;
  let paymentsClient: PaymentsServiceClient | null = null;
  let settingsClient: SettingsServiceClient | null = null;
  let waitlistsClient: WaitlistsServiceClient | null = null;
  let issueReportsClient: IssueReportsServiceClient | null = null;
  let internalOpsClient: InternalOperationsClient | null = null;
  let auditClient: AuditServiceClient | null = null;
  let analyticsClient: AnalyticsServiceClient | null = null;
  let mealplanningClient: MealPlanningServiceClient | null = null;

  const get = {
    auth: () => {
      if (!authClient) authClient = new AuthServiceClient(serverUrl, credentials);
      return authClient;
    },
    identity: () => {
      if (!identityClient) identityClient = new IdentityServiceClient(serverUrl, credentials);
      return identityClient;
    },
    oauth: () => {
      if (!oauthClient) oauthClient = new OAuthServiceClient(serverUrl, credentials);
      return oauthClient;
    },
    payments: () => {
      if (!paymentsClient) paymentsClient = new PaymentsServiceClient(serverUrl, credentials);
      return paymentsClient;
    },
    settings: () => {
      if (!settingsClient) settingsClient = new SettingsServiceClient(serverUrl, credentials);
      return settingsClient;
    },
    waitlists: () => {
      if (!waitlistsClient) waitlistsClient = new WaitlistsServiceClient(serverUrl, credentials);
      return waitlistsClient;
    },
    issueReports: () => {
      if (!issueReportsClient) issueReportsClient = new IssueReportsServiceClient(serverUrl, credentials);
      return issueReportsClient;
    },
    internalOps: () => {
      if (!internalOpsClient) internalOpsClient = new InternalOperationsClient(serverUrl, credentials);
      return internalOpsClient;
    },
    audit: () => {
      if (!auditClient) auditClient = new AuditServiceClient(serverUrl, credentials);
      return auditClient;
    },
    analytics: () => {
      if (!analyticsClient) analyticsClient = new AnalyticsServiceClient(serverUrl, credentials);
      return analyticsClient;
    },
    mealplanning: () => {
      if (!mealplanningClient) mealplanningClient = new MealPlanningServiceClient(serverUrl, credentials);
      return mealplanningClient;
    },
  };

  function authMetadata(token: string): Metadata {
    const m = new grpc.Metadata();
    m.add('authorization', `Bearer ${token}`);
    return m;
  }
  const emptyMetadata = new grpc.Metadata();

  return {
    authMetadata,

    adminLoginForToken: (request: AdminLoginForTokenRequest): Promise<LoginForTokenResponse> =>
      promisifyUnary<AdminLoginForTokenRequest, LoginForTokenResponse>(get.auth().adminLoginForToken.bind(get.auth()))(
        request,
        emptyMetadata,
      ),

    beginPasskeyAuthentication: (request: { username?: string }) =>
      promisifyUnary(get.auth().beginPasskeyAuthentication.bind(get.auth()))(
        { username: request.username ?? '' },
        emptyMetadata,
      ),
    finishPasskeyAuthentication: (request: { challenge: string; username: string; assertionResponse: Uint8Array }) =>
      promisifyUnary(get.auth().finishPasskeyAuthentication.bind(get.auth()))(request, emptyMetadata),

    adminListSessionsForUser: (
      token: string,
      request: AdminListSessionsForUserRequest,
    ): Promise<ListActiveSessionsResponse> =>
      promisifyUnary<AdminListSessionsForUserRequest, ListActiveSessionsResponse>(
        get.auth().adminListSessionsForUser.bind(get.auth()),
      )(request, authMetadata(token)),

    adminRevokeUserSession: (token: string, request: AdminRevokeUserSessionRequest): Promise<RevokeSessionResponse> =>
      promisifyUnary<AdminRevokeUserSessionRequest, RevokeSessionResponse>(
        get.auth().adminRevokeUserSession.bind(get.auth()),
      )(request, authMetadata(token)),

    adminRevokeAllUserSessions: (
      token: string,
      request: AdminRevokeAllUserSessionsRequest,
    ): Promise<RevokeAllOtherSessionsResponse> =>
      promisifyUnary<AdminRevokeAllUserSessionsRequest, RevokeAllOtherSessionsResponse>(
        get.auth().adminRevokeAllUserSessions.bind(get.auth()),
      )(request, authMetadata(token)),

    revokeCurrentSession: (token: string): Promise<RevokeCurrentSessionResponse> =>
      promisifyUnary<RevokeCurrentSessionRequest, RevokeCurrentSessionResponse>(
        get.auth().revokeCurrentSession.bind(get.auth()),
      )({}, authMetadata(token)),

    // Identity – request types are intentionally loose; callers pass proto-shaped objects
    getUser: (token: string, request: { userId: string }) =>
      promisifyUnary(get.identity().getUser.bind(get.identity()))(request as any, authMetadata(token)),
    getUsers: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().getUsers.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['getUsers']>[0],
        authMetadata(token),
      ),
    getAccount: (token: string, request: { accountId: string }) =>
      promisifyUnary(get.identity().getAccount.bind(get.identity()))(request as any, authMetadata(token)),
    getAccounts: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().getAccounts.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['getAccounts']>[0],
        authMetadata(token),
      ),
    searchForUsers: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().searchForUsers.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['searchForUsers']>[0],
        authMetadata(token),
      ),
    getUsersForAccount: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().getUsersForAccount.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['getUsersForAccount']>[0],
        authMetadata(token),
      ),
    getAccountsForUser: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().getAccountsForUser.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['getAccountsForUser']>[0],
        authMetadata(token),
      ),
    adminUpdateUserStatus: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().adminUpdateUserStatus.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['adminUpdateUserStatus']>[0],
        authMetadata(token),
      ),
    adminSetPasswordChangeRequired: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().adminSetPasswordChangeRequired.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['adminSetPasswordChangeRequired']>[0],
        authMetadata(token),
      ),
    updateUserDetails: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().updateUserDetails.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['updateUserDetails']>[0],
        authMetadata(token),
      ),
    updateAccount: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.identity().updateAccount.bind(get.identity()))(
        request as unknown as Parameters<IdentityServiceClient['updateAccount']>[0],
        authMetadata(token),
      ),

    // OAuth
    getOAuth2Clients: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.oauth().getOAuth2Clients.bind(get.oauth()))(
        request as unknown as Parameters<OAuthServiceClient['getOAuth2Clients']>[0],
        authMetadata(token),
      ),
    getOAuth2Client: (token: string, request: { oauth2ClientId: string }) =>
      promisifyUnary(get.oauth().getOAuth2Client.bind(get.oauth()))(request as any, authMetadata(token)),
    createOAuth2Client: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.oauth().createOAuth2Client.bind(get.oauth()))(
        request as unknown as Parameters<OAuthServiceClient['createOAuth2Client']>[0],
        authMetadata(token),
      ),
    archiveOAuth2Client: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.oauth().archiveOAuth2Client.bind(get.oauth()))(
        request as unknown as Parameters<OAuthServiceClient['archiveOAuth2Client']>[0],
        authMetadata(token),
      ),

    // Payments
    getProducts: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().getProducts.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['getProducts']>[0],
        authMetadata(token),
      ),
    getProduct: (token: string, request: { productId: string }) =>
      promisifyUnary(get.payments().getProduct.bind(get.payments()))(request as any, authMetadata(token)),
    createProduct: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().createProduct.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['createProduct']>[0],
        authMetadata(token),
      ),
    updateProduct: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().updateProduct.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['updateProduct']>[0],
        authMetadata(token),
      ),
    archiveProduct: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().archiveProduct.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['archiveProduct']>[0],
        authMetadata(token),
      ),
    getSubscription: (token: string, request: { subscriptionId: string }) =>
      promisifyUnary(get.payments().getSubscription.bind(get.payments()))(request as any, authMetadata(token)),
    createSubscription: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().createSubscription.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['createSubscription']>[0],
        authMetadata(token),
      ),
    updateSubscription: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().updateSubscription.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['updateSubscription']>[0],
        authMetadata(token),
      ),
    archiveSubscription: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().archiveSubscription.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['archiveSubscription']>[0],
        authMetadata(token),
      ),
    getSubscriptionsForAccount: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.payments().getSubscriptionsForAccount.bind(get.payments()))(
        request as unknown as Parameters<PaymentsServiceClient['getSubscriptionsForAccount']>[0],
        authMetadata(token),
      ),

    // Settings
    //
    // The catalog is administrative: a deployment decides which settings exist,
    // and these are the calls that decide it. A person's own answers are not
    // here — those are resolved for the signed-in user by the consumer client.
    getSettingDefinitions: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.settings().getSettingDefinitions.bind(get.settings()))(
        request as unknown as Parameters<SettingsServiceClient['getSettingDefinitions']>[0],
        authMetadata(token),
      ),
    getSettingDefinition: (token: string, request: { settingDefinitionId: string }) =>
      promisifyUnary(get.settings().getSettingDefinition.bind(get.settings()))(request as any, authMetadata(token)),
    getSettingDefinitionByName: (token: string, request: { settingName: string }) =>
      promisifyUnary(get.settings().getSettingDefinitionByName.bind(get.settings()))(
        request as any,
        authMetadata(token),
      ),
    createSettingDefinition: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.settings().createSettingDefinition.bind(get.settings()))(
        request as unknown as Parameters<SettingsServiceClient['createSettingDefinition']>[0],
        authMetadata(token),
      ),
    updateSettingDefinition: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.settings().updateSettingDefinition.bind(get.settings()))(
        request as unknown as Parameters<SettingsServiceClient['updateSettingDefinition']>[0],
        authMetadata(token),
      ),
    archiveSettingDefinition: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.settings().archiveSettingDefinition.bind(get.settings()))(
        request as unknown as Parameters<SettingsServiceClient['archiveSettingDefinition']>[0],
        authMetadata(token),
      ),
    // Who has overridden a setting. This is the administrative read that the old
    // per-account list stood in for, and it is service-admin only on the server.
    getSettingValuesForDefinition: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.settings().getSettingValuesForDefinition.bind(get.settings()))(
        request as unknown as Parameters<SettingsServiceClient['getSettingValuesForDefinition']>[0],
        authMetadata(token),
      ),

    // Waitlists
    getWaitlists: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.waitlists().getWaitlists.bind(get.waitlists()))(
        request as unknown as Parameters<WaitlistsServiceClient['getWaitlists']>[0],
        authMetadata(token),
      ),
    getWaitlist: (token: string, request: { waitlistId: string }) =>
      promisifyUnary(get.waitlists().getWaitlist.bind(get.waitlists()))(request as any, authMetadata(token)),
    getWaitlistSignupsForWaitlist: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.waitlists().getWaitlistSignupsForWaitlist.bind(get.waitlists()))(
        request as unknown as Parameters<WaitlistsServiceClient['getWaitlistSignupsForWaitlist']>[0],
        authMetadata(token),
      ),

    // Issue Reports
    getIssueReports: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.issueReports().getIssueReports.bind(get.issueReports()))(
        request as unknown as Parameters<IssueReportsServiceClient['getIssueReports']>[0],
        authMetadata(token),
      ),
    getIssueReport: (token: string, request: { issueReportId: string }) =>
      promisifyUnary(get.issueReports().getIssueReport.bind(get.issueReports()))(request as any, authMetadata(token)),
    updateIssueReport: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.issueReports().updateIssueReport.bind(get.issueReports()))(
        request as unknown as Parameters<IssueReportsServiceClient['updateIssueReport']>[0],
        authMetadata(token),
      ),
    archiveIssueReport: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.issueReports().archiveIssueReport.bind(get.issueReports()))(
        request as unknown as Parameters<IssueReportsServiceClient['archiveIssueReport']>[0],
        authMetadata(token),
      ),

    // Internal Ops
    testQueueMessage: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.internalOps().testQueueMessage.bind(get.internalOps()))(
        request as unknown as Parameters<InternalOperationsClient['testQueueMessage']>[0],
        authMetadata(token),
      ),

    // Audit
    getAuditLogEntriesForUser: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.audit().getAuditLogEntriesForUser.bind(get.audit()))(
        request as unknown as Parameters<AuditServiceClient['getAuditLogEntriesForUser']>[0],
        authMetadata(token),
      ),
    getAuditLogEntriesForAccount: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.audit().getAuditLogEntriesForAccount.bind(get.audit()))(
        request as unknown as Parameters<AuditServiceClient['getAuditLogEntriesForAccount']>[0],
        authMetadata(token),
      ),

    // Analytics
    trackEvent: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.analytics().trackEvent.bind(get.analytics()))(
        request as unknown as Parameters<AnalyticsServiceClient['trackEvent']>[0],
        authMetadata(token),
      ),
    trackAnonymousEvent: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.analytics().trackAnonymousEvent.bind(get.analytics()))(
        request as unknown as Parameters<AnalyticsServiceClient['trackAnonymousEvent']>[0],
        authMetadata(token),
      ),

    // Meal planning - recipes
    getRecipes: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getRecipes.bind(get.mealplanning()))(
        request as unknown as Parameters<MealPlanningServiceClient['getRecipes']>[0],
        authMetadata(token),
      ),
    getRecipe: (token: string, request: { recipeId: string }) =>
      promisifyUnary(get.mealplanning().getRecipe.bind(get.mealplanning()))(request as any, authMetadata(token)),
    searchForRecipes: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForRecipes.bind(get.mealplanning()))(
        request as unknown as Parameters<MealPlanningServiceClient['searchForRecipes']>[0],
        authMetadata(token),
      ),
    createRecipe: (token: string, request: CreateRecipeRequest): Promise<CreateRecipeResponse> =>
      promisifyUnary<CreateRecipeRequest, CreateRecipeResponse>(
        get.mealplanning().createRecipe.bind(get.mealplanning()),
      )(request, authMetadata(token)),

    // Meal planning - valid ingredients
    getValidIngredients: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidIngredients.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    searchForValidIngredients: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForValidIngredients.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    searchValidIngredientsByPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchValidIngredientsByPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidIngredient: (token: string, request: { validIngredientId: string }) =>
      promisifyUnary(get.mealplanning().getValidIngredient.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidIngredient: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidIngredient.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    updateValidIngredient: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().updateValidIngredient.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidIngredientMeasurementUnitsByIngredient: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidIngredientMeasurementUnitsByIngredient.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidIngredientMeasurementUnit: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidIngredientMeasurementUnit.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    archiveValidIngredientMeasurementUnit: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().archiveValidIngredientMeasurementUnit.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidIngredientPreparationsByIngredient: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidIngredientPreparationsByIngredient.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidIngredientPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidIngredientPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    archiveValidIngredientPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().archiveValidIngredientPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - valid instruments
    getValidInstruments: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidInstruments.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    searchForValidInstruments: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForValidInstruments.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidInstrument: (token: string, request: { validInstrumentId: string }) =>
      promisifyUnary(get.mealplanning().getValidInstrument.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidInstrument: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidInstrument.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    updateValidInstrument: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().updateValidInstrument.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidPreparationInstrumentsByInstrument: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidPreparationInstrumentsByInstrument.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidPreparationInstrument: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidPreparationInstrument.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    archiveValidPreparationInstrument: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().archiveValidPreparationInstrument.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - valid vessels
    getValidVessels: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidVessels.bind(get.mealplanning()))(request as any, authMetadata(token)),
    searchForValidVessels: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForValidVessels.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidVessel: (token: string, request: { validVesselId: string }) =>
      promisifyUnary(get.mealplanning().getValidVessel.bind(get.mealplanning()))(request as any, authMetadata(token)),
    createValidVessel: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidVessel.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    updateValidVessel: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().updateValidVessel.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidPreparationVesselsByVessel: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidPreparationVesselsByVessel.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidPreparationVessel: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidPreparationVessel.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    archiveValidPreparationVessel: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().archiveValidPreparationVessel.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - valid measurement units
    getValidMeasurementUnits: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidMeasurementUnits.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    searchForValidMeasurementUnits: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForValidMeasurementUnits.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidMeasurementUnit: (token: string, request: { validMeasurementUnitId: string }) =>
      promisifyUnary(get.mealplanning().getValidMeasurementUnit.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidMeasurementUnit: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidMeasurementUnit.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    updateValidMeasurementUnit: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().updateValidMeasurementUnit.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidMeasurementUnitConversionsForUnit: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidMeasurementUnitConversionsForUnit.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidMeasurementUnitConversion: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidMeasurementUnitConversion.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    archiveValidMeasurementUnitConversion: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().archiveValidMeasurementUnitConversion.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidMeasurementUnitConversionsForIngredients: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidMeasurementUnitConversionsForIngredients.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - valid ingredient states
    getValidIngredientStates: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidIngredientStates.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    searchForValidIngredientStates: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForValidIngredientStates.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidIngredientState: (token: string, request: { validIngredientStateId: string }) =>
      promisifyUnary(get.mealplanning().getValidIngredientState.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidIngredientState: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidIngredientState.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    updateValidIngredientState: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().updateValidIngredientState.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - valid preparations
    getValidPreparations: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidPreparations.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    searchForValidPreparations: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().searchForValidPreparations.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidPreparation: (token: string, request: { validPreparationId: string }) =>
      promisifyUnary(get.mealplanning().getValidPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    createValidPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().createValidPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    updateValidPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().updateValidPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidPreparationInstrumentsByPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidPreparationInstrumentsByPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidPreparationVesselsByPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidPreparationVesselsByPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidIngredientPreparationsByPreparation: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidIngredientPreparationsByPreparation.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - valid prep task configs
    getValidPrepTaskConfig: (token: string, request: { validPrepTaskConfigId: string }) =>
      promisifyUnary(get.mealplanning().getValidPrepTaskConfig.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
    getValidPrepTaskConfigs: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getValidPrepTaskConfigs.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),

    // Meal planning - conversion mismatches
    getMeasurementUnitConversionMismatches: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.mealplanning().getMeasurementUnitConversionMismatches.bind(get.mealplanning()))(
        request as any,
        authMetadata(token),
      ),
  };
}
