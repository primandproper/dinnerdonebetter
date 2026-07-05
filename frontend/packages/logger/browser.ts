import { provideLogger, type Logger } from '@primandproper/observability';

export type LoggerType = Logger;

/**
 * Client-side logger for use in components. Resolves to the console-backed provider via the
 * package's `browser` export condition, so it is safe to import into browser bundles.
 */
export function buildClientSideLogger(name: string): Logger {
  return provideLogger({ name });
}
