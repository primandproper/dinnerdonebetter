import type { Handle } from '@sveltejs/kit';
import { resolveOrRedirect } from '@primandproper/platform-client';
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

export const handle: Handle = async ({ event, resolve }) => {
  const timing = new ServerTiming();
  const totalEvent = timing.addEvent('total', 'Total request time');

  // Public pages get one too: signing in, signing out and the passkey doors all need it.
  event.locals.session = sessionFor(event.cookies);

  // Only whether a login is held is settled here. Whether it still works is settled by the
  // first call a page makes, which refreshes it if it has to; a page load that ended the
  // login while it ran goes to sign in, keeping the cookies it cleared.
  const response = await resolveOrRedirect(event.locals.session, event.request, async () => resolve(event), {
    isPublic: isPublicPath,
    loginPath: LOGIN_PATH,
  });
  totalEvent.end();
  response.headers.set(ServerTimingHeaderName, timing.headerValue());
  recordRequest(event.url.pathname, response.status, totalEvent.duration);
  return response;
};
