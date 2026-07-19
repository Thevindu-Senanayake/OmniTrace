package com.ieeeyp.ecommerce.payment.web;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/**
 * Liveness endpoint at the stack-wide conventional path {@code GET /health}
 * (used by the docker-compose healthcheck). Distinct from Actuator's
 * {@code /actuator/health}.
 */
@RestController
public class HealthController {

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok", "service", "payment-service");
    }
}
