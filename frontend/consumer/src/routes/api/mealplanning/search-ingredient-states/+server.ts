import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getValidIngredientStates, searchForValidIngredientStates } from '$lib/grpc/clients';
import { logger } from '$lib/logger';

const DEFAULT_LIST_FILTER = { maxResponseSize: 100 };

export const GET: RequestHandler = async ({ url, locals }) => {
  const token = locals.oauthToken;
  if (!token) {
    return json({ error: 'Unauthorized' }, { status: 401 });
  }

  const q = url.searchParams.get('q') ?? '';
  try {
    const res =
      q === ''
        ? await getValidIngredientStates(token, { filter: DEFAULT_LIST_FILTER })
        : await searchForValidIngredientStates(token, {
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
