import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { beginPasskeyRegistration } from '$lib/grpc/clients';

export const POST: RequestHandler = async ({ locals }) => {
  try {
    const res = await beginPasskeyRegistration(locals.session);
    return json({ options: Buffer.from(res.options).toString('base64') });
  } catch {
    return json({ error: 'failed to get passkey options' }, { status: 500 });
  }
};
