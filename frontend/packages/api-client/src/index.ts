/**
 * @dinnerdonebetter/api-client
 * gRPC client factory and types for the Dinner Done Better API.
 */

import * as grpc from '@grpc/grpc-js';
import type { Metadata } from '@grpc/grpc-js';

export { createGrpcClients, type GrpcClientConfig } from './create-clients.js';
export { createAdminGrpcClients } from './admin-clients.js';
export { QueryFilter, Pagination } from './primandproper/platform/filtering/v1/filtering.js';

/**
 * Metadata with Bearer token for authenticated gRPC calls.
 */
export function authMetadata(oauth2AccessToken: string): Metadata {
  const metadata = new grpc.Metadata();
  metadata.add('authorization', `Bearer ${oauth2AccessToken}`);
  return metadata;
}
