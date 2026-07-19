import { MiddlewareConsumer, Module, NestModule, RequestMethod } from '@nestjs/common';
import { HealthController } from './health/health.controller';
import { RequestIdMiddleware } from './request-id.middleware';
import { ProxyMiddleware } from './proxy.middleware';

@Module({
  controllers: [HealthController],
})
export class AppModule implements NestModule {
  configure(consumer: MiddlewareConsumer): void {
    // Ensure a request ID on every request (including /health).
    consumer.apply(RequestIdMiddleware).forRoutes('*');

    // Proxy only the backend routes; /health is served locally by the
    // HealthController and must not be proxied.
    consumer
      .apply(ProxyMiddleware)
      .exclude({ path: 'health', method: RequestMethod.ALL })
      .forRoutes(
        { path: 'order', method: RequestMethod.ALL },
        { path: 'products', method: RequestMethod.ALL },
        { path: 'products/*', method: RequestMethod.ALL },
      );
  }
}
