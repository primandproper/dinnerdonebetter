import type { PageServerLoad } from './$types';
import { searchForRecipes } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const res = (await searchForRecipes(session, {
      filter: QueryFilter.create({ maxResponseSize: 100 }),
    })) as { results?: Array<{ id?: string; name?: string }> };
    return { recipes: res?.results ?? [] };
  } catch (e) {
    return {
      recipes: [],
      error: e instanceof Error ? e.message : 'Failed to load recipes',
    };
  }
};
