/**
 * Every call the consumer app makes, platform's services and this repository's alike, goes
 * over one transport. An authenticated call goes through the request's Session (see
 * $lib/auth/session), which holds the login, refreshes it through platform's SignInService,
 * and retries a call the server refused once with the successor.
 */

import { env } from '$env/dynamic/private';
import { redirect } from '@sveltejs/kit';
import { CHANGE_PASSWORD_PATH, mustChangePassword } from '$lib/auth/required-actions';
import { MealPlanningServiceService, QueryFilter, createPlatformTransport } from '@dinnerdonebetter/api-client';
import { type CredentialStore, redirectOnNotSignedIn, Session, type UnaryMethod } from '@primandproper/platform-client';
import { IdentityServiceService } from '@primandproper/platform-client/identity/v1';
import { PasskeysServiceService } from '@primandproper/platform-client/passkeys/v1';
import { SignInServiceService } from '@primandproper/platform-client/signin/v1';
import { SettingsServiceService } from '@primandproper/platform-client/settings/v1';

const transport = createPlatformTransport({
  serverUrl: env.GRPC_API_SERVER_URL ?? 'localhost:50051',
  insecure: env.DEVELOPING_LOCALLY === 'true',
});

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
 * already cleared the cookie by then. One refused because they owe a password change sends
 * them to the form for it.
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

// QueryFilter is platform's schema, and its four timestamp windows are message fields
// rather than `optional` scalars, so a bare object literal is not one. The generated
// factory fills every field, which is what it is for.
const defaultSearchFilter = QueryFilter.create({ maxResponseSize: 20 });

/** filtered is `authed` for a list or search, filling in the default filter when none is given. */
function filtered<Req extends { filter: QueryFilter | undefined }, Res>(method: UnaryMethod<Req, Res>) {
  return (session: Session, request: Omit<Req, 'filter'> & { filter?: QueryFilter }) =>
    call(session, method, { ...request, filter: request.filter ?? defaultSearchFilter } as Req);
}

// Sign-in, sign-out, password reset, email verification and passkey sign-in are the
// library's own helpers (signIn, signOut, requestPasswordReset, passkeySignIn,
// beginPasskeyRegistration and the rest), called with the request's Session. What's here
// is the caller's own credentials and logins, which are plain calls on platform's services.
export const getSelf = async (session: Session) => (await call(session, SignInServiceService.getSelf, {})).user;
export const getPrincipal = (session: Session) => call(session, IdentityServiceService.getPrincipal, {});
export const getActiveAccount = async (session: Session) => (await getPrincipal(session)).activeAccount;
export const listPasskeys = async (session: Session) =>
  (await call(session, PasskeysServiceService.listPasskeys, {})).passkeys;
export const archivePasskey = authed(PasskeysServiceService.archivePasskey);
// A person's own logins: one per device or browser they've signed in on.
export const listSignIns = async (session: Session) =>
  (await call(session, SignInServiceService.listSignIns, { limit: 0 })).signIns;
export const endSignIn = authed(SignInServiceService.endSignIn);
export const endOtherSignIns = (session: Session) => call(session, SignInServiceService.endOtherSignIns, {});
// The handle and the address are on the sign-in surface rather than in the directory's
// profile update, and they ask for the password and a second factor: whoever holds the
// address can take the account through a password reset, so changing it is a credential
// change.
export const updateUsername = authed(SignInServiceService.updateUsername);
export const updateEmailAddress = authed(SignInServiceService.updateEmailAddress);
// The one call a person who owes a password change may still make.
export const updatePassword = authed(SignInServiceService.updatePassword);

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
