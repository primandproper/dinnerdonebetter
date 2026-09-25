import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getValidIngredientStates, searchForValidIngredientStates } from '$lib/grpc/clients';
import { logger } from '$lib/logger';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const GET: RequestHandler = async ({ url, locals }) => {
  const session = locals.session;
  const q = url.searchParams.get('q') ?? '';
  try {
    const res =
      q === ''
        ? await getValidIngredientStates(session, { filter: DEFAULT_LIST_FILTER })
        : await searchForValidIngredientStates(session, {
            filter: DEFAULT_LIST_FILTER,
            query: q,
            useSearchService: q.length > 2,
          });
    return json({ results: res.results ?? [] });
  } catch (e) {
    logger.error('searchForValidIngredientStates failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
