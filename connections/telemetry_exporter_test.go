package connections

import (
	"context"
	"testing"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
)

func TestResolveExporterTarget(t *testing.T) {
	cases := []struct {
		name string
		cfg  helperconfig.OtelConfig
		want exporterTarget
	}{
		{"empty endpoint defaults to Google", helperconfig.OtelConfig{},
			exporterTarget{Endpoint: "telemetry.googleapis.com:443", GoogleAuth: true}},
		{"google host without port", helperconfig.OtelConfig{Endpoint: "telemetry.googleapis.com"},
			exporterTarget{Endpoint: "telemetry.googleapis.com:443", GoogleAuth: true}},
		{"google https url", helperconfig.OtelConfig{Endpoint: "https://telemetry.googleapis.com/"},
			exporterTarget{Endpoint: "telemetry.googleapis.com:443", GoogleAuth: true}},
		{"google is never insecure", helperconfig.OtelConfig{Endpoint: "http://telemetry.googleapis.com:443", Insecure: true},
			exporterTarget{Endpoint: "telemetry.googleapis.com:443", GoogleAuth: true}},
		{"collector host:port", helperconfig.OtelConfig{Endpoint: "otel-collector:4317"},
			exporterTarget{Endpoint: "otel-collector:4317"}},
		{"collector insecure flag", helperconfig.OtelConfig{Endpoint: "otel-collector:4317", Insecure: true},
			exporterTarget{Endpoint: "otel-collector:4317", Insecure: true}},
		{"collector http url implies insecure", helperconfig.OtelConfig{Endpoint: "http://otel-collector:4317"},
			exporterTarget{Endpoint: "otel-collector:4317", Insecure: true}},
		{"collector without port gets 4317", helperconfig.OtelConfig{Endpoint: "otel-collector"},
			exporterTarget{Endpoint: "otel-collector:4317"}},
		{"whitespace trimmed", helperconfig.OtelConfig{Endpoint: "  otel-collector:4317 \n"},
			exporterTarget{Endpoint: "otel-collector:4317"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveExporterTarget(tc.cfg); got != tc.want {
				t.Fatalf("resolveExporterTarget(%+v) = %+v, want %+v", tc.cfg, got, tc.want)
			}
		})
	}
}

func TestResolveProjectID(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

	if got := resolveProjectID(helperconfig.OtelConfig{GCPProjectID: "cfg-proj"},
		env(map[string]string{"GOOGLE_CLOUD_PROJECT": "env-proj"})); got != "cfg-proj" {
		t.Errorf("config should win, got %q", got)
	}
	if got := resolveProjectID(helperconfig.OtelConfig{},
		env(map[string]string{"GOOGLE_CLOUD_PROJECT": " env-proj "})); got != "env-proj" {
		t.Errorf("env fallback (trimmed) expected, got %q", got)
	}
	if got := resolveProjectID(helperconfig.OtelConfig{GCPProjectID: "  "}, env(nil)); got != "" {
		t.Errorf("blank config and no env should be empty, got %q", got)
	}
}

func TestShouldExport(t *testing.T) {
	google := exporterTarget{Endpoint: "telemetry.googleapis.com:443", GoogleAuth: true}
	collector := exporterTarget{Endpoint: "otel-collector:4317", Insecure: true}

	if !shouldExport(google, "proj") {
		t.Error("google + project should export")
	}
	if shouldExport(google, "") {
		t.Error("google without project must not export")
	}
	if !shouldExport(collector, "") {
		t.Error("collector exports even without project")
	}
}

func resourceAttrs(t *testing.T, serviceName, version, project string) map[string]string {
	t.Helper()
	res, err := newResource(context.Background(), serviceName, version, project)
	if err != nil {
		t.Fatalf("newResource: %v", err)
	}
	out := map[string]string{}
	for _, kv := range res.Attributes() {
		out[string(kv.Key)] = kv.Value.Emit()
	}
	return out
}

func TestNewResource_IncludesServiceAndProject(t *testing.T) {
	a := resourceAttrs(t, "svc", "1.2.3", "proj-x")
	if a["service.name"] != "svc" || a["service.version"] != "1.2.3" || a["gcp.project_id"] != "proj-x" {
		t.Fatalf("unexpected attrs: %v", a)
	}
}

func TestNewResource_OmitsEmptyProject(t *testing.T) {
	a := resourceAttrs(t, "svc", "1.2.3", "")
	if _, ok := a["gcp.project_id"]; ok {
		t.Fatalf("gcp.project_id must be absent when project is empty: %v", a)
	}
}
