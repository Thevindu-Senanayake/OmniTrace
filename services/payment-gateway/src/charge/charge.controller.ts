import {
  BadRequestException,
  Body,
  Controller,
  Post,
} from '@nestjs/common';
import { ChargeService } from './charge.service';
import { ChargeRequest, ChargeResponse } from './charge.dto';

@Controller('charge')
export class ChargeController {
  constructor(private readonly chargeService: ChargeService) {}

  @Post()
  charge(@Body() body: ChargeRequest): ChargeResponse {
    if (!body?.orderId || typeof body.orderId !== 'string') {
      throw new BadRequestException('orderId is required');
    }
    if (typeof body.amount !== 'number' || body.amount <= 0) {
      throw new BadRequestException('amount must be a positive number');
    }
    return this.chargeService.charge(body);
  }
}
