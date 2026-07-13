package op

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const (
	backupTaskName        = "scheduled_backup"
	defaultBackupInterval = 24 * time.Hour
	defaultBackupKeepCount = 7
	backupDirName          = "backups"
	backupFilePrefix       = "backup-"
	backupFileSuffix       = ".db"
	backupTimestampFormat  = "20060102-150405"
)

// StartBackupTask registers the periodic SQLite backup task with the task
// scheduler. The default interval is 24 hours, configurable via the
// backup_interval setting (duration string like "24h"). Old backups beyond
// backup_keep_count (default 7) are automatically pruned by modification time.
func StartBackupTask() {
	if taskRegisterFn == nil {
		log.Warnf("backup task not registered: task register not set")
		return
	}
	interval := backupInterval()
	taskRegisterFn(backupTaskName, interval, false, runBackupTask)
	log.Infof("backup task registered with interval %v", interval)
}

func backupInterval() time.Duration {
	raw, err := SettingGetString(model.SettingKeyBackupInterval)
	if err != nil || strings.TrimSpace(raw) == "" {
		return defaultBackupInterval
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultBackupInterval
	}
	return d
}

func backupKeepCount() int {
	v, err := SettingGetInt(model.SettingKeyBackupKeepCount)
	if err != nil || v < 0 {
		return defaultBackupKeepCount
	}
	return v
}

// backupDir returns the directory where backup files are stored, derived
// from the SQLite database file location. It falls back to "data/backups"
// when the database path cannot be resolved.
func backupDir() (string, error) {
	dbPath := resolveSQLiteDBPath()
	if dbPath != "" {
		abs, err := filepath.Abs(dbPath)
		if err != nil {
			return filepath.Abs(filepath.Join(filepath.Dir(dbPath), backupDirName))
		}
		return filepath.Join(filepath.Dir(abs), backupDirName), nil
	}
	return filepath.Abs(filepath.Join("data", backupDirName))
}

// resolveSQLiteDBPath returns the current SQLite database file path, or an
// empty string when the database type is not SQLite.
func resolveSQLiteDBPath() string {
	dbType := strings.ToLower(strings.TrimSpace(db.GetCurrentDBType()))
	if dbType != "sqlite" {
		return ""
	}
	dsn := strings.TrimSpace(db.GetCurrentDSN())
	if dsn != "" {
		return dsn
	}
	return strings.TrimSpace(conf.AppConfig.Database.Path)
}

func runBackupTask() {
	dbType := strings.ToLower(strings.TrimSpace(db.GetCurrentDBType()))
	if dbType != "sqlite" {
		log.Infof("backup task skipped: database type %q is not sqlite", dbType)
		return
	}

	srcPath := resolveSQLiteDBPath()
	if srcPath == "" {
		log.Warnf("backup task skipped: sqlite database path is empty")
		return
	}

	srcAbs, err := filepath.Abs(srcPath)
	if err != nil {
		log.Errorf("backup task: failed to resolve source db path %q: %v", srcPath, err)
		return
	}

	if _, err := os.Stat(srcAbs); err != nil {
		log.Errorf("backup task: source db file not accessible %q: %v", srcAbs, err)
		return
	}

	dir, err := backupDir()
	if err != nil {
		log.Errorf("backup task: failed to resolve backup directory: %v", err)
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Errorf("backup task: failed to create backup directory %q: %v", dir, err)
		return
	}

	dstPath := filepath.Join(dir, backupFilePrefix+time.Now().Format(backupTimestampFormat)+backupFileSuffix)
	if err := copyFile(srcAbs, dstPath); err != nil {
		log.Errorf("backup task: failed to copy %q to %q: %v", srcAbs, dstPath, err)
		return
	}
	log.Infof("backup task: created backup %q", dstPath)

	if err := pruneOldBackups(dir, backupKeepCount()); err != nil {
		log.Warnf("backup task: failed to prune old backups: %v", err)
	}
}

// copyFile copies the contents of src to dst. The destination file is created
// with mode 0644. If dst already exists it is truncated.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy contents: %w", err)
	}
	return dstFile.Sync()
}

// pruneOldBackups removes the oldest backup files in dir so that at most
// keepCount files remain. Files are matched by the backup-*.db naming
// convention and sorted by modification time (newest first). When keepCount
// is 0 or negative, pruning is skipped (keep all).
func pruneOldBackups(dir string, keepCount int) error {
	if keepCount <= 0 {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read backup directory: %w", err)
	}

	type backupFile struct {
		path    string
		modTime time.Time
	}
	var files []backupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, backupFilePrefix) || !strings.HasSuffix(name, backupFileSuffix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, backupFile{
			path:    filepath.Join(dir, name),
			modTime: info.ModTime(),
		})
	}

	if len(files) <= keepCount {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	for _, f := range files[keepCount:] {
		if err := os.Remove(f.path); err != nil {
			log.Warnf("backup task: failed to remove old backup %q: %v", f.path, err)
		} else {
			log.Infof("backup task: pruned old backup %q", f.path)
		}
	}
	return nil
}
