package auth

import (
	"context"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
)

func TestVerifyJWTTokenPrimaryOnlyWhenSecondaryEmpty(t *testing.T) {
	setupAuthTestState(t)

	token, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() = false, want true for primary-signed token with empty secondary")
	}
}

func TestVerifyJWTTokenAcceptsOldTokenAfterRotation(t *testing.T) {
	setupAuthTestState(t)

	token, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() before rotation = false, want true")
	}

	if err := op.RotateAuthTokenSecret(context.Background()); err != nil {
		t.Fatalf("RotateAuthTokenSecret() error = %v", err)
	}

	if !VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() after rotation = false, want true (old token should verify via secondary)")
	}
}

func TestVerifyJWTTokenAcceptsNewTokenAfterRotation(t *testing.T) {
	setupAuthTestState(t)

	if err := op.RotateAuthTokenSecret(context.Background()); err != nil {
		t.Fatalf("RotateAuthTokenSecret() error = %v", err)
	}

	token, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() = false, want true for new primary-signed token after rotation")
	}
}

func TestVerifyJWTTokenRejectsTokenAfterBothSecretsChange(t *testing.T) {
	setupAuthTestState(t)

	token, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if !VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() before rotation = false, want true")
	}

	if err := op.RotateAuthTokenSecret(context.Background()); err != nil {
		t.Fatalf("RotateAuthTokenSecret() first rotation error = %v", err)
	}

	if !VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() after first rotation = false, want true (old primary now secondary)")
	}

	if err := op.RotateAuthTokenSecret(context.Background()); err != nil {
		t.Fatalf("RotateAuthTokenSecret() second rotation error = %v", err)
	}

	if VerifyJWTToken(token) {
		t.Fatal("VerifyJWTToken() after second rotation = true, want false (original signing key no longer primary or secondary)")
	}
}

func TestVerifyJWTTokenSigningUsesPrimaryKey(t *testing.T) {
	setupAuthTestState(t)

	primarySecret, err := JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() error = %v", err)
	}

	secondarySecret := make([]byte, len(primarySecret))
	copy(secondarySecret, primarySecret)
	secondarySecret[0] ^= 0xff

	if err := op.SettingSetString(model.SettingKeyAuthTokenSecretSecondary, string(secondarySecret)); err != nil {
		t.Fatalf("SettingSetString() secondary error = %v", err)
	}

	token, _, err := GenerateJWTToken(0)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	currentPrimary, err := JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() after generation error = %v", err)
	}

	if string(currentPrimary) != string(primarySecret) {
		t.Fatal("primary secret changed after token generation; signing should use primary")
	}

	if !verifyJWTTokenWithSecret(token, currentPrimary) {
		t.Fatal("newly generated token should verify with primary secret")
	}
}

func TestRotateAuthTokenSecretMovesPrimaryToSecondary(t *testing.T) {
	setupAuthTestState(t)

	originalPrimary, err := JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() error = %v", err)
	}

	if err := op.RotateAuthTokenSecret(context.Background()); err != nil {
		t.Fatalf("RotateAuthTokenSecret() error = %v", err)
	}

	newPrimary, err := JWTSigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSigningSecretForTests() after rotation error = %v", err)
	}
	newSecondary, err := JWTSecondarySigningSecretForTests()
	if err != nil {
		t.Fatalf("JWTSecondarySigningSecretForTests() after rotation error = %v", err)
	}

	if string(newPrimary) == string(originalPrimary) {
		t.Fatal("primary secret did not change after rotation")
	}
	if string(newSecondary) != string(originalPrimary) {
		t.Fatal("secondary secret should equal the original primary after rotation")
	}
}
