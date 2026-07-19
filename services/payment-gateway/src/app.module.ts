import { Module } from '@nestjs/common';
import { ChargeController } from './charge/charge.controller';
import { ChargeService } from './charge/charge.service';
import { HealthController } from './health/health.controller';

@Module({
  controllers: [ChargeController, HealthController],
  providers: [ChargeService],
})
export class AppModule {}
