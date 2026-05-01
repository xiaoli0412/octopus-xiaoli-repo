package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	serverauth "github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/auth"
	"github.com/golang-jwt/jwt/v5"
)

const importPreviewTokenTTL = 30 * time.Minute

type importPreviewFingerprint struct {
	DumpHash      string                `json:"dump_hash"`
	Mode          model.DBImportMode    `json:"mode"`
	ModelMappings map[string]string     `json:"model_mappings,omitempty"`
	ImportScopes  *model.DBImportScopes `json:"import_scopes,omitempty"`
}

type importPreviewTokenClaims struct {
	Digest string `json:"digest"`
	jwt.RegisteredClaims
}

func buildImportPreviewDigest(dump *model.DBDump, mode model.DBImportMode, options model.DBImportOptions) (string, error) {
	if dump == nil {
		return "", fmt.Errorf("empty import dump")
	}
	payload, err := json.Marshal(dump)
	if err != nil {
		return "", fmt.Errorf("marshal import dump: %w", err)
	}
	dumpHash := sha256.Sum256(payload)
	fingerprint := importPreviewFingerprint{
		DumpHash:      hex.EncodeToString(dumpHash[:]),
		Mode:          model.DefaultDBImportMode(string(mode)),
		ModelMappings: normalizeImportPreviewModelMappings(options.ModelMappings),
		ImportScopes:  normalizeImportPreviewScopes(options.ImportScopes),
	}
	fingerprintPayload, err := json.Marshal(fingerprint)
	if err != nil {
		return "", fmt.Errorf("marshal import preview fingerprint: %w", err)
	}
	fingerprintHash := sha256.Sum256(fingerprintPayload)
	return hex.EncodeToString(fingerprintHash[:]), nil
}

func signImportPreviewToken(digest string) (string, error) {
	now := time.Now().UTC()
	claims := importPreviewTokenClaims{
		Digest: digest,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(importPreviewTokenTTL)),
			Issuer:    conf.APP_NAME,
			Subject:   "import-preview",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(importPreviewTokenSecret())
}

func verifyImportPreviewToken(token string, digest string) error {
	trimmedToken := strings.TrimSpace(token)
	if trimmedToken == "" {
		return fmt.Errorf("preview_token is required")
	}
	claims := &importPreviewTokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(trimmedToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected preview token signing method")
		}
		return importPreviewTokenSecret(), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(conf.APP_NAME), jwt.WithSubject("import-preview"))
	if err != nil || parsedToken == nil || !parsedToken.Valid {
		return fmt.Errorf("preview_token is invalid or expired; rerun dry-run")
	}
	if claims.Digest != digest {
		return fmt.Errorf("preview_token does not match current import payload; rerun dry-run")
	}
	return nil
}

func importPreviewTokenSecret() []byte {
	baseSecret, err := serverauth.JWTSigningSecretForTests()
	if err != nil {
		return []byte(conf.APP_NAME + "\nimport-preview-token")
	}
	secret := strings.Join([]string{
		string(baseSecret),
		conf.APP_NAME,
		"import-preview-token",
	}, "\n")
	return []byte(secret)
}

func normalizeImportPreviewModelMappings(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]string, len(input))
	for source, target := range input {
		normalizedSource := strings.ToLower(strings.TrimSpace(source))
		normalizedTarget := strings.ToLower(strings.TrimSpace(target))
		if normalizedSource == "" || normalizedTarget == "" {
			continue
		}
		result[normalizedSource] = normalizedTarget
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func normalizeImportPreviewScopes(scopes *model.DBImportScopes) *model.DBImportScopes {
	if scopes == nil {
		return nil
	}
	if scopes.Routing && scopes.Models && scopes.APIKeys && scopes.Settings && scopes.Stats && scopes.Logs {
		return nil
	}
	return &model.DBImportScopes{
		Routing:  scopes.Routing,
		Models:   scopes.Models,
		APIKeys:  scopes.APIKeys,
		Settings: scopes.Settings,
		Stats:    scopes.Stats,
		Logs:     scopes.Logs,
	}
}
