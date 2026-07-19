package com.ieeeyp.ecommerce.inventory.repository;

import com.ieeeyp.ecommerce.inventory.domain.InventoryEvent;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;
import java.util.UUID;

public interface InventoryEventRepository extends JpaRepository<InventoryEvent, UUID> {

    /**
     * Idempotency check: has this order already had the given action applied?
     * Used to short-circuit duplicate Kafka deliveries before touching stock.
     */
    boolean existsByOrderIdAndAction(UUID orderId, String action);

    /**
     * Recovers the item and quantity a compensating release must restore, from
     * this service's own RESERVED audit row — payment.failed carries only the
     * orderId.
     */
    Optional<InventoryEvent> findByOrderIdAndAction(UUID orderId, String action);
}
