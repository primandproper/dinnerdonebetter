import { json } from '@sveltejs/kit';
import { beginPasskeyRegistration } from '@primandproper/platform-client';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ locals }) => {
  try {
    const options = await beginPasskeyRegistration(locals.session);
    // The options are JSON already, which the browser hands parseRegistrationOptions as they are.
    return new Response(new TextDecoder().decode(options), { headers: { 'content-type': 'application/json' } });
  } catch {
    return json({ error: 'failed to get passkey options' }, { status: 500 });
  }
};
