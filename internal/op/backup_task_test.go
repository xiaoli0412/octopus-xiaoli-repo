package op

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestBackupIntervalDefault(t *testing.T) {
	SetupOpTestDB(t)
	got := backupInterval()
	if got != defaultBackupInterval {
		t.Fatalf("default interval = %v, want %v", got, defaultBackupInterval)
	}
}

func TestBackupIntervalCustom(t *testing.T) {
	SetupOpTestDB(t)
	if err := SettingSetString(model.SettingKeyBackupInterval, "12h"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}
	got := backupInterval()
	if got != 12*time.Hour {
		t.Fatalf("custom interval = %v, want 12h", got)
	}
}

func TestBackupIntervalInvalidFallback(t *testing.T) {
	SetupOpTestDB(t)
	// Bypass validation by setting cache directly.
	settingCache.Set(model.SettingKeyBackupInterval, "not-a-duration")
	got := backupInterval()
	if got != defaultBackupInterval {
		t.Fatalf("invalid interval fallback = %v, want %v", got, defaultBackupInterval)
	}

	// Empty value falls back to default.
	settingCache.Set(model.SettingKeyBackupInterval, "")
	got = backupInterval()
	if got != defaultBackupInterval {
		t.Fatalf("empty interval fallback = %v, want %v", got, defaultBackupInterval)
	}

	// Zero or negative duration falls back to default.
	settingCache.Set(model.SettingKeyBackupInterval, "0s")
	got = backupInterval()
	if got != defaultBackupInterval {
		t.Fatalf("zero interval fallback = %v, want %v", got, defaultBackupInterval)
	}
}

func TestBackupKeepCountDefault(t *testing.T) {
	SetupOpTestDB(t)
	got := backupKeepCount()
	if got != defaultBackupKeepCount {
		t.Fatalf("default keep count = %d, want %d", got, defaultBackupKeepCount)
	}
}

func TestBackupKeepCountCustom(t *testing.T) {
	SetupOpTestDB(t)
	if err := SettingSetInt(model.SettingKeyBackupKeepCount, 3); err != nil {
		t.Fatalf("SettingSetInt error = %v", err)
	}
	got := backupKeepCount()
	if got != 3 {
		t.Fatalf("custom keep count = %d, want 3", got)
	}
}

func TestBackupKeepCountInvalidFallback(t *testing.T) {
	SetupOpTestDB(t)
	// Bypass validation by setting cache directly with a negative value.
	settingCache.Set(model.SettingKeyBackupKeepCount, "-5")
	got := backupKeepCount()
	if got != defaultBackupKeepCount {
		t.Fatalf("invalid keep count fallback = %d, want %d", got, defaultBackupKeepCount)
	}
}

func TestStartBackupTaskNilRegister(t *testing.T) {
	SetupOpTestDB(t)
	oldFn := taskRegisterFn
	taskRegisterFn = nil
	defer func() { taskRegisterFn = oldFn }()
	// Should not panic when task register is nil.
	StartBackupTask()
}

func TestStartBackupTaskRegistration(t *testing.T) {
	SetupOpTestDB(t)
	var registeredName string
	var registeredInterval time.Duration
	oldFn := taskRegisterFn
	taskRegisterFn = func(name string, interval time.Duration, runOnStart bool, fn func()) {
		registeredName = name
		registeredInterval = interval
	}
	defer func() { taskRegisterFn = oldFn }()

	StartBackupTask()

	if registeredName != backupTaskName {
		t.Fatalf("registered name = %q, want %q", registeredName, backupTaskName)
	}
	if registeredInterval != defaultBackupInterval {
		t.Fatalf("registered interval = %v, want %v", registeredInterval, defaultBackupInterval)
	}
}

func TestStartBackupTaskCustomInterval(t *testing.T) {
	SetupOpTestDB(t)
	if err := SettingSetString(model.SettingKeyBackupInterval, "6h"); err != nil {
		t.Fatalf("SettingSetString error = %v", err)
	}

	var registeredInterval time.Duration
	oldFn := taskRegisterFn
	taskRegisterFn = func(name string, interval time.Duration, runOnStart bool, fn func()) {
		registeredInterval = interval
	}
	defer func() { taskRegisterFn = oldFn }()

	StartBackupTask()

	if registeredInterval != 6*time.Hour {
		t.Fatalf("registered interval = %v, want 6h", registeredInterval)
	}
}

func TestResolveSQLiteDBPath(t *testing.T) {
	SetupOpTestDB(t)
	// After SetupOpTestDB, db.GetCurrentDBType() should be "sqlite".
	got := resolveSQLiteDBPath()
	if got == "" {
		t.Fatal("expected non-empty SQLite DB path after SetupOpTestDB")
	}
}

func TestBackupDir(t *testing.T) {
	SetupOpTestDB(t)
	dir, err := backupDir()
	if err != nil {
		t.Fatalf("backupDir() error = %v", err)
	}
	if dir == "" {
		t.Fatal("expected non-empty backup directory")
	}
	if !strings.HasSuffix(dir, backupDirName) {
		t.Fatalf("backup dir = %q, want suffix %q", dir, backupDirName)
	}
}

func TestRunBackupTaskCreatesBackup(t *testing.T) {
	SetupOpTestDB(t)

	runBackupTask()

	dir, err := backupDir()
	if err != nil {
		t.Fatalf("backupDir() error = %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", dir, err)
	}

	found := false
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, backupFilePrefix) && strings.HasSuffix(name, backupFileSuffix) {
			found = true
			info, err := entry.Info()
			if err != nil {
				t.Fatalf("entry.Info() error = %v", err)
			}
			if info.Size() == 0 {
				t.Fatalf("backup file %q is empty", name)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected at least one backup file to be created")
	}
}

func TestRunBackupTaskPrunesOldBackups(t *testing.T) {
	SetupOpTestDB(t)

	// Set keep count to 2.
	if err := SettingSetInt(model.SettingKeyBackupKeepCount, 2); err != nil {
		t.Fatalf("SettingSetInt error = %v", err)
	}

	dir, err := backupDir()
	if err != nil {
		t.Fatalf("backupDir() error = %v", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Pre-create 3 old backup files with descending mod times.
	base := time.Now().Add(-48 * time.Hour)
	for i := 0; i < 3; i++ {
		name := filepath.Join(dir, backupFilePrefix+base.Add(time.Duration(i)*time.Hour).Format(backupTimestampFormat)+backupFileSuffix)
		if err := os.WriteFile(name, []byte("old-backup"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		modTime := base.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(name, modTime, modTime); err != nil {
			t.Fatalf("Chtimes() error = %v", err)
		}
	}

	// Run backup task — should create 1 new backup and prune to keepCount=2.
	runBackupTask()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, backupFilePrefix) && strings.HasSuffix(name, backupFileSuffix) {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("after backup+prune, count = %d, want 2", count)
	}
}

func TestPruneOldBackups(t *testing.T) {
	SetupOpTestDB(t)
	dir := t.TempDir()

	// Create 5 backup files with distinct mod times (newest first after sort).
	base := time.Now()
	for i := 0; i < 5; i++ {
		name := filepath.Join(dir, backupFilePrefix+base.Add(-time.Duration(i)*time.Hour).Format(backupTimestampFormat)+backupFileSuffix)
		if err := os.WriteFile(name, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		modTime := base.Add(-time.Duration(i) * time.Hour)
		if err := os.Chtimes(name, modTime, modTime); err != nil {
			t.Fatalf("Chtimes() error = %v", err)
		}
	}

	// Keep only 2.
	if err := pruneOldBackups(dir, 2); err != nil {
		t.Fatalf("pruneOldBackups() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, backupFilePrefix) && strings.HasSuffix(name, backupFileSuffix) {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("after pruning, count = %d, want 2", count)
	}
}

func TestPruneOldBackupsKeepAll(t *testing.T) {
	SetupOpTestDB(t)
	dir := t.TempDir()

	for i := 0; i < 3; i++ {
		name := filepath.Join(dir, backupFilePrefix+time.Now().Add(-time.Duration(i)*time.Hour).Format(backupTimestampFormat)+backupFileSuffix)
		if err := os.WriteFile(name, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	// keepCount <= 0 means keep all.
	if err := pruneOldBackups(dir, 0); err != nil {
		t.Fatalf("pruneOldBackups() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, backupFilePrefix) && strings.HasSuffix(name, backupFileSuffix) {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("keep all, count = %d, want 3", count)
	}
}

func TestPruneOldBackupsIgnoresNonBackupFiles(t *testing.T) {
	SetupOpTestDB(t)
	dir := t.TempDir()

	// Create 2 backup files and 1 unrelated file.
	for i := 0; i < 2; i++ {
		name := filepath.Join(dir, backupFilePrefix+time.Now().Add(-time.Duration(i)*time.Hour).Format(backupTimestampFormat)+backupFileSuffix)
		if err := os.WriteFile(name, []byte("backup"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "other.txt"), []byte("other"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Keep 1 — only backup files should be counted/pruned.
	if err := pruneOldBackups(dir, 1); err != nil {
		t.Fatalf("pruneOldBackups() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	backupCount := 0
	otherExists := false
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, backupFilePrefix) && strings.HasSuffix(name, backupFileSuffix) {
			backupCount++
		}
		if name == "other.txt" {
			otherExists = true
		}
	}
	if backupCount != 1 {
		t.Fatalf("backup count = %d, want 1", backupCount)
	}
	if !otherExists {
		t.Fatal("non-backup file should not be removed")
	}
}

func TestCopyFile(t *testing.T) {
	SetupOpTestDB(t)
	src := filepath.Join(t.TempDir(), "src.txt")
	dst := filepath.Join(t.TempDir(), "dst.txt")
	content := []byte("hello backup")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("copied content = %q, want %q", string(got), string(content))
	}
}

func TestCopyFileOverwrite(t *testing.T) {
	SetupOpTestDB(t)
	src := filepath.Join(t.TempDir(), "src.txt")
	dst := filepath.Join(t.TempDir(), "dst.txt")
	if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(dst, []byte("old-content-that-is-longer"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("copied content = %q, want %q", string(got), "new")
	}
}
