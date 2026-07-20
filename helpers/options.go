package helpers

import (
	"sync"

	"github.com/radian-solusi/microservice-helpers/config"
	"github.com/radian-solusi/microservice-helpers/connections"
	"github.com/radian-solusi/microservice-helpers/timeutil"
	"github.com/radian-solusi/microservice-helpers/web"
)

type Helpers struct {
	configStart     string
	envLookup       func(string) string
	factory         connections.Factory
	errorCodeMapper web.ErrorCodeMapper
	migrationModels []any
	timeProvider    timeutil.TimeProvider

	config     config.MainConfig
	configOnce sync.Once

	database  connections.Database
	redis     connections.Redis
	pubsub    connections.GPubSub
	s3        connections.S3Client
	mongo     connections.MongoDB
	telemetry connections.Telemetry
	sftp      connections.SFTP

	client *web.Client

	mu          sync.RWMutex
	tokenActive string
	userActive  web.PayloadAuthorization
	userSession string
	baseURL     string
}

type Option func(*Helpers)

func WithFactory(f connections.Factory) Option {
	return func(h *Helpers) { h.factory = f }
}

func WithConfigStart(path string) Option {
	return func(h *Helpers) { h.configStart = path }
}

func WithEnvLookup(lookup func(string) string) Option {
	return func(h *Helpers) { h.envLookup = lookup }
}

func WithErrorCodeMapper(m web.ErrorCodeMapper) Option {
	return func(h *Helpers) { h.errorCodeMapper = m }
}

func WithMigrationModels(models ...any) Option {
	return func(h *Helpers) { h.migrationModels = models }
}

func WithTimeProvider(p timeutil.TimeProvider) Option {
	return func(h *Helpers) { h.timeProvider = p }
}
