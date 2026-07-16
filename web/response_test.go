package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSendResponseData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	SendResponseData(ctx, 200, "ok", map[string]string{"a": "b"})
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var body ResponseDefault
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Status || body.Message != "ok" {
		t.Fatalf("body %+v", body)
	}
}
