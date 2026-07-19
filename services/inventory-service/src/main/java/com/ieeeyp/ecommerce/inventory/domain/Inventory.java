package com.ieeeyp.ecommerce.inventory.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.time.OffsetDateTime;

/**
 * Current stock level for one catalog item. The reservation flow loads this
 * row with a pessimistic write lock (SELECT ... FOR UPDATE) so concurrent
 * orders for the same item genuinely serialize at the database.
 */
@Entity
@Table(name = "inventory")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
public class Inventory {

    @Id
    @Column(name = "item_id")
    private String itemId;

    @Column(nullable = false)
    private int stock;

    @Column(nullable = false)
    private int reserved;

    @Column(name = "updated_at")
    private OffsetDateTime updatedAt;
}
