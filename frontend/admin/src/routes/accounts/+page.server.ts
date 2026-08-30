import type { PageServerLoad } from './$types';
import { getAccounts } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const load: PageServerLoad = async ({ locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return { accounts: [], error: 'Not authenticated' };
  }
  try {
    const res = (await getAccounts(token, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string; name?: string }>;
    };
    return { accounts: res?.results ?? [] };
  } catch (e) {
    return {
      accounts: [],
      error: e instanceof Error ? e.message : 'Failed to load accounts',
    };
  }
};
