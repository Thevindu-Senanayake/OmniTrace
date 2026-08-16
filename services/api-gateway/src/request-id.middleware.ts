import { Injectable, Logger, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { REQUEST_ID_HEADER } from './config';
import { traceContext } from './trace-context';

/**
 * Ensures every inbound request carries an X-Request-ID. If the client did not
 * supply one, a UUID is generated. The header is set on both the incoming
 * request (so proxied backends receive it) and the response (so callers can
 * correlate). Downstream services log this value for request tracing.
 */
@Injectable()
export class RequestIdMiddleware implements NestMiddleware {
  private readonly logger = new Logger('gateway');

  use(req: Request, res: Response, next: NextFunction): void {
    let requestId = req.headers[REQUEST_ID_HEADER];
    if (!requestId || typeof requestId !== 'string') {
      requestId = uuidv4();
      req.headers[REQUEST_ID_HEADER] = requestId;
    }
    res.setHeader('X-Request-ID', requestId);
    const { trace_id, span_id } = traceContext();
    // A pure proxy never parses the body, so it cannot know the order id here.
    this.logger.log(
      `request_id=${requestId} method=${req.method} path=${req.originalUrl} order_id=unknown trace_id=${trace_id} span_id=${span_id} service=api-gateway`,
    );
    next();
  }
}
