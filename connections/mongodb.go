package connections

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	otelmongo "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

type mongoDBWrapper struct {
	client *mongo.Client
	dbName string
}

// mongoURI builds a mongodb:// connection string from cfg. If cfg.URI is set it
// is used verbatim. Otherwise host (with optional credentials) is required.
func mongoURI(cfg helperconfig.MongoDBConfig) (string, error) {
	if cfg.URI != "" {
		return cfg.URI, nil
	}
	if cfg.Host == "" {
		return "", errors.New("mongodb: host or uri must be configured")
	}
	port := cfg.Port
	if port == 0 {
		port = 27017
	}
	if cfg.User != "" && cfg.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d",
			url.QueryEscape(cfg.User), url.QueryEscape(cfg.Password), cfg.Host, port), nil
	}
	return fmt.Sprintf("mongodb://%s:%d", cfg.Host, port), nil
}

func NewMongoDB(ctx context.Context, cfg helperconfig.MongoDBConfig) (MongoDB, error) {
	uri, err := mongoURI(cfg)
	if err != nil {
		return nil, err
	}

	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second).
		SetMaxPoolSize(10).
		SetMinPoolSize(2).
		SetMonitor(otelmongo.NewMonitor())

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: failed to connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongodb: ping failed: %w", err)
	}

	dbName := cfg.DBName
	if dbName == "" {
		dbName = "app"
	}
	return &mongoDBWrapper{client: client, dbName: dbName}, nil
}

func (m *mongoDBWrapper) Client() *mongo.Client { return m.client }

func (m *mongoDBWrapper) Database() *mongo.Database { return m.client.Database(m.dbName) }

func (m *mongoDBWrapper) Collection(name string) *mongo.Collection {
	return m.client.Database(m.dbName).Collection(name)
}

func (m *mongoDBWrapper) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return m.client.Ping(pingCtx, readpref.Primary())
}

func (m *mongoDBWrapper) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
