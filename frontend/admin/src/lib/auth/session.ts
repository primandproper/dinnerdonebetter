/**
 * The admin login lives in one encrypted, HTTP-only cookie, and a Session per request reads
 * and writes it through the CredentialStore below, exactly as the consumer app's does. The
 * cookie holds the refresh token, so it never leaves the server.
 */

import type { Cookies } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { type CredentialStore, IssuedToken, type Session } from '@primandproper/platform-client';
import { newSession } from '$lib/grpc/clients';
import { encrypt, decrypt } from './crypto';

export function getCookieName(): string {
  return env.COOKIE_NAME ?? 'admin_webapp';
}

/**
 * cookieStore keeps an IssuedToken in the session cookie, which lives exactly as long as the
 * login can: until the refresh token stops being exchangeable.
 */
export function cookieStore(cookies: Cookies): CredentialStore {
  const name = getCookieName();

  return {
    async load() {
      const value = cookies.get(name);
      if (!value) {
        return undefined;
      }
      try {
        const token = IssuedToken.fromJSON(decrypt<unknown>(value));
        // Anything that decrypts but holds no token (a cookie from before this format) is no login.
        return token.token ? token : undefined;
      } catch {
        return undefined;
      }
    },
    async save(token) {
      cookies.set(name, encrypt(IssuedToken.toJSON(token)), {
        path: '/',
        httpOnly: true,
        secure: env.NODE_ENV === 'production',
        sameSite: 'lax',
        expires: token.refreshTokenExpiresAt ?? token.expiresAt,
      });
    },
    async clear() {
      cookies.delete(name, { path: '/' });
    },
  };
}

/** sessionFor is the Session a request holds its login through. */
export function sessionFor(cookies: Cookies): Session {
  return newSession(cookieStore(cookies));
}
