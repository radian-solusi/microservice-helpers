package strutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONConversions(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	data, err := StructToJSON(payload{Name: "radian"})
	if err != nil {
		t.Fatal(err)
	}
	var got payload
	if err := JSONToStruct(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "radian" {
		t.Fatalf("got %q", got.Name)
	}
	if err := InterfaceToStruct(map[string]any{"name": "investor"}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "investor" {
		t.Fatalf("got %q", got.Name)
	}
	if err := InterfaceToStruct(nil, &got); err == nil {
		t.Fatal("expected error")
	}
}

func TestFileHelpers(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name, body string
		want       bool
	}{
		{"object", `{"id":7}`, true},
		{"whitespace", "{\"id\":7}\n", true},
		{"empty", "", false},
		{"truncated", `{"id":`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name+".json")
			if err := os.WriteFile(path, []byte(tc.body), 0o600); err != nil {
				t.Fatal(err)
			}
			if got := IsFileComplete(path); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
	path := filepath.Join(dir, "read.json")
	if err := os.WriteFile(path, []byte(`{"id":7}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var got struct {
		ID int `json:"id"`
	}
	if err := ReadJSONFile(path, &got, 3, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if got.ID != 7 {
		t.Fatalf("got %d", got.ID)
	}
}
