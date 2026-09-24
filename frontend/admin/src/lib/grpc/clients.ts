/**
 * Admin gRPC clients: reads env, creates clients from @dinnerdonebetter/api-client createAdminGrpcClients.
 */

import { env } from '$env/dynamic/private';
import { createAdminGrpcClients, createPlatformClient } from '@dinnerdonebetter/api-client';
import { AuditServiceService, type ListEntriesRequest } from '@primandproper/platform-client/audit/v1';
import {
  BillingServiceService,
  type ListProductsRequest,
  type ListSubscriptionsForAccountRequest,
} from '@primandproper/platform-client/billing/v1';
import {
  IdentityServiceService,
  type GetAccountRequest,
  type GetUserRequest,
  type ListAccountMembersRequest,
  type ListAccountsForUserRequest,
  type ListAccountsRequest,
  type ListUsersRequest,
} from '@primandproper/platform-client/identity/v1';
import {
  IssueReportsServiceService,
  type GetReportRequest,
  type ListReportsRequest,
} from '@primandproper/platform-client/issuereports/v1';
import {
  OAuth2ClientsServiceService,
  type GetOAuth2ClientRequest,
  type ListOAuth2ClientsRequest,
} from '@primandproper/platform-client/oauth2clients/v1';
import { SettingsServiceService, type ListDefinitionsRequest } from '@primandproper/platform-client/settings/v1';
import {
  WaitlistsServiceService,
  type GetListRequest,
  type ListListsRequest,
} from '@primandproper/platform-client/waitlists/v1';

const config = {
  serverUrl: env.GRPC_API_SERVER_URL ?? 'localhost:50051',
  insecure: env.DEVELOPING_LOCALLY === 'true',
};

const clients = createAdminGrpcClients(config);
const platform = createPlatformClient(config);

export const authMetadata = clients.authMetadata;
export const adminLoginForToken = clients.adminLoginForToken;
export const beginPasskeyAuthentication = clients.beginPasskeyAuthentication;
export const finishPasskeyAuthentication = clients.finishPasskeyAuthentication;
export const testQueueMessage = clients.testQueueMessage;
export const trackEvent = clients.trackEvent;
export const trackAnonymousEvent = clients.trackAnonymousEvent;
export const getRecipes = clients.getRecipes;
export const getRecipe = clients.getRecipe;
export const searchForRecipes = clients.searchForRecipes;
export const createRecipe = clients.createRecipe;
export const getValidIngredients = clients.getValidIngredients;
export const searchForValidIngredients = clients.searchForValidIngredients;
export const searchValidIngredientsByPreparation = clients.searchValidIngredientsByPreparation;
export const getValidIngredient = clients.getValidIngredient;
export const createValidIngredient = clients.createValidIngredient;
export const updateValidIngredient = clients.updateValidIngredient;
export const getValidIngredientMeasurementUnitsByIngredient = clients.getValidIngredientMeasurementUnitsByIngredient;
export const createValidIngredientMeasurementUnit = clients.createValidIngredientMeasurementUnit;
export const archiveValidIngredientMeasurementUnit = clients.archiveValidIngredientMeasurementUnit;
export const getValidIngredientPreparationsByIngredient = clients.getValidIngredientPreparationsByIngredient;
export const createValidIngredientPreparation = clients.createValidIngredientPreparation;
export const archiveValidIngredientPreparation = clients.archiveValidIngredientPreparation;
export const getValidInstruments = clients.getValidInstruments;
export const searchForValidInstruments = clients.searchForValidInstruments;
export const getValidInstrument = clients.getValidInstrument;
export const createValidInstrument = clients.createValidInstrument;
export const updateValidInstrument = clients.updateValidInstrument;
export const getValidPreparationInstrumentsByInstrument = clients.getValidPreparationInstrumentsByInstrument;
export const createValidPreparationInstrument = clients.createValidPreparationInstrument;
export const archiveValidPreparationInstrument = clients.archiveValidPreparationInstrument;
export const getValidVessels = clients.getValidVessels;
export const searchForValidVessels = clients.searchForValidVessels;
export const getValidVessel = clients.getValidVessel;
export const createValidVessel = clients.createValidVessel;
export const updateValidVessel = clients.updateValidVessel;
export const getValidPreparationVesselsByVessel = clients.getValidPreparationVesselsByVessel;
export const createValidPreparationVessel = clients.createValidPreparationVessel;
export const archiveValidPreparationVessel = clients.archiveValidPreparationVessel;
export const getValidMeasurementUnits = clients.getValidMeasurementUnits;
export const searchForValidMeasurementUnits = clients.searchForValidMeasurementUnits;
export const getValidMeasurementUnit = clients.getValidMeasurementUnit;
export const createValidMeasurementUnit = clients.createValidMeasurementUnit;
export const updateValidMeasurementUnit = clients.updateValidMeasurementUnit;
export const getValidMeasurementUnitConversionsForUnit = clients.getValidMeasurementUnitConversionsForUnit;
export const createValidMeasurementUnitConversion = clients.createValidMeasurementUnitConversion;
export const archiveValidMeasurementUnitConversion = clients.archiveValidMeasurementUnitConversion;
export const getValidMeasurementUnitConversionsForIngredients =
  clients.getValidMeasurementUnitConversionsForIngredients;
export const getValidIngredientStates = clients.getValidIngredientStates;
export const searchForValidIngredientStates = clients.searchForValidIngredientStates;
export const getValidIngredientState = clients.getValidIngredientState;
export const createValidIngredientState = clients.createValidIngredientState;
export const updateValidIngredientState = clients.updateValidIngredientState;
export const getValidPreparations = clients.getValidPreparations;
export const searchForValidPreparations = clients.searchForValidPreparations;
export const getValidPreparation = clients.getValidPreparation;
export const createValidPreparation = clients.createValidPreparation;
export const updateValidPreparation = clients.updateValidPreparation;
export const getValidPreparationInstrumentsByPreparation = clients.getValidPreparationInstrumentsByPreparation;
export const getValidPreparationVesselsByPreparation = clients.getValidPreparationVesselsByPreparation;
export const getValidIngredientPreparationsByPreparation = clients.getValidIngredientPreparationsByPreparation;
export const getValidPrepTaskConfig = clients.getValidPrepTaskConfig;
export const getValidPrepTaskConfigs = clients.getValidPrepTaskConfigs;
export const getMeasurementUnitConversionMismatches = clients.getMeasurementUnitConversionMismatches;
export const adminListSessionsForUser = clients.adminListSessionsForUser;
export const adminRevokeUserSession = clients.adminRevokeUserSession;
export const adminRevokeAllUserSessions = clients.adminRevokeAllUserSessions;
export const revokeCurrentSession = clients.revokeCurrentSession;

// Platform's services, called through @primandproper/platform-client's stubs rather
// than any generated here. Only the reads a page calls are here: the writes these
// pages once had called services the server no longer runs, and nothing reached them.
export const getUser = (token: string, request: GetUserRequest) =>
  platform.call(IdentityServiceService.getUser, token, request);
export const listUsers = (token: string, request: ListUsersRequest) =>
  platform.call(IdentityServiceService.listUsers, token, request);
export const getAccount = (token: string, request: GetAccountRequest) =>
  platform.call(IdentityServiceService.getAccount, token, request);
export const listAccounts = (token: string, request: ListAccountsRequest) =>
  platform.call(IdentityServiceService.listAccounts, token, request);
export const listAccountMembers = (token: string, request: ListAccountMembersRequest) =>
  platform.call(IdentityServiceService.listAccountMembers, token, request);
export const listAccountsForUser = (token: string, request: ListAccountsForUserRequest) =>
  platform.call(IdentityServiceService.listAccountsForUser, token, request);
export const listOAuth2Clients = (token: string, request: ListOAuth2ClientsRequest) =>
  platform.call(OAuth2ClientsServiceService.listOAuth2Clients, token, request);
export const getOAuth2Client = (token: string, request: GetOAuth2ClientRequest) =>
  platform.call(OAuth2ClientsServiceService.getOAuth2Client, token, request);
export const listProducts = (token: string, request: ListProductsRequest) =>
  platform.call(BillingServiceService.listProducts, token, request);
export const listSubscriptionsForAccount = (token: string, request: ListSubscriptionsForAccountRequest) =>
  platform.call(BillingServiceService.listSubscriptionsForAccount, token, request);
export const listSettingDefinitions = (token: string, request: ListDefinitionsRequest) =>
  platform.call(SettingsServiceService.listDefinitions, token, request);
export const listWaitlists = (token: string, request: ListListsRequest) =>
  platform.call(WaitlistsServiceService.listLists, token, request);
export const getWaitlist = (token: string, request: GetListRequest) =>
  platform.call(WaitlistsServiceService.getList, token, request);
export const listIssueReports = (token: string, request: ListReportsRequest) =>
  platform.call(IssueReportsServiceService.listReports, token, request);
export const getIssueReport = (token: string, request: GetReportRequest) =>
  platform.call(IssueReportsServiceService.getReport, token, request);
// An operator's read is unscoped on the server, so a query by resource reaches every
// chain an entry about that user or account could be on.
export const listAuditEntries = (token: string, request: ListEntriesRequest) =>
  platform.call(AuditServiceService.listEntries, token, request);
