package com.ieeeyp.ecommerce.inventory.messaging.event;

/**
 * Outbound {@code inventory.failed} event consumed by the Notification Service
 * (and the Order Service, which moves the order to OUT_OF_STOCK). Reports the
 * single item that could not be reserved along with the shortfall.
 */
public record InventoryFailedEvent(
        String eventType,
        int version,
        String orderId,
        String reason,
        String itemId,
        int requested,
        int available,
        String timestamp) {

    public static InventoryFailedEvent outOfStock(OrderCreatedEvent order, int available, String timestamp) {
        return new InventoryFailedEvent(
                "inventory.failed",
                1,
                order.orderId(),
                "OUT_OF_STOCK",
                order.itemId(),
                order.quantity(),
                available,
                timestamp);
    }
}
