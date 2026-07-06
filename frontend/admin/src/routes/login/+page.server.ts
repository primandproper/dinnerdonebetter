import { fail, redirect } from '@sveltejs/kit';
import { messageOf } from '@primandproper/errors';
import type { Actions } from './$types';
import { adminLoginForToken } from '$lib/grpc/clients';
import { encodeSession, getCookieOptions } from '$lib/auth/session';

export const actions: Actions = {
  login: async ({ request, cookies }) => {
    const formData = await request.formData();
    const username = (formData.get('username') as string)?.trim() ?? '';
    const password = (formData.get('password') as string) ?? '';
    const totpToken = (formData.get('totpToken') as string)?.trim() ?? '';

    if (!username) {
      return fail(400, { error: 'Username is required', username });
    }
    if (!password) {
      return fail(400, { error: 'Password is required', username });
    }

    try {
      const response = await adminLoginForToken({
        input: {
          username,
          password,
          totpToken,
          desiredAccountId: '',
        },
      });
      const accessToken = response.result?.accessToken;
      if (!accessToken) {
        return fail(500, { error: 'No access token in response', username });
      }

      const encoded = encodeSession({ accessToken });
      const opts = getCookieOptions();
      cookies.set(opts.name, encoded, {
        path: opts.path,
        httpOnly: opts.httpOnly,
        secure: opts.secure,
        sameSite: opts.sameSite,
        maxAge: opts.maxAge,
      });
    } catch (err) {
      let message = 'Login failed';
      if (err instanceof Error) {
        message =
          err.message.includes('ECONNREFUSED') || err.message.includes('UNAVAILABLE')
            ? 'Cannot reach API server. Check GRPC_API_SERVER_URL and ensure the API is reachable (or port-forward for local dev).'
            : err.message;
      }
      const errMsg = messageOf(err);
      const totpRequired =
        errMsg.toLowerCase().includes('totp') ||
        errMsg.toLowerCase().includes('two factor') ||
        errMsg.toLowerCase().includes('2fa');
      return fail(401, { error: message, username, totpRequired });
    }

    throw redirect(302, '/');
  },
};
