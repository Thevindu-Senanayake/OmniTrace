package com.ieeeyp.ecommerce.inventory.messaging.event;

import java.util.List;

/**
 * Outbound {@code inventory.failed} event consumed by the Notification Service
 * (and the Order Service, which moves the order to OUT_OF_STOCK). Reports every
 * line that could not be reserved along with its shortfall.
 */
public record InventoryFailedEvent(
        String eventType,
        int version,
        String orderId,
        String reason,
        List<FailedItem> failedItems,
        String timestamp) {

    /** A line that could not be satisfied: how much was wanted vs available. */
    public record FailedItem(String itemId, int requested, int available) {
    }

    public static InventoryFailedEvent outOfStock(String orderId, List<FailedItem> failedItems, String timestamp) {
        return new InventoryFailedEvent(
                "inventory.failed",
                1,
                orderId,
                "OUT_OF_STOCK",
                failedItems,
                timestamp);
    }
}
