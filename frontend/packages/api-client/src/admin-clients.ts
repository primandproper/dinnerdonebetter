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
import { InternalOperationsClient } from './internal_ops/internal_ops_service.js';
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
  let internalOpsClient: InternalOperationsClient | null = null;
  let analyticsClient: AnalyticsServiceClient | null = null;
  let mealplanningClient: MealPlanningServiceClient | null = null;

  const get = {
    auth: () => {
      if (!authClient) authClient = new AuthServiceClient(serverUrl, credentials);
      return authClient;
    },
    internalOps: () => {
      if (!internalOpsClient) internalOpsClient = new InternalOperationsClient(serverUrl, credentials);
      return internalOpsClient;
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

    // Internal Ops
    testQueueMessage: (token: string, request: Record<string, unknown>) =>
      promisifyUnary(get.internalOps().testQueueMessage.bind(get.internalOps()))(
        request as unknown as Parameters<InternalOperationsClient['testQueueMessage']>[0],
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
