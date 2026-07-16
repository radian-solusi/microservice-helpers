package helpers

import (
	"context"
	"errors"

	"github.com/radian-solusi/microservice-helpers/strutil"
)

func (h *Helpers) SetCache(key string, value any, ttl int) error {
	if h.redis == nil {
		return errors.New("redis not initialized")
	}
	data, err := strutil.StructToJSON(value)
	if err != nil {
		return err
	}
	payload := string(data)
	if h.IsProduction() {
		enc, err := h.Encrypt(data, nil)
		if err != nil {
			return err
		}
		payload = enc
	}
	return h.redis.Set(context.Background(), key, payload, h.timeProvider.IntToDuration(ttl))
}

func (h *Helpers) GetCache(key string) (*string, error) {
	if h.redis == nil {
		return nil, errors.New("redis not initialized")
	}
	data, err := h.redis.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	if h.IsProduction() {
		dec, err := h.Decrypt(data, nil)
		if err != nil {
			return nil, err
		}
		data = string(dec)
	}
	return &data, nil
}

func (h *Helpers) DeleteCache(key string) error {
	if h.redis == nil {
		return errors.New("redis not initialized")
	}
	return h.redis.Clear(context.Background(), key)
}
func (h *Helpers) DeleteCachePattern(pattern string) error {
	if h.redis == nil {
		return errors.New("redis not initialized")
	}
	return h.redis.ClearPattern(context.Background(), pattern)
}
