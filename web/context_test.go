package web

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestContextAccessors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	SetTokenActiveCtx(ctx, "tok")
	if GetTokenActiveCtx(ctx) != "tok" {
		t.Fatal("token")
	}
	SetUserSessionCtx(ctx, "sess")
	if GetUserSessionCtx(ctx) != "sess" {
		t.Fatal("session")
	}
	SetUserActiveCtx(ctx, PayloadAuthorization{UserID: "u1"})
	var got any
	GetUserActiveCtx(ctx, &got)
	if got.(PayloadAuthorization).UserID != "u1" {
		t.Fatal("user")
	}
	empty, _ := gin.CreateTestContext(httptest.NewRecorder())
	var missing any
	GetUserActiveCtx(empty, &missing)
	if missing != nil {
		t.Fatal("missing should be zero")
	}
}
