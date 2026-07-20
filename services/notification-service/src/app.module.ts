import { Module } from '@nestjs/common';
import { HealthController } from './health/health.controller';
import { NotificationConsumer } from './notification.consumer';
import { NotifierService } from './notifier.service';

@Module({
  controllers: [HealthController],
  providers: [NotifierService, NotificationConsumer],
})
export class AppModule {}
