import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { createRecipe } from '$lib/grpc/clients';
import { logger } from '$lib/logger';
import type { RecipeCreationRequestInput } from '@dinnerdonebetter/api-client/mealplanning/mealplanning_service_types';

export const load: PageServerLoad = async ({ locals }) => {
  const token = locals.accessToken;
  if (!token) {
    throw redirect(302, '/login');
  }
  return {};
};

export const actions: Actions = {
  default: async ({ request, locals }) => {
    const token = locals.accessToken;
    if (!token) {
      throw redirect(302, '/login');
    }

    const formData = await request.formData();
    const recipeJson = formData.get('recipe') as string | null;
    if (!recipeJson) {
      return fail(400, { error: 'Missing recipe data' });
    }

    let input: RecipeCreationRequestInput;
    try {
      input = JSON.parse(recipeJson) as RecipeCreationRequestInput;
    } catch {
      return fail(400, { error: 'Invalid recipe data' });
    }

    // Ensure required fields
    if (!input.name?.trim()) {
      return fail(400, { error: 'Recipe name is required' });
    }
    if (!input.steps?.length) {
      return fail(400, { error: 'At least one step is required' });
    }
    if (input.steps.length < 2) {
      return fail(400, { error: 'Recipe must have at least 2 steps' });
    }

    try {
      const res = await createRecipe(token, { input });
      const created = res.created;
      if (created?.id) {
        throw redirect(302, `/recipes/${created.id}`);
      }
      return fail(500, { error: 'Recipe created but no ID returned' });
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      logger.error('createRecipe failed', e);
      return fail(500, {
        error: 'Failed to create recipe. Please try again.',
      });
    }
  },
};
