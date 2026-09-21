import type { PageServerLoad } from './$types';
import { listUsers, listAccounts, getRecipes } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals }) => {
  const token = locals.accessToken;
  if (!token) {
    return { userCount: '-', accountCount: '-', recipeCount: '-', error: 'Not authenticated' };
  }

  let userCount = '-';
  let accountCount = '-';
  let recipeCount = '-';

  try {
    const usersRes = (await listUsers(token, {})) as { results?: unknown[] };
    if (usersRes?.results) userCount = String(usersRes.results.length);
  } catch {
    // leave as '-'
  }
  try {
    const accountsRes = (await listAccounts(token, {})) as { results?: unknown[] };
    if (accountsRes?.results) accountCount = String(accountsRes.results.length);
  } catch {
    // leave as '-'
  }
  try {
    const recipesRes = (await getRecipes(token, { status: '' })) as { results?: unknown[] };
    if (recipesRes?.results) recipeCount = String(recipesRes.results.length);
  } catch {
    // leave as '-'
  }

  return { userCount, accountCount, recipeCount };
};
