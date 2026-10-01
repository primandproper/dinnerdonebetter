import { json } from '@sveltejs/kit';
import { PasskeyReason, PlatformError } from '@primandproper/platform-client';
import type { RequestHandler } from './$types';
import { finishPasskeyRegistration } from '$lib/grpc/clients';

export const POST: RequestHandler = async ({ request, locals }) => {
  let body: { attestationResponse?: unknown; friendlyName?: string };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid request' }, { status: 400 });
  }

  const attestationResponse = body.attestationResponse;
  if (!attestationResponse) {
    return json({ error: 'attestationResponse is required' }, { status: 400 });
  }
  const response = new TextEncoder().encode(
    typeof attestationResponse === 'string' ? attestationResponse : JSON.stringify(attestationResponse),
  );

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
