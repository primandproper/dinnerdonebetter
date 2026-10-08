/**
 * @dinnerdonebetter/session
 * The SvelteKit side of a platform-client Session, which both web apps hold their login in:
 * the cookie it is kept in, the browser it is held from, and the redirects an authenticated
 * call answers with.
 */

export { type Caller, createCaller, SIGN_IN_PATH } from './caller';
export { type ClientInfo, clientMetadata, clientOf } from './client-info';
export { type CookieStoreConfig, cookieStore } from './cookies';
export { CHANGE_PASSWORD_PATH, landingAfterSignIn, mustChangePassword } from './required-actions';
