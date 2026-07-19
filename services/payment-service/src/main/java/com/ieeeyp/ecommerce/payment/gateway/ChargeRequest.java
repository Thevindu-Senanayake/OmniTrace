package com.ieeeyp.ecommerce.payment.gateway;

import java.math.BigDecimal;

/** Request body for the gateway's POST /charge. */
public record ChargeRequest(String orderId, String customerId, BigDecimal amount) {
}
