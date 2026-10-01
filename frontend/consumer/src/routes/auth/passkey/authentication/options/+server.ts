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
    return json({ options: Buffer.from(options).toString('base64') });
  } catch {
    return json({ error: 'failed to get passkey options' }, { status: 500 });
  }
};
