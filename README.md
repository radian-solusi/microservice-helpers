# go-helpers

Shared Go helpers for Radian services.

Packages: `config`, `connections`, `cryptoutil`, `helpers`, `mask`, `otp`, `phone`, `strutil`, `timeutil`, `validate`, and `web`.

## Installation

Add the module to a Go service:

```bash
go get github.com/radian-solusi/go-helpers@latest
```

Then import only the focused packages the service needs:

```go
import (
    "github.com/radian-solusi/go-helpers/cryptoutil"
    "github.com/radian-solusi/go-helpers/phone"
    "github.com/radian-solusi/go-helpers/strutil"
)
```

For compatibility with the investor `HelperInterface`, use the facade:

```go
import "github.com/radian-solusi/go-helpers/helpers"
```

## Focused package usage

Prefer the focused packages for new code because their dependencies and state are explicit:

```go
normalized, err := phone.Validate("0812345678")
if err != nil {
    return err
}

slug := strutil.Slugify("Hello World")

key := []byte("12345678901234567890123456789012")
ciphertext, err := cryptoutil.EncryptLegacyCBC([]byte("secret"), key)
if err != nil {
    return err
}
plaintext, err := cryptoutil.DecryptLegacyCBC(ciphertext, key)
```

`cryptoutil.EncryptLegacyCBC` exists only for compatibility with existing ciphertext. Its wire format is `base64(IV || base64(zero-padded ciphertext))`. CBC is unauthenticated; new encrypted data should use an authenticated construction.

OTP and crypto APIs require explicit key, issuer, time, and TTL values:

```go
secret, otpauthURL, err := otp.GenerateTOTPSecret("E-IPO", "user@example.com")
code, err := otp.GenerateNumericOTP(6)
token, err := otp.GeneratePreAuthToken(userID, key, time.Now(), 5*time.Minute)
userID, err := otp.VerifyPreAuthToken(token, key, time.Now())
```

## Configuration

Configuration decoding remains the consumer's responsibility when using the focused packages:

```go
var cfg config.MainConfig
if err := config.Load("config/development.toml", &cfg); err != nil {
    return err
}
```

To follow the service convention, set `FILE_CONFIG` to a file name under `<project-root>/config` and use:

```go
var cfg config.MainConfig
err := config.LoadFromEnvironment("", os.Getenv, &cfg)
```

`LoadFromEnvironment` finds the project root by walking upward to `go.mod`. `FILE_CONFIG` must be a file name without path components.

## Facade (full parity)

`helpers.NewHelpers` provides the investor-compatible stateful facade. Configuration is loaded lazily and cached on first access.

```go
h := helpers.NewHelpers(
    helpers.WithConfigStart(""),
    helpers.WithEnvLookup(os.Getenv),
    helpers.WithMigrationModels(&models.User{}, &models.Session{}),
    helpers.WithErrorCodeMapper(func(err error) int {
        switch {
        case errors.Is(err, domain.ErrNotFound):
            return web.NotFound
        case errors.Is(err, domain.ErrUnauthorized):
            return web.Unauthorized
        default:
            return web.InternalServerError
        }
    }),
)

if err := h.InitializeSystem(); err != nil {
    return err
}
defer h.GetDatabase().Close()
defer h.GetRedisClient().Close()

if err := h.RunMigration(); err != nil {
    return err
}
```

Available options are:

- `WithFactory(connections.Factory)` replaces connection constructors, primarily for tests.
- `WithConfigStart(path)` sets the starting path used to find `go.mod`.
- `WithEnvLookup(func(string) string)` replaces environment lookup.
- `WithErrorCodeMapper(web.ErrorCodeMapper)` injects domain error-to-HTTP-code policy.
- `WithMigrationModels(models ...any)` injects GORM migration models.
- `WithTimeProvider(timeutil.TimeProvider)` injects clock behavior.

Connection clients are available after `InitializeSystem`:

```go
db := h.GetDatabase()
redisClient := h.GetRedisClient()
s3Client := h.GetS3Client()
pubsubClient := h.GetPubSub()
mongoClient := h.GetMongoDB()
```

Database initialization is required and returns an error. Pub/Sub, S3, MongoDB, and telemetry initialization are best-effort. Cache and migration methods return clear errors when called before initialization.

The facade also provides Gin responses/context state, JWT generation/parsing, outbound HTTP requests, password hashing, OTP, validation, masking-related delegations, time helpers, cache, migration, and telemetry spans.

```go
h.SetBaseUrl("https://api.example.com")
h.SetTokenActive(accessToken)
body, err := h.MakeRequest(http.MethodGet, "/v1/profile", nil)

jwtToken, err := h.GenerateJWTToken(
    map[string]string{"user_id": userID},
    time.Now().Add(time.Hour),
)
```

`GenderToString` and typed publisher adapters remain service-owned because they depend on domain constants and event types. Domain error mapping is injected with `WithErrorCodeMapper`.

## Local storage

Use the S3-compatible API with local storage during development:

```toml
[s3]
provider = "local"
local_path = "./uploads/users"
```

```go
client, err := connections.NewS3Client(context.Background(), cfg.S3)
err = client.UploadFile(context.Background(), "docs/a.txt", data, "text/plain")
```

Local object keys reject absolute paths and `..` traversal.
