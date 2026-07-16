package helpers

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/radian-solusi/go-helpers/connections"
)

func (h *Helpers) InitializeSystem() error {
	cfg := h.GetMainConfig()
	db, err := h.factory.NewDatabase(cfg.Database)
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	h.database = db
	h.redis = h.factory.NewRedis(cfg.Redis)

	if cfg.GPubSub.ProjectID != "" || cfg.GPubSub.EmulatorHost != "" {
		if ps, err := h.factory.NewGPubSub(context.Background(), cfg.GPubSub); err != nil {
			log.Println("pubsub init:", err)
		} else {
			h.pubsub = ps
		}
	}
	if cfg.S3.Provider != "" {
		if s3c, err := h.factory.NewS3(context.Background(), cfg.S3); err != nil {
			log.Println("s3 init:", err)
		} else {
			h.s3 = s3c
		}
	}
	if cfg.MongoDB.Host != "" || cfg.MongoDB.URI != "" {
		if m, err := h.factory.NewMongoDB(context.Background(), cfg.MongoDB); err != nil {
			log.Println("mongo init:", err)
		} else {
			h.mongo = m
		}
	}
	if cfg.Otel.Enabled {
		if tel, err := h.factory.NewTelemetry(context.Background(), cfg.Otel, h.IsProduction()); err != nil {
			log.Println("telemetry init:", err)
		} else {
			h.telemetry = tel
		}
	}
	return nil
}

func (h *Helpers) GetDatabase() connections.Database { return h.database }
func (h *Helpers) GetRedisClient() connections.Redis { return h.redis }
func (h *Helpers) GetPubSub() connections.GPubSub    { return h.pubsub }
func (h *Helpers) GetS3Client() connections.S3Client { return h.s3 }
func (h *Helpers) GetMongoDB() connections.MongoDB   { return h.mongo }

func (h *Helpers) SetGormProgressMode(enabled bool) {
	if h.database != nil {
		h.database.SetProgressMode(enabled)
	}
}
func (h *Helpers) RunMigration() error {
	if h.database == nil {
		return errors.New("database not initialized")
	}
	return h.database.Migrate(h.migrationModels...)
}
func (h *Helpers) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if h.telemetry != nil {
		return h.telemetry.StartSpan(ctx, name, opts...)
	}
	return otel.Tracer("go-helpers").Start(ctx, name, opts...)
}
func (h *Helpers) ShutdownTelemetry(ctx context.Context) error {
	if h.telemetry != nil {
		return h.telemetry.Shutdown(ctx)
	}
	return nil
}
