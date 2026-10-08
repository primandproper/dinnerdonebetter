import type { PageServerLoad } from './$types';
import { getValidIngredientStates, searchForValidIngredientStates } from '$lib/grpc/clients';
import { QueryFilter } from '@primandproper/platform-client/filtering/v1';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const load: PageServerLoad = async ({ locals, url }) => {
  const session = locals.session;
  const query = url.searchParams.get('q')?.trim() ?? '';
  try {
    const res =
      query === ''
        ? ((await getValidIngredientStates(session, { filter: DEFAULT_LIST_FILTER })) as {
            results?: Array<{ id?: string; name?: string }>;
          })
        : ((await searchForValidIngredientStates(session, {
            filter: DEFAULT_LIST_FILTER,
            query,
            useSearchService: false,
          })) as { results?: Array<{ id?: string; name?: string }> });
    return { items: res?.results ?? [] };
  } catch (e) {
    return {
      items: [],
      error: e instanceof Error ? e.message : 'Failed to load valid ingredient states',
    };
  }
};
