package helpers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/radian-solusi/go-helpers/config"
	"github.com/radian-solusi/go-helpers/connections"
	"github.com/radian-solusi/go-helpers/timeutil"
	"github.com/radian-solusi/go-helpers/web"
)

// HelperInterface mirrors the investor service's helper surface with module
// types. The compile-time assertion in helpers.go enforces completeness.
type HelperInterface interface {
	IsProduction() bool
	SetBaseUrl(url string)
	GetBaseUrl() string
	MakeRequest(method string, url string, params any) ([]byte, error)
	GetLastHeaderResponse() *http.Header
	GetLastStatusCode() int
	InitializeSystem() error
	GetDatabase() connections.Database
	GetS3Client() connections.S3Client
	GetRedisClient() connections.Redis
	GetPubSub() connections.GPubSub
	GetMongoDB() connections.MongoDB
	SetCache(key string, value any, ttl int) error
	GetCache(key string) (*string, error)
	DeleteCache(key string) error
	DeleteCachePattern(pattern string) error
	LoadConfig(conf any) error
	FindProjectRoot() (string, error)
	GetDefaultLimitData() int
	ErrorFatal(err error)
	ErrorResponse(ctx *gin.Context, code int, message string)
	HandleErrorResponse(ctx *gin.Context, err error)
	ErrorMessage(code int) string
	GetStringValue(ptr *string) string
	GetBoolValue(ptr *bool) string
	JSONToStruct(data []byte, v any) error
	StructToJSON(v any) ([]byte, error)
	InterfaceToStruct(data any, v any) error
	SendResponse(ctx *gin.Context, response web.ResponseDefault)
	SendResponseData(ctx *gin.Context, code int, message string, data any)
	GenerateRandomLabel(prefix string, n int) string
	FormatSize(bytes int64) string
	Encrypt(plaintext []byte, key *string) (string, error)
	Decrypt(ciphertextBase64 string, key *string) ([]byte, error)
	Unpad(src []byte) []byte
	ZeroUnpad(data []byte) []byte
	GetMainConfig() config.MainConfig
	SetTokenActive(token string)
	GetTokenActive() string
	SetUserActive(user web.PayloadAuthorization)
	GetUserActive() web.PayloadAuthorization
	SetUserSession(sessionID string)
	GetUserSession() string
	SetUserActiveCtx(c *gin.Context, user web.PayloadAuthorization)
	GetUserActiveCtx(c *gin.Context) web.PayloadAuthorization
	SetTokenActiveCtx(c *gin.Context, token string)
	GetTokenActiveCtx(c *gin.Context) string
	SetUserSessionCtx(c *gin.Context, sessionID string)
	GetUserSessionCtx(c *gin.Context) string
	IsFileComplete(filePath string) bool
	ReadJSONFile(filePath string, target any) error
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) bool
	GenerateSecureToken(length int) (string, error)
	GenerateAuthToken(userID string) (string, error)
	ValidatePasswordComplexity(password string) error
	GenerateTOTPSecret(accountName string) (secret string, otpauthURL string, err error)
	VerifyTOTPCode(secret, code string) bool
	GenerateNumericOTP(length int) (string, error)
	GeneratePreAuthToken(userID string) (string, error)
	VerifyPreAuthToken(token string) (string, error)
	LoadTimeLocale(locale string) *time.Location
	NormalizePhone(phonestring string) string
	DefaultValue(ptr *string, defaultValue string) string
	GetTimeProvider() timeutil.TimeProvider
	GenerateJWTToken(payLoad any, expires time.Time) (string, error)
	ParsingJWT(token string, payload any) error
	FormatValidationError(err error) string
	FormatValidationErrorFields(err error) map[string]string
	SetupLogging()
	ContainString(s, substr string) bool
	ConvertStringToInt64(s string) int64
	Slugify(text string) string
	ConvertInt64ToString(i int64) string
	ValidateSafeHTML(value string) error
	GenerateChecksum(data any) (string, error)
	ExtractIPFromMetadata(metadata *string) string
	RunMigration() error
	FindString(str string, list []string) bool
	ShutdownTelemetry(ctx context.Context) error
	StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	SetGormProgressMode(enabled bool)
	DeferString(strg *string) string
}
