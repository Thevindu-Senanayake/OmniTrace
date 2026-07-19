package com.ieeeyp.ecommerce.payment.domain;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.UUID;

/**
 * A single payment attempt for an order — the financial audit trail. order_id
 * is UNIQUE at the database level so a duplicate Kafka delivery cannot create a
 * second charge for the same order.
 */
@Entity
@Table(name = "payments")
@Getter
@Setter
@NoArgsConstructor
public class Payment {

    public enum Status {
        SUCCESS, FAILED
    }

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "order_id", nullable = false, unique = true)
    private UUID orderId;

    @Column(name = "customer_id", nullable = false)
    private String customerId;

    @Column(nullable = false)
    private BigDecimal amount;

    @Column(nullable = false, length = 32)
    private String status;

    @Column(name = "failure_reason", length = 64)
    private String failureReason;

    @Column(name = "transaction_id", length = 64)
    private String transactionId;

    @Column(name = "processing_ms")
    private Integer processingMs;

    // Populated by the DB default (now()); never written by the app.
    @Column(name = "created_at", insertable = false, updatable = false)
    private OffsetDateTime createdAt;
}
