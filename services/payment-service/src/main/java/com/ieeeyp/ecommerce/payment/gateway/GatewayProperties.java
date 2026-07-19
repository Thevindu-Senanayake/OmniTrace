package com.ieeeyp.ecommerce.payment.gateway;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * Binds the {@code payment.gateway.*} configuration. In production the url
 * points at the Toxiproxy listener, not the gateway directly.
 */
@ConfigurationProperties(prefix = "payment.gateway")
public record GatewayProperties(
        String url,
        int connectTimeoutMs,
        int readTimeoutMs) {
}
