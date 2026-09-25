/**
 * @dinnerdonebetter/api-client
 * gRPC method definitions and types for the Dinner Done Better API.
 */

export { createAdminGrpcClients } from './admin-clients.js';
export {
  createPlatformClient,
  createPlatformTransport,
  type PlatformClient,
  type PlatformTransportConfig,
} from './platform.js';
export { QueryFilter, Pagination } from './primandproper/platform/filtering/v1/filtering.js';
export { AnalyticsServiceService } from './analytics/analytics_service.js';
export { AuthServiceService } from './auth/auth_service.js';
export { MealPlanningServiceService } from './mealplanning/mealplanning_service.js';
