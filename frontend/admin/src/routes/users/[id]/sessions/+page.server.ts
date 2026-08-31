import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { adminListSessionsForUser, adminRevokeUserSession, adminRevokeAllUserSessions } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params, url }) => {
  const token = locals.accessToken;
  const userId = params.id;
  if (!token) {
    return { userId, sessions: [], error: 'Not authenticated', revoked: false, revokedAll: false };
  }

  try {
    const res = await adminListSessionsForUser(token, { userId });
    const sessions = res.sessions ?? [];
    const error = url.searchParams.get('error');
    const revoked = url.searchParams.get('revoked') === '1';
    const revokedAll = url.searchParams.get('revoked_all') === '1';
    return { userId, sessions, error, revoked, revokedAll };
  } catch {
    return { userId, sessions: [], error: 'server', revoked: false, revokedAll: false };
  }
};

export const actions: Actions = {
  'revoke': async ({ request, locals, params }) => {
    const token = locals.accessToken;
    if (!token) throw redirect(302, '/login');

    const userId = params.id;
    const formData = await request.formData();
    const sessionId = (formData.get('session_id') as string)?.trim() ?? '';
    if (!sessionId) {
      throw redirect(302, `/users/${userId}/sessions?error=invalid`);
    }

    try {
      await adminRevokeUserSession(token, { userId, sessionId });
      throw redirect(302, `/users/${userId}/sessions?revoked=1`);
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, `/users/${userId}/sessions?error=revoke_failed`);
    }
  },

  'revoke-all': async ({ locals, params }) => {
    const token = locals.accessToken;
    if (!token) throw redirect(302, '/login');

    const userId = params.id;

    try {
      await adminRevokeAllUserSessions(token, { userId });
      throw redirect(302, `/users/${userId}/sessions?revoked_all=1`);
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, `/users/${userId}/sessions?error=revoke_all_failed`);
    }
  },
};
