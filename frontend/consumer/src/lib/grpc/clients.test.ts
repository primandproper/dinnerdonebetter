import { randomUUID } from 'node:crypto';
import { describe, it, expect, vi } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { Code, Session, SignInReason } from '@primandproper/platform-client';
import { FakeTransport, MemoryCredentialStore, fakeIssuedToken, refusal } from '@primandproper/platform-client/testing';
import { SignInServiceService } from '@primandproper/platform-client/signin/v1';
import { CHANGE_PASSWORD_PATH } from '$lib/auth/required-actions';

const transport = vi.hoisted(() => ({ current: undefined as FakeTransport | undefined }));

vi.mock('$env/dynamic/private', () => ({ env: {} }));
vi.mock('@dinnerdonebetter/api-client', async (original) => ({
  ...(await original<typeof import('@dinnerdonebetter/api-client')>()),
  createPlatformTransport: () => ({
    unary: (...args: Parameters<FakeTransport['unary']>) => transport.current!.unary(...args),
  }),
}));

const { getSelf, newSession } = await import('./clients');

/** signedIn is a Session holding a login, over a transport that answers GetSelf with `answer`. */
function signedIn(answer: () => unknown): Session {
  transport.current = new FakeTransport().handle(SignInServiceService.getSelf, () => answer() as never);
  return newSession(new MemoryCredentialStore(fakeIssuedToken(new Date())));
}

async function failureOf(p: Promise<unknown>): Promise<unknown> {
  return p.then(
    () => expect.unreachable('the call succeeded'),
    (e: unknown) => e,
  );
}

describe('an authenticated call', () => {
  it('answers what the server answered', async () => {
    const id = randomUUID();
    const session = signedIn(() => ({ user: { id } }));

    expect((await getSelf(session))?.id).toBe(id);
  });

  it('sends somebody who owes a password change to the form for it', async () => {
    const session = signedIn(() => {
      throw refusal(Code.FAILED_PRECONDITION, randomUUID(), SignInReason.PASSWORD_CHANGE_REQUIRED);
    });

    const err = await failureOf(getSelf(session));

    expect(isRedirect(err) && err.location).toBe(CHANGE_PASSWORD_PATH);
  });

  it('sends somebody whose login is over to sign in', async () => {
    transport.current = new FakeTransport();
    const session = newSession(new MemoryCredentialStore());

    const err = await failureOf(getSelf(session));

    expect(isRedirect(err) && err.location).toBe('/login');
  });

  it('rethrows any other refusal rather than sending anybody anywhere', async () => {
    const session = signedIn(() => {
      throw refusal(Code.INTERNAL, randomUUID(), '');
    });

    const err = await failureOf(getSelf(session));

    expect(isRedirect(err)).toBe(false);
  });
});
