import { isRedirect, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { endOtherSignIns, endSignIn, listSignIns } from '$lib/grpc/clients';

// A person's live logins on platform's SignInService, password and passkey alike, most
// recently refreshed first.
export const load: PageServerLoad = async ({ locals, url }) => {
  try {
    const signIns = await listSignIns(locals.session);
    const error = url.searchParams.get('error');
    const revoked = url.searchParams.get('revoked') === '1';
    const revokedAll = url.searchParams.get('revoked_all') === '1';
    return { signIns, error, revoked, revokedAll };
  } catch {
    return { signIns: [], error: 'server', revoked: false, revokedAll: false };
  }
};

export const actions: Actions = {
  'revoke': async ({ request, locals }) => {
    const formData = await request.formData();
    const familyId = (formData.get('family_id') as string)?.trim() ?? '';

    if (!familyId) {
      throw redirect(302, '/account/sessions?error=invalid');
    }

    try {
      // It answers the same whether or not it ended anything, so the page just re-lists.
      await endSignIn(locals.session, { familyId });
      throw redirect(302, '/account/sessions?revoked=1');
    } catch (e) {
      if (isRedirect(e)) {
        throw e;
      }
      throw redirect(302, '/account/sessions?error=revoke_failed');
    }
  },
  'revoke-all': async ({ locals }) => {
    try {
      await endOtherSignIns(locals.session);
      throw redirect(302, '/account/sessions?revoked_all=1');
    } catch (e) {
      if (isRedirect(e)) {
        throw e;
      }
      throw redirect(302, '/account/sessions?error=revoke_all_failed');
    }
  },
};
