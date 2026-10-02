import { json } from '@sveltejs/kit';
import { passkeySignIn, PasskeyReason, PlatformError } from '@primandproper/platform-client';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, locals }) => {
  let body: { username?: string; assertionResponse?: unknown; totpCode?: string };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid request' }, { status: 400 });
  }

  const assertionResponse = body.assertionResponse;
  if (!assertionResponse) {
    return json({ error: 'assertionResponse is required' }, { status: 400 });
  }
  const response = new TextEncoder().encode(
    typeof assertionResponse === 'string' ? assertionResponse : JSON.stringify(assertionResponse),
  );

  try {
    // The login a passkey mints is the one a password does, refresh token and all, and the
    // Session writes it to the cookie.
    const result = await passkeySignIn(locals.session, {
      username: (body.username ?? '').trim(),
      response,
      totpCode: (body.totpCode ?? '').trim(),
    });
    if (result.kind === 'second_factor_required') {
      // The assertion's challenge is spent: the browser asks for the code, then for the key
      // again, and sends both.
      return json({ error: 'second factor required', totpRequired: true }, { status: 401 });
    }
    return json({ success: true, redirect: '/' });
  } catch (err) {
    if (err instanceof PlatformError && err.is(PasskeyReason.PASSKEY_SIGN_COUNT_REGRESSED)) {
      return json(
        { error: 'This passkey looks like a copy of another. Remove it, and add a new one.' },
        { status: 403 },
      );
    }
    return json({ error: 'authentication failed' }, { status: 401 });
  }
};
