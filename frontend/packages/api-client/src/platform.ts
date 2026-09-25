/**
 * Every call goes over @primandproper/platform-client's transport: platform's services
 * through the library's own stubs, and this repository's through the method definitions
 * generated here, which have the same shape.
 *
 * The consumer app holds its login in the library's Session, which signs in, refreshes and
 * signs out through platform's SignInService. The admin app still signs in through this
 * repository's AuthService, because platform's tokens don't record whether they came
 * through the administrative door until platform-go#887 is released, so it carries a token
 * it holds itself through createPlatformClient.
 */

import * as grpc from '@grpc/grpc-js';
import {
  bearerAuthorizer,
  type CallOptions,
  createGrpcJsTransport,
  type Transport,
  type UnaryMethod,
} from '@primandproper/platform-client';

export interface PlatformTransportConfig {
  serverUrl: string;
  insecure?: boolean;
}

/** createPlatformTransport dials the API server. One per process: it holds the channel. */
export function createPlatformTransport(config: PlatformTransportConfig): Transport {
  return createGrpcJsTransport({
    address: config.serverUrl,
    credentials: config.insecure ? grpc.credentials.createInsecure() : grpc.credentials.createSsl(),
  });
}

export interface PlatformClient {
  /** call makes an authenticated call, carrying `oauth2AccessToken` as a bearer credential. */
  call<Req, Res>(method: UnaryMethod<Req, Res>, oauth2AccessToken: string, request: Req): Promise<Res>;
  /** callAnonymous makes a call that carries no credential. */
  callAnonymous<Req, Res>(method: UnaryMethod<Req, Res>, request: Req): Promise<Res>;
}

export function createPlatformClient(config: PlatformTransportConfig): PlatformClient {
  const transport = createPlatformTransport(config);

  return {
    call: (method, oauth2AccessToken, request) => {
      const options: CallOptions = { metadata: bearerAuthorizer.credentials(oauth2AccessToken) };
      return transport.unary(method, request, options);
    },
    callAnonymous: (method, request) => transport.unary(method, request),
  };
}
