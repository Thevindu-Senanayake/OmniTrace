/** Request body for POST /charge. Amount is in major currency units. */
export interface ChargeRequest {
  orderId: string;
  customerId: string;
  amount: number;
}

/** Response from a successful charge. */
export interface ChargeResponse {
  transactionId: string;
  status: 'SUCCESS';
  amount: number;
  processingMs: number;
}
