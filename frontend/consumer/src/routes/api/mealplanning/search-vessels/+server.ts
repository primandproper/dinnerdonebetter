import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getValidPreparationVesselsByPreparation, getValidVessels, searchForValidVessels } from '$lib/grpc/clients';
import { logger } from '$lib/logger';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

export const GET: RequestHandler = async ({ url, locals }) => {
  const session = locals.session;
  const q = url.searchParams.get('q') ?? '';
  const preparationId = url.searchParams.get('preparationId') ?? '';

  try {
    if (preparationId) {
      const res = await getValidPreparationVesselsByPreparation(session, {
        validPreparationId: preparationId,
        filter: undefined,
      });
      const vpvs = res.results ?? [];
      const filtered =
        q.length > 0 ? vpvs.filter((vpv) => vpv.vessel?.name.toLowerCase().includes(q.toLowerCase())) : vpvs;
      return json({ results: filtered });
    }
    const res =
      q === ''
        ? await getValidVessels(session, { filter: DEFAULT_LIST_FILTER })
        : await searchForValidVessels(session, {
            filter: DEFAULT_LIST_FILTER,
            query: q,
            useSearchService: q.length > 2,
          });
    return json({ results: res.results ?? [] });
  } catch (e) {
    logger.error('vessel search failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
