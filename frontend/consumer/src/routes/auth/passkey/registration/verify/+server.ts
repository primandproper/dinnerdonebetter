import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { finishPasskeyRegistration } from '$lib/grpc/clients';

export const POST: RequestHandler = async ({ request, locals }) => {
  const session = locals.session;
  let body: { challenge?: string; attestationResponse?: string };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid request' }, { status: 400 });
  }

  const challenge = (body.challenge ?? '').trim();
  const attestationResponse = body.attestationResponse;

  if (!challenge || !attestationResponse) {
    return json({ error: 'attestation_response and challenge are required' }, { status: 400 });
  }

  const attestationBytes =
    typeof attestationResponse === 'string'
      ? new TextEncoder().encode(attestationResponse)
      : new TextEncoder().encode(JSON.stringify(attestationResponse));

  try {
    await finishPasskeyRegistration(session, {
      challenge,
      attestationResponse: attestationBytes,
    });
    return json({ success: true });
  } catch {
    return json({ error: 'failed to register passkey' }, { status: 400 });
  }
};
