package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/radian-solusi/microservice-helpers/cryptoutil"
)

type JWT struct {
	key []byte
}

func NewJWT(key []byte) (*JWT, error) {
	if len(key) != 32 {
		return nil, errors.New("jwt key must be 32 bytes")
	}
	return &JWT{key: key}, nil
}

func (j *JWT) Generate(payload any, expires time.Time) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}
	encrypted, err := cryptoutil.EncryptLegacyCBC(raw, j.key)
	if err != nil {
		return "", fmt.Errorf("encrypt payload: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": encrypted,
		"exp":  expires.Unix(),
	})
	signed, err := token.SignedString(j.key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (j *JWT) Parse(token string, payload any) error {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.key, nil
	})
	if err != nil || !parsed.Valid {
		return fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}
	encrypted, ok := claims["data"].(string)
	if !ok {
		return errors.New("token missing data claim")
	}
	decrypted, err := cryptoutil.DecryptLegacyCBC(encrypted, j.key)
	if err != nil {
		return fmt.Errorf("decrypt payload: %w", err)
	}
	if err := json.Unmarshal(decrypted, payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	return nil
}
