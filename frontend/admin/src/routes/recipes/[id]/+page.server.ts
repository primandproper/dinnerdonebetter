import type { PageServerLoad } from './$types';
import { getRecipe } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const session = locals.session;
  const id = params.id;
  try {
    const res = (await getRecipe(session, { recipeId: id })) as { result?: Record<string, unknown> };
    return { recipe: res?.result ?? null };
  } catch (e) {
    return {
      recipe: null,
      error: e instanceof Error ? e.message : 'Failed to load recipe',
    };
  }
};
