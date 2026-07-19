package com.ieeeyp.ecommerce.payment.service;

import com.ieeeyp.ecommerce.payment.domain.Payment;
import com.ieeeyp.ecommerce.payment.gateway.ChargeOutcome;
import com.ieeeyp.ecommerce.payment.gateway.PaymentGatewayClient;
import com.ieeeyp.ecommerce.payment.messaging.PaymentEventPublisher;
import com.ieeeyp.ecommerce.payment.messaging.event.InventoryReservedEvent;
import com.ieeeyp.ecommerce.payment.messaging.event.PaymentFailedEvent;
import com.ieeeyp.ecommerce.payment.messaging.event.PaymentSuccessEvent;
import com.ieeeyp.ecommerce.payment.repository.PaymentRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.UUID;

/**
 * Core payment logic: charge the gateway for a reserved order, record the
 * attempt in payments_db (financial audit trail) and publish the outcome.
 *
 * Idempotent on order_id — a duplicate inventory.reserved delivery for an order
 * already paid is a no-op. The slow, external gateway call is made OUTSIDE any
 * database transaction so a connection is never held open during a Toxiproxy
 * latency injection.
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class PaymentService {

    private final PaymentRepository repository;
    private final PaymentGatewayClient gateway;
    private final PaymentEventPublisher publisher;

    public void process(InventoryReservedEvent event) {
        UUID orderId = UUID.fromString(event.orderId());

        // Idempotency fast-path: already processed → skip before charging.
        if (repository.existsByOrderId(orderId)) {
            log.info("payment already processed, skipping order_id={} service=payment-service", event.orderId());
            return;
        }

        BigDecimal amount = event.totalAmount();
        ChargeOutcome outcome = gateway.charge(event.orderId(), event.customerId(), amount);

        Payment payment = new Payment();
        payment.setOrderId(orderId);
        payment.setCustomerId(event.customerId());
        payment.setAmount(amount);
        payment.setProcessingMs(outcome.processingMs());
        if (outcome.success()) {
            payment.setStatus(Payment.Status.SUCCESS.name());
            payment.setTransactionId(outcome.transactionId());
        } else {
            payment.setStatus(Payment.Status.FAILED.name());
            payment.setFailureReason(outcome.failureReason());
        }

        try {
            repository.saveAndFlush(payment);
        } catch (DataIntegrityViolationException e) {
            // A concurrent delivery inserted the row first (order_id UNIQUE).
            // The other worker owns publishing the outcome — skip.
            log.info("payment row already exists (concurrent delivery), skipping order_id={} service=payment-service",
                    event.orderId());
            return;
        }

        if (outcome.success()) {
            log.info("payment succeeded order_id={} amount={} transaction_id={} processing_ms={} service=payment-service",
                    event.orderId(), amount, outcome.transactionId(), outcome.processingMs());
            publisher.publish(PaymentEventPublisher.TOPIC_PAYMENT_SUCCESS, event.orderId(),
                    PaymentSuccessEvent.of(event.orderId(), payment.getId().toString(), amount, now()));
        } else {
            log.info("payment failed order_id={} amount={} reason={} processing_ms={} service=payment-service",
                    event.orderId(), amount, outcome.failureReason(), outcome.processingMs());
            publisher.publish(PaymentEventPublisher.TOPIC_PAYMENT_FAILED, event.orderId(),
                    PaymentFailedEvent.of(event.orderId(), outcome.failureReason(), amount, now()));
        }
    }

    private String now() {
        return OffsetDateTime.now(ZoneOffset.UTC).format(DateTimeFormatter.ISO_OFFSET_DATE_TIME);
    }
}
