import { isRedirect, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { endAllSignInsForUser, endSignInForUser, listSignInsForUser } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params, url }) => {
  const session = locals.session;
  const userId = params.id;

  try {
    const { signIns } = await listSignInsForUser(session, { userId, limit: 0 });
    const error = url.searchParams.get('error');
    const revoked = url.searchParams.get('revoked') === '1';
    const revokedAll = url.searchParams.get('revoked_all') === '1';
    return { userId, signIns, error, revoked, revokedAll };
  } catch {
    return { userId, signIns: [], error: 'server', revoked: false, revokedAll: false };
  }
};

export const actions: Actions = {
  'revoke': async ({ request, locals, params }) => {
    const session = locals.session;

    const userId = params.id;
    const formData = await request.formData();
    const familyId = (formData.get('family_id') as string)?.trim() ?? '';
    if (!familyId) {
      throw redirect(302, `/users/${userId}/sessions?error=invalid`);
    }

    try {
      await endSignInForUser(session, { userId, familyId });
      throw redirect(302, `/users/${userId}/sessions?revoked=1`);
    } catch (e) {
      if (isRedirect(e)) {
        throw e;
      }
      throw redirect(302, `/users/${userId}/sessions?error=revoke_failed`);
    }
  },

  'revoke-all': async ({ locals, params }) => {
    const session = locals.session;

    const userId = params.id;

    try {
      await endAllSignInsForUser(session, { userId });
      throw redirect(302, `/users/${userId}/sessions?revoked_all=1`);
    } catch (e) {
      if (isRedirect(e)) {
        throw e;
      }
      throw redirect(302, `/users/${userId}/sessions?error=revoke_all_failed`);
    }
  },
};
