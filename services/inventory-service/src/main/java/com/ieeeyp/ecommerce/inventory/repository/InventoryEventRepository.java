package com.ieeeyp.ecommerce.inventory.repository;

import com.ieeeyp.ecommerce.inventory.domain.InventoryEvent;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.UUID;

public interface InventoryEventRepository extends JpaRepository<InventoryEvent, UUID> {

    /**
     * Idempotency check: has this order already had the given action applied to
     * any line? Reservation is atomic, so one matching row means the whole
     * order was processed.
     */
    boolean existsByOrderIdAndAction(UUID orderId, String action);

    /**
     * Recovers the lines a compensating release must restore, from this
     * service's own RESERVED audit rows — payment.failed carries only the
     * orderId.
     */
    List<InventoryEvent> findByOrderIdAndAction(UUID orderId, String action);
}
