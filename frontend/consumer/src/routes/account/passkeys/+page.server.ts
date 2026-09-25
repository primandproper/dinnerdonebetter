import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { listPasskeys, archivePasskey } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, url }) => {
  const session = locals.session;
  try {
    const res = await listPasskeys(session);
    const passkeys = res.results ?? [];
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
      await archivePasskey(session, { credentialId });
      throw redirect(302, '/account/passkeys?deleted=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/passkeys?error=delete_failed');
    }
  },
};
