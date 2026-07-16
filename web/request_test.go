package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDoPostJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("missing auth")
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]string
		_ = json.Unmarshal(body, &got)
		if got["a"] != "b" {
			t.Errorf("body %v", got)
		}
		w.Header().Set("X-Test", "1")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	c.SetToken("tok")
	resp, err := c.Do(context.Background(), http.MethodPost, "/x", map[string]string{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}
	if c.LastStatusCode() != 201 || c.LastHeader().Get("X-Test") != "1" {
		t.Fatalf("status/header %d %v", c.LastStatusCode(), c.LastHeader())
	}
	if string(resp) != `{"ok":true}` {
		t.Fatalf("resp %s", resp)
	}
}

func TestClientDoGetEncodesQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("a") != "b" {
			t.Errorf("query missing: %v", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	if _, err := c.Do(context.Background(), http.MethodGet, "/y", map[string]any{"a": "b"}); err != nil {
		t.Fatal(err)
	}
}
