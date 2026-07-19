package com.ieeeyp.ecommerce.inventory.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.time.OffsetDateTime;
import java.util.UUID;

/**
 * Append-only audit record of a stock action. The unique (order_id, action)
 * constraint doubles as the idempotency guard: a duplicate Kafka delivery that
 * tries to insert the same RESERVED/RELEASED row is rejected by the database.
 */
@Entity
@Table(name = "inventory_events")
@Getter
@Setter
@NoArgsConstructor
public class InventoryEvent {

    public enum Action {
        RESERVED, RELEASED, REPLENISHED
    }

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "order_id", nullable = false)
    private UUID orderId;

    @Column(name = "item_id", nullable = false)
    private String itemId;

    @Column(nullable = false)
    private int quantity;

    @Column(nullable = false, length = 32)
    private String action;

    // Populated by the DB default (now()); never written by the app.
    @Column(name = "created_at", insertable = false, updatable = false)
    private OffsetDateTime createdAt;

    public InventoryEvent(UUID orderId, String itemId, int quantity, Action action) {
        this.orderId = orderId;
        this.itemId = itemId;
        this.quantity = quantity;
        this.action = action.name();
    }
}
