package com.ieeeyp.ecommerce.inventory.service;

import com.ieeeyp.ecommerce.inventory.domain.Inventory;
import com.ieeeyp.ecommerce.inventory.domain.InventoryEvent;
import com.ieeeyp.ecommerce.inventory.messaging.InventoryEventPublisher;
import com.ieeeyp.ecommerce.inventory.messaging.event.InventoryFailedEvent;
import com.ieeeyp.ecommerce.inventory.messaging.event.InventoryReservedEvent;
import com.ieeeyp.ecommerce.inventory.messaging.event.OrderCreatedEvent;
import com.ieeeyp.ecommerce.inventory.repository.InventoryEventRepository;
import com.ieeeyp.ecommerce.inventory.repository.InventoryRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.Optional;
import java.util.UUID;

/**
 * Core inventory business logic: reserve stock for an order and release it when
 * a downstream payment fails. All mutations run inside a single transaction
 * that holds a row-level lock on the inventory row, so concurrent orders for
 * the same item serialize at the database - no simulation code.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class InventoryService {

    private final InventoryRepository inventoryRepository;
    private final InventoryEventRepository eventRepository;
    private final InventoryEventPublisher publisher;

    /**
     * Handles order.created: reserve stock and publish inventory.reserved on
     * success, or inventory.failed when stock is insufficient / unknown item.
     * Idempotent - a redelivered order that was already reserved is a no-op.
     */
    @Transactional
    public void reserve(OrderCreatedEvent order) {
        UUID orderId = UUID.fromString(order.orderId());

        if (eventRepository.existsByOrderIdAndAction(orderId, InventoryEvent.Action.RESERVED.name())) {
            log.info("reservation already applied, skipping order_id={} service=inventory-service", order.orderId());
            return;
        }

        Optional<Inventory> maybeItem = inventoryRepository.findByIdForUpdate(order.itemId());
        if (maybeItem.isEmpty()) {
            log.warn("unknown item, failing reservation order_id={} item_id={} service=inventory-service",
                    order.orderId(), order.itemId());
            publishFailed(order, 0);
            return;
        }

        Inventory item = maybeItem.get();
        if (item.getStock() < order.quantity()) {
            log.info("insufficient stock order_id={} item_id={} requested={} available={} service=inventory-service",
                    order.orderId(), order.itemId(), order.quantity(), item.getStock());
            publishFailed(order, item.getStock());
            return;
        }

        item.setStock(item.getStock() - order.quantity());
        item.setReserved(item.getReserved() + order.quantity());
        item.setUpdatedAt(OffsetDateTime.now(ZoneOffset.UTC));
        inventoryRepository.save(item);

        eventRepository.save(new InventoryEvent(
                orderId, order.itemId(), order.quantity(), InventoryEvent.Action.RESERVED));

        log.info("stock reserved order_id={} item_id={} quantity={} remaining={} service=inventory-service",
                order.orderId(), order.itemId(), order.quantity(), item.getStock());

        publisher.publish(
                InventoryEventPublisher.TOPIC_INVENTORY_RESERVED,
                order.orderId(),
                InventoryReservedEvent.of(order, now()));
    }

    /**
     * Handles payment.failed: compensating transaction that returns previously
     * reserved stock. Item and quantity are recovered from our own RESERVED
     * audit row. Idempotent - a second release for the same order is skipped.
     */
    @Transactional
    public void release(String orderIdRaw) {
        UUID orderId = UUID.fromString(orderIdRaw);

        if (eventRepository.existsByOrderIdAndAction(orderId, InventoryEvent.Action.RELEASED.name())) {
            log.info("release already applied, skipping order_id={} service=inventory-service", orderIdRaw);
            return;
        }

        Optional<InventoryEvent> reserved =
                eventRepository.findByOrderIdAndAction(orderId, InventoryEvent.Action.RESERVED.name());
        if (reserved.isEmpty()) {
            // Nothing was ever reserved for this order (e.g. it failed at the
            // stock check). No compensation needed.
            log.info("no reservation to release order_id={} service=inventory-service", orderIdRaw);
            return;
        }

        InventoryEvent res = reserved.get();
        Inventory item = inventoryRepository.findByIdForUpdate(res.getItemId())
                .orElseThrow(() -> new IllegalStateException(
                        "reserved item missing from inventory: " + res.getItemId()));

        item.setStock(item.getStock() + res.getQuantity());
        item.setReserved(Math.max(0, item.getReserved() - res.getQuantity()));
        item.setUpdatedAt(OffsetDateTime.now(ZoneOffset.UTC));
        inventoryRepository.save(item);

        eventRepository.save(new InventoryEvent(
                orderId, res.getItemId(), res.getQuantity(), InventoryEvent.Action.RELEASED));

        log.info("stock released order_id={} item_id={} quantity={} restored_to={} service=inventory-service",
                orderIdRaw, res.getItemId(), res.getQuantity(), item.getStock());
    }

    /** Current stock levels for the admin endpoint and demo tooling. */
    @Transactional(readOnly = true)
    public java.util.List<Inventory> listStock() {
        return inventoryRepository.findAll();
    }

    /**
     * Admin/demo helper: add stock back to an item so a depleted demo can be
     * re-run. Records a REPLENISHED audit row. Uses a row lock for consistency
     * with the reservation path.
     */
    @Transactional
    public Inventory replenish(String itemId, int quantity) {
        if (quantity <= 0) {
            throw new IllegalArgumentException("quantity must be positive");
        }
        Inventory item = inventoryRepository.findByIdForUpdate(itemId)
                .orElseThrow(() -> new IllegalArgumentException("unknown item: " + itemId));
        item.setStock(item.getStock() + quantity);
        item.setUpdatedAt(OffsetDateTime.now(ZoneOffset.UTC));
        inventoryRepository.save(item);

        eventRepository.save(new InventoryEvent(
                UUID.randomUUID(), itemId, quantity, InventoryEvent.Action.REPLENISHED));

        log.info("stock replenished item_id={} quantity={} new_total={} service=inventory-service",
                itemId, quantity, item.getStock());
        return item;
    }

    private void publishFailed(OrderCreatedEvent order, int available) {
        publisher.publish(
                InventoryEventPublisher.TOPIC_INVENTORY_FAILED,
                order.orderId(),
                InventoryFailedEvent.outOfStock(order, available, now()));
    }

    private String now() {
        return OffsetDateTime.now(ZoneOffset.UTC).format(DateTimeFormatter.ISO_OFFSET_DATE_TIME);
    }
}
