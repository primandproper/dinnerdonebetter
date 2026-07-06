import { describe, it, expect, vi, afterEach } from 'vitest';
import { exchangeJwtForOAuth2Token } from './oauth2';

const AUTH_SERVER = 'https://auth.example.com';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('exchangeJwtForOAuth2Token', () => {
  it('exchanges a JWT for an OAuth2 access token', async () => {
    const fetchMock = vi.fn(async (input: string, _init?: RequestInit) => {
      const url = String(input);
      if (url.includes('/oauth2/authorize')) {
        return new Response(null, {
          status: 302,
          headers: { location: `${AUTH_SERVER}/?code=the-code&state=xyz` },
        });
      }
      if (url.includes('/oauth2/token')) {
        return new Response(
          JSON.stringify({ access_token: 'access-123', refresh_token: 'refresh-456', expires_in: 3600 }),
          { status: 200, headers: { 'content-type': 'application/json' } },
        );
      }
      throw new Error(`unexpected url ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    const result = await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');

    expect(result).toEqual({ accessToken: 'access-123', refreshToken: 'refresh-456', expiresIn: 3600 });

    // The authorize call carries the JWT as a bearer; the token call posts the code from the redirect.
    const authorizeCall = fetchMock.mock.calls.find(([u]) => String(u).includes('/oauth2/authorize'));
    expect(new Headers((authorizeCall?.[1] as RequestInit)?.headers).get('authorization')).toBe('Bearer jwt-token');
    const tokenCall = fetchMock.mock.calls.find(([u]) => String(u).includes('/oauth2/token'));
    expect(String((tokenCall?.[1] as RequestInit)?.body)).toContain('code=the-code');
  });

  it('throws when the authorize step returns no redirect location', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(null, { status: 200 })),
    );

    await expect(exchangeJwtForOAuth2Token(AUTH_SERVER, 'id', 'secret', 'jwt')).rejects.toThrow(
      /no redirect location/i,
    );
  });

  it('throws when the token exchange returns a non-2xx status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: string) => {
        const url = String(input);
        if (url.includes('/oauth2/authorize')) {
          return new Response(null, {
            status: 302,
            headers: { location: `${AUTH_SERVER}/?code=the-code` },
          });
        }
        // OAuth2 token endpoints return JSON error bodies (RFC 6749 §5.2).
        return new Response(JSON.stringify({ error: 'invalid_grant' }), {
          status: 400,
          headers: { 'content-type': 'application/json' },
        });
      }),
    );

    await expect(exchangeJwtForOAuth2Token(AUTH_SERVER, 'id', 'secret', 'jwt')).rejects.toThrow(
      /token exchange failed: 400/i,
    );
  });
});
