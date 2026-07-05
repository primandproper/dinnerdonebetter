import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getValidPreparationInstrumentsByPreparation } from '$lib/grpc/clients';
import { logger } from '$lib/logger';

export const GET: RequestHandler = async ({ url, locals }) => {
  const token = locals.oauthToken;
  if (!token) {
    return json({ error: 'Unauthorized' }, { status: 401 });
  }

  const preparationId = url.searchParams.get('preparationId') ?? '';
  if (!preparationId) {
    return json({ results: [] });
  }

  try {
    const res = await getValidPreparationInstrumentsByPreparation(token, {
      validPreparationId: preparationId,
      filter: undefined,
    });
    return json({ results: res.results ?? [] });
  } catch (e) {
    logger.error('getValidPreparationInstrumentsByPreparation failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
