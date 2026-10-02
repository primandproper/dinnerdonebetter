import type { PageServerLoad } from './$types';
import { getWaitlist } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const session = locals.session;
  const id = params.id;
  try {
    const res = (await getWaitlist(session, { listId: id })) as { result?: Record<string, unknown> };
    return { waitlist: res?.result ?? null };
  } catch (e) {
    return {
      waitlist: null,
      error: e instanceof Error ? e.message : 'Failed to load waitlist',
    };
  }
};
