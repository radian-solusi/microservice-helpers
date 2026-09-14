package web

import "github.com/gin-gonic/gin"

const (
	keyUserActive  = "user_active"
	keyTokenActive = "token_active"
	keyUserSession = "session_user"
	keyLanguage    = "request_language"
)

// SetLanguageCtx stores the resolved request language for downstream handlers.
func SetLanguageCtx(c *gin.Context, lang string) {
	c.Set(keyLanguage, lang)
}

// GetLanguageCtx returns the request language, defaulting to "en" when unset.
func GetLanguageCtx(c *gin.Context) string {
	lang := c.GetString(keyLanguage)
	if lang == "" {
		return "en"
	}
	return lang
}

func SetUserActiveCtx(c *gin.Context, user any) {
	c.Set(keyUserActive, user)
}

func GetUserActiveCtx(c *gin.Context, dataType *any) {
	v, ok := c.Get(keyUserActive)
	if !ok {
		return
	}
	*dataType = v
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
