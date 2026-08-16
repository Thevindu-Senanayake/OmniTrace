// Must be first: the OTel auto-instrumentations patch http and kafkajs at
// require time, so any module loaded before this one produces no spans.
import "./tracing";

import { NestFactory } from "@nestjs/core";
import { Logger } from "@nestjs/common";
import { AppModule } from "./app.module";
import { config } from "./config";

async function bootstrap() {
	const app = await NestFactory.create(AppModule);
	await app.listen(config.port);
	new Logger("bootstrap").log(`notification-service listening on ${config.port}`);
}

bootstrap();
