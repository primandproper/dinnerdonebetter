import type { PageServerLoad } from './$types';
import { getServiceSettings, searchForServiceSettings } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.accessToken;
  if (!token) {
    return { settings: [], error: 'Not authenticated' };
  }
  const query = url.searchParams.get('q')?.trim() ?? '';
  try {
    const res =
      query === ''
        ? ((await getServiceSettings(token, { filter: DEFAULT_LIST_FILTER })) as {
            results?: Array<{ id?: string; name?: string }>;
          })
        : ((await searchForServiceSettings(token, {
            filter: DEFAULT_LIST_FILTER,
            query,
          })) as { results?: Array<{ id?: string; name?: string }> });
    return { settings: res?.results ?? [] };
  } catch (e) {
    return {
      settings: [],
      error: e instanceof Error ? e.message : 'Failed to load settings',
    };
  }
};
