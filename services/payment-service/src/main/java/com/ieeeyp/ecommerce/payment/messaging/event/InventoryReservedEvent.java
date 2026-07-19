package com.ieeeyp.ecommerce.payment.messaging.event;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.math.BigDecimal;
import java.util.List;

/**
 * Inbound {@code inventory.reserved} event. Matches the shape emitted by the
 * Inventory Service: order/customer identity, the reserved line items, and the
 * total amount to charge. Unknown fields are ignored for forward compatibility.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public record InventoryReservedEvent(
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
