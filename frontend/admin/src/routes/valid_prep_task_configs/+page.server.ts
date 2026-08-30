import type { PageServerLoad } from './$types';
import { getValidPrepTaskConfigs } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const load: PageServerLoad = async ({ locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return { items: [], error: 'Not authenticated' };
  }
  try {
    const res = (await getValidPrepTaskConfigs(token, {
      filter: DEFAULT_LIST_FILTER,
    })) as {
      results?: Array<{
        id?: string;
        storageType?: string;
        ingredient?: { name?: string };
        preparation?: { name?: string };
      }>;
    };
    return { items: res?.results ?? [] };
  } catch (e) {
    return {
      items: [],
      error: e instanceof Error ? e.message : 'Failed to load valid prep task configs',
    };
  }
};
