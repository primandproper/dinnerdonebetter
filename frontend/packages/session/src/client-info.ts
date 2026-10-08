import type { RequestEvent } from '@sveltejs/kit';

/** ClientInfo is the browser a request came from. */
export interface ClientInfo {
  address?: string;
  userAgent?: string;
}

/**
 * clientOf is the browser behind a request. The address is the last X-Forwarded-For entry,
 * which Caddy writes and a browser cannot, and the connection's own address when there is none.
 */
export function clientOf(event: Pick<RequestEvent, 'request' | 'getClientAddress'>): ClientInfo {
  const forwarded = (event.request.headers.get('x-forwarded-for') ?? '')
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean);

  let address = forwarded.at(-1);
  if (!address) {
    try {
      address = event.getClientAddress();
    } catch {
      address = undefined;
    }
  }

  return { address, userAgent: event.request.headers.get('user-agent') ?? undefined };
}

/**
 * clientMetadata forwards the browser to the API, which records it beside every login a
 * Session signs in or renews, so "where you're signed in" names the browser rather than the
 * app's server. It is shown to the person whose login it is and decides nothing.
 */
export function clientMetadata(client: ClientInfo): Record<string, string> {
  const metadata: Record<string, string> = {};
  if (client.address) {
    metadata['x-client-address'] = client.address;
  }
  if (client.userAgent) {
    metadata['x-client-user-agent'] = client.userAgent;
  }
  return metadata;
}
