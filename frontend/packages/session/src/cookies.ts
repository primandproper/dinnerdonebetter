/**
 * The login lives in one sealed, HTTP-only cookie, and a Session per request reads and writes
 * it through platform-client's encryptedCredentialStore. The cookie holds the refresh token,
 * which is the credential worth stealing, so it never leaves the server: nothing in the
 * browser can read it.
 */

import type { Cookies } from '@sveltejs/kit';
import { type CredentialStore, encryptedCredentialStore } from '@primandproper/platform-client';

export interface CookieStoreConfig {
  /** name is the cookie's name. Each app has its own, so a browser can hold both logins. */
  name: string;
  /**
   * key is a base64-encoded 32-byte key, which the store checks the length of. It is a
   * string or nothing because that is what an app reads out of its environment.
   */
  key: string | undefined;
  /** secure keeps the cookie off plain HTTP, which local development is served over. */
  secure: boolean;
}

function decodeKey(encoded: string | undefined): Uint8Array {
  if (!encoded) {
    throw new Error('a cookie encryption key is required');
  }
  return Buffer.from(encoded, 'base64');
}

/**
 * cookieStore keeps an IssuedToken in the session cookie, which lives exactly as long as the
 * login can: until the refresh token stops being exchangeable, or, for a login with no
 * refresh token, until the access token expires. One that states neither lasts until the
 * browser closes, or until the server refuses it.
 */
export function cookieStore(cookies: Cookies, config: CookieStoreConfig): CredentialStore {
  const { name, secure } = config;

  return encryptedCredentialStore(decodeKey(config.key), {
    get: () => cookies.get(name),
    set: (value, expires) =>
      cookies.set(name, value, {
        path: '/',
        httpOnly: true,
        secure,
        sameSite: 'lax',
        expires,
      }),
    delete: () => cookies.delete(name, { path: '/' }),
  });
}
