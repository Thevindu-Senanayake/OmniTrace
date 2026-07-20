import { NestFactory } from "@nestjs/core";
import { Logger } from "@nestjs/common";
import { AppModule } from "./app.module";

async function bootstrap() {
	const app = await NestFactory.create(AppModule);
	const port = process.env.PORT;
	await app.listen(port);
	new Logger("bootstrap").log(`notification-service listening on ${port}`);
}

bootstrap();
