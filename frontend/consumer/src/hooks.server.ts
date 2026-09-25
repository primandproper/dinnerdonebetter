import { redirect } from '@sveltejs/kit';
import type { Handle, RequestEvent } from '@sveltejs/kit';
import { sessionFor } from '$lib/auth/session';
import { initServerOtel } from '$lib/otel/server';
import { recordRequest } from '$lib/otel/server-metrics';
import { ServerTiming, ServerTimingHeaderName } from '$lib/server-timing';

initServerOtel();

const LOGIN_PATH = '/login';

const PUBLIC_PATHS = [
  LOGIN_PATH,
  '/logout',
  '/forgot_password',
  '/reset_password',
  '/verify_email_address',
  '/terms-of-service',
  '/privacy-policy',
  '/accept_invitation',
  '/meal_plans',
  '/_ops_',
  '/.well-known',
  '/auth/passkey/authentication',
];

function isPublicPath(pathname: string): boolean {
  return PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`));
}

/**
 * loginEndedDuring reports whether a page load found the login over: a refresh the server
 * refused, which has already cleared the cookie. Most loads catch their own errors, so the
 * redirect the failed call threw doesn't always reach the response. A form action, a
 * client-side navigation or an API endpoint answers as it chose to, and the next page load
 * goes to sign-in.
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
  const timing = new ServerTiming();
  const totalEvent = timing.addEvent('total', 'Total request time');

  // Public pages get one too: signing in, signing out and the passkey doors all need it.
  event.locals.session = sessionFor(event.cookies);

  if (isPublicPath(event.url.pathname)) {
    const response = await resolve(event);
    totalEvent.end();
    response.headers.set(ServerTimingHeaderName, timing.headerValue());
    recordRequest(event.url.pathname, response.status, totalEvent.duration);
    return response;
  }

  // Only whether a login is held is settled here. Whether it still works is settled by the
  // first call a page makes, which refreshes it if it has to.
  const authEvent = timing.addEvent('auth', 'Session load');
  const held = await event.locals.session.held();
  authEvent.end();
  if (!held) {
    throw redirect(302, LOGIN_PATH);
  }

  let response = await resolve(event);
  if (loginEndedDuring(event)) {
    response = toLogin(response);
  }
  totalEvent.end();
  response.headers.set(ServerTimingHeaderName, timing.headerValue());
  recordRequest(event.url.pathname, response.status, totalEvent.duration);
  return response;
};
