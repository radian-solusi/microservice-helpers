package connections

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/sftp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type Database interface {
	DB() *gorm.DB
	Close() error
	Ping(context.Context) error
	Migrate(models ...any) error
	SetProgressMode(bool)
}

type Redis interface {
	Client() *redis.Client
	Ping(context.Context) error
	Set(context.Context, string, any, time.Duration) error
	Get(context.Context, string) (string, error)
	Clear(context.Context, string) error
	ClearPattern(context.Context, string) error
	Close() error
}

type MongoDB interface {
	Client() *mongo.Client
	Database() *mongo.Database
	Collection(string) *mongo.Collection
	Ping(context.Context) error
	Close(context.Context) error
}

type S3Client interface {
	Client() *s3.Client
	UploadFile(context.Context, string, []byte, string) error
	DownloadFile(context.Context, string) ([]byte, error)
	DeleteFile(context.Context, string) error
	ListFiles(context.Context, string) ([]string, error)
	FileExists(context.Context, string) (bool, error)
	GetFileURL(context.Context, string, time.Duration) (string, error)
	GetFileExtension(string) string
	IsConnected() bool
	Close() error
	IsLocalStorage() bool
	SetPathURL(string)
	GetPathURL() string
}

type GPubSub interface {
	Client() *pubsub.Client
	IsConnected() bool
	Publish(context.Context, string, []byte, map[string]string) (string, error)
	Subscribe(context.Context, string, func(*pubsub.Message)) error
	SubscribeAsync(context.Context, string, func(*pubsub.Message)) error
	StopSubscription(string) error
	CreateTopic(context.Context, string) error
	CreateSubscription(context.Context, string, string) error
	TopicExists(context.Context, string) (bool, error)
	SubscriptionExists(context.Context, string) (bool, error)
	DeleteTopic(context.Context, string) error
	DeleteSubscription(context.Context, string) error
	GetStats() SubscriptionStats
	Close() error
}

type SFTP interface {
	Client() *sftp.Client
	UploadFile(string, []byte, os.FileMode) error
	DownloadFile(string) ([]byte, error)
	DeleteFile(string) error
	FileExists(string) (bool, error)
	EnsureDir(string) error
	Close() error
	IsConnected() bool
}

type Telemetry interface {
	StartSpan(context.Context, string, ...trace.SpanStartOption) (context.Context, trace.Span)
	Shutdown(context.Context) error
}

type SubscriptionStats struct {
	MessagesReceived int64
	MessagesAcked    int64
	MessagesNacked   int64
	LastMessageTime  time.Time
}
