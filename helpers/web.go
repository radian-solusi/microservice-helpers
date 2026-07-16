package helpers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radian-solusi/microservice-helpers/web"
)

func (h *Helpers) SendResponse(ctx *gin.Context, r web.ResponseDefault) {
	web.SendResponse(ctx, r)
}
func (h *Helpers) SendResponseData(ctx *gin.Context, code int, msg string, data any) {
	web.SendResponseData(ctx, code, msg, data)
}
func (h *Helpers) ErrorResponse(ctx *gin.Context, code int, msg string) {
	web.ErrorResponse(ctx, code, msg)
}
func (h *Helpers) ErrorMessage(code int) string { return web.ErrorMessage(code) }
func (h *Helpers) HandleErrorResponse(ctx *gin.Context, err error) {
	web.HandleErrorResponse(ctx, err, h.errorCodeMapper)
}

func (h *Helpers) newJWT() (*web.JWT, error) {
	return web.NewJWT([]byte(h.GetMainConfig().App.AppKey))
}
func (h *Helpers) GenerateJWTToken(payload any, expires time.Time) (string, error) {
	j, err := h.newJWT()
	if err != nil {
		return "", err
	}
	return j.Generate(payload, expires)
}
func (h *Helpers) ParsingJWT(token string, payload any) error {
	j, err := h.newJWT()
	if err != nil {
		return err
	}
	return j.Parse(token, payload)
}

func (h *Helpers) MakeRequest(method string, url string, params any) ([]byte, error) {
	h.mu.Lock()
	if h.client == nil {
		h.client = web.NewClient(h.baseURL)
	}
	client := h.client
	token := h.tokenActive
	h.mu.Unlock()
	client.SetToken(token)
	return client.Do(context.Background(), method, url, params)
}

func (h *Helpers) GetLastStatusCode() int {
	h.mu.RLock()
	client := h.client
	h.mu.RUnlock()
	if client == nil {
		return 0
	}
	return client.LastStatusCode()
}
func (h *Helpers) GetLastHeaderResponse() *http.Header {
	h.mu.RLock()
	client := h.client
	h.mu.RUnlock()
	if client == nil {
		return nil
	}
	header := client.LastHeader()
	return &header
}
