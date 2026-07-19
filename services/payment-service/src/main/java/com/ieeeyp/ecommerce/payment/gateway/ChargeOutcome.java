package com.ieeeyp.ecommerce.payment.gateway;

/**
 * Result of a gateway charge attempt. On success carries the transactionId; on
 * failure carries a reason matching the payment.failed contract
 * (CONNECTION_ERROR | TIMEOUT | DECLINED).
 */
public record ChargeOutcome(boolean success, String transactionId, String failureReason, int processingMs) {

    public static ChargeOutcome ok(String transactionId, int processingMs) {
        return new ChargeOutcome(true, transactionId, null, processingMs);
    }

    public static ChargeOutcome failed(String reason, int processingMs) {
        return new ChargeOutcome(false, null, reason, processingMs);
    }
}
