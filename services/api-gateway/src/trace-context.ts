import { trace } from '@opentelemetry/api';

/**
 * Reads the active span's ids for log correlation. The OTel auto-instrumentation
 * makes the current span active for the duration of an HTTP request or Kafka
 * message, so business log sites can pick up trace_id/span_id with no plumbing.
 * Returns 'none' when there is no active span, matching the events.md contract.
 */
export function traceContext(): { trace_id: string; span_id: string } {
  const ctx = trace.getActiveSpan()?.spanContext();
  return { trace_id: ctx?.traceId ?? 'none', span_id: ctx?.spanId ?? 'none' };
}
