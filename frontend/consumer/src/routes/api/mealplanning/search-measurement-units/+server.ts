import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import {
  getValidIngredientMeasurementUnitsByIngredient,
  getValidMeasurementUnits,
  searchForValidMeasurementUnits,
} from '$lib/grpc/clients';
import { logger } from '$lib/logger';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const GET: RequestHandler = async ({ url, locals }) => {
  const session = locals.session;
  const q = url.searchParams.get('q') ?? '';
  const ingredientId = url.searchParams.get('ingredientId') ?? '';

  try {
    if (ingredientId) {
      // Use getValidIngredientMeasurementUnitsByIngredient for bridge IDs
      const res = await getValidIngredientMeasurementUnitsByIngredient(session, {
        validIngredientId: ingredientId,
        filter: QueryFilter.create({ maxResponseSize: 50 }),
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
        ? await getValidMeasurementUnits(session, { filter: DEFAULT_LIST_FILTER })
        : await searchForValidMeasurementUnits(session, {
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
