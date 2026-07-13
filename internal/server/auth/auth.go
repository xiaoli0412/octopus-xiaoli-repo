package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultJWTExpiry  = 15 * time.Minute
	rememberJWTExpiry = 30 * 24 * time.Hour
	maxJWTExpiry      = rememberJWTExpiry
)

func GenerateJWTToken(expiresMin int) (string, string, error) {
	now := time.Now()
	expiry, err := resolveJWTExpiry(expiresMin)
	if err != nil {
		return "", "", err
	}
	secret, err := jwtSigningSecret()
	if err != nil {
		return "", "", err
	}
	subject, sessionID, err := currentJWTSession(secret)
	if err != nil {
		return "", "", err
	}
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    conf.APP_NAME,
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		Subject:   subject,
		ID:        sessionID,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", "", err
	}
	return token, claims.ExpiresAt.Time.Format(time.RFC3339), nil
}

func resolveJWTExpiry(expiresMin int) (time.Duration, error) {
	switch {
	case expiresMin == -1:
		return rememberJWTExpiry, nil
	case expiresMin == 0:
		return defaultJWTExpiry, nil
	case expiresMin < -1:
		return 0, fmt.Errorf("expire must be -1, 0, or a positive number of minutes")
	default:
		expiry := time.Duration(expiresMin) * time.Minute
		if expiry > maxJWTExpiry {
			return maxJWTExpiry, nil
		}
		return expiry, nil
	}
}

func VerifyJWTToken(token string) bool {
	secret, err := jwtSigningSecret()
	if err != nil {
		return false
	}
	if verifyJWTTokenWithSecret(token, secret) {
		return true
	}
	secondary, err := jwtSecondarySigningSecret()
	if err != nil {
		return false
	}
	return verifyJWTTokenWithSecret(token, secondary)
}

func verifyJWTTokenWithSecret(token string, secret []byte) bool {
	claims := &jwt.RegisteredClaims{}
	jwtToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(conf.APP_NAME))
	if err != nil || !jwtToken.Valid {
		return false
	}
	expectedSubject, expectedSessionID, err := currentJWTSession(secret)
	if err != nil {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(claims.Subject), []byte(expectedSubject)) != 1 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(claims.ID), []byte(expectedSessionID)) != 1 {
		return false
	}
	return true
}

func currentJWTSession(secret []byte) (string, string, error) {
	user := op.UserGet()
	username := strings.TrimSpace(user.Username)
	passwordHash := strings.TrimSpace(user.Password)
	if username == "" || passwordHash == "" {
		return "", "", fmt.Errorf("admin user is not initialized")
	}

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(username))
	_, _ = mac.Write([]byte{'\n'})
	_, _ = mac.Write([]byte(passwordHash))
	_, _ = mac.Write([]byte{'\n'})
	_, _ = mac.Write([]byte(fmt.Sprintf("%t", op.UserMustChangePassword())))

	return username, hex.EncodeToString(mac.Sum(nil)), nil
}

func jwtSigningSecret() ([]byte, error) {
	secret, err := op.SettingGetString(model.SettingKeyAuthTokenSecret)
	if err != nil {
		return nil, fmt.Errorf("load auth token secret: %w", err)
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("auth token secret is empty")
	}
	return []byte(secret), nil
}

func jwtSecondarySigningSecret() ([]byte, error) {
	secret, err := op.SettingGetString(model.SettingKeyAuthTokenSecretSecondary)
	if err != nil {
		return nil, fmt.Errorf("load auth token secondary secret: %w", err)
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("auth token secondary secret is empty")
	}
	return []byte(secret), nil
}

func JWTSigningSecretForTests() ([]byte, error) {
	return jwtSigningSecret()
}

func JWTSecondarySigningSecretForTests() ([]byte, error) {
	return jwtSecondarySigningSecret()
}

func GenerateAPIKey() string {
	const keyChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 48)
	maxI := big.NewInt(int64(len(keyChars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, maxI)
		if err != nil {
			return ""
		}
		b[i] = keyChars[n.Int64()]
	}
	return "sk-" + conf.APP_NAME + "-" + string(b)
}
