import { Injectable, Logger, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { REQUEST_ID_HEADER } from './config';

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
    this.logger.log(
      `request_id=${requestId} method=${req.method} path=${req.originalUrl} service=api-gateway`,
    );
    next();
  }
}
