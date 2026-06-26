package update

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/client"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const (
	updateUrl                     = "https://github.com/xiaoli0412/octopus-xiaoli-repo/releases/latest/download"
	updateApiUrl                  = "https://api.github.com/repos/xiaoli0412/octopus-xiaoli-repo/releases/latest"
	updateChecksumFilename        = "SHA256SUMS"
	maxUpdateMetadataBytes  int64 = 2 << 20
	maxUpdateArchiveBytes   int64 = 128 << 20
	maxUpdateExpandedBytes  int64 = 512 << 20
	maxUpdateErrorBodyBytes int64 = 8 << 10
	maxUpdateChecksumBytes  int64 = 1 << 20
)

type LatestInfo struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
	Message     string `json:"message"`
}

type StatusInfo struct {
	Version                  string `json:"version"`
	SelfUpdateSupported      bool   `json:"self_update_supported"`
	SelfUpdateUnsupportedReason string `json:"self_update_unsupported_reason,omitempty"`
}

var github_pat = os.Getenv(strings.ToUpper(conf.APP_NAME) + "_GITHUB_PAT")

var (
	osRename = os.Rename
	osRemove = os.Remove
)

func GetStatusInfo() StatusInfo {
	status := StatusInfo{
		Version:             conf.Version,
		SelfUpdateSupported: currentRuntimeGOOS != "windows" && !isRunningInDocker(),
	}
	if !status.SelfUpdateSupported {
		if isRunningInDocker() {
			status.SelfUpdateUnsupportedReason = ErrUpdateInDocker.Error()
		} else {
			status.SelfUpdateUnsupportedReason = ErrUpdateUnsupportedPlatform.Error()
		}
	}
	return status
}

// doRequestWithFallback performs an HTTP GET request, first without proxy, then with proxy if failed.
func doRequestWithFallback(url string, maxBytes int64, resource string) ([]byte, error) {
	data, err := doRequest(url, false, maxBytes, resource)
	if err == nil {
		return data, nil
	}
	log.Warnf("direct request failed, trying with proxy: %v", err)
	return doRequest(url, true, maxBytes, resource)
}

func doRequest(url string, useProxy bool, maxBytes int64, resource string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hc, err := client.GetHTTPClientSystemProxy(useProxy)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Debugf("new request failed: %v", err)
		return nil, err
	}

	if github_pat != "" {
		req.Header.Set("Authorization", "Bearer "+github_pat)
	}

	resp, err := hc.Do(req)
	if err != nil {
		log.Debugf("request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := readUpdateResponseBody(resp, maxBytes, resource)
	if err != nil {
		log.Debugf("read body failed: %v", err)
		return nil, err
	}
	return data, nil
}

func GetLatestInfo() (*LatestInfo, error) {
	body, err := doRequestWithFallback(updateApiUrl, maxUpdateMetadataBytes, "update metadata")
	if err != nil {
		return nil, err
	}

	var latestInfo LatestInfo
	if err := json.Unmarshal(body, &latestInfo); err != nil {
		log.Debugf("unmarshal body failed: %v", err)
		return nil, err
	}
	if latestInfo.Message != "" {
		return nil, fmt.Errorf("failed to get latest info: %s", latestInfo.Message)
	}
	return &latestInfo, nil
}

func verifyArchiveChecksum(filename string, archive []byte) error {
	checksumURL := updateUrl + "/" + updateChecksumFilename
	manifest, err := doRequestWithFallback(checksumURL, maxUpdateChecksumBytes, "update checksum manifest")
	if err != nil {
		return err
	}
	return verifyArchiveChecksumFromManifest(filename, archive, manifest)
}

func verifyArchiveChecksumFromManifest(filename string, archive []byte, manifest []byte) error {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return fmt.Errorf("missing update archive filename")
	}

	expectedHex, err := findExpectedChecksum(filename, manifest)
	if err != nil {
		return err
	}

	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return fmt.Errorf("invalid checksum encoding for %s: %w", filename, err)
	}
	if len(expected) != sha256.Size {
		return fmt.Errorf("invalid checksum length for %s", filename)
	}

	actual := sha256.Sum256(archive)
	if subtle.ConstantTimeCompare(actual[:], expected) != 1 {
		return fmt.Errorf("update archive checksum mismatch for %s", filename)
	}
	return nil
}

func findExpectedChecksum(filename string, manifest []byte) (string, error) {
	for _, rawLine := range strings.Split(string(manifest), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		candidate := strings.TrimPrefix(fields[len(fields)-1], "./")
		candidate = filepath.Base(strings.ReplaceAll(candidate, "\\", "/"))
		if candidate != filename {
			continue
		}

		checksum := strings.TrimSpace(fields[0])
		if checksum == "" {
			break
		}
		return checksum, nil
	}

	return "", fmt.Errorf("checksum entry not found for %s", filename)
}

func readUpdateResponseBody(resp *http.Response, maxBytes int64, resource string) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("missing %s response body", resource)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxUpdateErrorBodyBytes))
		if err != nil {
			return nil, fmt.Errorf("read %s error response: %w", resource, err)
		}
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return nil, fmt.Errorf("%s request failed: %s", resource, message)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("%s response too large", resource)
	}
	return data, nil
}

func unzip(data []byte, dest string) error {
	return unzipWithLimit(data, dest, maxUpdateExpandedBytes)
}

func unzipExecutable(data []byte, execPath string) error {
	return unzipExecutableWithLimit(data, execPath, maxUpdateExpandedBytes)
}

func unzipExecutableWithLimit(data []byte, execPath string, maxExpandedBytes int64) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Debugf("new zip reader failed: %v", err)
		return err
	}

	targetName := filepath.Base(execPath)
	if targetName == "" || targetName == "." {
		return fmt.Errorf("invalid executable path: %s", execPath)
	}

	targetEntry, err := findExecutableEntry(r, targetName, maxExpandedBytes)
	if err != nil {
		log.Debugf("find executable entry failed: %v", err)
		return err
	}

	if err := ensureSafeExtractPath(execPath); err != nil {
		log.Debugf("unsafe executable path: %v", err)
		return err
	}

	_, err = extractFile(targetEntry, execPath, maxExpandedBytes)
	if err != nil {
		log.Debugf("extract executable failed: %v", err)
	}
	return err
}

func findExecutableEntry(r *zip.Reader, targetName string, maxExpandedBytes int64) (*zip.File, error) {
	if maxExpandedBytes <= 0 {
		return nil, fmt.Errorf("invalid update archive expanded size limit")
	}

	var targetEntry *zip.File
	for _, f := range r.File {
		info := f.FileInfo()
		if info.IsDir() {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("zip contains unsupported symlink entry: %s", f.Name)
		}

		cleanName := filepath.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		if cleanName == "." || cleanName == string(filepath.Separator) {
			continue
		}
		if strings.Contains(cleanName, "/") {
			continue
		}
		if cleanName != targetName {
			continue
		}
		if f.UncompressedSize64 > uint64(maxExpandedBytes) {
			return nil, fmt.Errorf("update archive entry too large: %s", f.Name)
		}
		if targetEntry != nil {
			return nil, fmt.Errorf("update archive contains duplicate executable entry: %s", targetName)
		}
		targetEntry = f
	}

	if targetEntry == nil {
		return nil, fmt.Errorf("update archive missing executable %s", targetName)
	}
	return targetEntry, nil
}

func unzipWithLimit(data []byte, dest string, maxExpandedBytes int64) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Debugf("new zip reader failed: %v", err)
		return err
	}

	if err := validateZipEntries(r, dest, maxExpandedBytes); err != nil {
		log.Debugf("validate zip entries failed: %v", err)
		return err
	}

	remainingBytes := maxExpandedBytes

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !isPathInDest(fpath, dest) {
			log.Debugf("invalid file path: %s", fpath)
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		info := f.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				log.Debugf("mkdir all failed: %v", err)
				return err
			}
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("zip contains unsupported symlink entry: %s", f.Name)
		}

		written, err := extractFile(f, fpath, remainingBytes)
		if err != nil {
			return err
		}
		remainingBytes -= written
	}
	return nil
}

func validateZipEntries(r *zip.Reader, dest string, maxExpandedBytes int64) error {
	if maxExpandedBytes <= 0 {
		return fmt.Errorf("invalid update archive expanded size limit")
	}

	var totalUncompressedBytes int64
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !isPathInDest(fpath, dest) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		info := f.FileInfo()
		if info.IsDir() {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("zip contains unsupported symlink entry: %s", f.Name)
		}
		if f.UncompressedSize64 > uint64(maxExpandedBytes) {
			return fmt.Errorf("update archive entry too large: %s", f.Name)
		}
		entryBytes := int64(f.UncompressedSize64)
		if totalUncompressedBytes > maxExpandedBytes-entryBytes {
			return fmt.Errorf("update archive expanded too large")
		}
		totalUncompressedBytes += entryBytes
	}

	return nil
}

func extractFile(f *zip.File, fpath string, remainingBytes int64) (int64, error) {
	if err := ensureSafeExtractPath(fpath); err != nil {
		log.Debugf("unsafe extract path: %v", err)
		return 0, err
	}

	dir := filepath.Dir(fpath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Debugf("mkdir all failed: %v", err)
		return 0, err
	}

	rc, err := f.Open()
	if err != nil {
		log.Debugf("open file failed: %v", err)
		return 0, err
	}
	defer rc.Close()

	tempPath, outFile, err := createExtractTempFile(dir, filepath.Base(fpath), f.Mode().Perm())
	if err != nil {
		log.Debugf("open file failed: %v", err)
		return 0, err
	}
	cleanupTemp := true
	defer func() {
		if outFile != nil {
			if closeErr := outFile.Close(); closeErr != nil {
				log.Debugf("close temp file failed: %v", closeErr)
			}
		}
		if cleanupTemp {
			if removeErr := os.Remove(tempPath); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Debugf("remove temp file failed: %v", removeErr)
			}
		}
	}()

	written, err := io.Copy(outFile, io.LimitReader(rc, remainingBytes+1))
	if err != nil {
		log.Debugf("copy failed: %v", err)
		return written, err
	}
	if written > remainingBytes {
		log.Debugf("update archive expanded too large while extracting: %s", f.Name)
		return written, fmt.Errorf("update archive expanded too large")
	}
	if err := outFile.Close(); err != nil {
		log.Debugf("close temp file failed: %v", err)
		return written, err
	}
	outFile = nil
	if err := replaceExtractedFile(tempPath, fpath); err != nil {
		log.Debugf("replace extracted file failed: %v", err)
		return written, err
	}
	cleanupTemp = false
	return written, nil
}

func createExtractTempFile(dir, name string, perm os.FileMode) (string, *os.File, error) {
	tempFile, err := os.CreateTemp(dir, name+".octopus-update-*")
	if err != nil {
		return "", nil, err
	}
	if err := tempFile.Chmod(perm); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return "", nil, err
	}
	return tempFile.Name(), tempFile, nil
}

func replaceExtractedFile(tempPath, destPath string) error {
	if err := osRename(tempPath, destPath); err == nil {
		return nil
	} else {
		info, statErr := os.Lstat(destPath)
		if statErr != nil {
			return err
		}
		if info.IsDir() {
			log.Debugf("refuse to replace directory with file: %s", destPath)
			return err
		}
		if !info.Mode().IsRegular() {
			log.Debugf("refuse to replace non-regular path: %s", destPath)
			return err
		}
		backupPath, backupErr := createReplaceBackupPath(filepath.Dir(destPath), filepath.Base(destPath))
		if backupErr != nil {
			log.Debugf("create replace backup path failed: %v", backupErr)
			return backupErr
		}
		if renameErr := osRename(destPath, backupPath); renameErr != nil {
			log.Debugf("backup existing file failed: %v", renameErr)
			_ = osRemove(backupPath)
			return renameErr
		}
		if renameErr := osRename(tempPath, destPath); renameErr != nil {
			if restoreErr := osRename(backupPath, destPath); restoreErr != nil {
				log.Debugf("restore existing file failed: %v", restoreErr)
				return fmt.Errorf("replace extracted file failed: %w; restore failed: %v", renameErr, restoreErr)
			}
			return renameErr
		}
		if removeErr := osRemove(backupPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Debugf("remove replace backup failed: %v", removeErr)
		}
		return nil
	}
}

func createReplaceBackupPath(dir, name string) (string, error) {
	backupFile, err := os.CreateTemp(dir, name+".octopus-backup-*")
	if err != nil {
		return "", err
	}
	backupPath := backupFile.Name()
	if err := backupFile.Close(); err != nil {
		_ = osRemove(backupPath)
		return "", err
	}
	if err := osRemove(backupPath); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return backupPath, nil
}

func isPathInDest(fpath, dest string) bool {
	rel, err := filepath.Rel(dest, fpath)
	if err != nil {
		return false
	}
	return filepath.IsLocal(rel)
}

func ensureSafeExtractPath(fpath string) error {
	resolvedPath, err := filepath.Abs(fpath)
	if err != nil {
		return err
	}

	current := filepath.VolumeName(resolvedPath) + string(os.PathSeparator)
	trimmedPath := strings.TrimPrefix(resolvedPath, current)
	for _, segment := range strings.Split(trimmedPath, string(os.PathSeparator)) {
		if segment == "" {
			continue
		}
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("extract path contains symlink: %s", current)
		}
		if pathIsSpecialLink(current, info) {
			return fmt.Errorf("extract path contains reparse link: %s", current)
		}
	}

	return nil
}

func pathIsSpecialLink(path string, info os.FileInfo) bool {
	if info == nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	if runtime.GOOS != "windows" {
		return false
	}
	sys := info.Sys()
	if sys == nil {
		return false
	}
	v := reflect.ValueOf(sys)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return false
	}
	attrs := v.FieldByName("FileAttributes")
	if !attrs.IsValid() || !attrs.CanUint() {
		return false
	}
	const fileAttributeReparsePoint = 0x0400
	return attrs.Uint()&fileAttributeReparsePoint != 0
}
