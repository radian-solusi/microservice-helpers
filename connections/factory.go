package connections

import (
	"context"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
)

type Factory interface {
	NewDatabase(helperconfig.DatabaseConfig) (Database, error)
	NewRedis(helperconfig.RedisConfig) Redis
	NewGPubSub(context.Context, helperconfig.GPubSubConfig) (GPubSub, error)
	NewS3(context.Context, helperconfig.S3Config) (S3Client, error)
	NewMongoDB(context.Context, helperconfig.MongoDBConfig) (MongoDB, error)
	NewSFTP(helperconfig.SftpConfig) (SFTP, error)
	NewTelemetry(context.Context, helperconfig.OtelConfig, bool) (Telemetry, error)
}

type DefaultFactory struct{}

func (DefaultFactory) NewDatabase(cfg helperconfig.DatabaseConfig) (Database, error) {
	return NewDatabase(cfg)
}

func (DefaultFactory) NewRedis(cfg helperconfig.RedisConfig) Redis {
	return NewRedis(cfg)
}

func (DefaultFactory) NewGPubSub(ctx context.Context, cfg helperconfig.GPubSubConfig) (GPubSub, error) {
	return NewGPubSub(ctx, cfg)
}

func (DefaultFactory) NewS3(ctx context.Context, cfg helperconfig.S3Config) (S3Client, error) {
	return NewS3Client(ctx, cfg)
}

func (DefaultFactory) NewMongoDB(ctx context.Context, cfg helperconfig.MongoDBConfig) (MongoDB, error) {
	return NewMongoDB(ctx, cfg)
}

func (DefaultFactory) NewSFTP(cfg helperconfig.SftpConfig) (SFTP, error) {
	return NewSFTPClient(cfg)
}

func (DefaultFactory) NewTelemetry(ctx context.Context, cfg helperconfig.OtelConfig, isProduction bool) (Telemetry, error) {
	return NewTelemetry(ctx, cfg, isProduction)
}
