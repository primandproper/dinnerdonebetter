/**
 * The login lives in one sealed, HTTP-only cookie, and a Session per request reads and writes
 * it through platform-client's encryptedCredentialStore. The cookie holds the refresh token,
 * which is the credential worth stealing, so it never leaves the server: nothing in the
 * browser can read it.
 */

import type { Cookies, RequestEvent } from '@sveltejs/kit';
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

/** ClientInfo is the browser a request came from. */
export interface ClientInfo {
  address?: string;
  userAgent?: string;
}

/**
 * clientOf is the browser behind a request. The address is the last X-Forwarded-For entry,
 * which Caddy writes and a browser cannot, and the connection's own address when there is none.
 */
export function clientOf(event: Pick<RequestEvent, 'request' | 'getClientAddress'>): ClientInfo {
  const forwarded = (event.request.headers.get('x-forwarded-for') ?? '')
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean);

  let address = forwarded.at(-1);
  if (!address) {
    try {
      address = event.getClientAddress();
    } catch {
      address = undefined;
    }
  }

  return { address, userAgent: event.request.headers.get('user-agent') ?? undefined };
}

/**
 * clientMetadata forwards the browser to the API, which records it beside every login this
 * Session signs in or renews, so "where you're signed in" names the browser rather than this
 * server. It is shown to the person whose login it is and decides nothing.
 */
export function clientMetadata(client: ClientInfo): Record<string, string> {
  const metadata: Record<string, string> = {};
  if (client.address) {
    metadata['x-client-address'] = client.address;
  }
  if (client.userAgent) {
    metadata['x-client-user-agent'] = client.userAgent;
  }
  return metadata;
}

/** sessionFor is the Session a request holds its login through. */
export function sessionFor(cookies: Cookies, client: ClientInfo = {}): Session {
  return newSession(cookieStore(cookies), clientMetadata(client));
}
