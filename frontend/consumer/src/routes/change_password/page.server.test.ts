import { randomUUID } from 'node:crypto';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { isActionFailure, isRedirect } from '@sveltejs/kit';
import { Code, SignInReason, toPlatformError } from '@primandproper/platform-client';
import { refusal } from '@primandproper/platform-client/testing';

// The server's answer is a plain function rather than a vi.fn, which reports a rejection it
// returned as the test's own failure even once the action has caught it.
const server = vi.hoisted(() => ({
  calls: [] as unknown[][],
  answer: (): Promise<unknown> => Promise.resolve({}),
}));
vi.mock('$lib/grpc/clients', () => ({
  updatePassword: (...args: unknown[]) => {
    server.calls.push(args);
    return server.answer();
  },
}));

const { actions } = await import('./+page.server');

type Event = Parameters<typeof actions.default>[0];

function submit(fields: Record<string, string>): Promise<unknown> {
  const body = new FormData();
  for (const [k, v] of Object.entries(fields)) body.set(k, v);
  const event = {
    request: new Request('http://localhost/change_password', { method: 'POST', body }),
    locals: { session },
  };
  return Promise.resolve(actions.default(event as unknown as Event)).catch((e: unknown) => e);
}

const session = { id: randomUUID() };

function fakeForm(): Record<string, string> {
  const newPassword = randomUUID();
  return { current_password: randomUUID(), new_password: newPassword, confirm_password: newPassword };
}

beforeEach(() => {
  server.calls = [];
  server.answer = () => Promise.resolve({});
});

describe('changing a password', () => {
  it('sends what was typed and goes home', async () => {
    const form = fakeForm();
    const result = await submit(form);

    expect(isRedirect(result) && result.location).toBe('/');
    expect(server.calls).toEqual([
      [session, { currentPassword: form.current_password, newPassword: form.new_password, totpCode: '' }],
    ]);
  });

  it('refuses two new passwords that differ without asking the server', async () => {
    const result = await submit({ ...fakeForm(), confirm_password: randomUUID() });

    expect(isActionFailure(result)).toBe(true);
    expect(server.calls).toEqual([]);
  });

  it('asks for the code when the server says a second factor is needed', async () => {
    server.answer = () =>
      Promise.reject(toPlatformError(refusal(Code.UNAUTHENTICATED, randomUUID(), SignInReason.SECOND_FACTOR_REQUIRED)));

    const result = await submit(fakeForm());

    expect(isActionFailure(result) && result.data).toEqual({
      error: 'Enter the code from your authenticator app as well.',
    });
  });
});
