import { randomBytes, randomUUID } from 'node:crypto';
import { describe, it, expect, vi } from 'vitest';
import type { Cookies } from '@sveltejs/kit';
import { IssuedToken } from '@primandproper/platform-client';

vi.mock('$env/dynamic/private', () => ({
  env: { COOKIE_ENCRYPTION_KEY: randomBytes(32).toString('base64'), COOKIE_NAME: `session_${randomUUID()}` },
}));
// The store never calls the API; this keeps the module from dialing it.
vi.mock('$lib/grpc/clients', () => ({ newSession: vi.fn() }));

const { clientMetadata, clientOf, cookieStore, getCookieName } = await import('./session');

interface SetCall {
  value: string;
  options: Parameters<Cookies['set']>[2];
}

/** fakeCookies is a jar holding one request's cookies, recording how each was set. */
function fakeCookies() {
  const jar = new Map<string, string>();
  const sets: SetCall[] = [];
  const cookies = {
    get: (name: string) => jar.get(name),
    set: (name: string, value: string, options: Parameters<Cookies['set']>[2]) => {
      jar.set(name, value);
      sets.push({ value, options });
    },
    delete: (name: string) => {
      jar.delete(name);
    },
  } as unknown as Cookies;
  return { cookies, jar, sets };
}

function inMinutes(minutes: number): Date {
  return new Date(Math.floor(Date.now() / 1000) * 1000 + minutes * 60_000);
}

function fakeToken(overrides: Partial<IssuedToken> = {}): IssuedToken {
  return IssuedToken.create({
    token: randomUUID(),
    tokenId: randomUUID(),
    familyId: randomUUID(),
    expiresAt: inMinutes(60),
    activeAccountId: randomUUID(),
    refreshToken: randomUUID(),
    refreshTokenExpiresAt: inMinutes(60 * 24 * 30),
    ...overrides,
  });
}

describe('cookieStore', () => {
  it('loads what it saved, dates included', async () => {
    const { cookies } = fakeCookies();
    const token = fakeToken();

    await cookieStore(cookies).save(token);

    expect(await cookieStore(cookies).load()).toEqual(token);
  });

  it('keeps no part of the token readable in the cookie', async () => {
    const { cookies, sets } = fakeCookies();
    const token = fakeToken();

    await cookieStore(cookies).save(token);

    expect(sets[0].value).not.toContain(token.refreshToken);
    expect(Buffer.from(sets[0].value, 'base64').toString('latin1')).not.toContain(token.refreshToken);
    expect(sets[0].options).toMatchObject({ httpOnly: true, path: '/' });
  });

  it('lasts as long as the refresh token does', async () => {
    const { cookies, sets } = fakeCookies();
    const token = fakeToken();

    await cookieStore(cookies).save(token);

    expect(sets[0].options.expires).toEqual(token.refreshTokenExpiresAt);
  });

  it('lasts as long as the access token does when there is no refresh token', async () => {
    const { cookies, sets } = fakeCookies();
    const token = fakeToken({ refreshToken: '', refreshTokenExpiresAt: undefined });

    await cookieStore(cookies).save(token);

    expect(sets[0].options.expires).toEqual(token.expiresAt);
  });

  it('lasts until the browser closes when the token states no expiry', async () => {
    const { cookies, sets } = fakeCookies();
    const token = fakeToken({ expiresAt: undefined, refreshToken: '', refreshTokenExpiresAt: undefined });

    await cookieStore(cookies).save(token);

    expect(sets[0].options.expires).toBeUndefined();
    expect(await cookieStore(cookies).load()).toEqual(token);
  });

  it('holds no login for a cookie that does not decrypt', async () => {
    const { cookies, jar } = fakeCookies();
    jar.set(getCookieName(), randomBytes(64).toString('base64'));

    expect(await cookieStore(cookies).load()).toBeUndefined();
  });

  it('holds no login after clear', async () => {
    const { cookies } = fakeCookies();
    const store = cookieStore(cookies);
    await store.save(fakeToken());

    await store.clear();

    expect(await store.load()).toBeUndefined();
  });
});

/** requestFrom is the part of a RequestEvent clientOf reads. */
function requestFrom(headers: Record<string, string>, socketAddress?: string) {
  return {
    request: new Request('http://localhost/', { headers }),
    getClientAddress: () => {
      if (!socketAddress) {
        throw new Error('no address');
      }
      return socketAddress;
    },
  };
}

describe('clientOf', () => {
  it('reads the address Caddy stamped, and the browser', () => {
    const stamped = `203.0.113.${Math.floor(Math.random() * 250)}`;
    const userAgent = `Mozilla/5.0 ${randomUUID()}`;

    const client = clientOf(
      requestFrom({ 'x-forwarded-for': `198.51.100.1, ${stamped}`, 'user-agent': userAgent }, '10.0.0.1'),
    );

    expect(client).toEqual({ address: stamped, userAgent });
  });

  it('falls back to the connection', () => {
    const socket = `10.0.0.${Math.floor(Math.random() * 250)}`;

    expect(clientOf(requestFrom({}, socket)).address).toBe(socket);
  });

  it('survives a request with neither', () => {
    expect(clientOf(requestFrom({}))).toEqual({ address: undefined, userAgent: undefined });
  });
});

describe('clientMetadata', () => {
  it('forwards what it knows', () => {
    const address = `203.0.113.${Math.floor(Math.random() * 250)}`;
    const userAgent = `Mozilla/5.0 ${randomUUID()}`;

    expect(clientMetadata({ address, userAgent })).toEqual({
      'x-client-address': address,
      'x-client-user-agent': userAgent,
    });
  });

  it('forwards nothing it does not', () => {
    expect(clientMetadata({})).toEqual({});
  });
});
