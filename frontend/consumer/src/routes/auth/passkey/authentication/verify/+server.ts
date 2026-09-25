import { json } from '@sveltejs/kit';
import { IssuedToken } from '@primandproper/platform-client';
import type { RequestHandler } from './$types';
import { finishPasskeyAuthentication } from '$lib/grpc/clients';
import { cookieStore } from '$lib/auth/session';

export const POST: RequestHandler = async ({ request, cookies }) => {
  let body: { challenge?: string; username?: string; assertionResponse?: unknown };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid request' }, { status: 400 });
  }

  const challenge = (body.challenge ?? '').trim();
  const username = (body.username ?? '').trim();
  const assertionResponse = body.assertionResponse;

  if (!challenge || !assertionResponse) {
    return json({ error: 'assertion_response and challenge are required' }, { status: 400 });
  }

  const assertionBytes =
    typeof assertionResponse === 'string'
      ? new TextEncoder().encode(assertionResponse)
      : new TextEncoder().encode(JSON.stringify(assertionResponse));

  try {
    const tokenRes = await finishPasskeyAuthentication({
      challenge,
      username,
      assertionResponse: assertionBytes,
    });
    const result = tokenRes.result;
    if (!result?.accessToken) {
      return json({ error: 'no access token' }, { status: 500 });
    }

    // Platform has no door a passkey can sign in through yet (platform-go#874), so this is
    // AuthService's token, and platform's refresh can't exchange AuthService's refresh token.
    // It is held without one: the login lasts as long as the access token does. That is the
    // server's to say, not expiresUtc's, which reads as the moment of issue whenever the
    // lifetime is left to the token issuer's default. With no expiry here the Session sends
    // the token until the server refuses it, and a refusal ends the login.
    await cookieStore(cookies).save(
      IssuedToken.create({
        token: result.accessToken,
        activeAccountId: result.accountId,
      }),
    );

    return json({ success: true, redirect: '/' });
  } catch {
    return json({ error: 'authentication failed' }, { status: 401 });
  }
};
