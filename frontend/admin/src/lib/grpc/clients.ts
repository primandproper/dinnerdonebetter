/**
 * Every call the admin app makes, platform's services and this repository's alike, goes over
 * one transport. An authenticated call goes through the request's Session (see
 * $lib/auth/session), which holds the operator's login, refreshes it through platform's
 * SignInService, and retries a call the server refused once with the successor.
 */

import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';
import { CHANGE_PASSWORD_PATH, mustChangePassword } from '$lib/auth/required-actions';
import {
  InternalOperationsService,
  MealPlanningServiceService,
  createPlatformTransport,
} from '@dinnerdonebetter/api-client';
import {
  type CredentialStore,
  type Metadata,
  redirectOnNotSignedIn,
  Session,
  type UnaryMethod,
} from '@primandproper/platform-client';
import { AuditServiceService } from '@primandproper/platform-client/audit/v1';
import { BillingServiceService } from '@primandproper/platform-client/billing/v1';
import { IdentityServiceService } from '@primandproper/platform-client/identity/v1';
import { IssueReportsServiceService } from '@primandproper/platform-client/issuereports/v1';
import { OAuth2ClientsServiceService } from '@primandproper/platform-client/oauth2clients/v1';
import { SettingsServiceService } from '@primandproper/platform-client/settings/v1';
import { SignInAdministrationServiceService, SignInServiceService } from '@primandproper/platform-client/signin/v1';
import { WaitlistsServiceService } from '@primandproper/platform-client/waitlists/v1';

const transport = createPlatformTransport({
  serverUrl: env.GRPC_API_SERVER_URL ?? 'localhost:50051',
  insecure: env.DEVELOPING_LOCALLY === 'true',
});

/**
 * newSession is a Session over `store`. Sessions built per request share the process's
 * exchange coordinator, so concurrent requests carrying the same cookie present its refresh
 * token once between them. That holds within one process only: more than one replica needs a
 * SharedExchangeCoordinator.
 *
 * `metadata` rides on every call the Session makes; see clientMetadata in $lib/auth/session.
 */
export function newSession(store: CredentialStore, metadata: Metadata = {}): Session {
  return new Session({ transport, store, metadata });
}

/**
 * call makes an authenticated call. A login that is over by the time it is made sends the
 * operator to sign in again; the Session has already cleared the cookie by then. One refused
 * because they owe a password change sends them to the form for it.
 */
async function call<Req, Res>(session: Session, method: UnaryMethod<Req, Res>, request: Req): Promise<Res> {
  try {
    return await redirectOnNotSignedIn(
      session,
      () => session.call(method, request),
      () => redirect(302, '/login'),
    );
  } catch (err) {
    if (mustChangePassword(err)) {
      redirect(302, CHANGE_PASSWORD_PATH);
    }
    throw err;
  }
}

function authed<Req, Res>(method: UnaryMethod<Req, Res>) {
  return (session: Session, request: Req) => call(session, method, request);
}

/**
 * loose is `authed` for the pages that build their requests from form fields as plain
 * objects, which is how every one of these was called before; their requests are not yet
 * typed.
 */
function loose<Req, Res>(method: UnaryMethod<Req, Res>) {
  return (session: Session, request: Record<string, unknown>) => call(session, method, request as Req);
}

// An operator's view of somebody else's logins is SignInAdministrationService: the
// caller's own RPCs name nobody, so there is no field an administrator could use.
export const listSignInsForUser = authed(SignInAdministrationServiceService.listSignInsForUser);
export const endSignInForUser = authed(SignInAdministrationServiceService.endSignInForUser);
export const endAllSignInsForUser = authed(SignInAdministrationServiceService.endAllSignInsForUser);

export const testQueueMessage = loose(InternalOperationsService.testQueueMessage);
export const createRecipe = authed(MealPlanningServiceService.createRecipe);
export const getRecipes = loose(MealPlanningServiceService.getRecipes);
export const getRecipe = loose(MealPlanningServiceService.getRecipe);
export const searchForRecipes = loose(MealPlanningServiceService.searchForRecipes);
export const getValidIngredients = loose(MealPlanningServiceService.getValidIngredients);
export const searchForValidIngredients = loose(MealPlanningServiceService.searchForValidIngredients);
export const searchValidIngredientsByPreparation = loose(
  MealPlanningServiceService.searchValidIngredientsByPreparation,
);
export const getValidIngredient = loose(MealPlanningServiceService.getValidIngredient);
export const createValidIngredient = loose(MealPlanningServiceService.createValidIngredient);
export const updateValidIngredient = loose(MealPlanningServiceService.updateValidIngredient);
export const getValidIngredientMeasurementUnitsByIngredient = loose(
  MealPlanningServiceService.getValidIngredientMeasurementUnitsByIngredient,
);
export const createValidIngredientMeasurementUnit = loose(
  MealPlanningServiceService.createValidIngredientMeasurementUnit,
);
export const archiveValidIngredientMeasurementUnit = loose(
  MealPlanningServiceService.archiveValidIngredientMeasurementUnit,
);
export const getValidIngredientPreparationsByIngredient = loose(
  MealPlanningServiceService.getValidIngredientPreparationsByIngredient,
);
export const createValidIngredientPreparation = loose(MealPlanningServiceService.createValidIngredientPreparation);
export const archiveValidIngredientPreparation = loose(MealPlanningServiceService.archiveValidIngredientPreparation);
export const getValidInstruments = loose(MealPlanningServiceService.getValidInstruments);
export const searchForValidInstruments = loose(MealPlanningServiceService.searchForValidInstruments);
export const getValidInstrument = loose(MealPlanningServiceService.getValidInstrument);
export const createValidInstrument = loose(MealPlanningServiceService.createValidInstrument);
export const updateValidInstrument = loose(MealPlanningServiceService.updateValidInstrument);
export const getValidPreparationInstrumentsByInstrument = loose(
  MealPlanningServiceService.getValidPreparationInstrumentsByInstrument,
);
export const createValidPreparationInstrument = loose(MealPlanningServiceService.createValidPreparationInstrument);
export const archiveValidPreparationInstrument = loose(MealPlanningServiceService.archiveValidPreparationInstrument);
export const getValidVessels = loose(MealPlanningServiceService.getValidVessels);
export const searchForValidVessels = loose(MealPlanningServiceService.searchForValidVessels);
export const getValidVessel = loose(MealPlanningServiceService.getValidVessel);
export const createValidVessel = loose(MealPlanningServiceService.createValidVessel);
export const updateValidVessel = loose(MealPlanningServiceService.updateValidVessel);
export const getValidPreparationVesselsByVessel = loose(MealPlanningServiceService.getValidPreparationVesselsByVessel);
export const createValidPreparationVessel = loose(MealPlanningServiceService.createValidPreparationVessel);
export const archiveValidPreparationVessel = loose(MealPlanningServiceService.archiveValidPreparationVessel);
export const getValidMeasurementUnits = loose(MealPlanningServiceService.getValidMeasurementUnits);
export const searchForValidMeasurementUnits = loose(MealPlanningServiceService.searchForValidMeasurementUnits);
export const getValidMeasurementUnit = loose(MealPlanningServiceService.getValidMeasurementUnit);
export const createValidMeasurementUnit = loose(MealPlanningServiceService.createValidMeasurementUnit);
export const updateValidMeasurementUnit = loose(MealPlanningServiceService.updateValidMeasurementUnit);
export const getValidMeasurementUnitConversionsForUnit = loose(
  MealPlanningServiceService.getValidMeasurementUnitConversionsForUnit,
);
export const createValidMeasurementUnitConversion = loose(
  MealPlanningServiceService.createValidMeasurementUnitConversion,
);
export const archiveValidMeasurementUnitConversion = loose(
  MealPlanningServiceService.archiveValidMeasurementUnitConversion,
);
export const getValidIngredientStates = loose(MealPlanningServiceService.getValidIngredientStates);
export const searchForValidIngredientStates = loose(MealPlanningServiceService.searchForValidIngredientStates);
export const getValidIngredientState = loose(MealPlanningServiceService.getValidIngredientState);
export const createValidIngredientState = loose(MealPlanningServiceService.createValidIngredientState);
export const updateValidIngredientState = loose(MealPlanningServiceService.updateValidIngredientState);
export const getValidPreparations = loose(MealPlanningServiceService.getValidPreparations);
export const searchForValidPreparations = loose(MealPlanningServiceService.searchForValidPreparations);
export const getValidPreparation = loose(MealPlanningServiceService.getValidPreparation);
export const createValidPreparation = loose(MealPlanningServiceService.createValidPreparation);
export const updateValidPreparation = loose(MealPlanningServiceService.updateValidPreparation);
export const getValidPreparationInstrumentsByPreparation = loose(
  MealPlanningServiceService.getValidPreparationInstrumentsByPreparation,
);
export const getValidPreparationVesselsByPreparation = loose(
  MealPlanningServiceService.getValidPreparationVesselsByPreparation,
);
export const getValidIngredientPreparationsByPreparation = loose(
  MealPlanningServiceService.getValidIngredientPreparationsByPreparation,
);
export const getValidPrepTaskConfig = loose(MealPlanningServiceService.getValidPrepTaskConfig);
export const getValidPrepTaskConfigs = loose(MealPlanningServiceService.getValidPrepTaskConfigs);
export const getMeasurementUnitConversionMismatches = loose(
  MealPlanningServiceService.getMeasurementUnitConversionMismatches,
);
export const getValidMeasurementUnitConversionsForIngredients = loose(
  MealPlanningServiceService.getValidMeasurementUnitConversionsForIngredients,
);

// The operator's own password. It is the one call an operator who owes a password change
// may still make.
export const updatePassword = authed(SignInServiceService.updatePassword);

// Platform's services, called through @primandproper/platform-client's stubs rather
// than any generated here. Only the reads a page calls are here.
export const getUser = authed(IdentityServiceService.getUser);
export const listUsers = authed(IdentityServiceService.listUsers);
export const getAccount = authed(IdentityServiceService.getAccount);
export const listAccounts = authed(IdentityServiceService.listAccounts);
export const listAccountMembers = authed(IdentityServiceService.listAccountMembers);
export const listAccountsForUser = authed(IdentityServiceService.listAccountsForUser);
export const listOAuth2Clients = authed(OAuth2ClientsServiceService.listOAuth2Clients);
export const getOAuth2Client = authed(OAuth2ClientsServiceService.getOAuth2Client);
export const listProducts = authed(BillingServiceService.listProducts);
export const listSubscriptionsForAccount = authed(BillingServiceService.listSubscriptionsForAccount);
export const listSettingDefinitions = authed(SettingsServiceService.listDefinitions);
export const listWaitlists = authed(WaitlistsServiceService.listLists);
export const getWaitlist = authed(WaitlistsServiceService.getList);
// The service's queue: every report anybody filed, which are all in the one global scope and
// paged by a service admin's issues.reports.triage.
export const listIssueReports = authed(IssueReportsServiceService.listReports);
export const getIssueReport = authed(IssueReportsServiceService.getReport);
// An operator's read is unscoped on the server, so a query by resource reaches every
// chain an entry about that user or account could be on.
export const listAuditEntries = authed(AuditServiceService.listEntries);
