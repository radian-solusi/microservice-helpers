package helpers

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/radian-solusi/microservice-helpers/validate"
	"github.com/radian-solusi/microservice-helpers/web"
)

func TestHelpersValidationErrorResponseDelegates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	ctx.Request.Header.Set("Accept-Language", "id")
	web.LanguageMiddleware()(ctx)

	h := NewHelpers()
	if got := h.GetRequestLanguage(ctx); got != "id" {
		t.Fatalf("lang %q", got)
	}
	h.ValidationErrorResponse(ctx, validate.NewFieldError("website", "url", errors.New("bad")))
	if rec.Code != web.ValidationError {
		t.Fatalf("code %d", rec.Code)
	}
	var body web.ResponseDefault
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	data, _ := body.Data.(map[string]any)
	if data["website"] != "website bukan URL yang valid" {
		t.Errorf("got %+v", body.Data)
	}
}

type refReq struct {
	MOT string `json:"mot" validate:"required"`
}

func newValidatorError(t *testing.T, value any) error {
	t.Helper()
	err := playgroundvalidator.New().Struct(value)
	if err == nil {
		t.Fatal("expected validation error")
	}
	return err
}

func TestHelpersValidationErrorResponseWithRefDelegates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	h := NewHelpers()
	err := newValidatorError(t, refReq{})
	h.ValidationErrorResponseWithRef(ctx, err, refReq{})
	var body web.ResponseDefault
	if e := json.Unmarshal(rec.Body.Bytes(), &body); e != nil {
		t.Fatal(e)
	}
	data, _ := body.Data.(map[string]any)
	if _, ok := data["mot"]; !ok {
		t.Errorf("expected json-tag key mot, got %+v", body.Data)
	}
}
