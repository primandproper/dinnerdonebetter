/**
 * Thin wrapper: reads env, creates gRPC clients from @dinnerdonebetter/api-client, re-exports API.
 */

import { env } from '$env/dynamic/private';
import { createGrpcClients, authMetadata as authMetadataFromPackage } from '@dinnerdonebetter/api-client';

const clients = createGrpcClients({
  serverUrl: env.GRPC_API_SERVER_URL ?? 'localhost:50051',
  insecure: env.DEVELOPING_LOCALLY === 'true',
});

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
export const getAccountsForUser = clients.getAccountsForUser;
export const getSentAccountInvitations = clients.getSentAccountInvitations;
export const createAccountInvitation = clients.createAccountInvitation;
export const cancelAccountInvitation = clients.cancelAccountInvitation;
export const updateAccountMemberPermissions = clients.updateAccountMemberPermissions;
export const updateAccount = clients.updateAccount;
export const updateUserUsername = clients.updateUserUsername;
export const updateUserDetails = clients.updateUserDetails;
export const uploadUserAvatar = clients.uploadUserAvatar;
export const getSettingDefinitions = clients.getSettingDefinitions;
export const getSettingValues = clients.getSettingValues;
export const resolveSettings = clients.resolveSettings;
export const setSettingValue = clients.setSettingValue;
export const clearSettingValue = clients.clearSettingValue;
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
