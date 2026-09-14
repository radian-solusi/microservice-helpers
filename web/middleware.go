package web

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// LanguageMiddleware resolves Accept-Language into a supported language
// ("en"/"id", default "en") and stores it via SetLanguageCtx for
// ValidationErrorResponse to pick up downstream.
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetLanguageCtx(c, ParseAcceptLanguage(c.GetHeader("Accept-Language")))
		c.Next()
	}
}

var supportedLangs = map[string]bool{"en": true, "id": true}

// ParseAcceptLanguage picks the highest-quality supported language from an
// Accept-Language header. Region subtags and case are ignored ("id-ID",
// "ID" -> "id"); entries with q=0 are excluded; anything unsupported or
// malformed falls back to "en".
func ParseAcceptLanguage(header string) string {
	best, bestQ := "en", -1.0
	for _, entry := range strings.Split(header, ",") {
		parts := strings.Split(entry, ";")
		lang := strings.ToLower(strings.TrimSpace(parts[0]))
		if i := strings.IndexAny(lang, "-_"); i >= 0 {
			lang = lang[:i]
		}
		if !supportedLangs[lang] {
			continue
		}
		q := 1.0
		for _, p := range parts[1:] {
			p = strings.TrimSpace(p)
			if v, ok := strings.CutPrefix(p, "q="); ok {
				parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
				if err != nil {
					q = -1 // malformed q: ignore this entry
					break
				}
				q = parsed
			}
		}
		if q <= 0 {
			continue
		}
		if q > bestQ {
			best, bestQ = lang, q
		}
	}
	return best
}
