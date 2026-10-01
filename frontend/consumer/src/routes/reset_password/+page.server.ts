import { fail, redirect } from '@sveltejs/kit';
import {
  completePasswordReset,
  PasswordResetReason,
  PlatformError,
  verifyPasswordResetToken,
} from '@primandproper/platform-client';
import type { Actions, PageServerLoad } from './$types';

const deadLinkMessage = 'This reset link is invalid or has expired. Please request a new one.';

export const load: PageServerLoad = async ({ url, locals }) => {
  const token = url.searchParams.get('t') ?? '';
  if (!token) {
    return { token, missingToken: true, deadLink: false };
  }
  // The link is checked before the form is shown, so nobody types a new password into a form
  // that was never going to take it.
  try {
    const result = await verifyPasswordResetToken(locals.session, token);
    return { token, missingToken: false, deadLink: result.kind === 'dead_link' };
  } catch {
    // Unreachable rather than refused: show the form, and let the submission say.
    return { token, missingToken: false, deadLink: false };
  }
};

export const actions: Actions = {
  default: async ({ request, locals }) => {
    const formData = await request.formData();
    const token = (formData.get('token') as string)?.trim() ?? '';
    const newPassword = (formData.get('new_password') as string) ?? '';
    const confirmPassword = (formData.get('confirm_password') as string) ?? '';

    if (!token) {
      return fail(400, {
        error: 'Missing reset token. Please use the link from your email.',
        token: '',
      });
    }

    if (!newPassword || newPassword.length < 8) {
      return fail(400, {
        error: 'Password must be at least 8 characters',
        token,
      });
    }

    if (newPassword !== confirmPassword) {
      return fail(400, {
        error: 'Passwords do not match',
        token,
      });
    }

    let result;
    try {
      result = await completePasswordReset(locals.session, token, newPassword);
    } catch (err) {
      if (err instanceof PlatformError && err.is(PasswordResetReason.REPLACEMENT_PASSWORD_REFUSED)) {
        return fail(400, { error: err.serverMessage || 'Choose a different password.', token });
      }
      return fail(500, { error: 'Something went wrong. Please try again.', token });
    }
    if (result.kind === 'dead_link') {
      return fail(400, { error: deadLinkMessage, token });
    }
    throw redirect(302, '/login?reset=success');
  },
};
