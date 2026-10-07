import type { PageServerLoad } from './$types';
import { listUsers } from '$lib/grpc/clients';
import { QueryFilter } from '@primandproper/platform-client/filtering/v1';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const res = (await listUsers(session, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string; username?: string; firstName?: string; lastName?: string }>;
    };
    return { users: res?.results ?? [] };
  } catch (e) {
    return {
      users: [],
      error: e instanceof Error ? e.message : 'Failed to load users',
    };
  }
};
