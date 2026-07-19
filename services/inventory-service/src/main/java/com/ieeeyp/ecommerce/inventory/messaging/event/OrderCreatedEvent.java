package com.ieeeyp.ecommerce.inventory.messaging.event;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.math.BigDecimal;
import java.util.List;

/**
 * Inbound {@code order.created} event. Carries one or more line items; the
 * whole order is reserved atomically. Unknown fields are ignored for forward
 * compatibility.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public record OrderCreatedEvent(
        String eventType,
        int version,
        String orderId,
        String customerId,
        List<Item> items,
        BigDecimal totalAmount,
        String timestamp) {

    @JsonIgnoreProperties(ignoreUnknown = true)
    public record Item(String itemId, int quantity, BigDecimal unitPrice) {
    }
}
