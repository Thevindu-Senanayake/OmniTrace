package com.ieeeyp.ecommerce.payment.gateway;

import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.io.InterruptedIOException;
import java.math.BigDecimal;
import java.net.ConnectException;
import java.net.SocketTimeoutException;

/**
 * Calls the Payment Gateway over synchronous HTTP. This outbound call is the
 * boundary Toxiproxy intercepts, so network faults are mapped to the reasons in
 * the payment.failed contract:
 *
 * <ul>
 *   <li>connection refused / reset → {@code CONNECTION_ERROR}</li>
 *   <li>read timeout (e.g. Toxiproxy 5s delay) → {@code TIMEOUT}</li>
 * </ul>
 *
 * There is no chaos code here — these faults are produced by Toxiproxy in front
 * of the gateway; this client only classifies them.
 */
@Slf4j
@Component
@EnableConfigurationProperties(GatewayProperties.class)
public class PaymentGatewayClient {

    public static final String REASON_CONNECTION_ERROR = "CONNECTION_ERROR";
    public static final String REASON_TIMEOUT = "TIMEOUT";
    public static final String REASON_DECLINED = "DECLINED";

    private final RestClient restClient;

    public PaymentGatewayClient(GatewayProperties props) {
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(props.connectTimeoutMs());
        factory.setReadTimeout(props.readTimeoutMs());
        this.restClient = RestClient.builder()
                .baseUrl(props.url())
                .requestFactory(factory)
                .build();
    }

    public ChargeOutcome charge(String orderId, String customerId, BigDecimal amount) {
        long startedAt = System.nanoTime();
        try {
            ChargeResponse resp = restClient.post()
                    .uri("/charge")
                    .body(new ChargeRequest(orderId, customerId, amount))
                    .retrieve()
                    .body(ChargeResponse.class);

            int ms = elapsedMs(startedAt);
            if (resp == null || resp.transactionId() == null) {
                log.warn("gateway returned no transaction order_id={} service=payment-service", orderId);
                return ChargeOutcome.failed(REASON_DECLINED, ms);
            }
            return ChargeOutcome.ok(resp.transactionId(), ms);

        } catch (Exception e) {
            int ms = elapsedMs(startedAt);
            String reason = classify(e);
            log.warn("gateway charge failed order_id={} reason={} error={} service=payment-service",
                    orderId, reason, e.getMessage());
            return ChargeOutcome.failed(reason, ms);
        }
    }

    /**
     * Maps a transport exception to a payment.failed reason. Read timeouts
     * (including Toxiproxy's injected delay) surface as SocketTimeoutException;
     * refused/reset connections as ConnectException.
     */
    private String classify(Throwable e) {
        for (Throwable t = e; t != null; t = t.getCause()) {
            if (t instanceof SocketTimeoutException || t instanceof InterruptedIOException) {
                return REASON_TIMEOUT;
            }
            if (t instanceof ConnectException) {
                return REASON_CONNECTION_ERROR;
            }
        }
        // Any other transport-level failure (reset, DNS, HTTP error) is treated
        // as a connection problem for the gateway.
        return REASON_CONNECTION_ERROR;
    }

    private int elapsedMs(long startedNanos) {
        return (int) ((System.nanoTime() - startedNanos) / 1_000_000);
    }
}
