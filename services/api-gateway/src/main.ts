import { NestFactory } from '@nestjs/core';
import { Logger } from '@nestjs/common';
import { AppModule } from './app.module';

async function bootstrap() {
  // Disable Nest's body parser: this gateway only proxies, and parsing the body
  // here would break http-proxy-middleware's streaming of POST bodies.
  const app = await NestFactory.create(AppModule, { bodyParser: false });
  const port = process.env.PORT ?? '8080';
  await app.listen(port);
  new Logger('bootstrap').log(`api-gateway listening on ${port}`);
}

bootstrap();
