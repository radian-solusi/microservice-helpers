package validate

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

type i18nItem struct {
	Fixed        string `json:"fixed" validate:"required"`
	MinimumOrder string `json:"minimum_order" validate:"required"`
}

type i18nSample struct {
	Website string     `json:"website" validate:"required,url"`
	Name    string     `json:"name" validate:"required"`
	Series  []i18nItem `json:"series" validate:"dive"`
}

type i18nInner struct {
	Code string `json:"code" validate:"required"`
}

type i18nOuter struct {
	i18nInner
	Name string `json:"name" validate:"required"`
	MOT  string `json:"mot" validate:"required"`
}

func TestFormatValidationErrorData_PlainValidator(t *testing.T) {
	err := validator.New().Struct(i18nSample{Website: "not-a-url", Name: ""})
	data := FormatValidationErrorData(err, "en")
	if _, ok := data["website"]; !ok {
		t.Fatalf("expected website key, got %+v", data)
	}
	if _, ok := data["name"]; !ok {
		t.Fatalf("expected name key, got %+v", data)
	}
}

func TestFormatValidationErrorData_PreservesNestedValidatorPath(t *testing.T) {
	err := validator.New().Struct(i18nSample{
		Website: "https://x.id", Name: "n",
		Series: []i18nItem{{Fixed: "f", MinimumOrder: ""}},
	})
	data := FormatValidationErrorData(err, "en")
	series, ok := data["series"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested series map, got %+v", data)
	}
	item0, ok := series["0"].(map[string]any)
	if !ok {
		t.Fatalf("expected index map, got %+v", series)
	}
	if _, ok := item0["minimum_order"]; !ok {
		t.Fatalf("expected minimum_order leaf, got %+v", item0)
	}
}

func TestFormatValidationErrorDataWithRef_HonorsJSONTagsAndEmbedded(t *testing.T) {
	err := validator.New().Struct(i18nOuter{})
	data := FormatValidationErrorDataWithRef(err, "en", i18nOuter{})
	if _, ok := data["code"]; !ok {
		t.Fatalf("embedded field must flatten to code, got %+v", data)
	}
	if _, ok := data["mot"]; !ok {
		t.Fatalf("json tag mot must win over Go name MOT, got %+v", data)
	}
	if _, ok := data["name"]; !ok {
		t.Fatalf("expected name key, got %+v", data)
	}
}

func TestFormatValidationErrorData_AppValidationError(t *testing.T) {
	err := NewFieldError("website", "url", errors.New("bad url"))
	want := map[string]any{"website": "website bukan URL yang valid"}
	if got := FormatValidationErrorData(err, "id"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestFormatValidationErrorData_NestedIndexedField(t *testing.T) {
	err := NewFieldError("allocations[0].fixed", "min", errors.New("too small"), "0")
	data := FormatValidationErrorData(err, "en")
	item0, ok := data["allocations"].(map[string]any)["0"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested index map, got %+v", data)
	}
	if item0["fixed"] != "fixed must be at least 0" {
		t.Errorf("got %+v", item0)
	}
}

func TestPrefixFieldErrors(t *testing.T) {
	err := validator.New().Struct(i18nItem{})
	prefixed := PrefixFieldErrors(err, "allocations[0]")
	if prefixed == nil || len(prefixed.Errors) == 0 {
		t.Fatal("expected prefixed errors")
	}
	for _, fe := range prefixed.Errors {
		if !strings.HasPrefix(fe.Field, "allocations[0].") {
			t.Errorf("missing prefix: %q", fe.Field)
		}
	}
	if !errors.As(prefixed, new(validator.ValidationErrors)) {
		t.Error("prefixed error must still unwrap to validator errors")
	}
	if PrefixFieldErrors(nil, "x") != nil {
		t.Error("nil error must give nil")
	}
}

func TestFormatValidationSummary_NoRepeatedField(t *testing.T) {
	err := NewFieldError("website", "url", errors.New("bad url"))
	msg := FormatValidationSummary(err, "en")
	if msg != "website is not a valid URL" {
		t.Errorf("got %q", msg)
	}
	if strings.Count(msg, "website") != 1 {
		t.Errorf("field repeated in summary: %q", msg)
	}
}

func TestFormatValidationErrorData_NilAndUnrelatedError(t *testing.T) {
	if data := FormatValidationErrorData(nil, "en"); len(data) != 0 {
		t.Errorf("nil should give empty map, got %+v", data)
	}
	dbDown := errors.New("db down")
	if data := FormatValidationErrorData(dbDown, "en"); len(data) != 0 {
		t.Errorf("unrelated error must not leak details, got %+v", data)
	}
	if msg := FormatValidationSummary(dbDown, "id"); msg != "Data masukan tidak valid" {
		t.Errorf("unknown error summary must be localized, got %q", msg)
	}
	if msg := FormatValidationSummary(nil, "en"); msg != "Invalid input data" {
		t.Errorf("nil summary must be localized fallback, got %q", msg)
	}
}

func TestFormatValidationErrorData_EmptyFieldPathSafe(t *testing.T) {
	err := &AppValidationError{Errors: []FieldError{{Field: "", Tag: "required"}}, Cause: errors.New("x")}
	_ = FormatValidationErrorData(err, "en") // must not panic
	_ = FormatValidationSummary(err, "en")   // must not panic
}
