package auth

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/golang-jwt/jwt/v5"
)

func reinitializeAuthTestState(t *testing.T) {
	t.Helper()

	if db.GetDB() != nil {
		if err := db.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}

	dbPath := filepath.Join(t.TempDir(), "octopus-auth.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	if err := op.InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	if err := op.SettingSetString(model.SettingKeyAuthTokenSecret, "auth-test-secret"); err != nil {
		t.Fatalf("SettingSetString() error = %v", err)
	}
	if err := op.UserInit(); err != nil {
		t.Fatalf("UserInit() error = %v", err)
	}
}

func setupAuthTestState(t *testing.T) {
	t.Helper()
	t.Setenv(op.BootstrapAdminUsernameEnv(), op.BootstrapAdminDefaultUsername())
	t.Setenv(op.BootstrapAdminPasswordEnv(), "admin")
	reinitializeAuthTestState(t)

	t.Cleanup(func() {
		if db.GetDB() != nil {
			if err := db.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		}
	})
}

func TestVerifyJWTTokenRejectsUnexpectedIssuer(t *testing.T) {
	setupAuthTestState(t)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "different-app",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	})

	secret, err := JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() error = %v", err)
	}
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() = true, want false for unexpected issuer")
	}
}

func TestVerifyJWTTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	setupAuthTestState(t)

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    conf.APP_NAME,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
	})

	secret, err := JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() error = %v", err)
	}
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() = true, want false for unexpected signing method")
	}
}

func TestVerifyJWTTokenAcceptsGeneratedToken(t *testing.T) {
	setupAuthTestState(t)

	signed, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	if !VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() = false, want true for generated token")
	}
}

func TestVerifyJWTTokenRejectsTokenAfterPasswordChange(t *testing.T) {
	setupAuthTestState(t)

	signed, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() before password change = false, want true")
	}
	if err := op.UserChangePassword("admin", "new-secret"); err != nil {
		t.Fatalf("UserChangePassword() error = %v", err)
	}
	if VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() after password change = true, want false")
	}
}

func TestVerifyJWTTokenRejectsTokenAfterUsernameChange(t *testing.T) {
	setupAuthTestState(t)

	signed, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() before username change = false, want true")
	}
	if err := op.UserChangeUsername("admin", "captain"); err != nil {
		t.Fatalf("UserChangeUsername() error = %v", err)
	}
	if VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() after username change = true, want false")
	}
}

func TestVerifyJWTTokenRejectsTokenAfterForcedPasswordChangeFlagClears(t *testing.T) {
	setupAuthTestState(t)
	t.Setenv(op.BootstrapAdminUsernameEnv(), "")
	t.Setenv(op.BootstrapAdminPasswordEnv(), "")
	reinitializeAuthTestState(t)

	signed, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() before force-change-password = false, want true")
	}
	if err := op.UserForceChangePassword("captain", "changed-secret"); err != nil {
		t.Fatalf("UserForceChangePassword() error = %v", err)
	}
	if VerifyJWTToken(signed) {
		t.Fatal("VerifyJWTToken() after force-change-password = true, want false")
	}
}
