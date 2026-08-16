// Must be first: the OTel auto-instrumentations patch http/express at require
// time, so any module loaded before this one produces no spans.
import './tracing';

import { NestFactory } from '@nestjs/core';
import { Logger } from '@nestjs/common';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: false });
  const port = process.env.PORT ?? '3005';
  await app.listen(port);
  new Logger('bootstrap').log(`payment-gateway listening on ${port}`);
}

bootstrap();
