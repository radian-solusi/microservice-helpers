package connections

import (
	"context"
	"errors"
	"fmt"
	"os"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type telemetryWrapper struct {
	tracer   trace.Tracer
	shutdown func(context.Context) error
}

func NewTelemetry(ctx context.Context, cfg helperconfig.OtelConfig, isProduction bool) (Telemetry, error) {
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "go-helpers"
	}
	serviceVersion := cfg.ServiceVersion
	if serviceVersion == "" {
		serviceVersion = "0.0.0"
	}

	projectID := resolveProjectID(cfg, os.Getenv)
	target := resolveExporterTarget(cfg)

	res, err := newResource(ctx, serviceName, serviceVersion, projectID)
	if err != nil {
		return nil, fmt.Errorf("otel: create resource: %w", err)
	}

	tpOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if isProduction && shouldExport(target, projectID) {
		exporter, err := newTraceExporter(ctx, target, projectID)
		if err != nil {
			return nil, fmt.Errorf("otel: create OTLP trace exporter: %w", err)
		}
		tpOpts = append(tpOpts, sdktrace.WithBatcher(exporter))
	}
	// Dev, or production without GCP project: spans stay local-only.
	tp := sdktrace.NewTracerProvider(tpOpts...)

	otel.SetTracerProvider(tp)

	return &telemetryWrapper{
		tracer:   tp.Tracer(serviceName),
		shutdown: tp.Shutdown,
	}, nil
}

func (t *telemetryWrapper) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

func (t *telemetryWrapper) Shutdown(ctx context.Context) error {
	if t.shutdown == nil {
		return nil
	}
	err := t.shutdown(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("otel: shutdown: %w", err)
	}
	return nil
}
