import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { getSelf, updateProfile, updateUserUsername } from '$lib/grpc/clients';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ locals, url }) => {
  const session = locals.session;
  try {
    const selfRes = await getSelf(session);
    const user = selfRes.result ?? null;
    const error = url.searchParams.get('error');
    const updated = url.searchParams.get('updated') === '1';
    const avatarMediaBaseUrl = env.PUBLIC_AVATAR_MEDIA_URL_PREFIX ?? '';
    return { user, error, updated, avatarMediaBaseUrl };
  } catch {
    return { user: null, error: 'server', updated: false, avatarMediaBaseUrl: '' };
  }
};

export const actions: Actions = {
  'update-username': async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const username = (formData.get('username') as string)?.trim() ?? '';
    const currentPassword = (formData.get('current_password') as string)?.trim() ?? '';
    const totpToken = (formData.get('totp_token') as string)?.trim() ?? '';

    if (!username) {
      throw redirect(302, '/account/profile?error=invalid_username');
    }
    if (!currentPassword) {
      throw redirect(302, '/account/profile?error=invalid_password');
    }

    try {
      await updateUserUsername(session, { newUsername: username, currentPassword, totpToken });
      throw redirect(302, '/account/profile?updated=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/profile?error=update_failed');
    }
  },
  'update-details': async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const firstName = (formData.get('first_name') as string)?.trim() ?? '';
    const lastName = (formData.get('last_name') as string)?.trim() ?? '';
    if (!firstName) {
      throw redirect(302, '/account/profile?error=invalid_first_name');
    }

    try {
      // One update where there were three, and each field absent unless it is being set —
      // so sending a first name no longer blanks a last one.
      //
      // No password or second factor is asked for. A name is not a credential, and the
      // directory's profile update asks for nothing; the two changes that *are* credential
      // changes — the handle and the address — are on the auth surface and are
      // re-authenticated there.
      await updateProfile(session, {
        input: {
          firstName,
          lastName,
        },
      });
      throw redirect(302, '/account/profile?updated=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/profile?error=update_failed');
    }
  },
  // 'update-avatar' is gone with the RPC it called. An avatar is a row in the upload
  // registry and a reference to it rather than a field of the directory's user, so
  // restoring it means an RPC on the media surface — see the api-client package.
};
