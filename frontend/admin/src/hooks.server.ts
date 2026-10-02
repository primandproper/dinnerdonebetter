import { redirect } from '@sveltejs/kit';
import type { Handle, RequestEvent } from '@sveltejs/kit';
import { sessionFor } from '$lib/auth/session';

const LOGIN_PATH = '/login';

const PUBLIC_PATHS = [LOGIN_PATH, '/logout', '/_ops_', '/auth/passkey/authentication'];

function isPublicPath(pathname: string): boolean {
  return PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`));
}

/**
 * loginEndedDuring reports whether a page load found the login over: a refresh the server
 * refused, which has already cleared the cookie. A load that caught the failed call's
 * redirect still sends the operator to sign in.
 */
function loginEndedDuring(event: RequestEvent): boolean {
  return (
    event.request.method === 'GET' &&
    !event.isDataRequest &&
    (event.request.headers.get('accept') ?? '').includes('text/html') &&
    event.locals.session.state === 'anonymous'
  );
}

function toLogin(resolved: Response): Response {
  const headers = new Headers({ location: LOGIN_PATH });
  for (const cookie of resolved.headers.getSetCookie()) {
    headers.append('set-cookie', cookie);
  }
  return new Response(null, { status: 302, headers });
}

export const handle: Handle = async ({ event, resolve }) => {
  // Public pages get one too: signing in and signing out need it.
  event.locals.session = sessionFor(event.cookies);

  if (isPublicPath(event.url.pathname)) {
    return resolve(event);
  }

  // Only whether a login is held is settled here. Whether it still works is settled by the
  // first call a page makes, which refreshes it if it has to.
  if (!(await event.locals.session.held())) {
    throw redirect(302, LOGIN_PATH);
  }

  const response = await resolve(event);
  return loginEndedDuring(event) ? toLogin(response) : response;
};
