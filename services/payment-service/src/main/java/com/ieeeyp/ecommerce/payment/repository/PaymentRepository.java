package com.ieeeyp.ecommerce.payment.repository;

import com.ieeeyp.ecommerce.payment.domain.Payment;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.UUID;

public interface PaymentRepository extends JpaRepository<Payment, UUID> {

    /**
     * Idempotency guard: has a payment already been recorded for this order?
     * Used to skip duplicate Kafka deliveries before charging the gateway.
     */
    boolean existsByOrderId(UUID orderId);
}
