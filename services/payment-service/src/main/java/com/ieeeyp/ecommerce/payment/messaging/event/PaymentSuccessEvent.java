package com.ieeeyp.ecommerce.payment.messaging.event;

import java.math.BigDecimal;

/**
 * Outbound {@code payment.success} event consumed by the Notification Service
 * and the Order Service (which moves the order to CONFIRMED).
 */
public record PaymentSuccessEvent(
        String eventType,
        int version,
        String orderId,
        String paymentId,
        BigDecimal amount,
        String timestamp) {

    public static PaymentSuccessEvent of(String orderId, String paymentId, BigDecimal amount, String timestamp) {
        return new PaymentSuccessEvent("payment.success", 1, orderId, paymentId, amount, timestamp);
    }
}
