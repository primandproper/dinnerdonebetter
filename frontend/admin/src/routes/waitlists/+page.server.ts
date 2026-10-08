import type { PageServerLoad } from './$types';
import { listWaitlists } from '$lib/grpc/clients';
import { QueryFilter } from '@primandproper/platform-client/filtering/v1';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const res = (await listWaitlists(session, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string; name?: string }>;
    };
    return { waitlists: res?.results ?? [] };
  } catch (e) {
    return {
      waitlists: [],
      error: e instanceof Error ? e.message : 'Failed to load waitlists',
    };
  }
};
