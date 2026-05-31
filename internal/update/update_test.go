package update

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/shutdown"
)

type testShutdownLogger struct{}

func (testShutdownLogger) Infof(string, ...interface{})  {}
func (testShutdownLogger) Errorf(string, ...interface{}) {}
func (testShutdownLogger) Warnf(string, ...interface{})  {}
func (testShutdownLogger) Debugf(string, ...interface{}) {}

func isMissingOrBlockedPathError(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "not a directory")
}

func isDirectoryCollisionError(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "not a directory") ||
		strings.Contains(errText, "is a directory") ||
		strings.Contains(errText, "access is denied") ||
		strings.Contains(errText, "cannot create") ||
		strings.Contains(errText, "permission denied") ||
		strings.Contains(errText, "file exists") ||
		os.IsExist(err)
}

func TestReadUpdateResponseBodyRejectsOversizedSuccessPayload(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxUpdateMetadataBytes)+1))),
	}

	_, err := readUpdateResponseBody(resp, maxUpdateMetadataBytes, "update metadata")
	if err == nil || !strings.Contains(err.Error(), "update metadata response too large") {
		t.Fatalf("readUpdateResponseBody() error = %v, want size limit error", err)
	}
}

func TestReadUpdateResponseBodyRejectsErrorStatusWithBodySnippet(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Status:     "502 Bad Gateway",
		Body:       io.NopCloser(strings.NewReader("upstream exploded")),
	}

	_, err := readUpdateResponseBody(resp, maxUpdateMetadataBytes, "update metadata")
	if err == nil || !strings.Contains(err.Error(), "update metadata request failed: upstream exploded") {
		t.Fatalf("readUpdateResponseBody() error = %v, want status error", err)
	}
}

func TestReadUpdateResponseBodyAcceptsSmallSuccessPayload(t *testing.T) {
	const body = `{"tag_name":"v0.1.3"}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	data, err := readUpdateResponseBody(resp, maxUpdateMetadataBytes, "update metadata")
	if err != nil {
		t.Fatalf("readUpdateResponseBody() error = %v", err)
	}
	if string(data) != body {
		t.Fatalf("payload = %q, want original body", string(data))
	}
}

func TestVerifyArchiveChecksumFromManifestAcceptsMatchingDigest(t *testing.T) {
	archive := []byte("release-binary")
	sum := sha256.Sum256(archive)
	manifest := []byte(hex.EncodeToString(sum[:]) + "  octopus-linux-x86_64.zip\n")

	if err := verifyArchiveChecksumFromManifest("octopus-linux-x86_64.zip", archive, manifest); err != nil {
		t.Fatalf("verifyArchiveChecksumFromManifest() error = %v", err)
	}
}

func TestVerifyArchiveChecksumFromManifestRejectsMismatch(t *testing.T) {
	archive := []byte("release-binary")
	other := sha256.Sum256([]byte("tampered"))
	manifest := []byte(hex.EncodeToString(other[:]) + "  octopus-linux-x86_64.zip\n")

	err := verifyArchiveChecksumFromManifest("octopus-linux-x86_64.zip", archive, manifest)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("verifyArchiveChecksumFromManifest() error = %v, want mismatch error", err)
	}
}

func TestVerifyArchiveChecksumFromManifestRejectsMissingEntry(t *testing.T) {
	archive := []byte("release-binary")
	sum := sha256.Sum256(archive)
	manifest := []byte(hex.EncodeToString(sum[:]) + "  octopus-linux-arm64.zip\n")

	err := verifyArchiveChecksumFromManifest("octopus-linux-x86_64.zip", archive, manifest)
	if err == nil || !strings.Contains(err.Error(), "checksum entry not found") {
		t.Fatalf("verifyArchiveChecksumFromManifest() error = %v, want missing-entry error", err)
	}
}

func TestVerifyArchiveChecksumFromManifestRejectsInvalidHex(t *testing.T) {
	archive := []byte("release-binary")
	manifest := []byte("not-hex  octopus-linux-x86_64.zip\n")

	err := verifyArchiveChecksumFromManifest("octopus-linux-x86_64.zip", archive, manifest)
	if err == nil || !strings.Contains(err.Error(), "invalid checksum encoding") {
		t.Fatalf("verifyArchiveChecksumFromManifest() error = %v, want invalid-hex error", err)
	}
}

func TestUpdateCoreRejectsConcurrentExecution(t *testing.T) {
	originalGOOS := currentRuntimeGOOS
	currentRuntimeGOOS = "linux"
	updateInProgress.Store(true)
	t.Cleanup(func() {
		currentRuntimeGOOS = originalGOOS
		updateInProgress.Store(false)
	})

	err := UpdateCore()
	if !errors.Is(err, ErrUpdateInProgress) {
		t.Fatalf("UpdateCore() error = %v, want %v", err, ErrUpdateInProgress)
	}
}

func TestUpdateCoreRejectsUnsupportedWindowsSelfUpdate(t *testing.T) {
	originalGOOS := currentRuntimeGOOS
	currentRuntimeGOOS = "windows"
	updateInProgress.Store(false)
	t.Cleanup(func() {
		currentRuntimeGOOS = originalGOOS
		updateInProgress.Store(false)
	})

	err := UpdateCore()
	if !errors.Is(err, ErrUpdateUnsupportedPlatform) {
		t.Fatalf("UpdateCore() error = %v, want %v", err, ErrUpdateUnsupportedPlatform)
	}
	if updateInProgress.Load() {
		t.Fatal("updateInProgress = true, want false after unsupported platform rejection")
	}
}

func TestRestartExecutableWindowsDoesNotExitWhenRestartFails(t *testing.T) {
	updateInProgress.Store(true)
	originalStart := startReplacementProcess
	originalExit := exitProcess
	exitCalled := false
	shutdown.Init(testShutdownLogger{})
	shutdownCalled := false
	shutdown.Register(func() error {
		shutdownCalled = true
		return nil
	})
	startReplacementProcess = func(execPath string, args []string) error {
		if shutdownCalled {
			t.Fatal("shutdown ran before replacement process start")
		}
		return errors.New("boom")
	}
	exitProcess = func(code int) {
		exitCalled = true
	}
	t.Cleanup(func() {
		startReplacementProcess = originalStart
		exitProcess = originalExit
		updateInProgress.Store(false)
	})

	restartExecutableForPlatform("windows", "octopus.exe", []string{"octopus.exe", "start"})
	if exitCalled {
		t.Fatalf("exitCalled = true, want false when replacement start fails")
	}
	if updateInProgress.Load() {
		t.Fatalf("updateInProgress = true, want false after restart failure")
	}
	if shutdownCalled {
		t.Fatal("shutdownCalled = true, want false when replacement start fails")
	}

	if runtime.GOOS != "windows" && exitCalled {
		t.Fatalf("exitCalled = true, want false on non-windows host test path")
	}
}

func TestRestartExecutableWindowsShutsDownAfterReplacementStarts(t *testing.T) {
	updateInProgress.Store(true)
	originalStart := startReplacementProcess
	originalExit := exitProcess
	shutdown.Init(testShutdownLogger{})
	shutdownCalled := false
	startReplacementProcess = func(execPath string, args []string) error {
		if shutdownCalled {
			t.Fatal("shutdown ran before replacement process start")
		}
		return nil
	}
	exitCode := -1
	exitProcess = func(code int) {
		exitCode = code
	}
	shutdown.Register(func() error {
		shutdownCalled = true
		return nil
	})
	t.Cleanup(func() {
		startReplacementProcess = originalStart
		exitProcess = originalExit
		updateInProgress.Store(false)
	})

	restartExecutableForPlatform("windows", "octopus.exe", []string{"octopus.exe", "start"})

	if !shutdownCalled {
		t.Fatal("shutdownCalled = false, want true after replacement start")
	}
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
}

func TestEnsureSafeExtractPathRejectsSymlinkAncestor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on this Windows host")
	}

	root := t.TempDir()
	target := t.TempDir()
	linkPath := filepath.Join(root, "linked")
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	err := ensureSafeExtractPath(filepath.Join(linkPath, "payload.bin"))
	if err == nil || !strings.Contains(err.Error(), "extract path contains symlink") {
		t.Fatalf("ensureSafeExtractPath() error = %v, want symlink rejection", err)
	}
}

func TestUnzipRejectsExistingSymlinkInDestination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on this Windows host")
	}

	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("linked/payload.bin")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := file.Write([]byte("payload")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	target := t.TempDir()
	linkPath := filepath.Join(root, "linked")
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	err = unzip(archive.Bytes(), root)
	if err == nil || !strings.Contains(err.Error(), "extract path contains symlink") {
		t.Fatalf("unzip() error = %v, want symlink rejection", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, "payload.bin")); !os.IsNotExist(statErr) {
		t.Fatalf("payload unexpectedly written outside destination, stat err = %v", statErr)
	}
}

func TestUnzipRejectsSymlinkArchiveEntry(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	header := &zip.FileHeader{Name: "linked", Method: zip.Store}
	header.SetMode(os.ModeSymlink | 0o777)
	file, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatalf("CreateHeader() error = %v", err)
	}
	if _, err := file.Write([]byte("payload.bin")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	err = unzip(archive.Bytes(), root)
	if err == nil || !strings.Contains(err.Error(), "unsupported symlink entry") {
		t.Fatalf("unzip() error = %v, want symlink entry rejection", err)
	}
	if _, statErr := os.Lstat(filepath.Join(root, "linked")); !os.IsNotExist(statErr) {
		t.Fatalf("symlink entry unexpectedly materialized, stat err = %v", statErr)
	}
}

func TestUnzipFailsWhenDirectoryEntryCollidesWithExistingFile(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	if _, err := writer.Create("nested/"); err != nil {
		t.Fatalf("Create(dir) error = %v", err)
	}
	file, err := writer.Create("nested/payload.bin")
	if err != nil {
		t.Fatalf("Create(file) error = %v", err)
	}
	if _, err := file.Write([]byte("payload")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	blockingPath := filepath.Join(root, "nested")
	if err := os.WriteFile(blockingPath, []byte("block"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err = unzip(archive.Bytes(), root)
	if err == nil {
		t.Fatal("unzip() error = nil, want mkdir collision error")
	}
	if !isDirectoryCollisionError(err) && !strings.Contains(strings.ToLower(err.Error()), "path specified") {
		t.Fatalf("unzip() error = %v, want directory collision error", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "nested", "payload.bin")); !isMissingOrBlockedPathError(statErr) {
		t.Fatalf("payload unexpectedly extracted through blocking file, stat err = %v", statErr)
	}
}

func TestUnzipRejectsEntryLargerThanExpandedLimit(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("payload.bin")
	if err != nil {
		t.Fatalf("Create(file) error = %v", err)
	}
	if _, err := file.Write([]byte("0123456789")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	err = unzipWithLimit(archive.Bytes(), root, 4)
	if err == nil || !strings.Contains(err.Error(), "update archive entry too large") {
		t.Fatalf("unzipWithLimit() error = %v, want entry-too-large rejection", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "payload.bin")); !os.IsNotExist(statErr) {
		t.Fatalf("payload unexpectedly extracted, stat err = %v", statErr)
	}
}

func TestUnzipRejectsArchiveWhoseExpandedSizeExceedsLimit(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for _, name := range []string{"first.bin", "second.bin"} {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := file.Write([]byte("12345")); err != nil {
			t.Fatalf("Write(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	err := unzipWithLimit(archive.Bytes(), root, 8)
	if err == nil || !strings.Contains(err.Error(), "update archive expanded too large") {
		t.Fatalf("unzipWithLimit() error = %v, want expanded-size rejection", err)
	}
	for _, name := range []string{"first.bin", "second.bin"} {
		if _, statErr := os.Stat(filepath.Join(root, name)); !os.IsNotExist(statErr) {
			t.Fatalf("%s unexpectedly extracted, stat err = %v", name, statErr)
		}
	}
}

func TestUnzipExecutableExtractsOnlyMatchingBinary(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	entries := map[string]string{
		"octopus.exe": "new-binary",
		"README.md":   "doc",
		"LICENSE":     "license",
	}
	for name, body := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := file.Write([]byte(body)); err != nil {
			t.Fatalf("Write(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	execPath := filepath.Join(root, "octopus.exe")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o600); err != nil {
		t.Fatalf("WriteFile(exec) error = %v", err)
	}

	if err := unzipExecutable(archive.Bytes(), execPath); err != nil {
		t.Fatalf("unzipExecutable() error = %v", err)
	}

	data, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("ReadFile(exec) error = %v", err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("exec contents = %q, want updated binary only", string(data))
	}
	for _, name := range []string{"README.md", "LICENSE"} {
		if _, statErr := os.Stat(filepath.Join(root, name)); !os.IsNotExist(statErr) {
			t.Fatalf("%s unexpectedly extracted, stat err = %v", name, statErr)
		}
	}
}

func TestUnzipExecutableRejectsArchiveMissingBinary(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("README.md")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := file.Write([]byte("doc")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	execPath := filepath.Join(root, "octopus.exe")
	err = unzipExecutable(archive.Bytes(), execPath)
	if err == nil || !strings.Contains(err.Error(), "missing executable octopus.exe") {
		t.Fatalf("unzipExecutable() error = %v, want missing executable rejection", err)
	}
	if _, statErr := os.Stat(execPath); !os.IsNotExist(statErr) {
		t.Fatalf("exec path unexpectedly created, stat err = %v", statErr)
	}
}

func TestUnzipExecutableRejectsDuplicateBinaryEntry(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for _, name := range []string{"octopus", "octopus"} {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := file.Write([]byte("payload")); err != nil {
			t.Fatalf("Write(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	execPath := filepath.Join(root, "octopus")
	err := unzipExecutable(archive.Bytes(), execPath)
	if err == nil || !strings.Contains(err.Error(), "duplicate executable entry") {
		t.Fatalf("unzipExecutable() error = %v, want duplicate entry rejection", err)
	}
}

func TestUnzipDoesNotReplaceExistingDirectoryWithFile(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("nested")
	if err != nil {
		t.Fatalf("Create(file) error = %v", err)
	}
	if _, err := file.Write([]byte("payload")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	blockingDir := filepath.Join(root, "nested")
	if err := os.Mkdir(blockingDir, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	err = unzip(archive.Bytes(), root)
	if err == nil {
		t.Fatal("unzip() error = nil, want file-vs-directory collision error")
	}
	if !isDirectoryCollisionError(err) {
		t.Fatalf("unzip() error = %v, want directory collision error", err)
	}
	info, statErr := os.Stat(blockingDir)
	if statErr != nil {
		t.Fatalf("Stat() error = %v, want preserved directory", statErr)
	}
	if !info.IsDir() {
		t.Fatalf("blocking path mode = %v, want directory to remain intact", info.Mode())
	}
}

func TestUnzipPreservesExistingFileWhenExpandedLimitExceededDuringExtraction(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create("payload.bin")
	if err != nil {
		t.Fatalf("Create(file) error = %v", err)
	}
	if _, err := file.Write([]byte("0123456789")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	root := t.TempDir()
	targetPath := filepath.Join(root, "payload.bin")
	if err := os.WriteFile(targetPath, []byte("stable"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err = unzipWithLimit(archive.Bytes(), root, 5)
	if err == nil || !strings.Contains(err.Error(), "update archive entry too large") {
		t.Fatalf("unzipWithLimit() error = %v, want size rejection", err)
	}
	data, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(data) != "stable" {
		t.Fatalf("payload contents = %q, want preserved original file", string(data))
	}
}

func TestReplaceExtractedFileRestoresOriginalWhenSecondRenameFails(t *testing.T) {
	root := t.TempDir()
	destPath := filepath.Join(root, "octopus.bin")
	tempPath := filepath.Join(root, "octopus.bin.tmp")
	if err := os.WriteFile(destPath, []byte("stable"), 0o600); err != nil {
		t.Fatalf("WriteFile(dest) error = %v", err)
	}
	if err := os.WriteFile(tempPath, []byte("incoming"), 0o600); err != nil {
		t.Fatalf("WriteFile(temp) error = %v", err)
	}

	originalRename := osRename
	originalRemove := osRemove
	var renameCalls int
	osRename = func(oldPath, newPath string) error {
		renameCalls++
		switch renameCalls {
		case 1:
			return errors.New("cross-device rename")
		case 2:
			return originalRename(oldPath, newPath)
		case 3:
			return errors.New("dest busy")
		default:
			return originalRename(oldPath, newPath)
		}
	}
	osRemove = originalRemove
	t.Cleanup(func() {
		osRename = originalRename
		osRemove = originalRemove
	})

	err := replaceExtractedFile(tempPath, destPath)
	if err == nil || !strings.Contains(err.Error(), "dest busy") {
		t.Fatalf("replaceExtractedFile() error = %v, want second rename failure", err)
	}

	data, readErr := os.ReadFile(destPath)
	if readErr != nil {
		t.Fatalf("ReadFile(dest) error = %v", readErr)
	}
	if string(data) != "stable" {
		t.Fatalf("destination contents = %q, want original file restored", string(data))
	}
	if _, statErr := os.Stat(tempPath); statErr != nil {
		t.Fatalf("temp file stat err = %v, want caller-cleanup temp file to remain", statErr)
	}

	entries, readDirErr := os.ReadDir(root)
	if readDirErr != nil {
		t.Fatalf("ReadDir() error = %v", readDirErr)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".octopus-backup-") {
			t.Fatalf("backup file unexpectedly remained: %s", entry.Name())
		}
	}
}

func TestEnsureSafeExtractPathRejectsJunctionAncestor(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("junction validation is Windows-specific")
	}

	root := t.TempDir()
	target := t.TempDir()
	linkPath := filepath.Join(root, "linked")
	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, target)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("mklink /J unavailable on this host: %v (%s)", err, strings.TrimSpace(string(output)))
	}

	err := ensureSafeExtractPath(filepath.Join(linkPath, "payload.bin"))
	if err == nil || !strings.Contains(err.Error(), "extract path contains reparse link") {
		t.Fatalf("ensureSafeExtractPath() error = %v, want junction rejection", err)
	}
}
