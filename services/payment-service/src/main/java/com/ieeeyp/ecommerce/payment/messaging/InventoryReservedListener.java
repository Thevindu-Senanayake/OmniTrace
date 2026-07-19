package com.ieeeyp.ecommerce.payment.messaging;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.ieeeyp.ecommerce.payment.messaging.event.InventoryReservedEvent;
import com.ieeeyp.ecommerce.payment.service.PaymentService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

/**
 * Kafka entry point for the Payment Service. Consumes {@code inventory.reserved}
 * and processes the payment.
 *
 * Offsets commit only after the handler returns (ack-mode: record), so a thrown
 * exception redelivers the message. The handler is idempotent on order_id,
 * making at-least-once delivery safe. A payload that can never succeed
 * (malformed JSON) is logged and swallowed to avoid a poison-pill loop.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class InventoryReservedListener {

    private final PaymentService paymentService;
    private final ObjectMapper objectMapper;

    @KafkaListener(topics = "inventory.reserved", groupId = "payment-service")
    public void onInventoryReserved(String payload) {
        InventoryReservedEvent event;
        try {
            event = objectMapper.readValue(payload, InventoryReservedEvent.class);
        } catch (Exception e) {
            log.error("malformed inventory.reserved, skipping service=payment-service error={}", e.getMessage());
            return;
        }
        paymentService.process(event);
    }
}
