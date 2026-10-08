/**
 * The admin login lives in one sealed, HTTP-only cookie; see @dinnerdonebetter/session for how it is
 * kept. What is this app's own is the cookie's name and where its key comes from.
 */

import type { Cookies } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { CredentialStore, Session } from '@primandproper/platform-client';
import { type ClientInfo, clientMetadata, cookieStore as sessionCookieStore } from '@dinnerdonebetter/session';
import { newSession } from '$lib/grpc/clients';

export { type ClientInfo, clientMetadata, clientOf } from '@dinnerdonebetter/session';

export function getCookieName(): string {
  return env.COOKIE_NAME ?? 'admin_webapp';
}

/** cookieStore keeps the login in this app's cookie, sealed with COOKIE_ENCRYPTION_KEY. */
export function cookieStore(cookies: Cookies): CredentialStore {
  return sessionCookieStore(cookies, {
    name: getCookieName(),
    key: env.COOKIE_ENCRYPTION_KEY,
    secure: env.NODE_ENV === 'production',
  });
}

/** sessionFor is the Session a request holds its login through. */
export function sessionFor(cookies: Cookies, client: ClientInfo = {}): Session {
  return newSession(cookieStore(cookies), clientMetadata(client));
}
