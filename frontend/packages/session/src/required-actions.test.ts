import { randomUUID } from 'node:crypto';
import { describe, it, expect } from 'vitest';
import { Code, Session, SignInReason } from '@primandproper/platform-client';
import { FakeTransport, MemoryCredentialStore, fakeIssuedToken, refusal } from '@primandproper/platform-client/testing';
import { type AuthStatus, SignInServiceService } from '@primandproper/platform-client/signin/v1';
import { CHANGE_PASSWORD_PATH, landingAfterSignIn, mustChangePassword } from './required-actions';

function fakeStatus(overrides: Partial<AuthStatus> = {}): AuthStatus {
  return {
    user: undefined,
    activeAccountId: randomUUID(),
    accountIds: [],
    hasPassword: true,
    twoFactorEnrolled: false,
    requiresPasswordChange: false,
    emailAddressVerified: true,
    ...overrides,
  };
}

/** signedIn is a Session holding a login, over a transport that answers GetAuthStatus with `answer`. */
function signedIn(answer: () => AuthStatus | Promise<AuthStatus>): Session {
  const transport = new FakeTransport().handle(SignInServiceService.getAuthStatus, async () => ({
    authenticated: true,
    status: await answer(),
  }));
  return new Session({ transport, store: new MemoryCredentialStore(fakeIssuedToken(new Date())) });
}

describe('mustChangePassword', () => {
  it('is true for a call refused because a password change is owed', async () => {
    const session = new Session({
      transport: new FakeTransport().handle(SignInServiceService.getSelf, () => {
        throw refusal(Code.FAILED_PRECONDITION, randomUUID(), SignInReason.PASSWORD_CHANGE_REQUIRED);
      }),
      store: new MemoryCredentialStore(fakeIssuedToken(new Date())),
    });

    const err = await session.call(SignInServiceService.getSelf, {}).catch((e: unknown) => e);

    expect(mustChangePassword(err)).toBe(true);
  });

  it('is false for any other refusal', async () => {
    const session = new Session({
      transport: new FakeTransport().handle(SignInServiceService.getSelf, () => {
        throw refusal(Code.FAILED_PRECONDITION, randomUUID(), SignInReason.SECOND_FACTOR_REQUIRED);
      }),
      store: new MemoryCredentialStore(fakeIssuedToken(new Date())),
    });

    const err = await session.call(SignInServiceService.getSelf, {}).catch((e: unknown) => e);

    expect(mustChangePassword(err)).toBe(false);
    expect(mustChangePassword(new Error(randomUUID()))).toBe(false);
  });
});

describe('landingAfterSignIn', () => {
  it('sends somebody who owes a password change to the form', async () => {
    const session = signedIn(() => fakeStatus({ requiresPasswordChange: true }));

    expect(await landingAfterSignIn(session)).toBe(CHANGE_PASSWORD_PATH);
  });

  it('sends everybody else home', async () => {
    const home = `/${randomUUID()}`;
    // An unverified address is owed too, but it is answered by the mailed link, not a form.
    const session = signedIn(() => fakeStatus({ emailAddressVerified: false }));

    expect(await landingAfterSignIn(session, home)).toBe(home);
  });

  it('sends them home when the status cannot be read', async () => {
    const home = `/${randomUUID()}`;
    const session = signedIn(() => {
      throw refusal(Code.UNAVAILABLE, randomUUID(), '');
    });

    expect(await landingAfterSignIn(session, home)).toBe(home);
  });
});
