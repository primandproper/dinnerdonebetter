/**
 * Thin wrapper: reads env, creates gRPC clients from @dinnerdonebetter/api-client, re-exports API.
 */

import { env } from '$env/dynamic/private';
import {
  createGrpcClients,
  createPlatformClient,
  authMetadata as authMetadataFromPackage,
} from '@dinnerdonebetter/api-client';
import {
  IdentityServiceService,
  type CancelInvitationRequest,
  type InviteRequest,
  type ListAccountMembersRequest,
  type ListAccountsForUserRequest,
  type ListInvitationsFromUserRequest,
  type SetMembershipRolesRequest,
  type UpdateAccountRequest,
  type UpdateProfileRequest,
} from '@primandproper/platform-client/identity/v1';
import {
  SettingsServiceService,
  type ResolveAllRequest,
  type SetValueRequest,
} from '@primandproper/platform-client/settings/v1';

const config = {
  serverUrl: env.GRPC_API_SERVER_URL ?? 'localhost:50051',
  insecure: env.DEVELOPING_LOCALLY === 'true',
};

const clients = createGrpcClients(config);
const platform = createPlatformClient(config);

export const authMetadata = authMetadataFromPackage;

export const loginForToken = clients.loginForToken;
export const requestPasswordResetToken = clients.requestPasswordResetToken;
export const redeemPasswordResetToken = clients.redeemPasswordResetToken;
export const verifyEmailAddress = clients.verifyEmailAddress;
export const beginPasskeyRegistration = clients.beginPasskeyRegistration;
export const finishPasskeyRegistration = clients.finishPasskeyRegistration;
export const beginPasskeyAuthentication = clients.beginPasskeyAuthentication;
export const finishPasskeyAuthentication = clients.finishPasskeyAuthentication;
export const getSelf = clients.getSelf;
export const getActiveAccount = clients.getActiveAccount;
export const listPasskeys = clients.listPasskeys;
export const archivePasskey = clients.archivePasskey;
export const listActiveSessions = clients.listActiveSessions;
export const revokeSession = clients.revokeSession;
export const revokeAllOtherSessions = clients.revokeAllOtherSessions;
export const revokeCurrentSession = clients.revokeCurrentSession;
export const exchangeToken = clients.exchangeToken;
export const updateUserUsername = clients.updateUserUsername;
export const updateUserEmailAddress = clients.updateUserEmailAddress;

// The directory and settings are platform's services, called through
// @primandproper/platform-client's stubs rather than any generated here.
export const listAccountsForUser = (token: string, request: ListAccountsForUserRequest) =>
  platform.call(IdentityServiceService.listAccountsForUser, token, request);
export const listAccountMembers = (token: string, request: ListAccountMembersRequest) =>
  platform.call(IdentityServiceService.listAccountMembers, token, request);
export const listInvitationsFromUser = (token: string, request: ListInvitationsFromUserRequest) =>
  platform.call(IdentityServiceService.listInvitationsFromUser, token, request);
export const invite = (token: string, request: InviteRequest) =>
  platform.call(IdentityServiceService.invite, token, request);
export const cancelInvitation = (token: string, request: CancelInvitationRequest) =>
  platform.call(IdentityServiceService.cancelInvitation, token, request);
export const setMembershipRoles = (token: string, request: SetMembershipRolesRequest) =>
  platform.call(IdentityServiceService.setMembershipRoles, token, request);
export const updateAccount = (token: string, request: UpdateAccountRequest) =>
  platform.call(IdentityServiceService.updateAccount, token, request);
export const updateProfile = (token: string, request: UpdateProfileRequest) =>
  platform.call(IdentityServiceService.updateProfile, token, request);
// A person's own preferences. The server refuses a subject other than the caller
// themselves, so every call names them — see settingsSubject.
export const resolveSettings = (token: string, request: ResolveAllRequest) =>
  platform.call(SettingsServiceService.resolveAll, token, request);
export const setSettingValue = (token: string, request: SetValueRequest) =>
  platform.call(SettingsServiceService.setValue, token, request);
export const getValidPreparations = clients.getValidPreparations;
export const searchForValidPreparations = clients.searchForValidPreparations;
export const getValidPreparationInstrumentsByPreparation = clients.getValidPreparationInstrumentsByPreparation;
export const getValidPreparationVesselsByPreparation = clients.getValidPreparationVesselsByPreparation;
export const searchValidIngredientsByPreparation = clients.searchValidIngredientsByPreparation;
export const searchValidMeasurementUnitsByIngredient = clients.searchValidMeasurementUnitsByIngredient;
export const getValidIngredientMeasurementUnitsByIngredient = clients.getValidIngredientMeasurementUnitsByIngredient;
export const getValidIngredientPreparationsByPreparation = clients.getValidIngredientPreparationsByPreparation;
export const getValidMeasurementUnits = clients.getValidMeasurementUnits;
export const searchForValidMeasurementUnits = clients.searchForValidMeasurementUnits;
export const getValidVessels = clients.getValidVessels;
export const searchForValidVessels = clients.searchForValidVessels;
export const getValidIngredientStates = clients.getValidIngredientStates;
export const searchForValidIngredientStates = clients.searchForValidIngredientStates;
export const createRecipe = clients.createRecipe;
export const searchForRecipes = clients.searchForRecipes;
export const trackEvent = clients.trackEvent;
export const trackAnonymousEvent = clients.trackAnonymousEvent;
