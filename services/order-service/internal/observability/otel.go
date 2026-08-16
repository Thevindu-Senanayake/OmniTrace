package observability

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup wires the three OTel signal providers (traces, metrics, logs) to export
// OTLP/gRPC to the collector named by OTEL_EXPORTER_OTLP_ENDPOINT, registers the
// W3C TraceContext + Baggage propagator (this is what carries the trace id across
// the HTTP and Kafka boundaries), and installs all three as process globals so
// otelhttp, otelpgx and the otelslog log bridge pick them up with no wiring.
//
// It must run before logging.New: the otelslog bridge binds to whatever the
// global LoggerProvider is at construction time, so the real provider has to be
// in place first or logs never reach the collector.
//
// The gRPC exporters dial lazily (non-blocking), so Setup returns immediately
// even when the collector is not up yet; signals buffer and the dial is retried.
// The returned shutdown flushes all three providers — the exit flush matters
// most for the last events before a container stops, which is when they count.
func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		// Picks up OTEL_RESOURCE_ATTRIBUTES (service.namespace, deployment.environment).
		resource.WithFromEnv(),
		// service.name is set explicitly so it is correct even if the env var is absent.
		resource.WithAttributes(attribute.String("service.name", serviceName)),
	)
	if err != nil {
		return nil, err
	}

	traceExp, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)

	metricExp, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, errors.Join(err, tp.Shutdown(ctx))
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)

	logExp, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, errors.Join(err, mp.Shutdown(ctx), tp.Shutdown(ctx))
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	logglobal.SetLoggerProvider(lp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx), lp.Shutdown(ctx))
	}, nil
}
