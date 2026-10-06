import { fail, redirect } from '@sveltejs/kit';
import { adminSignIn, isTransient, PlatformError, SignInReason } from '@primandproper/platform-client';
import { landingAfterSignIn } from '$lib/auth/required-actions';
import type { Actions } from './$types';

export const actions: Actions = {
  login: async ({ request, locals }) => {
    const formData = await request.formData();
    const username = (formData.get('username') as string)?.trim() ?? '';
    const password = (formData.get('password') as string) ?? '';
    const totpCode = (formData.get('totpToken') as string)?.trim() ?? '';

    if (!username) {
      return fail(400, { error: 'Username is required', username });
    }
    if (!password) {
      return fail(400, { error: 'Password is required', username });
    }

    try {
      // The administrative door: it refuses anybody who isn't an operator, before any token is
      // issued. The code is sent whenever one was typed, and ignored for anyone without one.
      const result = await adminSignIn(locals.session, { handle: { username }, password, totpCode });
      if (result.kind === 'second_factor_required') {
        // The form sends the password again with the code, so the resend isn't needed.
        return fail(401, { error: 'Enter the code from your authenticator app', username, totpRequired: true });
      }
    } catch (err) {
      return fail(401, { error: refusal(err), username, totpRequired: !!totpCode });
    }

    // Somebody an operator has made change their password is signed in all the same, and
    // goes to the form for it before anything else.
    throw redirect(302, await landingAfterSignIn(locals.session));
  },
};

/** refusal is what to tell the operator about a sign-in that didn't go through. */
function refusal(err: unknown): string {
  if (isTransient(err)) {
    return 'Cannot reach API server. Check GRPC_API_SERVER_URL and ensure the API is reachable (or port-forward for local dev).';
  }
  if (!(err instanceof PlatformError)) {
    return 'Sign-in failed';
  }
  if (err.is(SignInReason.INVALID_CREDENTIALS)) {
    return 'Invalid username, password or code';
  }
  if (err.is(SignInReason.NOT_AN_ADMINISTRATOR)) {
    return 'That account is not an administrator';
  }
  if (err.is(SignInReason.ADMIN_SIGNIN_UNAVAILABLE)) {
    return 'Administrative sign-in is turned off on this server';
  }
  return err.serverMessage || 'Login failed';
}
