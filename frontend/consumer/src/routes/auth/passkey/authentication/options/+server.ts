import { json } from '@sveltejs/kit';
import { beginPasskeySignIn } from '@primandproper/platform-client';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, locals }) => {
  let body: { username?: string };
  try {
    body = await request.json();
  } catch {
    return json({ error: 'invalid request' }, { status: 400 });
  }

  try {
    // An empty username is the discoverable login. Either way the answer looks the same
    // whether or not anybody holds the name, so the prompt does too.
    const options = await beginPasskeySignIn(locals.session, (body.username ?? '').trim());
    // The options are JSON already, which the browser hands parseAssertionOptions as they are.
    return new Response(new TextDecoder().decode(options), { headers: { 'content-type': 'application/json' } });
  } catch {
    return json({ error: 'failed to get passkey options' }, { status: 500 });
  }
};
