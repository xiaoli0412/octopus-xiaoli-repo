package op

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	transformerModel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
	"golang.org/x/net/proxy"
)

const importHealthCheckRequestTimeout = 30 * time.Second
const importHealthCheckMaxConcurrency = 2

func populateHealthPolicyMetadata(result *model.DBImportHealthCheckItem, channel *model.Channel, key model.ChannelKey, modelName string) {
	if result == nil {
		return
	}
	policy := ResolveRouteTargetPolicy(channel, key, modelName)
	result.SourceType = policy.SourceType
	result.BillingMode = string(policy.BillingMode)
	result.ProbePolicy = string(policy.ProbePolicy)
	result.PolicyBasis = policy.PolicyBasisSummary()
}

func normalizeHealthCheckChannel(channel *model.Channel) *model.Channel {
	if channel == nil {
		return nil
	}
	clone := *channel
	if clone.KeyManagementMode == "" {
		clone.KeyManagementMode = model.KeyManagementModePooled
	} else {
		clone.KeyManagementMode = model.NormalizeKeyManagementMode(clone.KeyManagementMode)
	}
	if clone.KeyRoutingPolicy == "" {
		clone.KeyRoutingPolicy = model.KeyRoutingPolicyRoundRobin
	} else {
		clone.KeyRoutingPolicy = model.NormalizeKeyRoutingPolicy(clone.KeyRoutingPolicy)
	}
	if clone.Keys != nil {
		keys := make([]model.ChannelKey, 0, len(clone.Keys))
		for _, key := range clone.Keys {
			key.SourceType = model.EffectiveChannelKeySourceType(key.SourceType)
			key.AllowedModels = model.NormalizeChannelKeyAllowedModels(key.AllowedModels)
			keys = append(keys, key)
		}
		clone.Keys = keys
	}
	return &clone
}

func CheckChannelModelHealthForConfig(ctx context.Context, channel *model.Channel, modelName string) model.DBImportHealthCheckItem {
	return CheckChannelModelHealth(ctx, normalizeHealthCheckChannel(channel), modelName)
}

type ChannelModelHealthCheckTarget struct {
	GroupID   int
	GroupName string
	ChannelID int
	ModelName string
}

func CheckChannelModelHealth(ctx context.Context, channel *model.Channel, modelName string) model.DBImportHealthCheckItem {
	channel = normalizeHealthCheckChannel(channel)
	result := model.DBImportHealthCheckItem{
		ChannelID:   0,
		ChannelName: "",
		Model:       modelName,
		CheckStage:  "preflight",
		SourceType:  model.ChannelKeySourceTypeUnknown,
		BillingMode: string(model.BillingModeUnknown),
		ProbePolicy: string(model.ProbePolicyPassiveOnly),
		PolicyBasis: model.RouteTargetPolicyBasisChannelKeyInheritance,
	}
	if channel == nil {
		result.Skipped = true
		result.Error = "channel not found"
		return result
	}
	result.ChannelID = channel.ID
	result.ChannelName = channel.Name

	if !channel.SupportsModel(modelName) {
		result.Skipped = true
		result.Error = fmt.Sprintf("channel:%s does not declare model:%s", channel.Name, modelName)
		return result
	}
	if !channel.Enabled {
		result.Skipped = true
		result.Error = fmt.Sprintf("channel:%s is disabled", channel.Name)
		return result
	}

	httpClient, err := channelHealthHTTPClient(channel)
	if err != nil {
		result.Error = "failed to create HTTP client: " + err.Error()
		return result
	}

	baseURL := channel.GetBaseUrl()
	if strings.TrimSpace(baseURL) == "" {
		result.Skipped = true
		result.Error = fmt.Sprintf("channel:%s has no base URL", channel.Name)
		return result
	}

	connectivityCtx, cancelConnectivity := context.WithTimeout(ctx, importHealthCheckRequestTimeout)
	defer cancelConnectivity()
	delay, err := getURLDelayForHealthCheck(httpClient, baseURL, connectivityCtx)
	if err != nil {
		result.CheckStage = "connectivity"
		result.Error = "connectivity test failed: " + err.Error()
		return result
	}
	result.Delay = delay

	channelKey := channel.GetChannelKeyForModel(modelName)
	populateHealthPolicyMetadata(&result, channel, channelKey, modelName)
	if strings.TrimSpace(channelKey.ChannelKey) == "" {
		result.Skipped = true
		result.CheckStage = "credentials"
		result.Error = fmt.Sprintf("channel:%s has no available key for model:%s", channel.Name, modelName)
		return result
	}

	outboundAdapter := transformerOutbound.GetForModel(transformerOutbound.OutboundType(channel.Type), modelName)
	if outboundAdapter == nil {
		result.Skipped = true
		result.CheckStage = "adapter"
		result.Error = fmt.Sprintf("channel:%s has unsupported outbound adapter for model:%s", channel.Name, modelName)
		return result
	}

	content := "1+1=?"
	maxTokens := int64(1)
	temperature := 0.0
	testReq := transformerModel.InternalLLMRequest{
		Model:       modelName,
		Messages:    []transformerModel.Message{{Role: "user", Content: transformerModel.MessageContent{Content: &content}}},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	requestCtx, cancelRequest := context.WithTimeout(ctx, importHealthCheckRequestTimeout)
	defer cancelRequest()
	outboundReq, err := outboundAdapter.TransformRequest(requestCtx, &testReq, baseURL, channelKey.ChannelKey)
	if err != nil {
		result.CheckStage = "request_build"
		result.Error = "failed to build request: " + err.Error()
		return result
	}
	for _, header := range channel.CustomHeader {
		if strings.TrimSpace(header.HeaderKey) == "" {
			continue
		}
		outboundReq.Header.Set(header.HeaderKey, header.HeaderValue)
	}

	resp, err := httpClient.Do(outboundReq.WithContext(requestCtx))
	if err != nil {
		result.CheckStage = "llm_request"
		result.Error = "LLM request failed: " + err.Error()
		return result
	}
	defer resp.Body.Close()

	result.CheckStage = "llm_request"
	result.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Passed = true
		return result
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		result.Passed = true
		result.RateLimited = true
		result.Error = "rate limited (429), but channel is reachable"
		return result
	}
	result.Error = "LLM returned status: " + resp.Status
	return result
}

func channelHealthHTTPClient(channel *model.Channel) (*http.Client, error) {
	if channel == nil {
		return nil, fmt.Errorf("channel is nil")
	}
	if !channel.Proxy {
		return newHealthHTTPClientNoProxy()
	}
	if channel.ChannelProxy == nil || strings.TrimSpace(*channel.ChannelProxy) == "" {
		proxyURL, err := SettingGetString(model.SettingKeyProxyURL)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(proxyURL) == "" {
			return nil, fmt.Errorf("proxy url is empty")
		}
		return newHealthHTTPClientCustomProxy(strings.TrimSpace(proxyURL))
	}
	return newHealthHTTPClientCustomProxy(strings.TrimSpace(*channel.ChannelProxy))
}

func newHealthHTTPClientNoProxy() (*http.Client, error) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("default transport is not *http.Transport")
	}
	cloned := transport.Clone()
	cloned.Proxy = nil
	return &http.Client{Transport: cloned}, nil
}

func newHealthHTTPClientCustomProxy(proxyURLStr string) (*http.Client, error) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("default transport is not *http.Transport")
	}
	cloned := transport.Clone()
	parsed, err := url.Parse(proxyURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy url: %w", err)
	}
	switch parsed.Scheme {
	case "http", "https":
		cloned.Proxy = http.ProxyURL(parsed)
	case "socks", "socks5":
		socksDialer, err := proxy.FromURL(parsed, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("invalid socks proxy: %w", err)
		}
		cloned.Proxy = nil
		cloned.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialProxyContext(ctx, socksDialer, network, addr)
		}
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", parsed.Scheme)
	}
	return &http.Client{Transport: cloned}, nil
}

func dialProxyContext(ctx context.Context, dialer proxy.Dialer, network, addr string) (net.Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
		return contextDialer.DialContext(ctx, network, addr)
	}

	type dialResult struct {
		conn net.Conn
		err  error
	}
	resultCh := make(chan dialResult, 1)
	go func() {
		conn, err := dialer.Dial(network, addr)
		if conn != nil && ctx.Err() != nil {
			_ = conn.Close()
		}
		resultCh <- dialResult{conn: conn, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		return result.conn, result.err
	}
}

func getURLDelayForHealthCheck(httpClient *http.Client, rawURL string, ctx context.Context) (int, error) {
	if httpClient == nil {
		return 0, fmt.Errorf("http client is nil")
	}
	if _, err := url.Parse(rawURL); err != nil {
		return 0, err
	}
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return int(time.Since(start).Milliseconds()), nil
}

func channelModelHealthCheckTargetKey(target ChannelModelHealthCheckTarget) string {
	return fmt.Sprintf("%d|%s|%d|%s", target.GroupID, strings.TrimSpace(target.GroupName), target.ChannelID, strings.TrimSpace(target.ModelName))
}

func sortChannelModelHealthCheckTargets(targets []ChannelModelHealthCheckTarget) {
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].GroupName != targets[j].GroupName {
			return targets[i].GroupName < targets[j].GroupName
		}
		if targets[i].GroupID != targets[j].GroupID {
			return targets[i].GroupID < targets[j].GroupID
		}
		if targets[i].ChannelID != targets[j].ChannelID {
			return targets[i].ChannelID < targets[j].ChannelID
		}
		return targets[i].ModelName < targets[j].ModelName
	})
}

func dedupeAndSortHealthCheckTargets(targets []ChannelModelHealthCheckTarget) []ChannelModelHealthCheckTarget {
	if len(targets) == 0 {
		return nil
	}
	deduped := make([]ChannelModelHealthCheckTarget, 0, len(targets))
	seen := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		target.GroupName = strings.TrimSpace(target.GroupName)
		target.ModelName = strings.TrimSpace(target.ModelName)
		if target.ChannelID == 0 || target.ModelName == "" {
			continue
		}
		key := channelModelHealthCheckTargetKey(target)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, target)
	}
	sortChannelModelHealthCheckTargets(deduped)
	return deduped
}

func RunImportHealthChecks(ctx context.Context, targets []ChannelModelHealthCheckTarget) *model.DBImportHealthCheckReport {
	targets = dedupeAndSortHealthCheckTargets(targets)
	if len(targets) == 0 {
		return nil
	}
	checks := make([]model.DBImportHealthCheckItem, len(targets))
	targetGroupSet := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		if strings.TrimSpace(target.GroupName) != "" {
			targetGroupSet[target.GroupName] = struct{}{}
		}
	}
	workerCount := importHealthCheckMaxConcurrency
	if workerCount > len(targets) {
		workerCount = len(targets)
	}
	if workerCount <= 0 {
		workerCount = 1
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				target := targets[idx]
				channel, err := loadChannelForHealthCheck(ctx, target.ChannelID)
				if err != nil {
					checks[idx] = model.DBImportHealthCheckItem{
						GroupName:  target.GroupName,
						ChannelID:  target.ChannelID,
						Model:      target.ModelName,
						Skipped:    true,
						CheckStage: "lookup",
						Error:      err.Error(),
					}
					continue
				}
				result := CheckChannelModelHealth(ctx, channel, target.ModelName)
				result.GroupName = target.GroupName
				checks[idx] = result
			}
		}()
	}
	for idx := range targets {
		jobs <- idx
	}
	close(jobs)
	wg.Wait()
	sort.Slice(checks, func(i, j int) bool {
		if checks[i].GroupName != checks[j].GroupName {
			return checks[i].GroupName < checks[j].GroupName
		}
		if checks[i].ChannelName != checks[j].ChannelName {
			return checks[i].ChannelName < checks[j].ChannelName
		}
		return checks[i].Model < checks[j].Model
	})
	report := &model.DBImportHealthCheckReport{Checks: checks}
	report.Summary = &model.DBImportHealthCheckSummary{TargetGroups: len(targetGroupSet), Targets: len(checks)}
	for _, check := range checks {
		if check.Skipped {
			report.Summary.Skipped++
			if check.CheckStage == "connectivity" {
				report.Summary.ConnectivityOnly++
			}
			continue
		}
		if check.Passed {
			report.Summary.Passed++
			if check.RateLimited {
				report.Summary.RateLimited++
			}
			continue
		}
		if check.CheckStage == "connectivity" {
			report.Summary.ConnectivityOnly++
		}
		report.Summary.Failed++
	}
	return report
}

func loadChannelForHealthCheck(ctx context.Context, channelID int) (*model.Channel, error) {
	var channel model.Channel
	if err := db.GetDB().WithContext(ctx).
		Preload("Keys").
		First(&channel, channelID).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}
