import type { PageServerLoad } from './$types';
import { getWaitlist } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const token = locals.accessToken;
  const id = params.id;
  if (!token) {
    return { waitlist: null, error: 'Not authenticated' };
  }
  try {
    const res = (await getWaitlist(token, { listId: id })) as { result?: Record<string, unknown> };
    return { waitlist: res?.result ?? null };
  } catch (e) {
    return {
      waitlist: null,
      error: e instanceof Error ? e.message : 'Failed to load waitlist',
    };
  }
};
