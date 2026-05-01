package op

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

const maxAIAutomationModelDiscoveryResponseBytes int64 = 2 << 20

func aiAutomationFetchModels(req model.AIModelsFetchRequest, ctx context.Context) (model.AIModelsFetchResult, error) {
	requestedRemote := aiAutomationFetchModelsRequestedRemote(req)
	config, err := aiAutomationConfigGetRaw(ctx)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	resolvedReq := req
	if strings.TrimSpace(resolvedReq.BaseURL) == "" {
		resolvedReq.BaseURL = config.BaseURL
	}
	if strings.TrimSpace(resolvedReq.APIKey) == "" {
		resolvedReq.APIKey = config.APIKey
	}
	if strings.TrimSpace(resolvedReq.ChannelType) == "" {
		resolvedReq.ChannelType = config.ChannelType
	}
	if resolvedReq.UseLocalDefault == nil {
		useLocalDefault := config.UseLocalDefault
		resolvedReq.UseLocalDefault = &useLocalDefault
	}

	remote, remoteErr := aiAutomationFetchModelsRemote(resolvedReq, ctx)
	if remoteErr == nil {
		if len(remote.Candidates) > 0 {
			return remote, nil
		}
		return model.AIModelsFetchResult{}, fmt.Errorf("remote model discovery returned no models")
	}
	if requestedRemote {
		return model.AIModelsFetchResult{}, remoteErr
	}
	return aiAutomationFetchModelsLocal(ctx)
}

func aiAutomationFetchModelsRequestedRemote(req model.AIModelsFetchRequest) bool {
	if strings.TrimSpace(req.BaseURL) != "" || strings.TrimSpace(req.APIKey) != "" || strings.TrimSpace(req.ChannelType) != "" {
		return true
	}
	if req.UseLocalDefault != nil && !*req.UseLocalDefault {
		return true
	}
	return false
}

func aiAutomationFetchModelsRemote(req model.AIModelsFetchRequest, ctx context.Context) (model.AIModelsFetchResult, error) {
	useLocalDefault := req.UseLocalDefault != nil && *req.UseLocalDefault
	baseURL, err := resolveAIAutomationBaseURL(req.BaseURL, useLocalDefault)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	channelType := strings.ToLower(strings.TrimSpace(req.ChannelType))
	if channelType == "" {
		channelType = model.DefaultAIAutomationChannelType
	}
	httpClient, err := newHealthHTTPClientNoProxy()
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	httpClient.Timeout = 20 * time.Second
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	applyAIAutomationAuthHeaders(httpReq, channelType, strings.TrimSpace(req.APIKey))
	httpReq.Close = true
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(httpResp.Body, 2048))
		return model.AIModelsFetchResult{}, fmt.Errorf("remote model discovery returned status %d: %s", httpResp.StatusCode, strings.TrimSpace(string(body)))
	}
	modelNames, err := decodeRemoteAIAutomationModels(httpResp, channelType)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	infos, err := LLMList(ctx)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	infoByName := make(map[string]model.LLMInfo, len(infos))
	for _, info := range infos {
		infoByName[strings.ToLower(strings.TrimSpace(info.Name))] = info
	}
	seen := make(map[string]struct{}, len(modelNames))
	candidates := make([]model.AIModelCandidate, 0, len(modelNames))
	for _, modelName := range modelNames {
		name := strings.TrimSpace(modelName)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidate := model.AIModelCandidate{Name: name, Source: model.AIAutomationModelSourceRemoteDiscovery, Available: true, SuccessRate: 1}
		if info, ok := infoByName[key]; ok {
			candidate.FreeLikely = aiAutomationFreeLikely(info)
		}
		candidate.Reason = aiModelCandidateReason(candidate)
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return model.AIModelsFetchResult{}, fmt.Errorf("remote model discovery returned no models")
	}
	return finalizeAIAutomationModelCandidates(model.AIAutomationModelSourceRemoteDiscovery, candidates), nil
}

func aiAutomationFetchModelsLocal(ctx context.Context) (model.AIModelsFetchResult, error) {
	channels, err := ChannelList(ctx)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	infos, err := LLMList(ctx)
	if err != nil {
		return model.AIModelsFetchResult{}, err
	}
	infoByName := make(map[string]model.LLMInfo, len(infos))
	for _, info := range infos {
		infoByName[strings.ToLower(strings.TrimSpace(info.Name))] = info
	}
	seen := make(map[string]model.AIModelCandidate)
	for _, channel := range channels {
		for _, modelName := range splitModelNames(channel.Model, channel.CustomModel) {
			channelCopy := channel
			candidate := modelCandidateFromChannel(modelName, &channelCopy, infoByName)
			key := strings.ToLower(candidate.Name)
			if existing, ok := seen[key]; ok {
				candidate = mergeModelCandidate(existing, candidate)
			}
			seen[key] = candidate
		}
	}
	for _, info := range infos {
		name := strings.TrimSpace(info.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = modelCandidateFromInfo(info)
	}
	candidates := make([]model.AIModelCandidate, 0, len(seen))
	for _, candidate := range seen {
		candidate.Reason = aiModelCandidateReason(candidate)
		candidates = append(candidates, candidate)
	}
	return finalizeAIAutomationModelCandidates(model.AIAutomationModelSourceLocalCache, candidates), nil
}

func finalizeAIAutomationModelCandidates(source string, candidates []model.AIModelCandidate) model.AIModelsFetchResult {
	sort.SliceStable(candidates, func(i, j int) bool { return aiModelCandidateRank(candidates[i]) > aiModelCandidateRank(candidates[j]) })
	selected := ""
	if len(candidates) > 0 {
		candidates[0].Recommended = true
		selected = candidates[0].Name
	}
	return model.AIModelsFetchResult{Source: source, Candidates: candidates, SelectedName: selected, Policy: model.AIAutomationDefaultSelectionPolicy}
}

func applyAIAutomationAuthHeaders(req *http.Request, channelType, apiKey string) {
	if req == nil || apiKey == "" {
		return
	}
	switch strings.ToLower(strings.TrimSpace(channelType)) {
	case "anthropic":
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Anthropic-Version", "2023-06-01")
	case "gemini":
		req.Header.Set("X-Goog-Api-Key", apiKey)
	default:
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func decodeRemoteAIAutomationModels(resp *http.Response, channelType string) ([]string, error) {
	rawPayload, err := decodeRemoteAIAutomationPayload(resp.Body)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(strings.TrimSpace(channelType)) {
	case "anthropic":
		var payload model.AnthropicModelList
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			return nil, err
		}
		out := make([]string, 0, len(payload.Data))
		for _, item := range payload.Data {
			if strings.TrimSpace(item.ID) != "" {
				out = append(out, item.ID)
			}
		}
		return out, nil
	case "gemini":
		var payload model.GeminiModelList
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			return nil, err
		}
		out := make([]string, 0, len(payload.Models))
		for _, item := range payload.Models {
			name := strings.TrimSpace(strings.TrimPrefix(item.Name, "models/"))
			if name != "" {
				out = append(out, name)
			}
		}
		return out, nil
	default:
		var payload model.OpenAIModelList
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			return nil, err
		}
		out := make([]string, 0, len(payload.Data))
		for _, item := range payload.Data {
			if strings.TrimSpace(item.ID) != "" {
				out = append(out, item.ID)
			}
		}
		return out, nil
	}
}

func decodeRemoteAIAutomationPayload(r io.Reader) ([]byte, error) {
	payload, err := io.ReadAll(io.LimitReader(r, maxAIAutomationModelDiscoveryResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxAIAutomationModelDiscoveryResponseBytes {
		return nil, fmt.Errorf("ai automation model discovery response too large")
	}
	return payload, nil
}
