import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getValidIngredientPreparationsByPreparation } from '$lib/grpc/clients';
import { logger } from '$lib/logger';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const GET: RequestHandler = async ({ url, locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return json({ error: 'Unauthorized' }, { status: 401 });
  }

  const preparationId = url.searchParams.get('preparationId') ?? '';
  if (!preparationId) {
    return json({ results: [] });
  }

  try {
    const res = await getValidIngredientPreparationsByPreparation(token, {
      validPreparationId: preparationId,
      filter: QueryFilter.create({ maxResponseSize: 50 }),
    });
    return json({ results: (res as { results?: unknown[] }).results ?? [] });
  } catch (e) {
    logger.error('getValidIngredientPreparationsByPreparation failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
