package validate

import (
	"fmt"
	"mime"
	"strings"
)

// FileInfo is the minimal shape needed to validate an uploaded file —
// Callers must provide a server-measured size and a MIME type detected from
// file contents (for example http.DetectContentType), not the client-supplied
// Content-Type header. This metadata check does not inspect file contents.
type FileInfo struct {
	SizeBytes   int64
	ContentType string
}

// FileConstraint describes the upload rules for one form field.
type FileConstraint struct {
	Required     bool
	MaxSizeBytes int64    // 0 = no limit
	AllowedMIME  []string // empty = any type allowed
}

// ValidateFile checks an uploaded file (nil info = absent) against a
// constraint, returning a *AppValidationError with a translatable tag
// (file_required / file_too_large / file_type_not_allowed) or nil if valid.
func ValidateFile(field string, info *FileInfo, c FileConstraint) error {
	if info == nil {
		if c.Required {
			return NewFieldError(field, "file_required", fmt.Errorf("%s: file is required", field))
		}
		return nil
	}
	if info.SizeBytes < 0 {
		return NewFieldError(field, "file_read_error", fmt.Errorf("%s: negative file size", field))
	}
	if c.MaxSizeBytes > 0 && info.SizeBytes > c.MaxSizeBytes {
		return NewFieldError(field, "file_too_large",
			fmt.Errorf("%s: size %d exceeds max %d", field, info.SizeBytes, c.MaxSizeBytes),
			formatBytes(c.MaxSizeBytes))
	}
	if len(c.AllowedMIME) > 0 && !mimeAllowed(c.AllowedMIME, info.ContentType) {
		return NewFieldError(field, "file_type_not_allowed",
			fmt.Errorf("%s: type %q not allowed", field, info.ContentType),
			strings.Join(c.AllowedMIME, ", "))
	}
	return nil
}

// mimeAllowed compares media types only — parameters such as
// "text/plain; charset=utf-8" are stripped before matching.
func mimeAllowed(allowed []string, contentType string) bool {
	got := mediaType(contentType)
	for _, m := range allowed {
		if strings.EqualFold(mediaType(m), got) {
			return true
		}
	}
	return false
}

func mediaType(contentType string) string {
	if parsed, _, err := mime.ParseMediaType(contentType); err == nil {
		return parsed
	}
	return strings.ToLower(strings.TrimSpace(contentType))
}

// formatBytes renders a byte count as a short human-readable size (B/KB/MB/GB),
// used as the {1} param in the file_too_large message.
func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%d%s", size/div, units[exp])
}
