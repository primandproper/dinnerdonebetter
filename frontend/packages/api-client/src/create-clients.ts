/**
 * gRPC client factory. Creates client singletons and promisified API from config.
 * Per-request auth is passed via call-level metadata, not per-client credentials.
 */

import * as grpc from '@grpc/grpc-js';
import { QueryFilter } from './primandproper/platform/filtering/v1/filtering.js';
import { AuthServiceClient } from './auth/auth_service.js';
import type {
  LoginForTokenRequest,
  LoginForTokenResponse,
  GetSelfRequest,
  GetSelfResponse,
  GetActiveAccountRequest,
  GetActiveAccountResponse,
  RequestPasswordResetTokenRequest,
  RequestPasswordResetTokenResponse,
  RedeemPasswordResetTokenRequest,
  RedeemPasswordResetTokenResponse,
  VerifyEmailAddressRequest,
  VerifyEmailAddressResponse,
  ListPasskeysRequest,
  ListPasskeysResponse,
  ArchivePasskeyRequest,
  ArchivePasskeyResponse,
  BeginPasskeyRegistrationRequest,
  BeginPasskeyRegistrationResponse,
  FinishPasskeyRegistrationRequest,
  FinishPasskeyRegistrationResponse,
  BeginPasskeyAuthenticationRequest,
  BeginPasskeyAuthenticationResponse,
  FinishPasskeyAuthenticationRequest,
  ListActiveSessionsRequest,
  ListActiveSessionsResponse,
  RevokeSessionRequest,
  RevokeSessionResponse,
  RevokeAllOtherSessionsRequest,
  RevokeAllOtherSessionsResponse,
  RevokeCurrentSessionRequest,
  RevokeCurrentSessionResponse,
  ExchangeTokenRequest,
  ExchangeTokenResponse,
  UpdateUserEmailAddressRequest,
  UpdateUserEmailAddressResponse,
  UpdateUserUsernameRequest,
  UpdateUserUsernameResponse,
} from './auth/auth_service_types.js';
// The directory is platform's, so its client and its messages come from the schema
// platform-go ships rather than from one of this repository's own protos. The RPC names
// changed with it: CreateAccountInvitation is Invite, GetAccountsForUser is
// ListAccountsForUser, and the three calls that each changed one field of a user are one
// UpdateProfile.
import { IdentityServiceClient } from './primandproper/platform/identity/v1/identity.js';
import type {
  ListAccountsForUserRequest,
  ListAccountsForUserResponse,
  ListAccountMembersRequest,
  ListAccountMembersResponse,
  ListInvitationsFromUserRequest,
  ListInvitationsFromUserResponse,
  InviteRequest,
  InviteResponse,
  CancelInvitationRequest,
  CancelInvitationResponse,
  SetMembershipRolesRequest,
  SetMembershipRolesResponse,
  UpdateAccountRequest,
  UpdateAccountResponse,
  UpdateProfileRequest,
  UpdateProfileResponse,
} from './primandproper/platform/identity/v1/identity.js';
import { SettingsServiceClient } from './settings/settings_service.js';
import type {
  GetSettingDefinitionsRequest,
  GetSettingDefinitionsResponse,
  GetSettingValuesRequest,
  GetSettingValuesResponse,
  ResolveSettingsRequest,
  ResolveSettingsResponse,
  SetSettingValueRequest,
  SetSettingValueResponse,
  ClearSettingValueRequest,
  ClearSettingValueResponse,
} from './settings/settings_service_types.js';
import { MealPlanningServiceClient } from './mealplanning/mealplanning_service.js';
import { AnalyticsServiceClient } from './analytics/analytics_service.js';
import type {
  SearchForValidPreparationsRequest,
  SearchForValidPreparationsResponse,
  GetValidPreparationInstrumentsByPreparationRequest,
  GetValidPreparationInstrumentsByPreparationResponse,
  GetValidPreparationVesselsByPreparationRequest,
  GetValidPreparationVesselsByPreparationResponse,
  SearchValidIngredientsByPreparationRequest,
  SearchValidIngredientsByPreparationResponse,
  SearchValidMeasurementUnitsByIngredientRequest,
  SearchValidMeasurementUnitsByIngredientResponse,
  GetValidIngredientMeasurementUnitsByIngredientRequest,
  GetValidIngredientMeasurementUnitsByIngredientResponse,
  GetValidIngredientPreparationsByPreparationRequest,
  GetValidIngredientPreparationsByPreparationResponse,
  SearchForValidMeasurementUnitsRequest,
  SearchForValidMeasurementUnitsResponse,
  SearchForValidVesselsRequest,
  SearchForValidVesselsResponse,
  SearchForValidIngredientStatesRequest,
  SearchForValidIngredientStatesResponse,
  GetValidVesselsRequest,
  GetValidVesselsResponse,
  GetValidPreparationsRequest,
  GetValidPreparationsResponse,
  GetValidMeasurementUnitsRequest,
  GetValidMeasurementUnitsResponse,
  GetValidIngredientStatesRequest,
  GetValidIngredientStatesResponse,
  CreateRecipeRequest,
  CreateRecipeResponse,
  SearchForRecipesRequest,
  SearchForRecipesResponse,
} from './mealplanning/mealplanning_service_types.js';
import type { Metadata } from '@grpc/grpc-js';

export interface GrpcClientConfig {
  serverUrl: string;
  insecure?: boolean;
}

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

// QueryFilter is platform's schema, and its four timestamp windows are message
// fields rather than `optional` scalars, so a bare object literal is not one.
// The generated factory fills every field, which is what it is for.
const defaultSearchFilter = QueryFilter.create({ maxResponseSize: 20 });

export function createGrpcClients(config: GrpcClientConfig) {
  const credentials = config.insecure ? grpc.credentials.createInsecure() : grpc.credentials.createSsl();
  const serverUrl = config.serverUrl;

  let authClient: AuthServiceClient | null = null;
  let identityClient: IdentityServiceClient | null = null;
  let settingsClient: SettingsServiceClient | null = null;
  let mealplanningClient: MealPlanningServiceClient | null = null;
  let analyticsClient: AnalyticsServiceClient | null = null;

  function getAuthClient(): AuthServiceClient {
    if (!authClient) authClient = new AuthServiceClient(serverUrl, credentials);
    return authClient;
  }
  function getIdentityClient(): IdentityServiceClient {
    if (!identityClient) identityClient = new IdentityServiceClient(serverUrl, credentials);
    return identityClient;
  }
  function getSettingsClient(): SettingsServiceClient {
    if (!settingsClient) settingsClient = new SettingsServiceClient(serverUrl, credentials);
    return settingsClient;
  }
  function getMealplanningClient(): MealPlanningServiceClient {
    if (!mealplanningClient) mealplanningClient = new MealPlanningServiceClient(serverUrl, credentials);
    return mealplanningClient;
  }
  function getAnalyticsClient(): AnalyticsServiceClient {
    if (!analyticsClient) analyticsClient = new AnalyticsServiceClient(serverUrl, credentials);
    return analyticsClient;
  }

  function authMetadata(oauth2AccessToken: string): Metadata {
    const m = new grpc.Metadata();
    m.add('authorization', `Bearer ${oauth2AccessToken}`);
    return m;
  }
  const emptyMetadata = new grpc.Metadata();

  return {
    authMetadata,

    loginForToken: (request: LoginForTokenRequest): Promise<LoginForTokenResponse> =>
      promisifyUnary<LoginForTokenRequest, LoginForTokenResponse>(getAuthClient().loginForToken.bind(getAuthClient()))(
        request,
        emptyMetadata,
      ),

    requestPasswordResetToken: (
      request: RequestPasswordResetTokenRequest,
    ): Promise<RequestPasswordResetTokenResponse> =>
      promisifyUnary<RequestPasswordResetTokenRequest, RequestPasswordResetTokenResponse>(
        getAuthClient().requestPasswordResetToken.bind(getAuthClient()),
      )(request, emptyMetadata),

    redeemPasswordResetToken: (request: RedeemPasswordResetTokenRequest): Promise<RedeemPasswordResetTokenResponse> =>
      promisifyUnary<RedeemPasswordResetTokenRequest, RedeemPasswordResetTokenResponse>(
        getAuthClient().redeemPasswordResetToken.bind(getAuthClient()),
      )(request, emptyMetadata),

    verifyEmailAddress: (request: VerifyEmailAddressRequest): Promise<VerifyEmailAddressResponse> =>
      promisifyUnary<VerifyEmailAddressRequest, VerifyEmailAddressResponse>(
        getAuthClient().verifyEmailAddress.bind(getAuthClient()),
      )(request, emptyMetadata),

    beginPasskeyRegistration: (oauth2Token: string): Promise<BeginPasskeyRegistrationResponse> =>
      promisifyUnary<BeginPasskeyRegistrationRequest, BeginPasskeyRegistrationResponse>(
        getAuthClient().beginPasskeyRegistration.bind(getAuthClient()),
      )({}, authMetadata(oauth2Token)),

    finishPasskeyRegistration: (
      oauth2Token: string,
      request: FinishPasskeyRegistrationRequest,
    ): Promise<FinishPasskeyRegistrationResponse> =>
      promisifyUnary<FinishPasskeyRegistrationRequest, FinishPasskeyRegistrationResponse>(
        getAuthClient().finishPasskeyRegistration.bind(getAuthClient()),
      )(request, authMetadata(oauth2Token)),

    beginPasskeyAuthentication: (
      request: BeginPasskeyAuthenticationRequest,
    ): Promise<BeginPasskeyAuthenticationResponse> =>
      promisifyUnary<BeginPasskeyAuthenticationRequest, BeginPasskeyAuthenticationResponse>(
        getAuthClient().beginPasskeyAuthentication.bind(getAuthClient()),
      )(request, emptyMetadata),

    finishPasskeyAuthentication: (request: FinishPasskeyAuthenticationRequest): Promise<LoginForTokenResponse> =>
      promisifyUnary<FinishPasskeyAuthenticationRequest, LoginForTokenResponse>(
        getAuthClient().finishPasskeyAuthentication.bind(getAuthClient()),
      )(request, emptyMetadata),

    getSelf: (oauth2Token: string): Promise<GetSelfResponse> =>
      promisifyUnary<GetSelfRequest, GetSelfResponse>(getAuthClient().getSelf.bind(getAuthClient()))(
        {},
        authMetadata(oauth2Token),
      ),

    getActiveAccount: (oauth2Token: string): Promise<GetActiveAccountResponse> =>
      promisifyUnary<GetActiveAccountRequest, GetActiveAccountResponse>(
        getAuthClient().getActiveAccount.bind(getAuthClient()),
      )({}, authMetadata(oauth2Token)),

    listPasskeys: (oauth2Token: string): Promise<ListPasskeysResponse> =>
      promisifyUnary<ListPasskeysRequest, ListPasskeysResponse>(getAuthClient().listPasskeys.bind(getAuthClient()))(
        {},
        authMetadata(oauth2Token),
      ),

    archivePasskey: (oauth2Token: string, request: ArchivePasskeyRequest): Promise<ArchivePasskeyResponse> =>
      promisifyUnary<ArchivePasskeyRequest, ArchivePasskeyResponse>(
        getAuthClient().archivePasskey.bind(getAuthClient()),
      )(request, authMetadata(oauth2Token)),

    listActiveSessions: (oauth2Token: string): Promise<ListActiveSessionsResponse> =>
      promisifyUnary<ListActiveSessionsRequest, ListActiveSessionsResponse>(
        getAuthClient().listActiveSessions.bind(getAuthClient()),
      )({ filter: undefined }, authMetadata(oauth2Token)),

    revokeSession: (oauth2Token: string, request: RevokeSessionRequest): Promise<RevokeSessionResponse> =>
      promisifyUnary<RevokeSessionRequest, RevokeSessionResponse>(getAuthClient().revokeSession.bind(getAuthClient()))(
        request,
        authMetadata(oauth2Token),
      ),

    revokeAllOtherSessions: (oauth2Token: string): Promise<RevokeAllOtherSessionsResponse> =>
      promisifyUnary<RevokeAllOtherSessionsRequest, RevokeAllOtherSessionsResponse>(
        getAuthClient().revokeAllOtherSessions.bind(getAuthClient()),
      )({}, authMetadata(oauth2Token)),

    revokeCurrentSession: (token: string): Promise<RevokeCurrentSessionResponse> =>
      promisifyUnary<RevokeCurrentSessionRequest, RevokeCurrentSessionResponse>(
        getAuthClient().revokeCurrentSession.bind(getAuthClient()),
      )({}, authMetadata(token)),

    // The handle and the address are on the auth surface rather than in the directory's
    // profile update, and they ask for the password and a second factor: whoever holds the
    // address can take the account through a password reset, so changing it is a
    // credential change.
    updateUserEmailAddress: (
      oauth2Token: string,
      request: UpdateUserEmailAddressRequest,
    ): Promise<UpdateUserEmailAddressResponse> =>
      promisifyUnary<UpdateUserEmailAddressRequest, UpdateUserEmailAddressResponse>(
        getAuthClient().updateUserEmailAddress.bind(getAuthClient()),
      )(request, authMetadata(oauth2Token)),

    updateUserUsername: (
      oauth2Token: string,
      request: UpdateUserUsernameRequest,
    ): Promise<UpdateUserUsernameResponse> =>
      promisifyUnary<UpdateUserUsernameRequest, UpdateUserUsernameResponse>(
        getAuthClient().updateUserUsername.bind(getAuthClient()),
      )(request, authMetadata(oauth2Token)),

    exchangeToken: (refreshToken: string, desiredAccountId?: string): Promise<ExchangeTokenResponse> =>
      promisifyUnary<ExchangeTokenRequest, ExchangeTokenResponse>(getAuthClient().exchangeToken.bind(getAuthClient()))(
        { refreshToken, desiredAccountId: desiredAccountId ?? '' },
        emptyMetadata,
      ),

    listAccountsForUser: (
      oauth2Token: string,
      request: ListAccountsForUserRequest,
    ): Promise<ListAccountsForUserResponse> =>
      promisifyUnary<ListAccountsForUserRequest, ListAccountsForUserResponse>(
        getIdentityClient().listAccountsForUser.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    listAccountMembers: (
      oauth2Token: string,
      request: ListAccountMembersRequest,
    ): Promise<ListAccountMembersResponse> =>
      promisifyUnary<ListAccountMembersRequest, ListAccountMembersResponse>(
        getIdentityClient().listAccountMembers.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    listInvitationsFromUser: (
      oauth2Token: string,
      request: ListInvitationsFromUserRequest,
    ): Promise<ListInvitationsFromUserResponse> =>
      promisifyUnary<ListInvitationsFromUserRequest, ListInvitationsFromUserResponse>(
        getIdentityClient().listInvitationsFromUser.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    invite: (oauth2Token: string, request: InviteRequest): Promise<InviteResponse> =>
      promisifyUnary<InviteRequest, InviteResponse>(getIdentityClient().invite.bind(getIdentityClient()))(
        request,
        authMetadata(oauth2Token),
      ),

    cancelInvitation: (oauth2Token: string, request: CancelInvitationRequest): Promise<CancelInvitationResponse> =>
      promisifyUnary<CancelInvitationRequest, CancelInvitationResponse>(
        getIdentityClient().cancelInvitation.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    setMembershipRoles: (
      oauth2Token: string,
      request: SetMembershipRolesRequest,
    ): Promise<SetMembershipRolesResponse> =>
      promisifyUnary<SetMembershipRolesRequest, SetMembershipRolesResponse>(
        getIdentityClient().setMembershipRoles.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    updateAccount: (oauth2Token: string, request: UpdateAccountRequest): Promise<UpdateAccountResponse> =>
      promisifyUnary<UpdateAccountRequest, UpdateAccountResponse>(
        getIdentityClient().updateAccount.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    // One call where there were three. A handle, a display name and the two parts of a
    // name are all fields of the same update, and each is absent unless it is being set —
    // so sending one no longer blanks the others.
    updateProfile: (oauth2Token: string, request: UpdateProfileRequest): Promise<UpdateProfileResponse> =>
      promisifyUnary<UpdateProfileRequest, UpdateProfileResponse>(
        getIdentityClient().updateProfile.bind(getIdentityClient()),
      )(request, authMetadata(oauth2Token)),

    // uploadUserAvatar is gone with the RPC it called.
    //
    // An avatar was never in the directory's User — it is a row in this application's
    // upload registry and a reference to it — so platform's identity surface has no
    // method for one, correctly. Restoring it means an RPC on the media surface and a
    // read that hydrates the reference, which is its own piece of work rather than part
    // of this cutover.

    getSettingDefinitions: (
      oauth2Token: string,
      request: GetSettingDefinitionsRequest,
    ): Promise<GetSettingDefinitionsResponse> =>
      promisifyUnary<GetSettingDefinitionsRequest, GetSettingDefinitionsResponse>(
        getSettingsClient().getSettingDefinitions.bind(getSettingsClient()),
      )(request, authMetadata(oauth2Token)),

    getSettingValues: (oauth2Token: string, request: GetSettingValuesRequest): Promise<GetSettingValuesResponse> =>
      promisifyUnary<GetSettingValuesRequest, GetSettingValuesResponse>(
        getSettingsClient().getSettingValues.bind(getSettingsClient()),
      )(request, authMetadata(oauth2Token)),

    // resolveSettings is the read a preferences page wants: every setting the
    // signed-in user may see, paired with the value that applies to them — their
    // own answer, or the setting's default where they have not answered.
    resolveSettings: (oauth2Token: string, request: ResolveSettingsRequest): Promise<ResolveSettingsResponse> =>
      promisifyUnary<ResolveSettingsRequest, ResolveSettingsResponse>(
        getSettingsClient().resolveSettings.bind(getSettingsClient()),
      )(request, authMetadata(oauth2Token)),

    // One call whether or not the user had answered before: the server converges
    // on the row, so a first answer and a changed one are the same write.
    setSettingValue: (oauth2Token: string, request: SetSettingValueRequest): Promise<SetSettingValueResponse> =>
      promisifyUnary<SetSettingValueRequest, SetSettingValueResponse>(
        getSettingsClient().setSettingValue.bind(getSettingsClient()),
      )(request, authMetadata(oauth2Token)),

    clearSettingValue: (oauth2Token: string, request: ClearSettingValueRequest): Promise<ClearSettingValueResponse> =>
      promisifyUnary<ClearSettingValueRequest, ClearSettingValueResponse>(
        getSettingsClient().clearSettingValue.bind(getSettingsClient()),
      )(request, authMetadata(oauth2Token)),

    getValidPreparations: (
      oauth2Token: string,
      request: GetValidPreparationsRequest,
    ): Promise<GetValidPreparationsResponse> =>
      promisifyUnary<GetValidPreparationsRequest, GetValidPreparationsResponse>(
        getMealplanningClient().getValidPreparations.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    searchForValidPreparations: (
      oauth2Token: string,
      request: Omit<SearchForValidPreparationsRequest, 'filter'> & {
        filter?: SearchForValidPreparationsRequest['filter'];
      },
    ): Promise<SearchForValidPreparationsResponse> =>
      promisifyUnary<SearchForValidPreparationsRequest, SearchForValidPreparationsResponse>(
        getMealplanningClient().searchForValidPreparations.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    getValidPreparationInstrumentsByPreparation: (
      oauth2Token: string,
      request: GetValidPreparationInstrumentsByPreparationRequest,
    ): Promise<GetValidPreparationInstrumentsByPreparationResponse> =>
      promisifyUnary<
        GetValidPreparationInstrumentsByPreparationRequest,
        GetValidPreparationInstrumentsByPreparationResponse
      >(getMealplanningClient().getValidPreparationInstrumentsByPreparation.bind(getMealplanningClient()))(
        { ...request, filter: request.filter ?? defaultSearchFilter },
        authMetadata(oauth2Token),
      ),

    getValidPreparationVesselsByPreparation: (
      oauth2Token: string,
      request: GetValidPreparationVesselsByPreparationRequest,
    ): Promise<GetValidPreparationVesselsByPreparationResponse> =>
      promisifyUnary<GetValidPreparationVesselsByPreparationRequest, GetValidPreparationVesselsByPreparationResponse>(
        getMealplanningClient().getValidPreparationVesselsByPreparation.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    searchValidIngredientsByPreparation: (
      oauth2Token: string,
      request: SearchValidIngredientsByPreparationRequest,
    ): Promise<SearchValidIngredientsByPreparationResponse> =>
      promisifyUnary<SearchValidIngredientsByPreparationRequest, SearchValidIngredientsByPreparationResponse>(
        getMealplanningClient().searchValidIngredientsByPreparation.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    searchValidMeasurementUnitsByIngredient: (
      oauth2Token: string,
      request: SearchValidMeasurementUnitsByIngredientRequest,
    ): Promise<SearchValidMeasurementUnitsByIngredientResponse> =>
      promisifyUnary<SearchValidMeasurementUnitsByIngredientRequest, SearchValidMeasurementUnitsByIngredientResponse>(
        getMealplanningClient().searchValidMeasurementUnitsByIngredient.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    getValidIngredientMeasurementUnitsByIngredient: (
      oauth2Token: string,
      request: GetValidIngredientMeasurementUnitsByIngredientRequest,
    ): Promise<GetValidIngredientMeasurementUnitsByIngredientResponse> =>
      promisifyUnary<
        GetValidIngredientMeasurementUnitsByIngredientRequest,
        GetValidIngredientMeasurementUnitsByIngredientResponse
      >(getMealplanningClient().getValidIngredientMeasurementUnitsByIngredient.bind(getMealplanningClient()))(
        { ...request, filter: request.filter ?? defaultSearchFilter },
        authMetadata(oauth2Token),
      ),

    getValidIngredientPreparationsByPreparation: (
      oauth2Token: string,
      request: GetValidIngredientPreparationsByPreparationRequest,
    ): Promise<GetValidIngredientPreparationsByPreparationResponse> =>
      promisifyUnary<
        GetValidIngredientPreparationsByPreparationRequest,
        GetValidIngredientPreparationsByPreparationResponse
      >(getMealplanningClient().getValidIngredientPreparationsByPreparation.bind(getMealplanningClient()))(
        { ...request, filter: request.filter ?? defaultSearchFilter },
        authMetadata(oauth2Token),
      ),

    getValidMeasurementUnits: (
      oauth2Token: string,
      request: GetValidMeasurementUnitsRequest,
    ): Promise<GetValidMeasurementUnitsResponse> =>
      promisifyUnary<GetValidMeasurementUnitsRequest, GetValidMeasurementUnitsResponse>(
        getMealplanningClient().getValidMeasurementUnits.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    searchForValidMeasurementUnits: (
      oauth2Token: string,
      request: Omit<SearchForValidMeasurementUnitsRequest, 'filter'> & {
        filter?: SearchForValidMeasurementUnitsRequest['filter'];
      },
    ): Promise<SearchForValidMeasurementUnitsResponse> =>
      promisifyUnary<SearchForValidMeasurementUnitsRequest, SearchForValidMeasurementUnitsResponse>(
        getMealplanningClient().searchForValidMeasurementUnits.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    getValidVessels: (oauth2Token: string, request: GetValidVesselsRequest): Promise<GetValidVesselsResponse> =>
      promisifyUnary<GetValidVesselsRequest, GetValidVesselsResponse>(
        getMealplanningClient().getValidVessels.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    searchForValidVessels: (
      oauth2Token: string,
      request: Omit<SearchForValidVesselsRequest, 'filter'> & { filter?: SearchForValidVesselsRequest['filter'] },
    ): Promise<SearchForValidVesselsResponse> =>
      promisifyUnary<SearchForValidVesselsRequest, SearchForValidVesselsResponse>(
        getMealplanningClient().searchForValidVessels.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    getValidIngredientStates: (
      oauth2Token: string,
      request: GetValidIngredientStatesRequest,
    ): Promise<GetValidIngredientStatesResponse> =>
      promisifyUnary<GetValidIngredientStatesRequest, GetValidIngredientStatesResponse>(
        getMealplanningClient().getValidIngredientStates.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    searchForValidIngredientStates: (
      oauth2Token: string,
      request: Omit<SearchForValidIngredientStatesRequest, 'filter'> & {
        filter?: SearchForValidIngredientStatesRequest['filter'];
      },
    ): Promise<SearchForValidIngredientStatesResponse> =>
      promisifyUnary<SearchForValidIngredientStatesRequest, SearchForValidIngredientStatesResponse>(
        getMealplanningClient().searchForValidIngredientStates.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    createRecipe: (oauth2Token: string, request: CreateRecipeRequest): Promise<CreateRecipeResponse> =>
      promisifyUnary<CreateRecipeRequest, CreateRecipeResponse>(
        getMealplanningClient().createRecipe.bind(getMealplanningClient()),
      )(request, authMetadata(oauth2Token)),

    searchForRecipes: (
      oauth2Token: string,
      request: Omit<SearchForRecipesRequest, 'filter'> & { filter?: SearchForRecipesRequest['filter'] },
    ): Promise<SearchForRecipesResponse> =>
      promisifyUnary<SearchForRecipesRequest, SearchForRecipesResponse>(
        getMealplanningClient().searchForRecipes.bind(getMealplanningClient()),
      )({ ...request, filter: request.filter ?? defaultSearchFilter }, authMetadata(oauth2Token)),

    trackEvent: async (event: string, properties: Record<string, string> = {}): Promise<void> => {
      await promisifyUnary(getAnalyticsClient().trackEvent.bind(getAnalyticsClient()))(
        { source: 'web', event, properties },
        emptyMetadata,
      );
    },

    trackAnonymousEvent: async (
      event: string,
      anonymousId: string,
      properties: Record<string, string> = {},
    ): Promise<void> => {
      await promisifyUnary(getAnalyticsClient().trackAnonymousEvent.bind(getAnalyticsClient()))(
        { source: 'web', event, anonymousId, properties },
        emptyMetadata,
      );
    },
  };
}
