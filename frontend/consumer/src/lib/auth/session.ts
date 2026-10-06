/**
 * The login lives in one sealed, HTTP-only cookie, and a Session per request reads and writes
 * it through platform-client's encryptedCredentialStore. The cookie holds the refresh token,
 * which is the credential worth stealing, so it never leaves the server: nothing in the
 * browser can read it.
 */

import type { Cookies } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { type CredentialStore, encryptedCredentialStore, type Session } from '@primandproper/platform-client';
import { newSession } from '$lib/grpc/clients';

export function getCookieName(): string {
  return env.COOKIE_NAME ?? 'consumer_session';
}

/** cookieKey is COOKIE_ENCRYPTION_KEY, a base64-encoded 32-byte key, which the store checks the length of. */
function cookieKey(): Uint8Array {
  const encoded = env.COOKIE_ENCRYPTION_KEY;
  if (!encoded) {
    throw new Error('COOKIE_ENCRYPTION_KEY is required');
  }
  return Buffer.from(encoded, 'base64');
}

/**
 * cookieStore keeps an IssuedToken in the session cookie, which lives exactly as long as the
 * login can: until the refresh token stops being exchangeable, or, for a login with no
 * refresh token, until the access token expires. One that states neither lasts until the
 * browser closes, or until the server refuses it.
 */
export function cookieStore(cookies: Cookies): CredentialStore {
  const name = getCookieName();

  return encryptedCredentialStore(cookieKey(), {
    get: () => cookies.get(name),
    set: (value, expires) =>
      cookies.set(name, value, {
        path: '/',
        httpOnly: true,
        secure: env.NODE_ENV === 'production',
        sameSite: 'lax',
        expires,
      }),
    delete: () => cookies.delete(name, { path: '/' }),
  });
}

/** sessionFor is the Session a request holds its login through. */
export function sessionFor(cookies: Cookies): Session {
  return newSession(cookieStore(cookies));
}
