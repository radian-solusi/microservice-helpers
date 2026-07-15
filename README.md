# go-helpers

Shared Go helpers for Radian services.

Packages: config, strutil, cryptoutil, otp, phone, validate, mask, timeutil.

`cryptoutil.EncryptLegacyCBC` exists only for compatibility with existing
ciphertext. New encrypted data should use an authenticated construction.

OTP and crypto APIs require explicit key, issuer, time, and TTL values.

Config decoding remains consumer responsibility.

```go
cfg := config.MainConfig{}
normalized, err := phone.Validate("0812345678")
ciphertext, err := cryptoutil.EncryptLegacyCBC(data, []byte(cfg.App.AppKey))
token, err := otp.GeneratePreAuthToken(userID, key, time.Now(), 5*time.Minute)
```
