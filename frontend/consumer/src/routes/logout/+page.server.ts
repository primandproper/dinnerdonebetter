import { redirect } from '@sveltejs/kit';
import { signOut } from '@primandproper/platform-client';
import type { PageServerLoad } from './$types';
import { revokeCurrentSession } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals }) => {
  const held = await locals.session.held().catch(() => undefined);
  // A login with no refresh token came through a passkey, which platform has no door for yet
  // (platform-go#874): it is one of AuthService's sessions, and ending it is AuthService's.
  if (held && !held.refreshToken) {
    await revokeCurrentSession(locals.session).catch(() => {
      /* best-effort: the cookie is cleared either way */
    });
  }
  // signOut never rejects, and clears the cookie whether or not the server heard it.
  await signOut(locals.session);
  throw redirect(302, '/login');
};
