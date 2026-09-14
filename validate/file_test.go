package validate

import (
	"errors"
	"testing"
)

func TestValidateFile_Required(t *testing.T) {
	err := ValidateFile("final_memo", nil, FileConstraint{Required: true})
	var appErr *AppValidationError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *AppValidationError, got %v", err)
	}
	if appErr.Errors[0].Field != "final_memo" || appErr.Errors[0].Tag != "file_required" {
		t.Fatalf("unexpected: %+v", appErr.Errors[0])
	}
}

func TestValidateFile_NotRequiredAndAbsent(t *testing.T) {
	if err := ValidateFile("logo", nil, FileConstraint{Required: false}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateFile_TooLarge(t *testing.T) {
	info := &FileInfo{SizeBytes: 11 * 1024 * 1024, ContentType: "application/pdf"}
	err := ValidateFile("prospectus", info, FileConstraint{
		MaxSizeBytes: 10 * 1024 * 1024,
		AllowedMIME:  []string{"application/pdf"},
	})
	var appErr *AppValidationError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *AppValidationError, got %v", err)
	}
	fe := appErr.Errors[0]
	if fe.Field != "prospectus" || fe.Tag != "file_too_large" || fe.Param != "10MB" {
		t.Fatalf("unexpected: %+v", fe)
	}
}

func TestValidateFile_TypeNotAllowed(t *testing.T) {
	info := &FileInfo{SizeBytes: 1024, ContentType: "image/gif"}
	err := ValidateFile("logo", info, FileConstraint{
		MaxSizeBytes: 10 * 1024 * 1024,
		AllowedMIME:  []string{"image/png", "image/jpeg"},
	})
	var appErr *AppValidationError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *AppValidationError, got %v", err)
	}
	fe := appErr.Errors[0]
	if fe.Field != "logo" || fe.Tag != "file_type_not_allowed" || fe.Param != "image/png, image/jpeg" {
		t.Fatalf("unexpected: %+v", fe)
	}
}

func TestValidateFile_Valid(t *testing.T) {
	info := &FileInfo{SizeBytes: 2048, ContentType: "application/pdf"}
	err := ValidateFile("final_memo", info, FileConstraint{
		Required:     true,
		MaxSizeBytes: 10 * 1024 * 1024,
		AllowedMIME:  []string{"application/pdf"},
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateFile_MIMEParametersIgnored(t *testing.T) {
	info := &FileInfo{SizeBytes: 1024, ContentType: "text/plain; charset=utf-8"}
	if err := ValidateFile("notes", info, FileConstraint{AllowedMIME: []string{"text/plain"}}); err != nil {
		t.Fatalf("charset parameter must not break matching: %v", err)
	}
}

func TestValidateFile_NegativeSize(t *testing.T) {
	info := &FileInfo{SizeBytes: -1, ContentType: "application/pdf"}
	err := ValidateFile("prospectus", info, FileConstraint{MaxSizeBytes: 1024})
	var appErr *AppValidationError
	if !errors.As(err, &appErr) || appErr.Errors[0].Tag != "file_read_error" {
		t.Fatalf("expected file_read_error, got %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	cases := map[int64]string{
		500:                    "500B",
		2048:                   "2KB",
		10 * 1024 * 1024:       "10MB",
		2 * 1024 * 1024 * 1024: "2GB",
	}
	for size, want := range cases {
		if got := formatBytes(size); got != want {
			t.Errorf("formatBytes(%d) = %q, want %q", size, got, want)
		}
	}
}

func TestTranslateFileMessages(t *testing.T) {
	if got := Translate("file_too_large", "en", "prospectus", "10MB"); got != "prospectus must not exceed 10MB" {
		t.Errorf("got %q", got)
	}
	if got := Translate("file_too_large", "id", "prospectus", "10MB"); got != "prospectus tidak boleh melebihi 10MB" {
		t.Errorf("got %q", got)
	}
	if got := Translate("file_type_not_allowed", "id", "logo", "image/png"); got != "logo harus salah satu dari tipe berikut: image/png" {
		t.Errorf("got %q", got)
	}
}
