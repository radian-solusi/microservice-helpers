package web

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorResponseMapsWrongCredential(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ErrorResponse(ctx, WrongCredential, "bad")
	if rec.Code != Unauthorized {
		t.Fatalf("code %d", rec.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("expected abort")
	}
}

func TestHandleErrorResponseUsesMapper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	sentinel := errors.New("nope")
	HandleErrorResponse(ctx, sentinel, func(err error) int {
		if errors.Is(err, sentinel) {
			return Forbidden
		}
		return InternalServerError
	})
	if rec.Code != Forbidden {
		t.Fatalf("code %d", rec.Code)
	}
}

func TestHandleErrorResponseNilMapperDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	HandleErrorResponse(ctx, errors.New("x"), nil)
	if rec.Code != InternalServerError {
		t.Fatalf("code %d", rec.Code)
	}
}
