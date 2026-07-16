package helpers

import (
	"os"

	"github.com/radian-solusi/go-helpers/connections"
	"github.com/radian-solusi/go-helpers/timeutil"
	"github.com/radian-solusi/go-helpers/web"
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
