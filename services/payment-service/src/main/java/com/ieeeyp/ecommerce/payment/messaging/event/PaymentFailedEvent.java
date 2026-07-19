package com.ieeeyp.ecommerce.payment.messaging.event;

import java.math.BigDecimal;

/**
 * Outbound {@code payment.failed} event consumed by the Notification Service,
 * the Order Service (→ PAYMENT_FAILED) and the Inventory Service (→ release
 * reserved stock). reason ∈ CONNECTION_ERROR | TIMEOUT | DECLINED.
 */
public record PaymentFailedEvent(
        String eventType,
        int version,
        String orderId,
        String reason,
        BigDecimal amount,
        String timestamp) {

    public static PaymentFailedEvent of(String orderId, String reason, BigDecimal amount, String timestamp) {
        return new PaymentFailedEvent("payment.failed", 1, orderId, reason, amount, timestamp);
    }
}
