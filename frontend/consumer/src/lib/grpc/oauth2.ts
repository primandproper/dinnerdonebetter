/**
 * OAuth2 authorization code flow. Replicates backend/pkg/client/client.go WithOAuth2Credentials.
 * Uses JWT (from LoginForToken) as Bearer token to obtain OAuth2 access token for gRPC.
 */

import { randomBytes } from 'node:crypto';
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
    retry: { maxAttempts: 3, baseDelayMs: 100, maxDelayMs: 30_000, jitter: 0.1 },
    fetch: manualRedirectFetch,
  },
  observabilityDeps,
);

/**
 * Exchanges JWT (from LoginForToken) for OAuth2 access token via authorization code flow.
 * 1. GET /oauth2/authorize with Bearer JWT -> redirect with ?code=...
 * 2. POST /oauth2/token with code -> OAuth2 access token
 */
export async function exchangeJwtForOAuth2Token(
  authServerUrl: string,
  clientId: string,
  clientSecret: string,
  jwt: string,
): Promise<OAuth2TokenResult> {
  const state = randomBytes(32).toString('base64url');
  const authUrl = new URL('/oauth2/authorize', authServerUrl);
  authUrl.searchParams.set('response_type', 'code');
  authUrl.searchParams.set('client_id', clientId);
  authUrl.searchParams.set('redirect_uri', authServerUrl);
  authUrl.searchParams.set('state', state);
  authUrl.searchParams.set('code_challenge_method', 'plain');

  const authRes = await oauthHttp.get(authUrl.toString(), {
    headers: { Authorization: `Bearer ${jwt}` },
  });

  const location = authRes.headers.get('location');
  if (!location) {
    throw new PlatformError('oauth2/no-redirect', 'No redirect location from oauth2 authorize');
  }

  const redirectUrl = new URL(location, authServerUrl);
  const code = redirectUrl.searchParams.get('code');
  if (!code) {
    throw new PlatformError('oauth2/no-code', 'Code not returned from oauth2 redirect');
  }

  const tokenUrl = `${authServerUrl.replace(/\/$/, '')}/oauth2/token`;
  const tokenRes = await oauthHttp.post<OAuth2TokenResponse>(
    tokenUrl,
    new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: authServerUrl,
      client_id: clientId,
      client_secret: clientSecret,
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
