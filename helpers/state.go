package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/radian-solusi/microservice-helpers/web"
)

func (h *Helpers) SetBaseUrl(url string) {
	h.mu.Lock()
	h.baseURL = url
	if h.client != nil {
		h.client.SetBaseURL(url)
	}
	h.mu.Unlock()
}
func (h *Helpers) GetBaseUrl() string { h.mu.RLock(); defer h.mu.RUnlock(); return h.baseURL }

func (h *Helpers) SetTokenActive(token string) {
	h.mu.Lock()
	h.tokenActive = token
	h.mu.Unlock()
}
func (h *Helpers) GetTokenActive() string { h.mu.RLock(); defer h.mu.RUnlock(); return h.tokenActive }

func (h *Helpers) SetUserActive(c *gin.Context, user any) {
	if c != nil {
		h.SetUserActiveCtx(c, user)
	}
	h.mu.Lock()
	h.userActive = user
	h.mu.Unlock()
}
func (h *Helpers) GetUserActive(c *gin.Context, dataType *any) {
	if c != nil {
		h.GetUserActiveCtx(c, dataType)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	*dataType = h.userActive
	return
}

func (h *Helpers) SetUserSession(c *gin.Context, sessionID string) {
	if c != nil {
		h.SetUserSessionCtx(c, sessionID)
	}
	h.mu.Lock()
	h.userSession = sessionID
	h.mu.Unlock()
}
func (h *Helpers) GetUserSession(c *gin.Context) string {
	if c != nil {
		return h.GetUserSessionCtx(c)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.userSession
}

func (h *Helpers) SetUserActiveCtx(c *gin.Context, user any) {
	web.SetUserActiveCtx(c, user)
}
func (h *Helpers) GetUserActiveCtx(c *gin.Context, dataType *any) {
	web.GetUserActiveCtx(c, dataType)
}
func (h *Helpers) SetTokenActiveCtx(c *gin.Context, token string) {
	web.SetTokenActiveCtx(c, token)
}
func (h *Helpers) GetTokenActiveCtx(c *gin.Context) string { return web.GetTokenActiveCtx(c) }
func (h *Helpers) SetUserSessionCtx(c *gin.Context, sessionID string) {
	web.SetUserSessionCtx(c, sessionID)
}
func (h *Helpers) GetUserSessionCtx(c *gin.Context) string { return web.GetUserSessionCtx(c) }
