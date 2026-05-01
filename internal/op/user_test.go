package op

import (
	"path/filepath"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
)

func setupUserInitTestDB(t *testing.T) {
	t.Helper()

	resetOpTestState()
	if db.GetDB() != nil {
		_ = db.Close()
	}

	dbPath := filepath.Join(t.TempDir(), "octopus-test.db")
	if err := db.InitDB("sqlite", dbPath, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	t.Cleanup(func() {
		if db.GetDB() != nil {
			_ = db.Close()
		}
		resetOpTestState()
	})
}

func TestUserInitUsesConfiguredBootstrapCredentials(t *testing.T) {
	setupUserInitTestDB(t)
	t.Setenv(bootstrapAdminUsernameEnv, "bootstrap-admin")
	t.Setenv(bootstrapAdminPasswordEnv, "bootstrap-secret")

	if err := UserInit(); err != nil {
		t.Fatalf("UserInit() error = %v", err)
	}
	if got := UserGet().Username; got != "bootstrap-admin" {
		t.Fatalf("username = %q, want %q", got, "bootstrap-admin")
	}
	if err := UserVerify("bootstrap-admin", "bootstrap-secret"); err != nil {
		t.Fatalf("UserVerify() error = %v", err)
	}
}

func TestUserInitUsesBuiltInBootstrapPasswordWhenUnset(t *testing.T) {
	setupUserInitTestDB(t)
	t.Setenv(bootstrapAdminUsernameEnv, "")
	t.Setenv(bootstrapAdminPasswordEnv, "")
	var loggedUsername string
	var loggedPassword string
	var loggedBuiltInDefault bool
	originalLogBootstrap := logBootstrapAdminCredentials
	logBootstrapAdminCredentials = func(username, password string, usingBuiltInDefaultPassword bool) {
		loggedUsername = username
		loggedPassword = password
		loggedBuiltInDefault = usingBuiltInDefaultPassword
	}
	t.Cleanup(func() {
		logBootstrapAdminCredentials = originalLogBootstrap
	})

	if err := UserInit(); err != nil {
		t.Fatalf("UserInit() error = %v", err)
	}
	if got := UserGet().Username; got != bootstrapAdminDefaultUsername {
		t.Fatalf("username = %q, want %q", got, bootstrapAdminDefaultUsername)
	}
	if loggedPassword != "" {
		t.Fatalf("built-in bootstrap password should be redacted before reaching log hook, got %q", loggedPassword)
	}
	if err := UserVerify(bootstrapAdminDefaultUsername, bootstrapAdminDefaultPassword); err != nil {
		t.Fatalf("UserVerify(built-in password) error = %v", err)
	}
	if !loggedBuiltInDefault {
		t.Fatalf("built-in bootstrap password should be marked for forced rotation")
	}
	if loggedUsername != bootstrapAdminDefaultUsername {
		t.Fatalf("logged username = %q, want %q", loggedUsername, bootstrapAdminDefaultUsername)
	}
	if !UserMustChangePassword() {
		t.Fatalf("UserMustChangePassword() = false, want true")
	}
}

func TestUserInitDoesNotExposeConfiguredBootstrapPasswordToLogHook(t *testing.T) {
	setupUserInitTestDB(t)
	t.Setenv(bootstrapAdminUsernameEnv, "bootstrap-admin")
	t.Setenv(bootstrapAdminPasswordEnv, "bootstrap-secret")
	var loggedUsername string
	var loggedPassword string
	var loggedGenerated bool
	originalLogBootstrap := logBootstrapAdminCredentials
	logBootstrapAdminCredentials = func(username, password string, generatedPassword bool) {
		loggedUsername = username
		loggedPassword = password
		loggedGenerated = generatedPassword
	}
	t.Cleanup(func() {
		logBootstrapAdminCredentials = originalLogBootstrap
	})

	if err := UserInit(); err != nil {
		t.Fatalf("UserInit() error = %v", err)
	}
	if loggedUsername != "bootstrap-admin" {
		t.Fatalf("logged username = %q, want %q", loggedUsername, "bootstrap-admin")
	}
	if loggedGenerated {
		t.Fatalf("configured bootstrap password should not be marked as generated")
	}
	if loggedPassword != "" {
		t.Fatalf("configured bootstrap password should be redacted before reaching log hook, got %q", loggedPassword)
	}
	if UserMustChangePassword() {
		t.Fatalf("UserMustChangePassword() = true, want false when env password is provided")
	}
}

func TestUserForceChangePasswordClearsRequiredFlag(t *testing.T) {
	setupUserInitTestDB(t)
	t.Setenv(bootstrapAdminUsernameEnv, "")
	t.Setenv(bootstrapAdminPasswordEnv, "")

	if err := UserInit(); err != nil {
		t.Fatalf("UserInit() error = %v", err)
	}
	if !UserMustChangePassword() {
		t.Fatalf("UserMustChangePassword() = false, want true")
	}

	if err := UserForceChangePassword("captain", "changed-secret"); err != nil {
		t.Fatalf("UserForceChangePassword() error = %v", err)
	}
	if UserMustChangePassword() {
		t.Fatalf("UserMustChangePassword() = true, want false")
	}
	if got := UserGet().Username; got != "captain" {
		t.Fatalf("username = %q, want %q", got, "captain")
	}
	if err := UserVerify("captain", "changed-secret"); err != nil {
		t.Fatalf("UserVerify(changed-secret) error = %v", err)
	}
}
