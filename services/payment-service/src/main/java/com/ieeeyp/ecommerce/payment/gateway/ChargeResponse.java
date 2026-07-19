package com.ieeeyp.ecommerce.payment.gateway;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

/** Response body from the gateway's POST /charge on approval. */
@JsonIgnoreProperties(ignoreUnknown = true)
public record ChargeResponse(String transactionId, String status, int processingMs) {
}
