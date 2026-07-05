import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import {
  getValidIngredientMeasurementUnitsByIngredient,
  getValidMeasurementUnits,
  searchForValidMeasurementUnits,
} from '$lib/grpc/clients';
import { logger } from '$lib/logger';

const DEFAULT_LIST_FILTER = { maxResponseSize: 100 };

export const GET: RequestHandler = async ({ url, locals }) => {
  const token = locals.oauthToken;
  if (!token) {
    return json({ error: 'Unauthorized' }, { status: 401 });
  }

  const q = url.searchParams.get('q') ?? '';
  const ingredientId = url.searchParams.get('ingredientId') ?? '';

  try {
    if (ingredientId) {
      // Use getValidIngredientMeasurementUnitsByIngredient for bridge IDs
      const res = await getValidIngredientMeasurementUnitsByIngredient(token, {
        validIngredientId: ingredientId,
        filter: { maxResponseSize: 50 },
      });
      const vimus = res.results ?? [];
      const filtered =
        q.length > 0
          ? vimus.filter((vimu) => vimu.measurementUnit?.name.toLowerCase().includes(q.toLowerCase()))
          : vimus;
      return json({ results: filtered });
    }
    const res =
      q === ''
        ? await getValidMeasurementUnits(token, { filter: DEFAULT_LIST_FILTER })
        : await searchForValidMeasurementUnits(token, {
            filter: DEFAULT_LIST_FILTER,
            query: q,
            useSearchService: q.length > 2,
          });
    return json({ results: res.results ?? [] });
  } catch (e) {
    logger.error('measurement unit search failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
