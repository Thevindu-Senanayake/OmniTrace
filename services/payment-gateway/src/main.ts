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
