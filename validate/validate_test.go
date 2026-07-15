package validate

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type testValidation struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18,lte=120"`
}

func TestFormatValidationError(t *testing.T) {
	if FormatValidationError(nil) != "" {
		t.Error("nil should be empty")
	}
	if got := FormatValidationError(errValidation("some error")); got != "Invalid input data" {
		t.Errorf("non-validator error: got %q", got)
	}

	v := validator.New()
	s := testValidation{Name: "", Email: "bad", Age: 5}
	err := v.Struct(s)
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := FormatValidationError(err)
	if msg == "" || msg == "Invalid input data" {
		t.Errorf("unexpected message: %q", msg)
	}
}

type errValidation string

func (e errValidation) Error() string { return string(e) }

func TestFormatValidationErrorFields(t *testing.T) {
	if m := FormatValidationErrorFields(nil); len(m) != 0 {
		t.Error("nil should be empty")
	}
	v := validator.New()
	s := testValidation{Name: "", Email: "", Age: 5}
	err := v.Struct(s)
	fields := FormatValidationErrorFields(err)
	if len(fields) == 0 {
		t.Fatal("expected multiple fields")
	}
	for _, f := range []string{"Name", "Email", "Age"} {
		if fields[f] == "" {
			t.Errorf("missing field %s", f)
		}
	}
}

func TestSafeHTML(t *testing.T) {
	if err := SafeHTML(""); err != nil {
		t.Errorf("empty should pass: %v", err)
	}
	if err := SafeHTML("<b>bold</b>"); err != nil {
		t.Errorf("allowed tag: %v", err)
	}
	if err := SafeHTML("<script>alert(1)</script>"); err == nil {
		t.Error("script tag should fail")
	}
	if err := SafeHTML("<iframe src=x></iframe>"); err == nil {
		t.Error("iframe should fail")
	}
	if err := SafeHTML("<b onclick='x'>x</b>"); err == nil {
		t.Error("event handler should fail")
	}
	// safe style/class passes
	if err := SafeHTML("<span style='color:red' class='x'>x</span>"); err != nil {
		t.Errorf("safe attribute should pass: %v", err)
	}
	if err := SafeHTML("<p style='expression(alert(1))'>x</p>"); err == nil {
		t.Error("expression in style should fail")
	}
	if err := SafeHTML("<p style='javascript:void(0)'>x</p>"); err == nil {
		t.Error("javascript: in style should fail")
	}
	if err := SafeHTML("<p style='vbscript:msgbox'>x</p>"); err == nil {
		t.Error("vbscript: in style should fail")
	}
	if err := SafeHTML("<p style='data:image'>x</p>"); err == nil {
		t.Error("data: in style should fail")
	}
	if err := SafeHTML("<p style='-moz-binding:url(x)'>x</p>"); err == nil {
		t.Error("-moz-binding should fail")
	}
}

func TestPasswordComplexity(t *testing.T) {
	cases := []struct {
		pw      string
		wantErr bool
	}{
		{"Abc1@", true},     // too short
		{"abcdefgh", true},  // missing upper, number, special
		{"ABCDEFGH", true},  // missing lower? actually has upper but no number/special
		{"Abcdefgh", true},  // missing number, special
		{"Abcdef1h", true},  // missing special
		{"Abcd1@ef", false}, // valid
		{"Str0ng!x", false}, // valid
	}
	for _, c := range cases {
		if err := PasswordComplexity(c.pw); (err != nil) != c.wantErr {
			t.Errorf("PasswordComplexity(%q) err=%v wantErr=%v", c.pw, err, c.wantErr)
		}
	}
}
