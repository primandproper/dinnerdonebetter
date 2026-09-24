import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { getSelf, resolveSettings, setSettingValue } from '$lib/grpc/clients';
import { configurableSettings, settingsSubject, typedValue } from '$lib/settings/resolutions';

/** The signed-in person's user id, which every settings call names as its subject. */
async function selfId(token: string): Promise<string> {
  const self = await getSelf(token);
  const id = self.result?.id;
  if (!id) throw new Error('no user for this session');
  return id;
}

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.oauthToken;
  if (!token) {
    return { configurableSettings: [], error: null, updated: false };
  }

  try {
    // One call. The catalog and the user's answers used to be two requests
    // joined here, with the fallback to a setting's default reimplemented
    // alongside; the server does both now. It also decides which settings this
    // user may see, so admin-only ones are absent rather than filtered out here.
    const resolved = await resolveSettings(token, { subject: settingsSubject(await selfId(token)) });

    const error = url.searchParams.get('error');
    const updated = url.searchParams.get('updated') === '1';

    return { configurableSettings: configurableSettings(resolved.resolutions ?? []), error, updated };
  } catch {
    return {
      configurableSettings: [],
      error: 'server',
      updated: false,
    };
  }
};

const _errorMessages: Record<string, string> = {
  invalid: 'Invalid input. Please try again.',
  update_failed: 'Failed to save preference. Please try again.',
  server: 'Something went wrong. Please try again.',
};

export const actions: Actions = {
  update: async ({ request, locals }) => {
    const token = locals.oauthToken;
    if (!token) throw redirect(302, '/login');

    const formData = await request.formData();
    const settingName = (formData.get('setting_name') as string)?.trim() ?? '';
    const raw = (formData.get('value') as string)?.trim() ?? '';
    const value = typedValue(Number(formData.get('setting_kind')), raw);

    if (!settingName || !value) {
      throw redirect(302, '/account/preferences?error=invalid');
    }

    try {
      // One call whether or not they had answered before: the server converges on
      // the row, so a first answer and a changed one are the same write and this
      // no longer has to know which it is making.
      await setSettingValue(token, { subject: settingsSubject(await selfId(token)), name: settingName, value });
      throw redirect(302, '/account/preferences?updated=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/preferences?error=update_failed');
    }
  },
};
