package helpers

import (
	"time"

	"github.com/radian-solusi/go-helpers/config"
	"github.com/radian-solusi/go-helpers/cryptoutil"
	"github.com/radian-solusi/go-helpers/otp"
	"github.com/radian-solusi/go-helpers/phone"
	"github.com/radian-solusi/go-helpers/strutil"
	"github.com/radian-solusi/go-helpers/timeutil"
	"github.com/radian-solusi/go-helpers/validate"
)

// --- config ---

func (h *Helpers) FindProjectRoot() (string, error) {
	return config.FindProjectRoot(h.configStart)
}

func (h *Helpers) LoadConfig(conf any) error {
	return config.LoadFromEnvironment(h.configStart, h.envLookup, conf)
}

func (h *Helpers) GetMainConfig() config.MainConfig {
	h.configOnce.Do(func() {
		_ = config.LoadFromEnvironment(h.configStart, h.envLookup, &h.config)
	})
	return h.config
}

func (h *Helpers) GetDefaultLimitData() int {
	if l := h.GetMainConfig().App.LimitData; l > 0 {
		return l
	}
	return 50
}

// --- pure delegations ---

func (h *Helpers) GetStringValue(ptr *string) string { return strutil.StringValue(ptr) }
func (h *Helpers) GetBoolValue(ptr *bool) string     { return strutil.BoolString(ptr) }
func (h *Helpers) DeferString(strg *string) string   { return strutil.StringValue(strg) }
func (h *Helpers) DefaultValue(ptr *string, defaultValue string) string {
	return strutil.DefaultValue(ptr, defaultValue)
}

func (h *Helpers) GenerateRandomLabel(prefix string, n int) string {
	label, err := strutil.RandomLabel(prefix, n)
	if err != nil {
		return prefix
	}
	return label
}

func (h *Helpers) FormatSize(bytes int64) string       { return strutil.FormatSize(bytes) }
func (h *Helpers) ContainString(s, substr string) bool { return strutil.ContainsSubstr(s, substr) }
func (h *Helpers) ConvertStringToInt64(s string) int64 { return strutil.ParseInt64Default(s, 0) }
func (h *Helpers) Slugify(text string) string          { return strutil.Slugify(text) }
func (h *Helpers) ConvertInt64ToString(i int64) string { return strutil.Int64ToString(i) }
func (h *Helpers) FindString(str string, list []string) bool {
	return strutil.Contains(list, str)
}
func (h *Helpers) ExtractIPFromMetadata(metadata *string) string {
	return strutil.ExtractIPFromMetadata(metadata)
}

func (h *Helpers) JSONToStruct(data []byte, v any) error { return strutil.JSONToStruct(data, v) }
func (h *Helpers) StructToJSON(v any) ([]byte, error)    { return strutil.StructToJSON(v) }
func (h *Helpers) InterfaceToStruct(data any, v any) error {
	return strutil.InterfaceToStruct(data, v)
}
func (h *Helpers) IsFileComplete(filePath string) bool { return strutil.IsFileComplete(filePath) }
func (h *Helpers) ReadJSONFile(filePath string, target any) error {
	return strutil.ReadJSONFile(filePath, target, 3, 50*time.Millisecond)
}

func (h *Helpers) Unpad(src []byte) []byte      { return cryptoutil.Unpad(src) }
func (h *Helpers) ZeroUnpad(data []byte) []byte { return cryptoutil.ZeroUnpad(data) }

func (h *Helpers) HashPassword(password string) (string, error) {
	return cryptoutil.HashPassword(password)
}
func (h *Helpers) VerifyPassword(hashedPassword, password string) bool {
	return cryptoutil.VerifyPassword(hashedPassword, password)
}
func (h *Helpers) GenerateSecureToken(length int) (string, error) {
	return cryptoutil.GenerateSecureToken(length)
}
func (h *Helpers) GenerateAuthToken(userID string) (string, error) {
	return cryptoutil.GenerateSecureToken(32)
}
func (h *Helpers) GenerateChecksum(data any) (string, error) { return cryptoutil.Checksum(data) }

func (h *Helpers) ValidatePasswordComplexity(password string) error {
	return validate.PasswordComplexity(password)
}
func (h *Helpers) ValidateSafeHTML(value string) error { return validate.SafeHTML(value) }
func (h *Helpers) FormatValidationError(err error) string {
	return validate.FormatValidationError(err)
}
func (h *Helpers) FormatValidationErrorFields(err error) map[string]string {
	return validate.FormatValidationErrorFields(err)
}

func (h *Helpers) NormalizePhone(phonestring string) string { return phone.Normalize(phonestring) }

func (h *Helpers) LoadTimeLocale(locale string) *time.Location {
	loc, err := timeutil.LoadLocation(locale)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (h *Helpers) GetTimeProvider() timeutil.TimeProvider { return h.timeProvider }

func (h *Helpers) IsProduction() bool { return isProduction() }

func (h *Helpers) ErrorFatal(err error) {
	if err != nil {
		panic(err)
	}
}

// --- encrypt/decrypt with AppKey fallback ---

func (h *Helpers) Encrypt(plaintext []byte, key *string) (string, error) {
	return cryptoutil.EncryptLegacyCBC(plaintext, h.resolveKey(key))
}
func (h *Helpers) Decrypt(ciphertextBase64 string, key *string) ([]byte, error) {
	return cryptoutil.DecryptLegacyCBC(ciphertextBase64, h.resolveKey(key))
}
func (h *Helpers) resolveKey(key *string) []byte {
	if key != nil {
		return []byte(*key)
	}
	return []byte(h.GetMainConfig().App.AppKey)
}

// --- OTP delegations ---

func (h *Helpers) GenerateTOTPSecret(accountName string) (string, string, error) {
	issuer := h.GetMainConfig().App.OtpIssuer
	if issuer == "" {
		issuer = "E-IPO"
	}
	return otp.GenerateTOTPSecret(issuer, accountName)
}
func (h *Helpers) VerifyTOTPCode(secret, code string) bool {
	return otp.VerifyTOTPCode(secret, code)
}
func (h *Helpers) GenerateNumericOTP(length int) (string, error) {
	return otp.GenerateNumericOTP(length)
}
func (h *Helpers) GeneratePreAuthToken(userID string) (string, error) {
	return otp.GeneratePreAuthToken(userID, []byte(h.GetMainConfig().App.AppKey), h.timeProvider.Now(), 5*time.Minute)
}
func (h *Helpers) VerifyPreAuthToken(token string) (string, error) {
	return otp.VerifyPreAuthToken(token, []byte(h.GetMainConfig().App.AppKey), h.timeProvider.Now())
}
