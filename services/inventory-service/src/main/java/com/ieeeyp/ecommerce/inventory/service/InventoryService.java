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
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Core inventory business logic: reserve stock for a multi-item order and
 * release it when a downstream payment fails. All mutations run inside a single
 * transaction that holds row-level locks on the affected inventory rows, so
 * concurrent orders for the same item serialize at the database — no simulation
 * code. Reservation is all-or-nothing: either every line is reserved or none is.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class InventoryService {

    private final InventoryRepository inventoryRepository;
    private final InventoryEventRepository eventRepository;
    private final InventoryEventPublisher publisher;

    /**
     * Handles order.created: reserve stock for all lines and publish
     * inventory.reserved on success, or inventory.failed (with per-line
     * shortfalls) when any line is short / unknown. Idempotent — a redelivered
     * order that was already reserved is a no-op.
     */
    @Transactional
    public void reserve(OrderCreatedEvent order) {
        UUID orderId = UUID.fromString(order.orderId());
        List<OrderCreatedEvent.Item> items = order.items() == null ? List.of() : order.items();

        if (items.isEmpty()) {
            log.warn("order with no items, failing order_id={} service=inventory-service", order.orderId());
            publisher.publish(InventoryEventPublisher.TOPIC_INVENTORY_FAILED, order.orderId(),
                    InventoryFailedEvent.outOfStock(order.orderId(), List.of(), now()));
            return;
        }

        // Idempotency: if any line was already reserved for this order, the
        // whole order was processed (reservation is atomic). Skip.
        if (eventRepository.existsByOrderIdAndAction(orderId, InventoryEvent.Action.RESERVED.name())) {
            log.info("reservation already applied, skipping order_id={} service=inventory-service", order.orderId());
            return;
        }

        // Lock every item row in a deterministic order (sorted by itemId) so
        // concurrent multi-item orders can never deadlock against each other.
        List<OrderCreatedEvent.Item> ordered = new ArrayList<>(items);
        ordered.sort(Comparator.comparing(OrderCreatedEvent.Item::itemId));

        List<Inventory> lockedRows = new ArrayList<>(ordered.size());
        List<InventoryFailedEvent.FailedItem> shortfalls = new ArrayList<>();
        for (OrderCreatedEvent.Item line : ordered) {
            Optional<Inventory> maybe = inventoryRepository.findByIdForUpdate(line.itemId());
            if (maybe.isEmpty()) {
                shortfalls.add(new InventoryFailedEvent.FailedItem(line.itemId(), line.quantity(), 0));
                continue;
            }
            Inventory row = maybe.get();
            if (row.getStock() < line.quantity()) {
                shortfalls.add(new InventoryFailedEvent.FailedItem(line.itemId(), line.quantity(), row.getStock()));
            }
            lockedRows.add(row);
        }

        if (!shortfalls.isEmpty()) {
            // All-or-nothing: touch no stock, let the transaction roll back the
            // locks, and report every shortfall.
            log.info("insufficient stock, failing order order_id={} shortfalls={} service=inventory-service",
                    order.orderId(), shortfalls.size());
            publisher.publish(InventoryEventPublisher.TOPIC_INVENTORY_FAILED, order.orderId(),
                    InventoryFailedEvent.outOfStock(order.orderId(), shortfalls, now()));
            return;
        }

        // Every line is satisfiable — commit the reservation for all of them.
        OffsetDateTime ts = OffsetDateTime.now(ZoneOffset.UTC);
        for (OrderCreatedEvent.Item line : ordered) {
            Inventory row = lockedRows.stream()
                    .filter(r -> r.getItemId().equals(line.itemId()))
                    .findFirst()
                    .orElseThrow();
            row.setStock(row.getStock() - line.quantity());
            row.setReserved(row.getReserved() + line.quantity());
            row.setUpdatedAt(ts);
            inventoryRepository.save(row);

            eventRepository.save(new InventoryEvent(
                    orderId, line.itemId(), line.quantity(), InventoryEvent.Action.RESERVED));

            log.info("stock reserved order_id={} item_id={} quantity={} remaining={} service=inventory-service",
                    order.orderId(), line.itemId(), line.quantity(), row.getStock());
        }

        publisher.publish(InventoryEventPublisher.TOPIC_INVENTORY_RESERVED, order.orderId(),
                InventoryReservedEvent.of(order, now()));
    }

    /**
     * Handles payment.failed: compensating transaction that returns all stock
     * reserved for the order. Items and quantities are recovered from this
     * service's own RESERVED audit rows. Idempotent — a second release for the
     * same order is skipped.
     */
    @Transactional
    public void release(String orderIdRaw) {
        UUID orderId = UUID.fromString(orderIdRaw);

        if (eventRepository.existsByOrderIdAndAction(orderId, InventoryEvent.Action.RELEASED.name())) {
            log.info("release already applied, skipping order_id={} service=inventory-service", orderIdRaw);
            return;
        }

        List<InventoryEvent> reserved =
                eventRepository.findByOrderIdAndAction(orderId, InventoryEvent.Action.RESERVED.name());
        if (reserved.isEmpty()) {
            // Nothing was ever reserved for this order (e.g. it failed at the
            // stock check). No compensation needed.
            log.info("no reservation to release order_id={} service=inventory-service", orderIdRaw);
            return;
        }

        // Restore in sorted itemId order, consistent with the reserve path.
        reserved.sort(Comparator.comparing(InventoryEvent::getItemId));
        OffsetDateTime ts = OffsetDateTime.now(ZoneOffset.UTC);
        for (InventoryEvent res : reserved) {
            Inventory row = inventoryRepository.findByIdForUpdate(res.getItemId())
                    .orElseThrow(() -> new IllegalStateException(
                            "reserved item missing from inventory: " + res.getItemId()));
            row.setStock(row.getStock() + res.getQuantity());
            row.setReserved(Math.max(0, row.getReserved() - res.getQuantity()));
            row.setUpdatedAt(ts);
            inventoryRepository.save(row);

            eventRepository.save(new InventoryEvent(
                    orderId, res.getItemId(), res.getQuantity(), InventoryEvent.Action.RELEASED));

            log.info("stock released order_id={} item_id={} quantity={} restored_to={} service=inventory-service",
                    orderIdRaw, res.getItemId(), res.getQuantity(), row.getStock());
        }
    }

    /** Current stock levels for the admin endpoint and demo tooling. */
    @Transactional(readOnly = true)
    public List<Inventory> listStock() {
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

    private String now() {
        return OffsetDateTime.now(ZoneOffset.UTC).format(DateTimeFormatter.ISO_OFFSET_DATE_TIME);
    }
}
