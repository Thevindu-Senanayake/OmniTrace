import { Injectable, Logger } from "@nestjs/common";
import { randomUUID } from "crypto";
import { ChargeRequest, ChargeResponse } from "./charge.dto";

/**
 * Mock payment processor. Clean business logic only: it always approves the
 * charge. Latency and failures are injected at the network layer by Toxiproxy
 * sitting in front of this service - there is deliberately no chaos code here.
 */
@Injectable()
export class ChargeService {
	private readonly logger = new Logger(ChargeService.name);

	charge(req: ChargeRequest): ChargeResponse {
		const startedAt = Date.now();
		// A real processor would call the card network here; the mock approves
		// immediately. processingMs reflects only in-process time.
		const response: ChargeResponse = {
			transactionId: `txn_${randomUUID()}`,
			status: "SUCCESS",
			amount: req.amount,
			processingMs: Date.now() - startedAt,
		};
		this.logger.log(
			`charge approved order_id=${req.orderId} amount=${req.amount} transaction_id=${response.transactionId} service=payment-gateway`,
		);
		return response;
	}
}
