package com.ieeeyp.ecommerce.inventory.messaging;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.ieeeyp.ecommerce.inventory.messaging.event.OrderCreatedEvent;
import com.ieeeyp.ecommerce.inventory.messaging.event.PaymentFailedEvent;
import com.ieeeyp.ecommerce.inventory.service.InventoryService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

/**
 * Kafka entry point for the Inventory Service.
 *
 * <ul>
 *   <li>{@code order.created}  → reserve stock (saga forward step)</li>
 *   <li>{@code payment.failed} → release stock (saga compensation)</li>
 * </ul>
 *
 * Offsets commit only after the handler returns (ack-mode: record), so a thrown
 * exception redelivers the message. Handlers are idempotent, making at-least-
 * once delivery safe. A payload that can never succeed (malformed JSON) is
 * logged and swallowed to avoid a poison-pill loop.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class OrderEventListener {

    private final InventoryService inventoryService;
    private final ObjectMapper objectMapper;

    @KafkaListener(topics = "order.created", groupId = "inventory-service")
    public void onOrderCreated(String payload) {
        OrderCreatedEvent event;
        try {
            event = objectMapper.readValue(payload, OrderCreatedEvent.class);
        } catch (Exception e) {
            log.error("malformed order.created, skipping service=inventory-service error={}", e.getMessage());
            return;
        }
        inventoryService.reserve(event);
    }

    @KafkaListener(topics = "payment.failed", groupId = "inventory-service")
    public void onPaymentFailed(String payload) {
        PaymentFailedEvent event;
        try {
            event = objectMapper.readValue(payload, PaymentFailedEvent.class);
        } catch (Exception e) {
            log.error("malformed payment.failed, skipping service=inventory-service error={}", e.getMessage());
            return;
        }
        inventoryService.release(event.orderId());
    }
}
