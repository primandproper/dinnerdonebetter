import { randomUUID } from 'node:crypto';
import { describe, it, expect } from 'vitest';
import { clientMetadata, clientOf } from './client-info';

/** requestFrom is the part of a RequestEvent clientOf reads. */
function requestFrom(headers: Record<string, string>, socketAddress?: string) {
  return {
    request: new Request('http://localhost/', { headers }),
    getClientAddress: () => {
      if (!socketAddress) {
        throw new Error('no address');
      }
      return socketAddress;
    },
  };
}

describe('clientOf', () => {
  it('reads the address Caddy stamped, and the browser', () => {
    const stamped = `203.0.113.${Math.floor(Math.random() * 250)}`;
    const userAgent = `Mozilla/5.0 ${randomUUID()}`;

    const client = clientOf(
      requestFrom({ 'x-forwarded-for': `198.51.100.1, ${stamped}`, 'user-agent': userAgent }, '10.0.0.1'),
    );

    expect(client).toEqual({ address: stamped, userAgent });
  });

  it('falls back to the connection', () => {
    const socket = `10.0.0.${Math.floor(Math.random() * 250)}`;

    expect(clientOf(requestFrom({}, socket)).address).toBe(socket);
  });

  it('survives a request with neither', () => {
    expect(clientOf(requestFrom({}))).toEqual({ address: undefined, userAgent: undefined });
  });
});

describe('clientMetadata', () => {
  it('forwards what it knows', () => {
    const address = `203.0.113.${Math.floor(Math.random() * 250)}`;
    const userAgent = `Mozilla/5.0 ${randomUUID()}`;

    expect(clientMetadata({ address, userAgent })).toEqual({
      'x-client-address': address,
      'x-client-user-agent': userAgent,
    });
  });

  it('forwards nothing it does not', () => {
    expect(clientMetadata({})).toEqual({});
  });
});
