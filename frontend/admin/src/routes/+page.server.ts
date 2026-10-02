import type { PageServerLoad } from './$types';
import { listUsers, listAccounts, getRecipes } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;

  let userCount = '-';
  let accountCount = '-';
  let recipeCount = '-';

  try {
    const usersRes = (await listUsers(session, { filter: undefined })) as { results?: unknown[] };
    if (usersRes?.results) userCount = String(usersRes.results.length);
  } catch {
    // leave as '-'
  }
  try {
    const accountsRes = (await listAccounts(session, { filter: undefined })) as { results?: unknown[] };
    if (accountsRes?.results) accountCount = String(accountsRes.results.length);
  } catch {
    // leave as '-'
  }
  try {
    const recipesRes = (await getRecipes(session, { status: '' })) as { results?: unknown[] };
    if (recipesRes?.results) recipeCount = String(recipesRes.results.length);
  } catch {
    // leave as '-'
  }

  return { userCount, accountCount, recipeCount };
};
