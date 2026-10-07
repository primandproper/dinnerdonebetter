import type { Handle } from '@sveltejs/kit';
import { resolveOrRedirect } from '@primandproper/platform-client';
import { clientOf, sessionFor } from '$lib/auth/session';

const LOGIN_PATH = '/login';

const PUBLIC_PATHS = [LOGIN_PATH, '/logout', '/_ops_', '/auth/passkey/authentication'];

function isPublicPath(pathname: string): boolean {
  return PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`));
}

// Public pages get a Session too: signing in and signing out need it. Only whether a login
// is held is settled here. Whether it still works is settled by the first call a page makes,
// which refreshes it if it has to; a page load that ended the login while it ran goes to
// sign in, keeping the cookies it cleared.
export const handle: Handle = ({ event, resolve }) => {
  event.locals.session = sessionFor(event.cookies, clientOf(event));
  return resolveOrRedirect(event.locals.session, event.request, async () => resolve(event), {
    isPublic: isPublicPath,
    loginPath: LOGIN_PATH,
  });
};
