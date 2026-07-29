import http from "k6/http";
import { check } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

/**
 * Load generator for the e-commerce saga.
 *
 * This drives the system from outside — it is a client, not a deployed service.
 * The chaos it produces is *natural*: real row-lock contention from concurrent
 * reservations against the same item, and real stock depletion. Network-level
 * faults are Toxiproxy's job, not this script's (see toxiproxy/scenarios.md);
 * the two compose, and the most interesting runs use both at once.
 *
 * Usage:
 *   k6 run -e SCENARIO=smoke load-gen/k6-script.js
 *   k6 run -e SCENARIO=spike -e BASE_URL=http://127.0.0.1:8080 load-gen/k6-script.js
 *
 * Without a local k6 install:
 *   docker run --rm -i --network host \
 *     -v "${PWD}/load-gen:/scripts" grafana/k6 run -e SCENARIO=smoke /scripts/k6-script.js
 */

const BASE_URL = __ENV.BASE_URL || "http://127.0.0.1:8080";
const SCENARIO = __ENV.SCENARIO || "smoke";

// Catalog item IDs and their seeded stock. Stock levels are what make the
// scenarios behave differently — see docs/events.md.
const ITEMS = {
  plentiful: "00000000-0000-0000-0000-000000000001", // 500
  moderate: "00000000-0000-0000-0000-000000000002", //  100
  hot: "00000000-0000-0000-0000-000000000003", //   20
  flash: "00000000-0000-0000-0000-000000000004", //    5
  ghost: "00000000-0000-0000-0000-000000000099", //    0, always out of stock
};

const scenarios = {
  // Is the wiring alive? One user, minimal load.
  smoke: { executor: "constant-vus", vus: 1, duration: "30s" },

  // Steady-state baseline. Run this with no toxics to get the numbers that
  // later chaos runs are compared against.
  load: {
    executor: "ramping-vus",
    startVUs: 0,
    stages: [
      { duration: "30s", target: 20 },
      { duration: "2m", target: 20 },
      { duration: "30s", target: 0 },
    ],
  },

  // Concurrency against the two scarcest items. This is where PostgreSQL
  // SELECT ... FOR UPDATE genuinely serialises: 50 VUs contend for the same
  // rows, so lock wait time shows up as db.postgresql span duration rather
  // than anything this script simulates.
  spike: { executor: "constant-vus", vus: 50, duration: "1m" },

  // Every outcome the saga can produce, in one run: CONFIRMED, OUT_OF_STOCK,
  // and — with a Toxiproxy toxic active — PAYMENT_FAILED.
  mixed: { executor: "constant-vus", vus: 10, duration: "2m" },
};

if (!scenarios[SCENARIO]) {
  throw new Error(
    `unknown SCENARIO "${SCENARIO}" — expected one of: ${Object.keys(scenarios).join(", ")}`,
  );
}

const accepted = new Counter("orders_accepted");
const rejectedOutOfCatalog = new Counter("orders_unknown_product");
const acceptRate = new Rate("order_accept_rate");
const orderLatency = new Trend("order_latency_ms", true);

export const options = {
  scenarios: { [SCENARIO]: { ...scenarios[SCENARIO], gracefulStop: "10s" } },
  thresholds: {
    // The gateway must stay responsive even while the saga fails downstream.
    // Deliberately not asserting on saga outcomes: under chaos, orders are
    // *meant* to fail, and a threshold on that would make chaos runs "fail".
    http_req_failed: ["rate<0.05"],
    order_latency_ms: ["p(95)<2000"],
  },
};

/** Picks the item mix for the current scenario. */
function pickItems() {
  switch (SCENARIO) {
    case "spike":
      // Hammer the scarce items so reservations collide on the same rows.
      return [
        {
          itemId: Math.random() < 0.5 ? ITEMS.hot : ITEMS.flash,
          quantity: randomIntBetween(1, 3),
        },
      ];

    case "mixed": {
      // Weighted so most orders succeed and a minority go out of stock,
      // which is roughly what a real catalog looks like.
      const roll = Math.random();
      if (roll < 0.1) return [{ itemId: ITEMS.ghost, quantity: 1 }];
      if (roll < 0.3) return [{ itemId: ITEMS.hot, quantity: randomIntBetween(1, 2) }];
      if (roll < 0.5) {
        // Multi-item order: exercises the atomic all-or-nothing reservation.
        return [
          { itemId: ITEMS.plentiful, quantity: 1 },
          { itemId: ITEMS.moderate, quantity: 1 },
        ];
      }
      return [{ itemId: ITEMS.plentiful, quantity: randomIntBetween(1, 3) }];
    }

    case "load":
      return [{ itemId: ITEMS.plentiful, quantity: randomIntBetween(1, 2) }];

    default:
      return [{ itemId: ITEMS.plentiful, quantity: 1 }];
  }
}

export default function () {
  const payload = JSON.stringify({
    customerId: `cust-${__VU}-${__ITER}`,
    items: pickItems(),
  });

  const res = http.post(`${BASE_URL}/order`, payload, {
    headers: { "Content-Type": "application/json" },
    tags: { endpoint: "place_order" },
  });

  orderLatency.add(res.timings.duration);
  acceptRate.add(res.status === 202);

  if (res.status === 202) accepted.add(1);
  if (res.status === 404) rejectedOutOfCatalog.add(1);

  // 202 means accepted, not fulfilled — the saga decides the terminal status
  // asynchronously, so it is deliberately not asserted here. Query orders_db
  // or watch the notification-service logs for outcomes.
  check(res, {
    "order accepted (202)": (r) => r.status === 202,
    "returned an orderId": (r) => {
      if (r.status !== 202) return false;
      try {
        return Boolean(r.json("orderId"));
      } catch {
        return false;
      }
    },
  });
}

export function handleSummary(data) {
  const m = data.metrics;
  const line = (label, value) => `  ${label.padEnd(28)} ${value}`;
  const metric = (name, field, unit = "") => {
    const v = m[name]?.values?.[field];
    return v === undefined ? "n/a" : `${Math.round(v * 100) / 100}${unit}`;
  };

  return {
    stdout: [
      "",
      `scenario: ${SCENARIO}  ->  ${BASE_URL}`,
      line("orders accepted", metric("orders_accepted", "count")),
      line("accept rate", metric("order_accept_rate", "rate")),
      line("unknown product (404)", metric("orders_unknown_product", "count")),
      line("latency p95", metric("order_latency_ms", "p(95)", "ms")),
      line("latency max", metric("order_latency_ms", "max", "ms")),
      "",
      "Terminal saga outcomes are asynchronous — check orders_db for final status.",
      "",
    ].join("\n"),
  };
}
