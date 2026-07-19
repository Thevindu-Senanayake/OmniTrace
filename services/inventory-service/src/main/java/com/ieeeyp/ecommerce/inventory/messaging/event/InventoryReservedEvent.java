package com.ieeeyp.ecommerce.inventory.messaging.event;

import java.math.BigDecimal;
import java.util.List;

/**
 * Outbound {@code inventory.reserved} event consumed by the Payment Service.
 * Carries the full item list and amount forward so payment can proceed without
 * querying another service's database.
 */
public record InventoryReservedEvent(
        String eventType,
        int version,
        String orderId,
        String customerId,
        List<OrderCreatedEvent.Item> items,
        BigDecimal totalAmount,
        String timestamp) {

    public static InventoryReservedEvent of(OrderCreatedEvent order, String timestamp) {
        return new InventoryReservedEvent(
                "inventory.reserved",
                1,
                order.orderId(),
                order.customerId(),
                order.items(),
                order.totalAmount(),
                timestamp);
    }
}
