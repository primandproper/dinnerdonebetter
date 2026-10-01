import type { PageServerLoad } from './$types';
import { getValidIngredient } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const session = locals.session;
  const id = params.id;
  try {
    const res = (await getValidIngredient(session, { validIngredientId: id })) as {
      result?: Record<string, unknown>;
    };
    return { item: res?.result ?? null };
  } catch (e) {
    return {
      item: null,
      error: e instanceof Error ? e.message : 'Failed to load ingredient',
    };
  }
};
