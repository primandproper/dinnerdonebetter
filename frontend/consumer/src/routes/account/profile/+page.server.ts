import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { getSelf, updateUserUsername, updateUserDetails, uploadUserAvatar } from '$lib/grpc/clients';
import { logger } from '$lib/logger';
import { env } from '$env/dynamic/private';

const MAX_AVATAR_SIZE_BYTES = 5 * 1024 * 1024; // 5 MB
const ALLOWED_AVATAR_TYPES = ['image/png', 'image/jpeg', 'image/gif'];

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.oauthToken;
  if (!token) {
    return { user: null, error: null, updated: false, avatarMediaBaseUrl: '' };
  }

  try {
    const selfRes = await getSelf(token);
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
    const token = locals.oauthToken;
    if (!token) {
      throw redirect(302, '/login');
    }

    const formData = await request.formData();
    const username = (formData.get('username') as string)?.trim() ?? '';

    if (!username) {
      throw redirect(302, '/account/profile?error=invalid_username');
    }

    try {
      await updateUserUsername(token, { newUsername: username });
      throw redirect(302, '/account/profile?updated=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/profile?error=update_failed');
    }
  },
  'update-details': async ({ request, locals }) => {
    const token = locals.oauthToken;
    if (!token) {
      throw redirect(302, '/login');
    }

    const formData = await request.formData();
    const firstName = (formData.get('first_name') as string)?.trim() ?? '';
    const lastName = (formData.get('last_name') as string)?.trim() ?? '';
    const currentPassword = (formData.get('current_password') as string)?.trim() ?? '';
    const totpToken = (formData.get('totp_token') as string)?.trim() ?? '';
    const birthdayStr = (formData.get('birthday') as string)?.trim() ?? '';

    if (!firstName) {
      throw redirect(302, '/account/profile?error=invalid_first_name');
    }
    if (!currentPassword) {
      throw redirect(302, '/account/profile?error=invalid_password');
    }

    let birthday: Date | undefined;
    if (birthdayStr) {
      const parsed = new Date(birthdayStr);
      if (!isNaN(parsed.getTime())) {
        birthday = parsed;
      }
    }

    try {
      await updateUserDetails(token, {
        input: {
          firstName,
          lastName,
          birthday,
          currentPassword,
          totpToken,
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
  'update-avatar': async ({ request, locals }) => {
    const token = locals.oauthToken;
    if (!token) {
      throw redirect(302, '/login');
    }

    const formData = await request.formData();
    const file = formData.get('avatar') as File | null;
    if (!file || !(file instanceof File) || file.size === 0) {
      throw redirect(302, '/account/profile?error=avatar_upload_failed');
    }

    if (!ALLOWED_AVATAR_TYPES.includes(file.type)) {
      throw redirect(302, '/account/profile?error=avatar_upload_failed');
    }
    if (file.size > MAX_AVATAR_SIZE_BYTES) {
      throw redirect(302, '/account/profile?error=avatar_upload_failed');
    }

    const arrayBuffer = await file.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);
    const filename = file.name || 'avatar';

    try {
      const apiResponse = await uploadUserAvatar(token, buffer, filename, file.type);
      const storagePath =
        apiResponse?.created?.storagePath ??
        (apiResponse?.created as { storage_path?: string } | undefined)?.storage_path;
      logger.with({ storagePath: storagePath ?? null }).debug('avatar upload returning');
      if (storagePath) {
        return { updated: true, avatarStoragePath: storagePath };
      }
      return { updated: true };
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/profile?error=avatar_upload_failed');
    }
  },
};
