/**
 * Analytics wrapper for the gRPC analytics proxy.
 * Events are sent to the backend, which forwards to Segment/PostHog/Rudderstack.
 * Use from server-side code (load functions, form actions, API routes).
 * For client-side tracking, call the /api/analytics/track endpoint.
 */

import { trackEvent as grpcTrackEvent, trackAnonymousEvent as grpcTrackAnonymousEvent } from '$lib/grpc/clients';
import { logger } from '$lib/logger';

/**
 * Track an event for an identified user.
 * Call from server-side only (load, form action, API route).
 */
export async function trackEvent(event: string, properties: Record<string, string> = {}): Promise<void> {
  try {
    await grpcTrackEvent(event, properties);
  } catch (err) {
    logger.error('Analytics trackEvent failed', err);
  }
}

/**
 * Track an event for an anonymous user.
 * Call from server-side only (load, form action, API route).
 */
export async function trackAnonymousEvent(
  event: string,
  anonymousId: string,
  properties: Record<string, string> = {},
): Promise<void> {
  try {
    await grpcTrackAnonymousEvent(event, anonymousId, properties);
  } catch (err) {
    logger.error('Analytics trackAnonymousEvent failed', err);
  }
}
