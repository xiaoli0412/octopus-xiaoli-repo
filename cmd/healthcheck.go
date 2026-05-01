package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf"
	"github.com/spf13/cobra"
)

var healthcheckURL string
var healthcheckHTTPClient = &http.Client{Timeout: 5 * time.Second}

func resolveHealthcheckTarget() (string, error) {
	target := strings.TrimSpace(healthcheckURL)
	if target != "" {
		return target, nil
	}

	if err := conf.Load(cfgFile); err != nil {
		return "", err
	}

	host := strings.TrimSpace(conf.AppConfig.Server.Host)
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}

	return fmt.Sprintf("http://%s:%d/healthz", host, conf.AppConfig.Server.Port), nil
}

func isLocalHealthcheckTarget(target string) bool {
	parsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(parsed.Hostname())) {
	case "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}

func isWindowsServiceProviderFailure(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "无法加载或初始化请求的服务提供程序") ||
		strings.Contains(message, "requested service provider") ||
		strings.Contains(message, "service provider could not be loaded or initialized")
}

func wrapHealthcheckRequestError(target string, err error) error {
	if err == nil {
		return nil
	}

	if isLocalHealthcheckTarget(target) && isWindowsServiceProviderFailure(err) {
		return fmt.Errorf("localhost healthcheck for %s hit a Windows service-provider initialization failure; treat this as a host networking blocker instead of an Octopus regression: %w", target, err)
	}

	return err
}

var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check whether the Octopus HTTP server is healthy",
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolveHealthcheckTarget()
		if err != nil {
			return err
		}

		resp, err := healthcheckHTTPClient.Get(target)
		if err != nil {
			return wrapHealthcheckRequestError(target, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("healthcheck returned status %d", resp.StatusCode)
		}
		cmd.Println("ok")
		return nil
	},
}

func init() {
	healthcheckCmd.Flags().StringVar(&healthcheckURL, "url", "", "healthcheck URL override")
	rootCmd.AddCommand(healthcheckCmd)
}
