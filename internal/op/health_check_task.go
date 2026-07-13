package op

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xstrings"
)

const (
	healthCheckTaskName        = "health_check"
	defaultHealthCheckInterval = 5 * time.Minute
	defaultHealthCheckTimeout  = 10 * time.Second
	healthCheckBatchTimeout    = 30 * time.Minute

	responseCacheEvictTaskName        = "response_cache_evict"
	defaultResponseCacheEvictInterval = 60 * time.Second
)

// taskRegisterFn is set by the task package during initialization to avoid
// a circular import between op and task.
var taskRegisterFn func(name string, interval time.Duration, runOnStart bool, fn func())

// SetTaskRegister registers the task scheduler's Register function. It is
// called by the task package during Init() so that op can register periodic
// tasks without importing task directly.
func SetTaskRegister(fn func(name string, interval time.Duration, runOnStart bool, fn func())) {
	taskRegisterFn = fn
}

// circuitBreakerResetForChannelFn is set by the balancer package during
// initialization to avoid a circular import. It resets all circuit breakers
// for the given channel and returns the number of breakers that were reset
// from a non-closed state.
var circuitBreakerResetForChannelFn func(channelID int) int

// SetCircuitBreakerResetForChannel registers the balancer's circuit breaker
// reset function. It is called by the balancer package during init so that
// the health check task can auto-recover tripped circuit breakers without
// importing balancer directly.
func SetCircuitBreakerResetForChannel(fn func(channelID int) int) {
	circuitBreakerResetForChannelFn = fn
}

// StartHealthCheckTask registers the periodic health check task with the
// task scheduler. The default interval is 5 minutes, configurable via the
// health_check_interval setting (duration string like "5m").
func StartHealthCheckTask() {
	if taskRegisterFn == nil {
		log.Warnf("health check task not registered: task register not set")
		return
	}
	interval := healthCheckInterval()
	taskRegisterFn(healthCheckTaskName, interval, false, runHealthCheckTask)
	log.Infof("health check task registered with interval %v", interval)
}

// StartResponseCacheEvictTask 注册响应缓存过期清理后台任务，默认每 60 秒执行一次。
// 同时会同步缓存最大条目数配置。
func StartResponseCacheEvictTask() {
	if taskRegisterFn == nil {
		log.Warnf("response cache evict task not registered: task register not set")
		return
	}
	SyncResponseCacheConfig()
	taskRegisterFn(responseCacheEvictTaskName, defaultResponseCacheEvictInterval, false, func() {
		SyncResponseCacheConfig()
		GetResponseCache().EvictExpired()
	})
	log.Infof("response cache evict task registered with interval %v", defaultResponseCacheEvictInterval)
}

func healthCheckInterval() time.Duration {
	raw, err := SettingGetString(model.SettingKeyHealthCheckInterval)
	if err != nil || strings.TrimSpace(raw) == "" {
		return defaultHealthCheckInterval
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultHealthCheckInterval
	}
	return d
}

func healthCheckTimeout() time.Duration {
	raw, err := SettingGetString(model.SettingKeyHealthCheckTimeout)
	if err != nil || strings.TrimSpace(raw) == "" {
		return defaultHealthCheckTimeout
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultHealthCheckTimeout
	}
	return d
}

func healthCheckLLMProbeEnabled() bool {
	v, err := SettingGetBool(model.SettingKeyHealthCheckLLMProbe)
	if err != nil {
		return false
	}
	return v
}

func runHealthCheckTask() {
	ctx, cancel := context.WithTimeout(context.Background(), healthCheckBatchTimeout)
	defer cancel()

	channels, err := ChannelList(ctx)
	if err != nil {
		log.Errorf("health check: failed to list channels: %v", err)
		return
	}

	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}
		ch := channel
		if err := probeChannel(ch); err != nil {
			log.Warnf("health check: channel %d (%s) probe failed: %v", ch.ID, ch.Name, err)
		}
	}
}

// probeChannel performs a lightweight HEAD request to the channel's primary
// base URL, records the HTTP latency, updates the base URL delay, and
// triggers circuit breaker auto-recovery on success. An optional 1-token
// LLM probe is executed when the health_check_llm_probe setting is enabled.
func probeChannel(channel model.Channel) error {
	baseURL := channel.GetBaseUrl()
	if strings.TrimSpace(baseURL) == "" {
		return fmt.Errorf("channel %s has no base URL", channel.Name)
	}

	httpClient, err := channelHealthHTTPClient(&channel)
	if err != nil {
		return fmt.Errorf("failed to create http client: %w", err)
	}

	timeout := healthCheckTimeout()
	probeCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	delay, err := getURLDelayForHealthCheck(httpClient, baseURL, probeCtx)
	if err != nil {
		_ = StatsChannelUpdate(channel.ID, model.StatsMetrics{RequestFailed: 1})
		return fmt.Errorf("connectivity probe failed: %w", err)
	}

	// Update the probed base URL delay so routing decisions reflect fresh latency.
	if err := updateChannelBaseUrlDelay(channel.ID, baseURL, delay, context.Background()); err != nil {
		log.Warnf("health check: failed to update base url delay for channel %d: %v", channel.ID, err)
	}

	// Circuit breaker auto-recovery: reset all tripped breakers for this channel.
	if circuitBreakerResetForChannelFn != nil {
		if reset := circuitBreakerResetForChannelFn(channel.ID); reset > 0 {
			log.Infof("health check: recovered %d circuit breaker(s) for channel %d (%s)", reset, channel.ID, channel.Name)
		}
	}

	// Optional LLM probe (1-token call) when explicitly enabled.
	if healthCheckLLMProbeEnabled() {
		if err := llmProbeChannel(context.Background(), &channel); err != nil {
			log.Warnf("health check: LLM probe for channel %d (%s): %v", channel.ID, channel.Name, err)
		}
	}

	_ = StatsChannelUpdate(channel.ID, model.StatsMetrics{RequestSuccess: 1, WaitTime: int64(delay)})
	log.Debugf("health check: channel %d (%s) ok, delay=%dms", channel.ID, channel.Name, delay)
	return nil
}

// updateChannelBaseUrlDelay persists the measured delay for the probed base
// URL without disturbing other base URLs in the channel.
func updateChannelBaseUrlDelay(channelID int, probedURL string, delay int, ctx context.Context) error {
	ch, ok := channelCache.Get(channelID)
	if !ok {
		return fmt.Errorf("channel not found in cache")
	}
	baseUrls := make([]model.BaseUrl, len(ch.BaseUrls))
	copy(baseUrls, ch.BaseUrls)
	updated := false
	for i, bu := range baseUrls {
		if bu.URL == probedURL {
			baseUrls[i].Delay = delay
			updated = true
			break
		}
	}
	if !updated {
		return nil
	}
	return ChannelBaseUrlUpdate(channelID, baseUrls, ctx)
}

// llmProbeChannel performs an optional 1-token LLM call to verify end-to-end
// model availability. It reuses the existing CheckChannelModelHealth logic.
func llmProbeChannel(ctx context.Context, channel *model.Channel) error {
	models := xstrings.SplitTrimCompact(",", channel.Model, channel.CustomModel)
	if len(models) == 0 {
		return nil
	}
	modelName := models[0]

	probeCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout())
	defer cancel()

	result := CheckChannelModelHealth(probeCtx, channel, modelName)
	if result.Skipped {
		return nil
	}
	if result.Error != "" {
		return fmt.Errorf("%s", result.Error)
	}
	if !result.Passed {
		return fmt.Errorf("LLM probe did not pass")
	}
	return nil
}
