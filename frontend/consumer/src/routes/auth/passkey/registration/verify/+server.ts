import { json } from '@sveltejs/kit';
import { finishPasskeyRegistration, PasskeyReason, PlatformError } from '@primandproper/platform-client';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, locals }) => {
  let body: { attestationResponse?: string; friendlyName?: string };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid request' }, { status: 400 });
  }

  // The browser sends the credential as platform-client's webauthn bridge serialized it.
  const attestationResponse = body.attestationResponse;
  if (typeof attestationResponse !== 'string' || !attestationResponse) {
    return json({ error: 'attestationResponse is required' }, { status: 400 });
  }
  const response = new TextEncoder().encode(attestationResponse);

  try {
    await finishPasskeyRegistration(locals.session, { friendlyName: (body.friendlyName ?? '').trim(), response });
    return json({ success: true });
  } catch (err) {
    if (err instanceof PlatformError && err.is(PasskeyReason.PASSKEY_ALREADY_REGISTERED)) {
      return json({ error: 'This passkey is already on your account.' }, { status: 409 });
    }
    return json({ error: 'failed to register passkey' }, { status: 400 });
  }
};
