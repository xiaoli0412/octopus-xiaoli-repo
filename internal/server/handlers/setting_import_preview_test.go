package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestVerifyImportPreviewTokenAcceptsSignedToken(t *testing.T) {
	setupHandlerTest(t)

	const digest = "digest-accept"
	token, err := signImportPreviewToken(digest)
	if err != nil {
		t.Fatalf("signImportPreviewToken() error = %v", err)
	}

	if err := verifyImportPreviewToken(token, digest); err != nil {
		t.Fatalf("verifyImportPreviewToken() error = %v, want nil", err)
	}
}

func TestVerifyImportPreviewTokenRejectsUnexpectedIssuer(t *testing.T) {
	setupHandlerTest(t)

	now := time.Now().UTC()
	token := signCustomImportPreviewTokenForTest(t, "digest-issuer", jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "different-app",
		Subject:   "import-preview",
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	})

	err := verifyImportPreviewToken(token, "digest-issuer")
	if err == nil || !strings.Contains(err.Error(), "preview_token is invalid or expired") {
		t.Fatalf("verifyImportPreviewToken() error = %v, want invalid or expired", err)
	}
}

func TestVerifyImportPreviewTokenRejectsUnexpectedSubject(t *testing.T) {
	setupHandlerTest(t)

	now := time.Now().UTC()
	token := signCustomImportPreviewTokenForTest(t, "digest-subject", jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    conf.APP_NAME,
		Subject:   "user-session",
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	})

	err := verifyImportPreviewToken(token, "digest-subject")
	if err == nil || !strings.Contains(err.Error(), "preview_token is invalid or expired") {
		t.Fatalf("verifyImportPreviewToken() error = %v, want invalid or expired", err)
	}
}

func TestVerifyImportPreviewTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	setupHandlerTest(t)

	now := time.Now().UTC()
	token := signCustomImportPreviewTokenForTest(t, "digest-method", jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    conf.APP_NAME,
		Subject:   "import-preview",
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	})

	err := verifyImportPreviewToken(token, "digest-method")
	if err == nil || !strings.Contains(err.Error(), "preview_token is invalid or expired") {
		t.Fatalf("verifyImportPreviewToken() error = %v, want invalid or expired", err)
	}
}

func TestBuildImportPreviewDigestRejectsInvalidMode(t *testing.T) {
	setupHandlerTest(t)

	_, err := buildImportPreviewDigest(&model.DBDump{Version: 1}, model.DBImportMode("surprise"), model.DBImportOptions{})
	if err == nil {
		t.Fatal("buildImportPreviewDigest() expected unsupported import mode error")
	}
	if !strings.Contains(err.Error(), "unsupported import mode") {
		t.Fatalf("buildImportPreviewDigest() error = %v, want unsupported import mode", err)
	}
}

func TestBuildImportPreviewDigestAcceptsValidMode(t *testing.T) {
	setupHandlerTest(t)

	if _, err := buildImportPreviewDigest(&model.DBDump{Version: 1}, model.DBImportModeIncremental, model.DBImportOptions{}); err != nil {
		t.Fatalf("buildImportPreviewDigest() error = %v, want nil", err)
	}
}

func TestBuildImportPreviewDigestRejectsBlankModelMappings(t *testing.T) {
	setupHandlerTest(t)

	tests := []struct {
		name    string
		mapping map[string]string
	}{
		{name: "blank source", mapping: map[string]string{"   ": "gpt-4o"}},
		{name: "blank target", mapping: map[string]string{"legacy-model": "   "}},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildImportPreviewDigest(&model.DBDump{Version: 1}, model.DBImportModeMap, model.DBImportOptions{ModelMappings: tc.mapping})
			if err == nil || !strings.Contains(err.Error(), "invalid model_mappings") {
				t.Fatalf("buildImportPreviewDigest() error = %v, want invalid model_mappings", err)
			}
		})
	}
}

func signCustomImportPreviewTokenForTest(t *testing.T, digest string, method jwt.SigningMethod, claims jwt.RegisteredClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(method, importPreviewTokenClaims{
		Digest:           digest,
		RegisteredClaims: claims,
	})
	signed, err := token.SignedString(importPreviewTokenSecret())
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}
