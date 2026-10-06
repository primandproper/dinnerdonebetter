import { fail, isRedirect, redirect } from '@sveltejs/kit';
import { getAuthStatus, PlatformError, SignInReason } from '@primandproper/platform-client';
import type { Actions, PageServerLoad } from './$types';
import { updatePassword } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals }) => {
  // GetAuthStatus is one of the calls a forced change leaves open, so this reads whether the
  // change is owed, which the form says, rather than being refused for owing it.
  try {
    const status = await getAuthStatus(locals.session);
    return { required: status.authenticated && status.requiredActions.includes('change_password') };
  } catch {
    return { required: false };
  }
};

export const actions: Actions = {
  default: async ({ request, locals }) => {
    const formData = await request.formData();
    const currentPassword = (formData.get('current_password') as string) ?? '';
    const newPassword = (formData.get('new_password') as string) ?? '';
    const confirmPassword = (formData.get('confirm_password') as string) ?? '';
    const totpCode = (formData.get('totp_code') as string)?.trim() ?? '';

    if (!currentPassword || !newPassword) {
      return fail(400, { error: 'Enter your current password and a new one.' });
    }
    if (newPassword !== confirmPassword) {
      return fail(400, { error: 'The new passwords do not match.' });
    }

    try {
      await updatePassword(locals.session, { currentPassword, newPassword, totpCode });
    } catch (err) {
      if (isRedirect(err)) {
        throw err;
      }
      return fail(400, { error: refusal(err) });
    }

    throw redirect(302, '/');
  },
};

/** refusal is what to tell the person about a change that didn't go through. */
function refusal(err: unknown): string {
  if (err instanceof PlatformError) {
    if (err.is(SignInReason.INVALID_CREDENTIALS)) {
      return 'That password or code was not right.';
    }
    if (err.is(SignInReason.SECOND_FACTOR_REQUIRED)) {
      return 'Enter the code from your authenticator app as well.';
    }
    if (err.is(SignInReason.PASSWORD_REFUSED)) {
      return err.serverMessage || 'Choose a different password.';
    }
  }
  return 'Your password was not changed. Please try again.';
}
