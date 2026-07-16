package connections

import (
	"context"
	"testing"

	helperconfig "github.com/radian-solusi/go-helpers/config"
)

func TestNewRedisOptions(t *testing.T) {
	r := NewRedis(helperconfig.RedisConfig{Host: "127.0.0.1", Port: 6379, Password: "x", DB: 2})
	defer r.Close()
	opts := r.Client().Options()
	if opts.Addr != "127.0.0.1:6379" || opts.DB != 2 || opts.Protocol != 2 {
		t.Fatalf("options: %+v", opts)
	}
}

func TestMongoURI(t *testing.T) {
	tests := []struct {
		name    string
		cfg     helperconfig.MongoDBConfig
		want    string
		wantErr bool
	}{
		{"uri", helperconfig.MongoDBConfig{URI: "mongodb://example:27017"}, "mongodb://example:27017", false},
		{"host", helperconfig.MongoDBConfig{Host: "localhost"}, "mongodb://localhost:27017", false},
		{"credentials", helperconfig.MongoDBConfig{Host: "localhost", Port: 27018, User: "u", Password: "p"}, "mongodb://u:p@localhost:27018", false},
		{"missing", helperconfig.MongoDBConfig{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mongoURI(tt.cfg)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("got %q,%v", got, err)
			}
		})
	}
}

func TestLocalS3RoundTrip(t *testing.T) {
	c, err := NewS3Client(context.Background(), helperconfig.S3Config{Provider: helperconfig.S3ProviderLocal, LocalPath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.UploadFile(context.Background(), "docs/a.txt", []byte("hello"), "text/plain"); err != nil {
		t.Fatal(err)
	}
	got, err := c.DownloadFile(context.Background(), "docs/a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q", got)
	}
	if _, err := c.DownloadFile(context.Background(), "../escape"); err == nil {
		t.Fatal("expected traversal rejection")
	}
}

func TestNewGPubSubDisabled(t *testing.T) {
	ps, err := NewGPubSub(context.Background(), helperconfig.GPubSubConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if ps.IsConnected() {
		t.Fatal("must be disconnected")
	}
	if _, err := ps.Publish(context.Background(), "topic", []byte("x"), nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewSFTPClientValidation(t *testing.T) {
	for _, cfg := range []helperconfig.SftpConfig{{}, {Host: "x"}, {Host: "x", User: "u"}} {
		if _, err := NewSFTPClient(cfg); err == nil {
			t.Fatalf("expected error for %+v", cfg)
		}
	}
}

func TestTelemetryLifecycle(t *testing.T) {
	telemetry, err := NewTelemetry(context.Background(), helperconfig.OtelConfig{ServiceName: "test", ServiceVersion: "1"}, false)
	if err != nil {
		t.Fatal(err)
	}
	_, span := telemetry.StartSpan(context.Background(), "unit")
	span.End()
	if err := telemetry.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
