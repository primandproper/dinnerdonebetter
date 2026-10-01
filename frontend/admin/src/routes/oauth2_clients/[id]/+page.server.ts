import type { PageServerLoad } from './$types';
import { getOAuth2Client } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const session = locals.session;
  const id = params.id;
  try {
    const res = (await getOAuth2Client(session, { oauth2ClientId: id })) as {
      result?: Record<string, unknown>;
    };
    return { client: res?.result ?? null };
  } catch (e) {
    return {
      client: null,
      error: e instanceof Error ? e.message : 'Failed to load client',
    };
  }
};
