package helpers

import (
	"os"

	"github.com/radian-solusi/microservice-helpers/connections"
	"github.com/radian-solusi/microservice-helpers/timeutil"
	"github.com/radian-solusi/microservice-helpers/web"
)

var _ HelperInterface = (*Helpers)(nil)

func NewHelpers(opts ...Option) *Helpers {
	h := &Helpers{
		envLookup:    os.Getenv,
		factory:      connections.DefaultFactory{},
		timeProvider: &timeutil.DefaultTimeProvider{},
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.errorCodeMapper == nil {
		h.errorCodeMapper = func(error) int { return web.InternalServerError }
	}
	return h
}
