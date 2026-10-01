import { isRedirect, redirect } from '@sveltejs/kit';
import { PasskeyReason, PlatformError } from '@primandproper/platform-client';
import type { Actions, PageServerLoad } from './$types';
import { listPasskeys, archivePasskey } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, url }) => {
  const session = locals.session;
  try {
    const passkeys = await listPasskeys(session);
    const error = url.searchParams.get('error');
    const deleted = url.searchParams.get('deleted') === '1';
    return { passkeys, error, deleted };
  } catch {
    return { passkeys: [], error: 'server', deleted: false };
  }
};

export const actions: Actions = {
  delete: async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const credentialId = (formData.get('credential_id') as string)?.trim() ?? '';
    if (!credentialId) {
      throw redirect(302, '/account/passkeys?error=invalid');
    }

    try {
      await archivePasskey(session, { id: credentialId });
      throw redirect(302, '/account/passkeys?deleted=1');
    } catch (e) {
      if (isRedirect(e)) {
        throw e;
      }
      if (e instanceof PlatformError && e.is(PasskeyReason.LAST_PASSKEY)) {
        throw redirect(302, '/account/passkeys?error=last_passkey');
      }
      throw redirect(302, '/account/passkeys?error=delete_failed');
    }
  },
};
