package op

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"gorm.io/gorm"
)

var userCache model.User
var userCacheLock sync.RWMutex

var (
	ErrInvalidUsername          = errors.New("new username cannot be empty")
	ErrIncorrectCurrentPassword = errors.New("incorrect current password")
	ErrIncorrectOldPassword     = errors.New("incorrect old password")
	ErrInvalidNewPassword       = errors.New("new password cannot be empty")
	ErrSameUsername             = errors.New("new username is the same as the old username")
	ErrForcePasswordNotRequired = errors.New("force password change is not required")
)

const (
	bootstrapAdminDefaultUsername = "admin"
	bootstrapAdminUsernameEnv     = "OCTOPUS_ADMIN_USERNAME"
	bootstrapAdminPasswordEnv     = "OCTOPUS_ADMIN_PASSWORD"
)

var logBootstrapAdminCredentials = func(username, password string, usingBuiltInDefaultPassword bool) {
	if usingBuiltInDefaultPassword {
		log.Warnf("initial administrator created with generated password: %s — sign in and change the password before using the console", password)
		return
	}
	log.Warnf("initial administrator created from %s/%s; rotate these bootstrap credentials immediately", bootstrapAdminUsernameEnv, bootstrapAdminPasswordEnv)
}

func generateRandomPassword() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random password: %v", err))
	}
	return hex.EncodeToString(b)
}

func parseBoolSettingValue(value string) bool {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return parsed
}

func persistForcePasswordChange(required bool) error {
	value := "false"
	if required {
		value = "true"
	}

	setting := model.Setting{Key: model.SettingKeyForcePasswordChange, Value: value}
	if err := setting.Validate(); err != nil {
		return err
	}

	result := db.GetDB().Save(&setting)
	if result.Error != nil {
		return fmt.Errorf("persist force password change: %w", result.Error)
	}

	settingCache.Set(model.SettingKeyForcePasswordChange, value)
	return nil
}

func UserMustChangePassword() bool {
	if value, ok := settingCache.Get(model.SettingKeyForcePasswordChange); ok {
		return parseBoolSettingValue(value)
	}

	var setting model.Setting
	if err := db.GetDB().First(&setting, "key = ?", model.SettingKeyForcePasswordChange).Error; err == nil {
		settingCache.Set(model.SettingKeyForcePasswordChange, setting.Value)
		return parseBoolSettingValue(setting.Value)
	}

	return false
}

func BootstrapAdminDefaultUsername() string {
	return bootstrapAdminDefaultUsername
}

func BootstrapAdminUsernameEnv() string {
	return bootstrapAdminUsernameEnv
}

func BootstrapAdminPasswordEnv() string {
	return bootstrapAdminPasswordEnv
}

func resolveBootstrapAdminCredentials() (string, string, bool, error) {
	username := strings.TrimSpace(os.Getenv(bootstrapAdminUsernameEnv))
	if username == "" {
		username = bootstrapAdminDefaultUsername
	}

	password := os.Getenv(bootstrapAdminPasswordEnv)
	if strings.TrimSpace(password) != "" {
		return username, password, false, nil
	}
	return username, generateRandomPassword(), true, nil
}

func UserInit() error {
	var loaded model.User
	if err := db.GetDB().Order("id asc").First(&loaded).Error; err == nil {
		userCacheLock.Lock()
		userCache = loaded
		userCacheLock.Unlock()
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	username, password, usingBuiltInDefaultPassword, err := resolveBootstrapAdminCredentials()
	if err != nil {
		return err
	}
	adminUser := model.User{Username: username, Password: password}
	if err := adminUser.HashPassword(); err != nil {
		return err
	}
	if err := db.GetDB().Create(&adminUser).Error; err != nil {
		return err
	}
	userCacheLock.Lock()
	userCache = adminUser
	userCacheLock.Unlock()
	if err := persistForcePasswordChange(usingBuiltInDefaultPassword); err != nil {
		return err
	}
	logBootstrapAdminCredentials(username, password, usingBuiltInDefaultPassword)
	return nil
}

func UserChangePassword(oldPassword, newPassword string) error {
	userCacheLock.Lock()
	defer userCacheLock.Unlock()
	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return ErrInvalidNewPassword
	}
	if err := userCache.ComparePassword(oldPassword); err != nil {
		return ErrIncorrectOldPassword
	}

	updatedUser := userCache
	updatedUser.Password = newPassword
	if err := updatedUser.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := db.GetDB().Model(&userCache).Update("password", updatedUser.Password).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	userCache.Password = updatedUser.Password
	if err := persistForcePasswordChange(false); err != nil {
		return err
	}

	return nil
}

func UserForceChangePassword(newUsername, newPassword string) error {
	userCacheLock.Lock()
	defer userCacheLock.Unlock()
	newUsername = strings.TrimSpace(newUsername)
	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return ErrInvalidNewPassword
	}
	if !UserMustChangePassword() {
		return ErrForcePasswordNotRequired
	}

	updatedUser := userCache
	updatedUser.Password = newPassword
	if err := updatedUser.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := db.GetDB().Model(&userCache).Update("password", updatedUser.Password).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	userCache.Password = updatedUser.Password

	if newUsername != "" && newUsername != userCache.Username {
		if err := db.GetDB().Model(&userCache).Update("username", newUsername).Error; err != nil {
			return fmt.Errorf("failed to update username: %w", err)
		}
		userCache.Username = newUsername
	}
	if err := persistForcePasswordChange(false); err != nil {
		return err
	}

	return nil
}

func UserChangeUsername(currentPassword, newUsername string) error {
	userCacheLock.Lock()
	defer userCacheLock.Unlock()
	newUsername = strings.TrimSpace(newUsername)
	if newUsername == "" {
		return ErrInvalidUsername
	}
	if userCache.Username == newUsername {
		return ErrSameUsername
	}
	if err := userCache.ComparePassword(currentPassword); err != nil {
		return ErrIncorrectCurrentPassword
	}
	if err := db.GetDB().Model(&userCache).Update("username", newUsername).Error; err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}
	userCache.Username = newUsername
	return nil
}

func UserVerify(username, password string) error {
	userCacheLock.RLock()
	defer userCacheLock.RUnlock()
	if username != userCache.Username {
		return fmt.Errorf("incorrect username")
	}
	if err := userCache.ComparePassword(password); err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}

func UserGet() model.User {
	userCacheLock.RLock()
	defer userCacheLock.RUnlock()
	return userCache
}
