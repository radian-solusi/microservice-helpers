# microservice-helpers

Shared Go helpers for Radian microservices.

Packages: `config`, `connections`, `cryptoutil`, `helpers`, `mask`, `otp`, `phone`, `strutil`, `timeutil`, `validate`, and `web`.

## Installation

Add the module to a Go service:

```bash
go get github.com/radian-solusi/microservice-helpers@latest
```

Then import only the focused packages the service needs:

```go
import (
    "github.com/radian-solusi/microservice-helpers/cryptoutil"
    "github.com/radian-solusi/microservice-helpers/phone"
    "github.com/radian-solusi/microservice-helpers/strutil"
)
```

For compatibility with the investor `HelperInterface`, use the facade:

```go
import "github.com/radian-solusi/microservice-helpers/helpers"
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

### Telemetry (OpenTelemetry)

Traces are exported with OTLP gRPC. Default target is `telemetry.googleapis.com:443` using Application Default Credentials (Workload Identity / runtime service account). JSON key files are not used. `endpoint` and `insecure` are optional, for pointing to an OTel Collector.

```toml
[opentelemetry]
enabled = true
service_name = "my-service"
service_version = "1.0.0"
gcp_project_id = "my-gcp-project"   # fallback: env GOOGLE_CLOUD_PROJECT
# endpoint = ""      # opsional: OTLP gRPC host:port (mis. "otel-collector:4317"); kosong = telemetry.googleapis.com:443
# insecure = false   # opsional: true = gRPC tanpa TLS (hanya untuk collector internal); diabaikan untuk Google
```

| Mode | Target | Project ID | Result |
|---|---|---|---|
| Dev (`isProduction=false`) | any | any | Tracer provider only, no exporter, no network |
| Production | Google (`endpoint` empty or `telemetry.googleapis.com`) | empty | No exporter, no error |
| Production | Google | set | OTLP gRPC to `telemetry.googleapis.com:443`, TLS, ADC per-RPC credentials, header `x-goog-user-project`, resource `gcp.project_id`. Missing ADC returns an error (logged by `InitializeSystem`) |
| Production | Collector (non-Google `endpoint`) | any | OTLP gRPC to `endpoint` (default port `4317`), no Google credentials, TLS unless `insecure = true`. `gcp.project_id` resource attribute only when project ID is set |

GCP prerequisite: `telemetry.googleapis.com` must be enabled, and the runtime service account needs `roles/telemetry.writer` and `roles/serviceusage.serviceUsageConsumer`. Without them exports fail (`PermissionDenied` / `SERVICE_DISABLED`) and traces are lost.

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

## GCS storage

GCS is accessed through its S3-compatible XML API (`https://storage.googleapis.com`) using a
[Cloud Storage HMAC key](https://cloud.google.com/storage/docs/authentication/hmackeys) — the same
`connections.NewS3Client` / AWS SDK path used for `aws`/`minio`, just pointed at Google's endpoint.

```toml
[s3]
provider = "gcs"
bucket_name = "my-bucket"
access_key_id = "GOOG..."       # HMAC access ID
secret_access_key = "..."       # HMAC secret
# endpoint = "http://127.0.0.1:4443" # optional; defaults to https://storage.googleapis.com (e.g. fake-gcs-server)
# region   = "us-east-1"             # optional; defaults to "auto"
```

Path-style addressing is used automatically. `GetFileURL` returns a SigV4 presigned URL, same as `aws`/`minio`.
