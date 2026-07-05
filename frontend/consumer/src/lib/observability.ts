import type { ObservabilityDeps } from '@primandproper/observability';
import { logger } from '$lib/logger';

/**
 * Observability deps injected into platform-ts providers (httpclient, retry). Only the logger is
 * supplied; the platform's default tracer/meter providers read the global OpenTelemetry API that
 * initServerOtel() registers in hooks.server.ts, so spans land on the existing pipeline.
 *
 * Server-side only — pulls in the pino-backed logger.
 */
export const observabilityDeps: ObservabilityDeps = { logger };
