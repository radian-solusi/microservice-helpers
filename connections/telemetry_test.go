package connections

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakeOTLPReceiver struct {
	coltracepb.UnimplementedTraceServiceServer
	mu   sync.Mutex
	reqs []*coltracepb.ExportTraceServiceRequest
	mds  []metadata.MD
}

func (r *fakeOTLPReceiver) Export(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	r.mu.Lock()
	r.reqs = append(r.reqs, req)
	r.mds = append(r.mds, md)
	r.mu.Unlock()
	return &coltracepb.ExportTraceServiceResponse{}, nil
}

// spanNames returns every span name received, in order.
func (r *fakeOTLPReceiver) spanNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var names []string
	for _, req := range r.reqs {
		for _, rs := range req.GetResourceSpans() {
			for _, ss := range rs.GetScopeSpans() {
				for _, s := range ss.GetSpans() {
					names = append(names, s.GetName())
				}
			}
		}
	}
	return names
}

// resourceAttr returns the string value of key on the first received resource.
func (r *fakeOTLPReceiver) resourceAttr(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, req := range r.reqs {
		for _, rs := range req.GetResourceSpans() {
			for _, kv := range rs.GetResource().GetAttributes() {
				if kv.GetKey() == key {
					return kv.GetValue().GetStringValue()
				}
			}
		}
	}
	return ""
}

// header returns the first value of a gRPC metadata key on the first request.
func (r *fakeOTLPReceiver) header(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.mds) == 0 || len(r.mds[0].Get(key)) == 0 {
		return ""
	}
	return r.mds[0].Get(key)[0]
}

func startFakeOTLPReceiver(t *testing.T) (string, *fakeOTLPReceiver) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	rec := &fakeOTLPReceiver{}
	coltracepb.RegisterTraceServiceServer(srv, rec)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String(), rec
}

func TestNewTelemetry_ProductionCollectorExportsSpans(t *testing.T) {
	addr, rec := startFakeOTLPReceiver(t)
	cfg := helperconfig.OtelConfig{ServiceName: "svc", ServiceVersion: "1.2.3",
		GCPProjectID: "proj-x", Endpoint: addr, Insecure: true}

	tel, err := NewTelemetry(context.Background(), cfg, true)
	if err != nil {
		t.Fatalf("NewTelemetry: %v", err)
	}
	_, span := tel.StartSpan(context.Background(), "op")
	span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tel.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	if names := rec.spanNames(); len(names) != 1 || names[0] != "op" {
		t.Fatalf("received spans = %v, want [op]", names)
	}
	if got := rec.resourceAttr("service.name"); got != "svc" {
		t.Errorf("service.name = %q, want svc", got)
	}
	if got := rec.resourceAttr("gcp.project_id"); got != "proj-x" {
		t.Errorf("gcp.project_id = %q, want proj-x", got)
	}
	if got := rec.header("x-goog-user-project"); got != "" {
		t.Errorf("collector target must not send x-goog-user-project, got %q", got)
	}
}

func TestNewTelemetry_ProductionGoogleWithoutCredentialsFails(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/nonexistent/adc.json")
	tel, err := NewTelemetry(context.Background(),
		helperconfig.OtelConfig{ServiceName: "svc", GCPProjectID: "proj-x"}, true)
	if err == nil {
		t.Fatal("expected error when ADC is unavailable")
	}
	if tel != nil {
		t.Fatal("telemetry must be nil on error")
	}
}

func TestNewTelemetry_ProductionWithoutProjectIDHasNoExporter(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/nonexistent/adc.json") // proves ADC is never touched
	tel, err := NewTelemetry(context.Background(), helperconfig.OtelConfig{ServiceName: "svc"}, true)
	if err != nil {
		t.Fatalf("NewTelemetry: %v", err)
	}
	_, span := tel.StartSpan(context.Background(), "op")
	span.End()
	if err := tel.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestNewTelemetry_DevIgnoresEndpoint(t *testing.T) {
	addr, rec := startFakeOTLPReceiver(t)
	tel, err := NewTelemetry(context.Background(),
		helperconfig.OtelConfig{ServiceName: "svc", GCPProjectID: "proj-x", Endpoint: addr, Insecure: true}, false)
	if err != nil {
		t.Fatalf("NewTelemetry: %v", err)
	}
	_, span := tel.StartSpan(context.Background(), "op")
	span.End()
	_ = tel.Shutdown(context.Background())
	if names := rec.spanNames(); len(names) != 0 {
		t.Fatalf("dev mode must not export, got %v", names)
	}
}

func TestNewTelemetry_ShutdownBoundedWhenCollectorDown(t *testing.T) {
	cfg := helperconfig.OtelConfig{ServiceName: "svc", Endpoint: "127.0.0.1:1", Insecure: true}
	tel, err := NewTelemetry(context.Background(), cfg, true)
	if err != nil {
		t.Fatalf("construction must not dial eagerly: %v", err)
	}
	_, span := tel.StartSpan(context.Background(), "op")
	span.End()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	_ = tel.Shutdown(ctx) // an error is acceptable; hanging is not
	if elapsed := time.Since(start); elapsed > 4*time.Second {
		t.Fatalf("Shutdown took %v, want <= ~2s (ctx deadline)", elapsed)
	}
}

func TestNewTraceExporter_GoogleAuthWithoutCredentialsFails(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/nonexistent/adc.json")
	_, err := newTraceExporter(context.Background(),
		exporterTarget{Endpoint: googleTelemetryEndpoint, GoogleAuth: true}, "proj-x")
	if err == nil {
		t.Fatal("expected ADC error")
	}
}
