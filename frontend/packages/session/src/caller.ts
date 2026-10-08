/**
 * Every call an app makes, platform's services and this repository's alike, goes over one
 * transport. An authenticated call goes through the request's Session, which holds the login,
 * refreshes it through platform's SignInService, and retries a call the server refused once
 * with the successor.
 */

import { redirect } from '@sveltejs/kit';
import {
  type CredentialStore,
  type Metadata,
  redirectOnNotSignedIn,
  Session,
  type Transport,
  type UnaryMethod,
} from '@primandproper/platform-client';
import { CHANGE_PASSWORD_PATH, mustChangePassword } from './required-actions';

/** SIGN_IN_PATH is where somebody whose login is over is sent. */
export const SIGN_IN_PATH = '/login';

export interface Caller {
  /**
   * newSession is a Session over `store`. Sessions built per request share the process's
   * exchange coordinator, so concurrent requests carrying the same cookie present its refresh
   * token once between them. That holds within one process only: more than one replica needs
   * a SharedExchangeCoordinator.
   *
   * `metadata` rides on every call the Session makes; see clientMetadata.
   */
  newSession(store: CredentialStore, metadata?: Metadata): Session;

  /**
   * call makes an authenticated call. A login that is over by the time it is made — never
   * held, lapsed, or refused a refresh — sends the person to sign in again; the Session has
   * already cleared the cookie by then. One refused because they owe a password change sends
   * them to the form for it.
   */
  call<Req, Res>(session: Session, method: UnaryMethod<Req, Res>, request: Req): Promise<Res>;

  /** authed is `call` bound to one method, which is how an app lists the calls it makes. */
  authed<Req, Res>(method: UnaryMethod<Req, Res>): (session: Session, request: Req) => Promise<Res>;
}

/** createCaller is the Caller over `transport`. One per process: the transport holds the channel. */
export function createCaller(transport: Transport): Caller {
  async function call<Req, Res>(session: Session, method: UnaryMethod<Req, Res>, request: Req): Promise<Res> {
    try {
      return await redirectOnNotSignedIn(
        session,
        () => session.call(method, request),
        () => redirect(302, SIGN_IN_PATH),
      );
    } catch (err) {
      if (mustChangePassword(err)) {
        redirect(302, CHANGE_PASSWORD_PATH);
      }
      throw err;
    }
  }

  return {
    newSession: (store, metadata = {}) => new Session({ transport, store, metadata }),
    call,
    authed: (method) => (session, request) => call(session, method, request),
  };
}
