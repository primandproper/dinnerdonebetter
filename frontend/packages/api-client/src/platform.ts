/**
 * Every call goes over @primandproper/platform-client's transport: platform's services
 * through the library's own stubs, and this repository's through the method definitions
 * generated here, which have the same shape. Both web apps hold their login in the
 * library's Session, which signs in, refreshes and signs out through platform's
 * SignInService, and make every authenticated call through it.
 */

import * as grpc from '@grpc/grpc-js';
import { createGrpcJsTransport, type Transport } from '@primandproper/platform-client';

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
