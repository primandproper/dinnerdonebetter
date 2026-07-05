import { provideLogger, type Logger } from '@primandproper/observability';

export type LoggerType = Logger;

/**
 * Server-side logger. Resolves to the pino-backed provider via the package's `node` export
 * condition. The `pretty` flag is retained for call-site compatibility; pino emits structured
 * JSON in every environment.
 */
export function buildServerSideLogger(name: string, _pretty = false): Logger {
  return provideLogger({ name });
}
