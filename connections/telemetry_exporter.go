package connections

import (
	"context"
	"net"
	"strings"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	googleTelemetryHost     = "telemetry.googleapis.com"
	googleTelemetryEndpoint = "telemetry.googleapis.com:443"
	defaultCollectorPort    = "4317"
)

type exporterTarget struct {
	Endpoint   string // host:port, never a URL
	GoogleAuth bool   // attach ADC per-RPC credentials + x-goog-user-project
	Insecure   bool   // plaintext gRPC (collector only)
}

func resolveExporterTarget(cfg helperconfig.OtelConfig) exporterTarget {
	e := strings.TrimSpace(cfg.Endpoint)
	if e == "" {
		return exporterTarget{Endpoint: googleTelemetryEndpoint, GoogleAuth: true}
	}

	plaintext := false
	switch {
	case strings.HasPrefix(e, "https://"):
		e = strings.TrimPrefix(e, "https://")
	case strings.HasPrefix(e, "http://"):
		e = strings.TrimPrefix(e, "http://")
		plaintext = true
	}
	e = strings.TrimSuffix(e, "/")

	host, port, err := net.SplitHostPort(e)
	if err != nil {
		host, port = e, ""
	}

	// Google is never insecure: gRPC refuses per-RPC creds over plaintext.
	if host == googleTelemetryHost {
		if port == "" {
			port = "443"
		}
		return exporterTarget{Endpoint: net.JoinHostPort(host, port), GoogleAuth: true}
	}

	if port == "" {
		port = defaultCollectorPort
	}
	return exporterTarget{
		Endpoint: net.JoinHostPort(host, port),
		Insecure: cfg.Insecure || plaintext,
	}
}

func resolveProjectID(cfg helperconfig.OtelConfig, lookupEnv func(string) string) string {
	if id := strings.TrimSpace(cfg.GCPProjectID); id != "" {
		return id
	}
	return strings.TrimSpace(lookupEnv("GOOGLE_CLOUD_PROJECT"))
}

func shouldExport(t exporterTarget, projectID string) bool {
	return !t.GoogleAuth || projectID != ""
}

func newResource(ctx context.Context, serviceName, serviceVersion, projectID string) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
	}
	if projectID != "" {
		attrs = append(attrs, attribute.String("gcp.project_id", projectID))
	}
	return resource.New(ctx, resource.WithAttributes(attrs...))
}
