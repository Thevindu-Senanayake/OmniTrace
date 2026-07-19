package com.ieeeyp.ecommerce.inventory.messaging.event;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.math.BigDecimal;

/**
 * Inbound {@code order.created} event. Matches the flat, single-item shape
 * actually emitted by the Order Service (itemId/quantity/totalAmount at the
 * top level). Unknown fields are ignored for forward compatibility.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public record OrderCreatedEvent(
        String eventType,
        int version,
        String orderId,
        String customerId,
        String itemId,
        int quantity,
        BigDecimal totalAmount,
        String timestamp) {
}
