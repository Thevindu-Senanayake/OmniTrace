/**
 * OpenTelemetry bootstrap.
 *
 * This module must be imported before anything else in main.ts: the auto
 * instrumentations patch modules like http and express at require time, so
 * anything loaded first is never instrumented and silently produces no spans.
 *
 * Exporting is OTLP to the collector only — the service knows nothing about
 * Tempo, Loki or Prometheus.
 */
import { NodeSDK } from "@opentelemetry/sdk-node";
import { getNodeAutoInstrumentations } from "@opentelemetry/auto-instrumentations-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-grpc";
import { OTLPLogExporter } from "@opentelemetry/exporter-logs-otlp-grpc";
import { OTLPMetricExporter } from "@opentelemetry/exporter-metrics-otlp-grpc";
import { PeriodicExportingMetricReader } from "@opentelemetry/sdk-metrics";
import { BatchLogRecordProcessor } from "@opentelemetry/sdk-logs";
import { resourceFromAttributes } from "@opentelemetry/resources";
import {
	ATTR_SERVICE_NAME,
	ATTR_SERVICE_VERSION,
} from "@opentelemetry/semantic-conventions";

const serviceName = process.env.OTEL_SERVICE_NAME ?? "api-gateway";

const sdk = new NodeSDK({
	resource: resourceFromAttributes({
		[ATTR_SERVICE_NAME]: serviceName,
		[ATTR_SERVICE_VERSION]: process.env.SERVICE_VERSION ?? "0.1.0",
	}),
	traceExporter: new OTLPTraceExporter(),
	logRecordProcessors: [
		new BatchLogRecordProcessor({ exporter: new OTLPLogExporter() }),
	],
	metricReader: new PeriodicExportingMetricReader({
		exporter: new OTLPMetricExporter(),
		exportIntervalMillis: 15000,
	}),
	instrumentations: [
		getNodeAutoInstrumentations({
			// Noisy and of no diagnostic value here — every file read would
			// otherwise become a span.
			"@opentelemetry/instrumentation-fs": { enabled: false },
		}),
	],
});

sdk.start();

// Flush buffered spans on shutdown; without this the final requests before a
// container stops are lost, which is exactly when they matter most.
for (const signal of ["SIGTERM", "SIGINT"] as const) {
	process.on(signal, () => {
		sdk
			.shutdown()
			.catch((err: unknown) => console.error("otel shutdown failed", err))
			.finally(() => process.exit(0));
	});
}
