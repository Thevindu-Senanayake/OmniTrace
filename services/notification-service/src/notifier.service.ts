import { Injectable, Logger } from '@nestjs/common';

/**
 * Turns a consumed saga event into a simulated customer notification. Nothing
 * is actually sent — the "notification" is a structured log line carrying the
 * correlation fields (order_id, trace_id) so it can be tied back to the trace.
 */
@Injectable()
export class NotifierService {
  private readonly logger = new Logger('notification');

  notify(topic: string, event: Record<string, unknown>, traceId: string): void {
    const orderId = String(event.orderId ?? 'unknown');
    const message = this.render(topic, event);
    this.logger.log(
      `notification sent order_id=${orderId} trace_id=${traceId} channel=email topic=${topic} message="${message}" service=notification-service`,
    );
  }

  private render(topic: string, event: Record<string, unknown>): string {
    switch (topic) {
      case 'payment.success':
        return `Your order is confirmed. Payment ${event.paymentId ?? ''} of ${event.amount ?? ''} succeeded.`;
      case 'payment.failed':
        return `Your payment could not be processed (${event.reason ?? 'unknown'}). Please try again.`;
      case 'inventory.failed':
        return `Sorry, some items are out of stock and your order was cancelled.`;
      default:
        return `Update on your order.`;
    }
  }
}
