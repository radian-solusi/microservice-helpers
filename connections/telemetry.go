package connections

import (
	"context"
	"errors"
	"fmt"
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	helperconfig "github.com/radian-solusi/go-helpers/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: create resource: %w", err)
	}

	var tp *sdktrace.TracerProvider
	if isProduction {
		projectID := cfg.GCPProjectID
		if projectID == "" {
			projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
		if projectID == "" {
			// Graceful fallback: production without GCP project → no exporter
			tp = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		} else {
			exporter, err := texporter.New(texporter.WithProjectID(projectID))
			if err != nil {
				return nil, fmt.Errorf("otel: create GCP trace exporter: %w", err)
			}
			tp = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(exporter),
				sdktrace.WithResource(res),
			)
		}
	} else {
		// Dev: no exporter, spans are local-only.
		tp = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
	}

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
