package com.ieeeyp.ecommerce.inventory.messaging;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

/**
 * Publishes inventory outcome events. The Kafka message key is always the
 * orderId, so every event for one order lands on the same partition and
 * preserves per-order ordering.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class InventoryEventPublisher {

    public static final String TOPIC_INVENTORY_RESERVED = "inventory.reserved";
    public static final String TOPIC_INVENTORY_FAILED = "inventory.failed";

    private final KafkaTemplate<String, String> kafkaTemplate;
    private final ObjectMapper objectMapper;

    public void publish(String topic, String orderId, Object event) {
        String payload;
        try {
            payload = objectMapper.writeValueAsString(event);
        } catch (JsonProcessingException e) {
            // A payload we produced ourselves failing to serialize is a bug,
            // not a transient fault - fail loudly rather than swallow it.
            throw new IllegalStateException("failed to serialize " + topic + " event", e);
        }
        kafkaTemplate.send(topic, orderId, payload);
        log.info("published event topic={} order_id={} service=inventory-service", topic, orderId);
    }
}
