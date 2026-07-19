package com.ieeeyp.ecommerce.inventory.messaging.event;

import java.math.BigDecimal;

/**
 * Outbound {@code inventory.reserved} event consumed by the Payment Service.
 * Carries forward the item and amount so payment can proceed without querying
 * another service's database. Single-item shape mirrors order.created.
 */
public record InventoryReservedEvent(
        String eventType,
        int version,
        String orderId,
        String customerId,
        String itemId,
        int quantity,
        BigDecimal totalAmount,
        String timestamp) {

    public static InventoryReservedEvent of(OrderCreatedEvent order, String timestamp) {
        return new InventoryReservedEvent(
                "inventory.reserved",
                1,
                order.orderId(),
                order.customerId(),
                order.itemId(),
                order.quantity(),
                order.totalAmount(),
                timestamp);
    }
}
