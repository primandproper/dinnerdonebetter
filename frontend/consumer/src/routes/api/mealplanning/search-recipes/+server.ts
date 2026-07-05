import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { searchForRecipes } from '$lib/grpc/clients';
import { logger } from '$lib/logger';

export const GET: RequestHandler = async ({ url, locals }) => {
  const token = locals.oauthToken;
  if (!token) {
    return json({ error: 'Unauthorized' }, { status: 401 });
  }

  const q = url.searchParams.get('q') ?? '';
  try {
    const res = await searchForRecipes(token, {
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
