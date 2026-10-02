import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { searchForRecipes } from '$lib/grpc/clients';
import { logger } from '$lib/logger';

export const GET: RequestHandler = async ({ url, locals }) => {
  const session = locals.session;
  const q = url.searchParams.get('q') ?? '';
  try {
    const res = await searchForRecipes(session, {
      query: q,
      useSearchService: q.length > 2,
    });
    const results = (res.results ?? []).map((r) => ({
      id: r.id,
      name: r.name ?? '',
      slug: r.slug ?? '',
    }));
    return json({ results });
  } catch (e) {
    logger.error('searchForRecipes failed', e);
    return json({ error: 'Search failed' }, { status: 500 });
  }
};
