import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { searchValidIngredientsByPreparation } from '$lib/grpc/clients';
import { logger } from '$lib/logger';

export const GET: RequestHandler = async ({ url, locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return json({ error: 'Unauthorized' }, { status: 401 });
  }

  const q = url.searchParams.get('q') ?? '';
  const preparationId = url.searchParams.get('preparationId') ?? '';
  if (!preparationId) {
    return json({ results: [] });
  }

  try {
    const res = await searchValidIngredientsByPreparation(token, {
      query: q,
      validPreparationId: preparationId,
      filter: undefined,
    });
    return json({ results: (res as { results?: unknown[] }).results ?? [] });
  } catch (e) {
    logger.error('searchValidIngredientsByPreparation failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
