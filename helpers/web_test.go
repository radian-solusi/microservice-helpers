package helpers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/radian-solusi/microservice-helpers/web"
)

func TestFacadeGinDelegations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	h := NewHelpers()
	h.SetUserActiveCtx(ctx, web.PayloadAuthorization{UserID: "u"})
	if h.GetUserActiveCtx(ctx).UserID != "u" {
		t.Fatal("context")
	}
	h.SendResponseData(ctx, web.Success, "ok", nil)
	if rec.Code != web.Success {
		t.Fatalf("code %d", rec.Code)
	}
}
