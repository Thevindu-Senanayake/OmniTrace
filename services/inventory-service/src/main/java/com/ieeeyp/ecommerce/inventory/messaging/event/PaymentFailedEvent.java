package com.ieeeyp.ecommerce.inventory.messaging.event;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

/**
 * Inbound {@code payment.failed} event. Per the event contract this payload
 * only reliably carries {@code orderId}; the item and quantity to release are
 * recovered from this service's own RESERVED inventory_events row, not the
 * payload. Unknown fields (reason, amount, timestamp) are ignored.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public record PaymentFailedEvent(String orderId) {
}
