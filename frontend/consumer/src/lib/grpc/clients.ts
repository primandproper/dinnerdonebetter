/**
 * Every call the consumer app makes, platform's services and this repository's alike, goes
 * over one transport. An authenticated call goes through the request's Session (see
 * $lib/auth/session), which holds the login, refreshes it through platform's SignInService,
 * and retries a call the server refused once with the successor.
 */

import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';
import {
  AnalyticsServiceService,
  AuthServiceService,
  MealPlanningServiceService,
  QueryFilter,
  createPlatformTransport,
} from '@dinnerdonebetter/api-client';
import {
  type CredentialStore,
  NotSignedInError,
  Session,
  TokenCaller,
  type UnaryMethod,
} from '@primandproper/platform-client';
import { IdentityServiceService } from '@primandproper/platform-client/identity/v1';
import { SettingsServiceService } from '@primandproper/platform-client/settings/v1';

const transport = createPlatformTransport({
  serverUrl: env.GRPC_API_SERVER_URL ?? 'localhost:50051',
  insecure: env.DEVELOPING_LOCALLY === 'true',
});

const anonymous = new TokenCaller({ transport });

/**
 * newSession is a Session over `store`. Sessions built per request share the process's
 * exchange coordinator, so concurrent requests carrying the same cookie present its refresh
 * token once between them. That holds within one process only: more than one replica needs a
 * SharedExchangeCoordinator.
 */
export function newSession(store: CredentialStore): Session {
  return new Session({ transport, store });
}

/**
 * call makes an authenticated call. A login that is over by the time it is made — never
 * held, lapsed, or refused a refresh — sends the person to sign in again; the Session has
 * already cleared the cookie by then.
 */
async function call<Req, Res>(session: Session, method: UnaryMethod<Req, Res>, request: Req): Promise<Res> {
  try {
    return await session.call(method, request);
  } catch (err) {
    if (err instanceof NotSignedInError || session.state === 'anonymous') {
      throw redirect(302, '/login');
    }
    throw err;
  }
}

function authed<Req, Res>(method: UnaryMethod<Req, Res>) {
  return (session: Session, request: Req) => call(session, method, request);
}

function unauthed<Req, Res>(method: UnaryMethod<Req, Res>) {
  return (request: Req) => anonymous.callAnonymous(method, request);
}

// QueryFilter is platform's schema, and its four timestamp windows are message fields
// rather than `optional` scalars, so a bare object literal is not one. The generated
// factory fills every field, which is what it is for.
const defaultSearchFilter = QueryFilter.create({ maxResponseSize: 20 });

/** filtered is `authed` for a list or search, filling in the default filter when none is given. */
function filtered<Req extends { filter: QueryFilter | undefined }, Res>(method: UnaryMethod<Req, Res>) {
  return (session: Session, request: Omit<Req, 'filter'> & { filter?: QueryFilter }) =>
    call(session, method, { ...request, filter: request.filter ?? defaultSearchFilter } as Req);
}

// Sign-in is platform's, through signIn and signOut on the Session; what AuthService still
// does is everything platform's SignInService doesn't expose yet.
export const requestPasswordResetToken = unauthed(AuthServiceService.requestPasswordResetToken);
export const redeemPasswordResetToken = unauthed(AuthServiceService.redeemPasswordResetToken);
export const verifyEmailAddress = unauthed(AuthServiceService.verifyEmailAddress);
export const beginPasskeyAuthentication = unauthed(AuthServiceService.beginPasskeyAuthentication);
export const finishPasskeyAuthentication = unauthed(AuthServiceService.finishPasskeyAuthentication);
export const beginPasskeyRegistration = authed(AuthServiceService.beginPasskeyRegistration);
export const finishPasskeyRegistration = authed(AuthServiceService.finishPasskeyRegistration);
export const getSelf = (session: Session) => call(session, AuthServiceService.getSelf, {});
export const getActiveAccount = (session: Session) => call(session, AuthServiceService.getActiveAccount, {});
export const listPasskeys = (session: Session) => call(session, AuthServiceService.listPasskeys, {});
export const archivePasskey = authed(AuthServiceService.archivePasskey);
export const listActiveSessions = (session: Session) =>
  call(session, AuthServiceService.listActiveSessions, { filter: undefined });
export const revokeSession = authed(AuthServiceService.revokeSession);
export const revokeAllOtherSessions = (session: Session) =>
  call(session, AuthServiceService.revokeAllOtherSessions, {});
export const revokeCurrentSession = (session: Session) => call(session, AuthServiceService.revokeCurrentSession, {});
// The handle and the address are on the auth surface rather than in the directory's profile
// update, and they ask for the password and a second factor: whoever holds the address can
// take the account through a password reset, so changing it is a credential change.
export const updateUserUsername = authed(AuthServiceService.updateUserUsername);
export const updateUserEmailAddress = authed(AuthServiceService.updateUserEmailAddress);

// The directory and settings are platform's services, called through
// @primandproper/platform-client's stubs rather than any generated here.
export const listAccountsForUser = authed(IdentityServiceService.listAccountsForUser);
export const listAccountMembers = authed(IdentityServiceService.listAccountMembers);
export const listInvitationsFromUser = authed(IdentityServiceService.listInvitationsFromUser);
export const invite = authed(IdentityServiceService.invite);
export const cancelInvitation = authed(IdentityServiceService.cancelInvitation);
export const setMembershipRoles = authed(IdentityServiceService.setMembershipRoles);
export const updateAccount = authed(IdentityServiceService.updateAccount);
export const updateProfile = authed(IdentityServiceService.updateProfile);
// A person's own preferences. The server refuses a subject other than the caller
// themselves, so every call names them — see settingsSubject.
export const resolveSettings = authed(SettingsServiceService.resolveAll);
export const setSettingValue = authed(SettingsServiceService.setValue);

export const getValidPreparations = filtered(MealPlanningServiceService.getValidPreparations);
export const searchForValidPreparations = filtered(MealPlanningServiceService.searchForValidPreparations);
export const getValidPreparationInstrumentsByPreparation = filtered(
  MealPlanningServiceService.getValidPreparationInstrumentsByPreparation,
);
export const getValidPreparationVesselsByPreparation = filtered(
  MealPlanningServiceService.getValidPreparationVesselsByPreparation,
);
export const searchValidIngredientsByPreparation = filtered(
  MealPlanningServiceService.searchValidIngredientsByPreparation,
);
export const searchValidMeasurementUnitsByIngredient = filtered(
  MealPlanningServiceService.searchValidMeasurementUnitsByIngredient,
);
export const getValidIngredientMeasurementUnitsByIngredient = filtered(
  MealPlanningServiceService.getValidIngredientMeasurementUnitsByIngredient,
);
export const getValidIngredientPreparationsByPreparation = filtered(
  MealPlanningServiceService.getValidIngredientPreparationsByPreparation,
);
export const getValidMeasurementUnits = filtered(MealPlanningServiceService.getValidMeasurementUnits);
export const searchForValidMeasurementUnits = filtered(MealPlanningServiceService.searchForValidMeasurementUnits);
export const getValidVessels = filtered(MealPlanningServiceService.getValidVessels);
export const searchForValidVessels = filtered(MealPlanningServiceService.searchForValidVessels);
export const getValidIngredientStates = filtered(MealPlanningServiceService.getValidIngredientStates);
export const searchForValidIngredientStates = filtered(MealPlanningServiceService.searchForValidIngredientStates);
export const createRecipe = authed(MealPlanningServiceService.createRecipe);
export const searchForRecipes = filtered(MealPlanningServiceService.searchForRecipes);

export const trackEvent = (event: string, properties: Record<string, string> = {}) =>
  anonymous.callAnonymous(AnalyticsServiceService.trackEvent, { source: 'web', event, properties });
export const trackAnonymousEvent = (event: string, anonymousId: string, properties: Record<string, string> = {}) =>
  anonymous.callAnonymous(AnalyticsServiceService.trackAnonymousEvent, {
    source: 'web',
    event,
    anonymousId,
    properties,
  });
