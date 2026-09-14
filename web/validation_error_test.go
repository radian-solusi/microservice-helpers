package web

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/radian-solusi/microservice-helpers/validate"
)

func newValidationCtx(t *testing.T, lang string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	SetLanguageCtx(ctx, lang)
	return ctx, rec
}

func TestValidationErrorResponseShape(t *testing.T) {
	ctx, rec := newValidationCtx(t, "id")
	err := validate.NewFieldError("website", "url", errors.New("bad"))
	ValidationErrorResponse(ctx, err)
	if rec.Code != ValidationError {
		t.Fatalf("code %d", rec.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("expected abort")
	}
	var body ResponseDefault
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status {
		t.Fatal("status must be false")
	}
	data, _ := body.Data.(map[string]any)
	if data["website"] != "website bukan URL yang valid" {
		t.Errorf("got %+v", body.Data)
	}
	if body.Message != "website bukan URL yang valid" {
		t.Errorf("summary got %q", body.Message)
	}
}

func TestValidationErrorResponseEnglish(t *testing.T) {
	ctx, rec := newValidationCtx(t, "en")
	ValidationErrorResponse(ctx, validate.NewFieldError("website", "url", errors.New("bad")))
	var body ResponseDefault
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	data, _ := body.Data.(map[string]any)
	if data["website"] != "website is not a valid URL" {
		t.Errorf("got %+v", body.Data)
	}
}

func TestValidationErrorResponseUnknownErrorNoRawDetails(t *testing.T) {
	ctx, rec := newValidationCtx(t, "en")
	ValidationErrorResponse(ctx, errors.New("db connection refused at 10.0.0.1"))
	if rec.Code != ValidationError {
		t.Fatalf("code %d", rec.Code)
	}
	body := rec.Body.String()
	if containsAny(body, "db connection refused", "10.0.0.1") {
		t.Errorf("raw detail leaked: %s", body)
	}
	var parsed ResponseDefault
	_ = json.Unmarshal(rec.Body.Bytes(), &parsed)
	if parsed.Message != "Invalid input data" {
		t.Errorf("got message %q", parsed.Message)
	}
}

func TestValidationErrorResponseNilError(t *testing.T) {
	ctx, rec := newValidationCtx(t, "en")
	ValidationErrorResponse(ctx, nil)
	if rec.Code != ValidationError {
		t.Fatalf("code %d", rec.Code)
	}
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if n != "" && len(haystack) >= len(n) {
			for i := 0; i+len(n) <= len(haystack); i++ {
				if haystack[i:i+len(n)] == n {
					return true
				}
			}
		}
	}
	return false
}
