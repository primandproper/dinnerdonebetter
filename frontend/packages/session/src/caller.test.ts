import { randomUUID } from 'node:crypto';
import { describe, it, expect } from 'vitest';
import { isRedirect } from '@sveltejs/kit';
import { Code, SignInReason } from '@primandproper/platform-client';
import { FakeTransport, MemoryCredentialStore, fakeIssuedToken, refusal } from '@primandproper/platform-client/testing';
import { SignInServiceService } from '@primandproper/platform-client/signin/v1';
import { createCaller, SIGN_IN_PATH } from './caller';
import { CHANGE_PASSWORD_PATH } from './required-actions';

/** signedIn is a caller's GetSelf over a held login, on a transport that answers it with `answer`. */
function signedIn(answer: () => unknown) {
  const transport = new FakeTransport().handle(SignInServiceService.getSelf, () => answer() as never);
  const { authed, newSession } = createCaller(transport);
  const getSelf = authed(SignInServiceService.getSelf);
  const session = newSession(new MemoryCredentialStore(fakeIssuedToken(new Date())));
  return () => getSelf(session, {});
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
    const getSelf = signedIn(() => ({ user: { id } }));

    expect((await getSelf()).user?.id).toBe(id);
  });

  it('sends somebody who owes a password change to the form for it', async () => {
    const getSelf = signedIn(() => {
      throw refusal(Code.FAILED_PRECONDITION, randomUUID(), SignInReason.PASSWORD_CHANGE_REQUIRED);
    });

    const err = await failureOf(getSelf());

    expect(isRedirect(err) && err.location).toBe(CHANGE_PASSWORD_PATH);
  });

  it('sends somebody whose login is over to sign in', async () => {
    const { call, newSession } = createCaller(new FakeTransport());
    const session = newSession(new MemoryCredentialStore());

    const err = await failureOf(call(session, SignInServiceService.getSelf, {}));

    expect(isRedirect(err) && err.location).toBe(SIGN_IN_PATH);
  });

  it('rethrows any other refusal rather than sending anybody anywhere', async () => {
    const getSelf = signedIn(() => {
      throw refusal(Code.INTERNAL, randomUUID(), '');
    });

    const err = await failureOf(getSelf());

    expect(isRedirect(err)).toBe(false);
  });

  it('carries the metadata the session was built with', async () => {
    const address = randomUUID();
    const transport = new FakeTransport().handle(SignInServiceService.getSelf, () => ({ user: undefined }));
    const { call, newSession } = createCaller(transport);
    const session = newSession(new MemoryCredentialStore(fakeIssuedToken(new Date())), {
      'x-client-address': address,
    });

    await call(session, SignInServiceService.getSelf, {});

    expect(transport.callsTo(SignInServiceService.getSelf)[0].options.metadata).toMatchObject({
      'x-client-address': address,
    });
  });
});
