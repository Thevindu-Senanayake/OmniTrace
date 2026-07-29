# Chaos Scenarios

Toxiproxy sits between Payment Service and Payment Gateway. Routing is
declarative (`toxiproxy-config.json`, loaded at startup); **toxics are applied at
runtime** through the REST API, so chaos is toggled live during a demo without a
redeploy or a pod restart.

The proxy starts with **no toxics**, which makes it a transparent wire. That is
deliberate: a chaos setup you cannot fully switch off gives you nothing to
compare against.

## Reaching the API

| Deployment | API address |
|---|---|
| Docker Compose | `http://127.0.0.1:8474` |
| Kubernetes | `kubectl port-forward svc/toxiproxy 8474:8474` |

Every command below assumes `TOXI=http://127.0.0.1:8474`.

```bash
export TOXI=http://127.0.0.1:8474
```

Confirm the proxy is up and currently clean:

```bash
curl -s $TOXI/proxies | jq
```

---

## Scenario 1 — Slow gateway (latency, still succeeds)

Adds latency below the client's read timeout. Payments still succeed, but the
latency is visible as a slow span. This is the "latency spike you click into"
case the dashboard is built around.

```bash
curl -s -X POST $TOXI/proxies/payment-gateway/toxics -d '{
  "name": "latency_downstream",
  "type": "latency",
  "stream": "downstream",
  "toxicity": 1.0,
  "attributes": { "latency": 1500, "jitter": 500 }
}'
```

Expected: orders still reach `CONFIRMED`; `payments.processing_ms` jumps from
~200ms to ~1500-2000ms.

## Scenario 2 — Gateway timeout (TIMEOUT)

Delays past the client's read timeout, so Payment Service gives up.

```bash
curl -s -X POST $TOXI/proxies/payment-gateway/toxics -d '{
  "name": "timeout_downstream",
  "type": "latency",
  "stream": "downstream",
  "toxicity": 1.0,
  "attributes": { "latency": 10000, "jitter": 0 }
}'
```

Expected: `payment.failed` with `reason=TIMEOUT`, order → `PAYMENT_FAILED`,
inventory releases the reserved stock.

Note this is an *ambiguous* failure: the gateway may still approve the charge
after the client stops listening. Watch the payment-gateway logs during this
scenario — you will see `charge approved` for an order that Payment Service has
already recorded as failed. That divergence is real, and is what idempotency
keys and reconciliation exist to solve.

## Scenario 3 — Gateway down (CONNECTION_ERROR)

Drops connections outright.

```bash
curl -s -X POST $TOXI/proxies/payment-gateway/toxics -d '{
  "name": "drop_all",
  "type": "reset_peer",
  "stream": "downstream",
  "toxicity": 1.0,
  "attributes": { "timeout": 0 }
}'
```

Expected: `payment.failed` with `reason=CONNECTION_ERROR`.

For a partial outage — the more realistic and more interesting case — set
`toxicity` to `0.2` so only 20% of connections are dropped. Under k6 load this
produces a mix of successes and failures rather than a clean binary.

## Scenario 4 — Bandwidth throttle

Useful when demonstrating that slowness is not always a clean timeout.

```bash
curl -s -X POST $TOXI/proxies/payment-gateway/toxics -d '{
  "name": "slow_pipe",
  "type": "bandwidth",
  "stream": "downstream",
  "toxicity": 1.0,
  "attributes": { "rate": 1 }
}'
```

`rate` is in KB/s.

---

## Inspecting and resetting

List active toxics:

```bash
curl -s $TOXI/proxies/payment-gateway | jq '.toxics'
```

Remove one toxic:

```bash
curl -s -X DELETE $TOXI/proxies/payment-gateway/toxics/latency_downstream
```

Reset everything back to a clean wire — **run this between scenarios**, or the
next measurement inherits the previous one's faults:

```bash
curl -s $TOXI/proxies/payment-gateway \
  | jq -r '.toxics[].name' \
  | xargs -I{} curl -s -X DELETE $TOXI/proxies/payment-gateway/toxics/{}
```

Simulate the proxy itself vanishing (distinct from the gateway failing):

```bash
curl -s -X POST $TOXI/proxies/payment-gateway -d '{"enabled": false}'
curl -s -X POST $TOXI/proxies/payment-gateway -d '{"enabled": true}'
```

---

## `stream` direction

`downstream` is gateway → client (the response). `upstream` is client → gateway
(the request). For latency it rarely matters which you pick, but for bandwidth
and slow-close it does. The scenarios above use `downstream` because a slow or
failing *response* is what a struggling payment processor actually looks like.
