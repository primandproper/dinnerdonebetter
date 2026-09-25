import { fail, redirect } from '@sveltejs/kit';
import { PlatformError, SignInReason, signIn } from '@primandproper/platform-client';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
  const resetSuccess = url.searchParams.get('reset') === 'success';
  return { resetSuccess };
};

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
      // The code is sent whenever the person typed one: it is ignored for anyone without a
      // second factor, so the form needs no branch to know which they are.
      const result = await signIn(locals.session, { handle: { username }, password, totpCode });
      if (result.kind === 'second_factor_required') {
        // The form sends the password again with the code, so the resend isn't needed.
        return fail(401, { error: 'Enter the code from your authenticator app', username, totpRequired: true });
      }
    } catch (err) {
      return fail(401, { error: refusal(err), username, totpRequired: !!totpCode });
    }

    throw redirect(302, '/');
  },
};

/** refusal is what to tell the person about a sign-in that didn't go through. */
function refusal(err: unknown): string {
  if (!(err instanceof PlatformError)) {
    return 'Cannot reach the API server. Check GRPC_API_SERVER_URL and that it is reachable.';
  }
  if (err.is(SignInReason.INVALID_CREDENTIALS)) {
    return 'Invalid username, password or code';
  }
  return err.serverMessage || 'Sign-in failed';
}
