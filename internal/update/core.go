package update

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/shutdown"
)

var ErrUpdateInProgress = errors.New("update already in progress")
var ErrUpdateUnsupportedPlatform = errors.New("self-update is not supported on Windows; replace the binary or container image manually")
var ErrUpdateInDocker = errors.New("self-update is not supported inside a Docker container; update the container image instead")

var updateInProgress atomic.Bool
var currentRuntimeGOOS = runtime.GOOS
var exitProcess = os.Exit
var startReplacementProcess = func(execPath string, args []string) error {
	cmd := exec.Command(execPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func UpdateCore() error {
	if currentRuntimeGOOS == "windows" {
		return ErrUpdateUnsupportedPlatform
	}
	if isRunningInDocker() {
		return ErrUpdateInDocker
	}
	if !updateInProgress.CompareAndSwap(false, true) {
		return ErrUpdateInProgress
	}
	releaseUpdateGuard := true
	defer func() {
		if releaseUpdateGuard {
			updateInProgress.Store(false)
		}
	}()

	log.Infof("start update core")

	filename, err := getDownloadFilename()
	if err != nil {
		log.Warnf("update core failed: %v", err)
		return err
	}

	downloadUrl := updateUrl + "/" + filename
	log.Infof("download url: %s", downloadUrl)
	data, err := doRequestWithFallback(downloadUrl, maxUpdateArchiveBytes, "update archive")
	if err != nil {
		log.Warnf("download failed: %v", err)
		return err
	}
	if err := verifyArchiveChecksum(filename, data); err != nil {
		log.Warnf("checksum verify failed: %v", err)
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Warnf("get executable path failed: %v", err)
		return err
	}

	if err := unzipExecutable(data, execPath); err != nil {
		log.Warnf("unzip failed: %v", err)
		return err
	}

	log.Infof("update core success")
	releaseUpdateGuard = false
	go restartExecutable(execPath)
	return nil
}

// isRunningInDocker 检测当前进程是否运行在 Docker 容器中。
// 它通过检查 /.dockerenv 文件以及 /proc/1/cgroup 内容中是否包含 docker/kube 标记来判断。
func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	file, err := os.Open("/proc/1/cgroup")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		if strings.Contains(line, "docker") || strings.Contains(line, "kubepods") {
			return true
		}
	}
	return false
}

func getDownloadFilename() (string, error) {
	arch := runtime.GOARCH
	goos := runtime.GOOS

	switch goos {
	case "windows":
		switch arch {
		case "386":
			return "octopus-windows-x86.zip", nil
		case "amd64":
			return "octopus-windows-x86_64.zip", nil
		}
	case "darwin":
		switch arch {
		case "amd64":
			return "octopus-darwin-x86_64.zip", nil
		case "arm64":
			return "octopus-darwin-arm64.zip", nil
		}
	case "linux":
		switch arch {
		case "386":
			return "octopus-linux-x86.zip", nil
		case "amd64":
			return "octopus-linux-x86_64.zip", nil
		case "arm":
			return "octopus-linux-armv7.zip", nil
		case "arm64":
			return "octopus-linux-arm64.zip", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", goos, arch)
}

func restartExecutable(execPath string) {
	restartExecutableForPlatform(runtime.GOOS, execPath, os.Args)
}

func restartExecutableForPlatform(goos, execPath string, args []string) {
	argv := append([]string(nil), args...)
	if len(argv) == 0 {
		argv = []string{execPath}
	}
	childArgs := []string{}
	if len(argv) > 1 {
		childArgs = argv[1:]
	}

	log.Infof("restarting: %q %q", execPath, childArgs)

	if goos == "windows" {
		if err := startReplacementProcess(execPath, childArgs); err != nil {
			updateInProgress.Store(false)
			log.Errorf("restarting failed: %v", err)
			return
		}
		shutdown.Shutdown()
		exitProcess(0)
		return
	}

	shutdown.Shutdown()
	if err := syscall.Exec(execPath, argv, os.Environ()); err != nil {
		updateInProgress.Store(false)
		log.Errorf("restarting failed: %v", err)
	}
}

func CurrentRuntimeGOOSForTest() string {
	return currentRuntimeGOOS
}

func SetCurrentRuntimeGOOSForTest(value string) {
	currentRuntimeGOOS = value
}
