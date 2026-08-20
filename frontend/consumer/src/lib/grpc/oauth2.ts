/**
 * OAuth2 authorization code flow. Replicates backend/pkg/client/client.go WithOAuth2Credentials.
 * Uses JWT (from LoginForToken) as Bearer token to obtain OAuth2 access token for gRPC.
 */

import { createHash, randomBytes } from 'node:crypto';
import { PlatformError } from '@primandproper/errors';
import { FetchHttpClient, type FetchLike } from '@primandproper/httpclient';
import { observabilityDeps } from '$lib/observability';

export interface OAuth2TokenResult {
  accessToken: string;
  refreshToken?: string;
  expiresIn?: number;
}

interface OAuth2TokenResponse {
  access_token: string;
  refresh_token?: string;
  expires_in?: number;
}

/**
 * The authorize step returns a 302 whose Location header carries the code. `httpclient` has no
 * per-request redirect control, so we force manual redirect at the fetch layer for this
 * dedicated client; both OAuth calls then flow through it, gaining OTel spans and retry on
 * transient network failures.
 */
const manualRedirectFetch: FetchLike = (input, init) => fetch(input, { ...init, redirect: 'manual' });

const oauthHttp = new FetchHttpClient(
  {
    headers: {},
    timeoutMs: 30_000,
    retry: { maxAttempts: 3, baseDelayMs: 100, maxDelayMs: 30_000, jitter: 0.1, maxElapsedMs: 0 },
    fetch: manualRedirectFetch,
  },
  observabilityDeps,
);

/**
 * Generates a PKCE (RFC 7636) verifier and its S256 challenge. The authorization server requires
 * a challenge and accepts S256 only — a missing method is refused rather than defaulted, because
 * RFC 7636 defaults it to `plain`, which this server does not accept at all.
 *
 * 32 random bytes encode to 43 unpadded base64url characters, which is both a legal verifier and
 * the exact shape the server checks an S256 challenge for.
 */
function generatePkcePair(): { verifier: string; challenge: string } {
  const verifier = randomBytes(32).toString('base64url');

  return { verifier, challenge: createHash('sha256').update(verifier).digest('base64url') };
}

/**
 * Exchanges JWT (from LoginForToken) for OAuth2 access token via authorization code flow.
 * 1. POST /authorize with Bearer JWT -> redirect with ?code=...
 * 2. POST /token with the code and the PKCE verifier -> OAuth2 access token
 */
export async function exchangeJwtForOAuth2Token(
  authServerUrl: string,
  clientId: string,
  clientSecret: string,
  jwt: string,
): Promise<OAuth2TokenResult> {
  const state = randomBytes(32).toString('base64url');
  const { verifier, challenge } = generatePkcePair();

  const authUrl = new URL('/authorize', authServerUrl);
  authUrl.searchParams.set('response_type', 'code');
  authUrl.searchParams.set('client_id', clientId);
  // Matched byte for byte against the client's registered redirect_uris here, and again at
  // /token against the URI the code was issued for — not by hostname, not ignoring ports.
  // `ddb-bootstrap init` registers `--api-server-url` for every first-party client, so this
  // string has to be that one exactly, trailing slash included.
  authUrl.searchParams.set('redirect_uri', authServerUrl);
  authUrl.searchParams.set('state', state);
  authUrl.searchParams.set('code_challenge', challenge);
  authUrl.searchParams.set('code_challenge_method', 'S256');

  // POST rather than GET: a GET renders the login form — the answer for a browser arriving
  // without a session — and only a POST runs the authenticator that reads this bearer token.
  // The authorization parameters stay in the query string on both, so the request that issues
  // the code is the one that was validated. No body travels with it; the content type keeps the
  // request well-formed for a server that parses the form on every request.
  const authRes = await oauthHttp.post(authUrl.toString(), undefined, {
    headers: {
      'Authorization': `Bearer ${jwt}`,
      'Content-Type': 'application/x-www-form-urlencoded',
    },
  });

  const location = authRes.headers.get('location');
  if (!location) {
    throw new PlatformError('oauth2/no-redirect', 'No redirect location from oauth2 authorize');
  }

  // The redirect also carries `iss` (RFC 9207) alongside the code and state. Nothing here reads
  // it; it is not an error parameter.
  const redirectUrl = new URL(location, authServerUrl);
  const code = redirectUrl.searchParams.get('code');
  if (!code) {
    throw new PlatformError('oauth2/no-code', 'Code not returned from oauth2 redirect');
  }

  const tokenUrl = `${authServerUrl.replace(/\/$/, '')}/token`;
  const tokenRes = await oauthHttp.post<OAuth2TokenResponse>(
    tokenUrl,
    new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: authServerUrl,
      client_id: clientId,
      client_secret: clientSecret,
      code_verifier: verifier,
    }),
    { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } },
  );

  if (!tokenRes.ok) {
    const text = await tokenRes.text();
    throw new PlatformError('oauth2/token-exchange', `OAuth2 token exchange failed: ${tokenRes.status} ${text}`);
  }

  const tokenData = tokenRes.data;

  return {
    accessToken: tokenData.access_token,
    refreshToken: tokenData.refresh_token,
    expiresIn: tokenData.expires_in,
  };
}
