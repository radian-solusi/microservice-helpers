package validate

import (
	"strings"
	"sync"
	"testing"
)

func TestTranslateBuiltinTags(t *testing.T) {
	cases := []struct{ tag, lang, field, param, want string }{
		{"required", "en", "website", "", "website is required"},
		{"required", "id", "website", "", "website wajib diisi"},
		{"url", "en", "website", "", "website is not a valid URL"},
		{"url", "id", "website", "", "website bukan URL yang valid"},
		{"min", "en", "name", "3", "name must be at least 3"},
		{"min", "id", "name", "3", "name minimal 3"},
	}
	for _, c := range cases {
		t.Run(c.tag+"/"+c.lang, func(t *testing.T) {
			if got := Translate(c.tag, c.lang, c.field, c.param); got != c.want {
				t.Errorf("Translate(%q,%q,%q,%q) = %q, want %q",
					c.tag, c.lang, c.field, c.param, got, c.want)
			}
		})
	}
}

func TestTranslateUnknownTagUsesLocalizedFallback(t *testing.T) {
	for lang, want := range map[string]string{"en": "field is invalid", "id": "field tidak valid"} {
		got := Translate("nonexistent_tag_xyz", lang, "field", "")
		if got != want {
			t.Errorf("lang %q: got %q, want %q", lang, got, want)
		}
		if strings.Contains(got, "nonexistent_tag_xyz") {
			t.Errorf("lang %q: must not leak raw tag: %q", lang, got)
		}
	}
}

func TestTranslateUnknownLangFallsBackToEnglish(t *testing.T) {
	if got, want := Translate("required", "fr", "website", ""), Translate("required", "en", "website", ""); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestTranslateLangCaseInsensitive(t *testing.T) {
	if got := Translate("required", "ID", "website", ""); got != "website wajib diisi" {
		t.Errorf("got %q", got)
	}
}

func TestRegisterMessageCustomKey(t *testing.T) {
	RegisterMessage("offering_before_book_building_test",
		"offering period cannot start before book building period",
		"periode penawaran tidak boleh dimulai sebelum periode book building")
	if got := Translate("offering_before_book_building_test", "id", "", ""); got != "periode penawaran tidak boleh dimulai sebelum periode book building" {
		t.Errorf("got %q", got)
	}
}

func TestTranslateConcurrentSafe(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Translate("required", "id", "f", "")
			RegisterMessage("race_key", "en msg", "id msg")
			Translate("race_key", "en", "f", "")
		}()
	}
	wg.Wait()
}
