import type { PageServerLoad } from './$types';
import { getValidInstruments, searchForValidInstruments } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.accessToken;
  if (!token) {
    return { items: [], error: 'Not authenticated' };
  }
  const query = url.searchParams.get('q')?.trim() ?? '';
  try {
    const res =
      query === ''
        ? ((await getValidInstruments(token, { filter: DEFAULT_LIST_FILTER })) as {
            results?: Array<{ id?: string; name?: string }>;
          })
        : ((await searchForValidInstruments(token, {
            filter: DEFAULT_LIST_FILTER,
            query,
            useSearchService: false,
          })) as { results?: Array<{ id?: string; name?: string }> });
    return { items: res?.results ?? [] };
  } catch (e) {
    return {
      items: [],
      error: e instanceof Error ? e.message : 'Failed to load valid instruments',
    };
  }
};
