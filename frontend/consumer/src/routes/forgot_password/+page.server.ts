import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { requestPasswordReset } from '@primandproper/platform-client';

export const load: PageServerLoad = async () => {
  return {};
};

export const actions: Actions = {
  default: async ({ request, locals }) => {
    const formData = await request.formData();
    const email = (formData.get('email') as string)?.trim() ?? '';

    if (!email) {
      return fail(400, {
        error: 'Email is required',
        success: false,
      });
    }

    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(email)) {
      return fail(400, {
        error: 'Please enter a valid email address',
        success: false,
      });
    }

    try {
      await requestPasswordReset(locals.session, email);
    } catch {
      // The same screen either way: one that differed would say which addresses hold an account.
    }
    return { success: true, error: null };
  },
};
