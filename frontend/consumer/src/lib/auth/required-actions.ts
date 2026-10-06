/**
 * A person an operator has forced to change their password is still signed in: the
 * alternative is somebody who cannot reach the form. The server holds them there instead,
 * refusing every call but the change itself with PASSWORD_CHANGE_REQUIRED, so routing them to
 * the form is this app's job, both straight after signing in and on any refusal after that.
 */

import { getAuthStatus, PlatformError, type Session, SignInReason } from '@primandproper/platform-client';

export const CHANGE_PASSWORD_PATH = '/change_password';

/** mustChangePassword reports whether a call was refused because a password change is owed first. */
export function mustChangePassword(err: unknown): boolean {
  return err instanceof PlatformError && err.is(SignInReason.PASSWORD_CHANGE_REQUIRED);
}

/**
 * landingAfterSignIn is where somebody who just signed in goes: the change-password form when
 * they owe a change, and `home` otherwise. A status that cannot be read sends them home, where
 * the first refused call brings them back to the form.
 */
export async function landingAfterSignIn(session: Session, home = '/'): Promise<string> {
  try {
    const status = await getAuthStatus(session);
    return status.authenticated && status.requiredActions.includes('change_password') ? CHANGE_PASSWORD_PATH : home;
  } catch {
    return home;
  }
}
