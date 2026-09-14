package web

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func runMiddleware(t *testing.T, acceptLanguage string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	if acceptLanguage != "" {
		ctx.Request.Header.Set("Accept-Language", acceptLanguage)
	}
	LanguageMiddleware()(ctx)
	return GetLanguageCtx(ctx)
}

func TestLanguageMiddlewareDefaultsToEnglish(t *testing.T) {
	if got := runMiddleware(t, ""); got != "en" {
		t.Fatalf("got %q", got)
	}
}

func TestLanguageMiddlewareReadsAcceptLanguage(t *testing.T) {
	if got := runMiddleware(t, "id"); got != "id" {
		t.Fatalf("got %q", got)
	}
}

func TestLanguageMiddlewareSupportedCases(t *testing.T) {
	cases := []struct{ header, want string }{
		{"id-ID", "id"},
		{"ID", "id"},
		{"en-US,en;q=0.9", "en"},
		{"id-ID,id;q=0.9,en;q=0.8", "id"},
		{"fr-FR,fr;q=0.9", "en"},
		{"en;q=0.3,id;q=0.9", "id"},
		{"id;q=0.5,en;q=0.5", "id"},
		{"id;q=0", "en"},
		{"id;q=0,en;q=0", "en"},
		{"*", "en"},
		{"  id-ID  ", "id"},
	}
	for _, c := range cases {
		t.Run(c.header, func(t *testing.T) {
			if got := runMiddleware(t, c.header); got != c.want {
				t.Errorf("header %q: got %q want %q", c.header, got, c.want)
			}
		})
	}
}

func TestSetGetLanguageCtx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	SetLanguageCtx(ctx, "id")
	if got := GetLanguageCtx(ctx); got != "id" {
		t.Fatalf("got %q", got)
	}
	empty, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := GetLanguageCtx(empty); got != "en" {
		t.Fatalf("unset should default to en, got %q", got)
	}
}
