package web

import "github.com/gin-gonic/gin"

const (
	keyUserActive  = "user_active"
	keyTokenActive = "token_active"
	keyUserSession = "session_user"
)

func SetUserActiveCtx(c *gin.Context, user PayloadAuthorization) {
	c.Set(keyUserActive, user)
}

func GetUserActiveCtx(c *gin.Context) PayloadAuthorization {
	v, ok := c.Get(keyUserActive)
	if !ok {
		return PayloadAuthorization{}
	}
	u, ok := v.(PayloadAuthorization)
	if !ok {
		return PayloadAuthorization{}
	}
	return u
}

func SetTokenActiveCtx(c *gin.Context, token string) {
	c.Set(keyTokenActive, token)
}

func GetTokenActiveCtx(c *gin.Context) string {
	return c.GetString(keyTokenActive)
}

func SetUserSessionCtx(c *gin.Context, sessionID string) {
	c.Set(keyUserSession, sessionID)
}

func GetUserSessionCtx(c *gin.Context) string {
	return c.GetString(keyUserSession)
}
