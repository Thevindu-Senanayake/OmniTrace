import { Injectable, Logger, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import {
  createProxyMiddleware,
  RequestHandler,
} from 'http-proxy-middleware';
import { config } from './config';

/**
 * Proxies gateway routes to the backend services. /order goes to the Order
 * Service; /products (and /products/:id) go to the Product Catalog. The gateway
 * does no business logic — it only forwards, carrying the X-Request-ID header
 * (already ensured by RequestIdMiddleware) through to the backend.
 */
@Injectable()
export class ProxyMiddleware implements NestMiddleware {
  private readonly logger = new Logger('gateway-proxy');

  private readonly orderProxy: RequestHandler = createProxyMiddleware({
    target: config.orderServiceUrl,
    changeOrigin: true,
    on: {
      error: (err, _req, res) => this.onError(err, res as Response),
    },
  });

  private readonly catalogProxy: RequestHandler = createProxyMiddleware({
    target: config.productCatalogUrl,
    changeOrigin: true,
    on: {
      error: (err, _req, res) => this.onError(err, res as Response),
    },
  });

  use(req: Request, res: Response, next: NextFunction): void {
    if (req.path === '/order') {
      void this.orderProxy(req, res, next);
      return;
    }
    if (req.path === '/products' || req.path.startsWith('/products/')) {
      void this.catalogProxy(req, res, next);
      return;
    }
    next();
  }

  private onError(err: Error, res: Response): void {
    this.logger.error(`proxy error: ${err.message} service=api-gateway`);
    if (!res.headersSent) {
      res.writeHead(502, { 'Content-Type': 'application/json' });
    }
    res.end(JSON.stringify({ error: 'bad gateway' }));
  }
}
