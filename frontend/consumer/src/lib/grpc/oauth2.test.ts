import { createHash } from 'node:crypto';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { exchangeJwtForOAuth2Token } from './oauth2';

const AUTH_SERVER = 'https://auth.example.com';

afterEach(() => {
  vi.unstubAllGlobals();
});

/** Stubs the pair of calls the flow makes, returning the recorded fetch mock. */
function stubOAuthServer(tokenResponse?: () => Response) {
  const fetchMock = vi.fn(async (input: string, _init?: RequestInit) => {
    const url = String(input);
    if (url.includes('/authorize')) {
      // The 302 carries the code, the echoed state, and `iss` (RFC 9207).
      return new Response(null, {
        status: 302,
        headers: {
          location: `${AUTH_SERVER}/?code=the-code&state=xyz&iss=${encodeURIComponent(AUTH_SERVER)}`,
        },
      });
    }
    if (url.includes('/token')) {
      return (
        tokenResponse?.() ??
        new Response(JSON.stringify({ access_token: 'access-123', refresh_token: 'refresh-456', expires_in: 900 }), {
          status: 200,
          headers: { 'content-type': 'application/json' },
        })
      );
    }
    throw new Error(`unexpected url ${url}`);
  });
  vi.stubGlobal('fetch', fetchMock);

  return fetchMock;
}

type FetchCall = [string, RequestInit | undefined];

function callFor(fetchMock: ReturnType<typeof stubOAuthServer>, path: string): FetchCall {
  const call = fetchMock.mock.calls.find(([u]) => new URL(String(u)).pathname === path);
  if (!call) {
    throw new Error(`no request to ${path}`);
  }

  return call as unknown as FetchCall;
}

describe('exchangeJwtForOAuth2Token', () => {
  it('exchanges a JWT for an OAuth2 access token', async () => {
    const fetchMock = stubOAuthServer();

    const result = await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');

    expect(result).toEqual({ accessToken: 'access-123', refreshToken: 'refresh-456', expiresIn: 900 });

    // The authorize call carries the JWT as a bearer; the token call posts the code from the redirect.
    const [, authorizeInit] = callFor(fetchMock, '/authorize');
    expect(new Headers(authorizeInit?.headers).get('authorization')).toBe('Bearer jwt-token');
    const [, tokenInit] = callFor(fetchMock, '/token');
    expect(new URLSearchParams(String(tokenInit?.body)).get('code')).toBe('the-code');
  });

  it('authorizes at POST /authorize, not GET /oauth2/authorize', async () => {
    // A GET renders the login form; only a POST runs the authenticator that reads the bearer.
    const fetchMock = stubOAuthServer();

    await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');

    const [authorizeUrl, authorizeInit] = callFor(fetchMock, '/authorize');
    expect(authorizeInit?.method).toBe('POST');
    // The authorization parameters stay in the query string, and nothing travels in a body.
    expect(new URL(authorizeUrl).searchParams.get('response_type')).toBe('code');
    expect(authorizeInit?.body).toBeUndefined();

    const [tokenUrl] = callFor(fetchMock, '/token');
    expect(tokenUrl).toBe(`${AUTH_SERVER}/token`);
  });

  it('sends an S256 PKCE challenge and redeems it with the matching verifier', async () => {
    const fetchMock = stubOAuthServer();

    await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');

    const [authorizeUrl] = callFor(fetchMock, '/authorize');
    const params = new URL(authorizeUrl).searchParams;
    expect(params.get('code_challenge_method')).toBe('S256');

    const challenge = params.get('code_challenge') ?? '';
    // 43 unpadded base64url characters is the shape the server checks a challenge for.
    expect(challenge).toMatch(/^[A-Za-z0-9_-]{43}$/);

    const [, tokenInit] = callFor(fetchMock, '/token');
    const verifier = new URLSearchParams(String(tokenInit?.body)).get('code_verifier') ?? '';
    expect(verifier).toMatch(/^[A-Za-z0-9_-]{43}$/);
    expect(createHash('sha256').update(verifier).digest('base64url')).toBe(challenge);
  });

  it('sends a fresh verifier and state on each exchange', async () => {
    const fetchMock = stubOAuthServer();

    await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');
    const first = new URL(callFor(fetchMock, '/authorize')[0]).searchParams;
    const firstVerifier = new URLSearchParams(String(callFor(fetchMock, '/token')[1]?.body)).get('code_verifier');
    fetchMock.mockClear();

    await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');
    const second = new URL(callFor(fetchMock, '/authorize')[0]).searchParams;
    const secondVerifier = new URLSearchParams(String(callFor(fetchMock, '/token')[1]?.body)).get('code_verifier');

    expect(second.get('code_challenge')).not.toBe(first.get('code_challenge'));
    expect(secondVerifier).not.toBe(firstVerifier);
    expect(second.get('state')).not.toBe(first.get('state'));
  });

  it('sends the same registered redirect URI at both steps', async () => {
    // Matching is byte for byte at /authorize and again at /token, so the two must agree
    // exactly with the URI registered for the client.
    const fetchMock = stubOAuthServer();

    await exchangeJwtForOAuth2Token(AUTH_SERVER, 'client-id', 'client-secret', 'jwt-token');

    const authorizeRedirect = new URL(callFor(fetchMock, '/authorize')[0]).searchParams.get('redirect_uri');
    const tokenRedirect = new URLSearchParams(String(callFor(fetchMock, '/token')[1]?.body)).get('redirect_uri');

    expect(authorizeRedirect).toBe(AUTH_SERVER);
    expect(tokenRedirect).toBe(AUTH_SERVER);
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
    // OAuth2 token endpoints return JSON error bodies (RFC 6749 §5.2).
    stubOAuthServer(
      () =>
        new Response(JSON.stringify({ error: 'invalid_grant' }), {
          status: 400,
          headers: { 'content-type': 'application/json' },
        }),
    );

    await expect(exchangeJwtForOAuth2Token(AUTH_SERVER, 'id', 'secret', 'jwt')).rejects.toThrow(
      /token exchange failed: 400/i,
    );
  });
});
