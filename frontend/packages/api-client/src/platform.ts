/**
 * Calls to the services platform-go ships — identity, settings, waitlists, billing and the
 * rest — go through @primandproper/platform-client rather than through stubs generated
 * here. Their schema is platform's, so the stubs are too: this package generates only what
 * this repository's own protos describe.
 *
 * Only the transport and the method definitions are adopted. The library's Session, which
 * holds and refreshes the login, refreshes through platform's SignInService, and this
 * server does not register it: sign-in is still this repository's AuthService, so tokens
 * are still minted, exchanged and carried by the apps themselves.
 */

import * as grpc from '@grpc/grpc-js';
import {
  bearerAuthorizer,
  type CallOptions,
  createGrpcJsTransport,
  type UnaryMethod,
} from '@primandproper/platform-client';
import type { GrpcClientConfig } from './create-clients.js';

export interface PlatformClient {
  /** call makes an authenticated call, carrying `oauth2AccessToken` as a bearer credential. */
  call<Req, Res>(method: UnaryMethod<Req, Res>, oauth2AccessToken: string, request: Req): Promise<Res>;
  /** callAnonymous makes a call that carries no credential. */
  callAnonymous<Req, Res>(method: UnaryMethod<Req, Res>, request: Req): Promise<Res>;
}

export function createPlatformClient(config: GrpcClientConfig): PlatformClient {
  const transport = createGrpcJsTransport({
    address: config.serverUrl,
    credentials: config.insecure ? grpc.credentials.createInsecure() : grpc.credentials.createSsl(),
  });

  return {
    call: (method, oauth2AccessToken, request) => {
      const options: CallOptions = { metadata: bearerAuthorizer.credentials(oauth2AccessToken) };
      return transport.unary(method, request, options);
    },
    callAnonymous: (method, request) => transport.unary(method, request),
  };
}
