import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { listActiveSessions, revokeSession, revokeAllOtherSessions } from '$lib/grpc/clients';

// These are AuthService's sessions: the ones passkey sign-ins open. A password sign-in is a
// login on platform's SignInService, which can't list a person's logins yet
// (platform-go#886), so it doesn't appear here, and nothing here is marked current for it.
export const load: PageServerLoad = async ({ locals, url }) => {
  try {
    const res = await listActiveSessions(locals.session);
    const sessions = res.sessions ?? [];
    const error = url.searchParams.get('error');
    const revoked = url.searchParams.get('revoked') === '1';
    const revokedAll = url.searchParams.get('revoked_all') === '1';
    return { sessions, error, revoked, revokedAll };
  } catch {
    return { sessions: [], error: 'server', revoked: false, revokedAll: false };
  }
};

export const actions: Actions = {
  'revoke': async ({ request, locals }) => {
    const formData = await request.formData();
    const sessionId = (formData.get('session_id') as string)?.trim() ?? '';
    if (!sessionId) {
      throw redirect(302, '/account/sessions?error=invalid');
    }

    try {
      await revokeSession(locals.session, { sessionId });
      throw redirect(302, '/account/sessions?revoked=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/sessions?error=revoke_failed');
    }
  },

  'revoke-all': async ({ locals }) => {
    try {
      await revokeAllOtherSessions(locals.session);
      throw redirect(302, '/account/sessions?revoked_all=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/sessions?error=revoke_all_failed');
    }
  },
};
