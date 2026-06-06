package relay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/helper"
	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
	transformerModel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

func allowsRacingByDefault(sourceType string) bool {
	return normalizeSourceType(sourceType) == dbmodel.ChannelKeySourceTypePublicFree
}

func allowsRacingByPolicy(policy dbmodel.RouteTargetResolvedPolicy) bool {
	sourceType := normalizeSourceType(policy.SourceType)
	billingMode := dbmodel.NormalizeBillingMode(policy.BillingMode)
	probePolicy := dbmodel.NormalizeProbePolicy(policy.ProbePolicy)

	switch billingMode {
	case dbmodel.BillingModePerRequest, dbmodel.BillingModePerQuota:
		return false
	case dbmodel.BillingModePerToken:
		return probePolicy == dbmodel.ProbePolicyConcurrent && policy.ProbeConcurrencyLimit > 1
	case dbmodel.BillingModeFree, dbmodel.BillingModeFlat:
		return true
	default:
		return allowsRacingByDefault(sourceType)
	}
}

func allowsRacingByModel(modelName string) bool {
	if strings.TrimSpace(modelName) == "" {
		return true
	}
	info, err := op.LLMGet(strings.TrimSpace(modelName))
	if err != nil {
		return true
	}

	switch dbmodel.NormalizeBillingMode(info.BillingMode) {
	case dbmodel.BillingModePerRequest, dbmodel.BillingModePerQuota:
		return false
	case dbmodel.BillingModePerToken:
		return dbmodel.NormalizeProbePolicy(info.ProbePolicy) == dbmodel.ProbePolicyConcurrent && info.ProbeConcurrencyLimit > 1
	case dbmodel.BillingModeFree, dbmodel.BillingModeFlat:
		return true
	default:
		return dbmodel.NormalizeProbePolicy(info.ProbePolicy) == dbmodel.ProbePolicyConcurrent && info.ProbeConcurrencyLimit > 1
	}
}

func isStreamingRequest(req *transformerModel.InternalLLMRequest) bool {
	return req != nil && req.Stream != nil && *req.Stream
}

func shouldEscalateToRace(group dbmodel.Group, policy dbmodel.RouteTargetResolvedPolicy, consecutiveFails int, tuning dynamicRoutingTuning) bool {
	if group.Mode != dbmodel.GroupModeFailover {
		return false
	}
	if group.FailoverWindowSec <= 0 {
		return false
	}
	if tuning.RaceConcurrency <= 1 {
		return false
	}
	threshold := defaultPositiveInt(tuning.RaceAfterFails, 2)
	if consecutiveFails < threshold {
		return false
	}
	return allowsRacingByPolicy(policy)
}

type raceCandidate struct {
	order         int
	channel       *dbmodel.Channel
	usedKey       dbmodel.ChannelKey
	internalReq   *transformerModel.InternalLLMRequest
	outAdapter    transformerModel.Outbound
	releaseBudget raceBudgetRelease
}

type raceFallbackOutcome struct {
	result          attemptResult
	executed        bool
	consumedToIndex int
}

func cloneInternalRequestForRace(req *transformerModel.InternalLLMRequest, modelName string) *transformerModel.InternalLLMRequest {
	if req == nil {
		return &transformerModel.InternalLLMRequest{Model: modelName}
	}
	cloned := req.DeepClone()
	cloned.Model = modelName
	return cloned
}

func runRaceProbe(ctx context.Context, req *relayRequest, channel *dbmodel.Channel, usedKey dbmodel.ChannelKey, outAdapter transformerModel.Outbound, internalReq *transformerModel.InternalLLMRequest) attemptResult {
	start := time.Now()
	result := attemptResult{Channel: channel, UsedKey: usedKey}
	defer func() {
		result.Duration = time.Since(start)
	}()

	if channel == nil {
		result.Err = fmt.Errorf("race probe channel unavailable")
		return result
	}
	if internalReq == nil {
		internalReq = &transformerModel.InternalLLMRequest{Model: req.requestModel}
	}
	result.ActualModel = internalReq.Model
	if outAdapter == nil {
		result.Err = fmt.Errorf("race probe adapter unavailable")
		return result
	}

	outReq, err := outAdapter.TransformRequest(ctx, internalReq, channel.GetBaseUrl(), usedKey.ChannelKey)
	if err != nil {
		result.Err = fmt.Errorf("race probe request build failed: %w", err)
		return result
	}

	ra := &relayAttempt{relayRequest: req, channel: channel}
	ra.copyHeaders(outReq)

	httpClient, err := helper.ChannelHttpClient(channel)
	if err != nil {
		result.Err = err
		return result
	}

	resp, err := httpClient.Do(outReq)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			result.Err = context.Canceled
			return result
		}
		result.Err = fmt.Errorf("race probe request failed: %w", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamErrorBodyBytes))
		if readErr != nil {
			result.Err = fmt.Errorf("failed to read race probe response body: %w", readErr)
			return result
		}
		result.Err = fmt.Errorf("race probe upstream error: %d: %s", resp.StatusCode, string(body))
		return result
	}

	internalResp, err := outAdapter.TransformResponse(ctx, resp)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			result.Err = context.Canceled
			return result
		}
		result.Err = fmt.Errorf("race probe transform failed: %w", err)
		return result
	}

	result.Success = true
	result.InternalResponse = internalResp
	if internalResp != nil && strings.TrimSpace(internalResp.Model) != "" {
		result.ActualModel = internalResp.Model
	}
	return result
}

func (req *relayRequest) finalizeRaceFallbackSuccess(result attemptResult) error {
	if req == nil || req.iter == nil || result.Channel == nil {
		return fmt.Errorf("race fallback result unavailable")
	}
	actualModel := strings.TrimSpace(result.ActualModel)
	if actualModel == "" {
		actualModel = req.requestModel
	}
	if result.InternalResponse != nil {
		req.metrics.SetInternalResponse(result.InternalResponse, actualModel)
	}
	body, err := req.inAdapter.TransformResponse(req.c.Request.Context(), result.InternalResponse)
	if err != nil {
		req.iter.Record(result.Channel.ID, result.UsedKey.ID, result.Channel.Name, actualModel, dbmodel.AttemptFailed, result.StatusCode, result.Duration, fmt.Sprintf("race fallback inbound transform failed: %v", err))
		return err
	}

	usedKey := result.UsedKey
	usedKey.StatusCode = result.StatusCode
	usedKey.LastUseTimeStamp = time.Now().Unix()
	usedKey.TotalCost += req.metrics.Stats.InputCost + req.metrics.Stats.OutputCost
	if err := op.ChannelKeyUpdate(usedKey); err != nil {
		log.Warnf("failed to update race winner key state: %v", err)
	}

	op.StatsChannelUpdate(result.Channel.ID, dbmodel.StatsMetrics{
		WaitTime:       result.Duration.Milliseconds(),
		RequestSuccess: 1,
	})
	balancer.RecordSuccess(result.Channel.ID, result.UsedKey.ID, actualModel)
	recordDynamicRouteLearning(req.c.Request.Context(), op.DynamicRouteLearningObservation{
		ChannelID:  result.Channel.ID,
		KeyID:      result.UsedKey.ID,
		ModelName:  actualModel,
		Success:    true,
		RaceWinner: true,
		LatencyMs:  result.Duration.Milliseconds(),
	})
	balancer.SetSticky(req.apiKeyID, req.requestModel, result.Channel.ID, result.UsedKey.ID)
	if !result.AttemptRecorded {
		req.iter.Record(result.Channel.ID, result.UsedKey.ID, result.Channel.Name, actualModel, dbmodel.AttemptSuccess, result.StatusCode, result.Duration, "race fallback winner")
	}
	req.c.Data(http.StatusOK, "application/json", body)
	return nil
}

func runRaceFallback(req *relayRequest, iter *balancer.Iterator, deadline time.Time, maxConcurrency int) (attemptResult, bool) {
	if req == nil || iter == nil {
		return attemptResult{}, false
	}
	remaining := iter.Len() - iter.Index() - 1
	if remaining <= 0 {
		return attemptResult{}, false
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	if maxConcurrency > remaining {
		maxConcurrency = remaining
	}

	candidates, _, _ := buildRaceCandidateBatch(req, iter, iter.Index()+1, maxConcurrency)

	if len(candidates) == 0 {
		return attemptResult{}, true
	}
	result, ok := executeRaceCandidateBatch(req, iter, candidates, deadline)
	if ok {
		return result, true
	}

	return attemptResult{}, true
}

func runRaceFallbackWindow(req *relayRequest, iter *balancer.Iterator, deadline time.Time, maxConcurrency int) raceFallbackOutcome {
	if req == nil || iter == nil {
		return raceFallbackOutcome{}
	}
	remaining := iter.Len() - iter.Index() - 1
	if remaining <= 0 {
		return raceFallbackOutcome{consumedToIndex: iter.Index()}
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	lastConsumedIndex := iter.Index()
	for startIdx := iter.Index() + 1; startIdx < iter.Len(); {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return raceFallbackOutcome{executed: true, consumedToIndex: lastConsumedIndex}
		}

		candidates, nextStart, batchConsumed := buildRaceCandidateBatch(req, iter, startIdx, maxConcurrency)
		if batchConsumed > lastConsumedIndex {
			lastConsumedIndex = batchConsumed
		}
		startIdx = nextStart
		if len(candidates) == 0 {
			continue
		}

		result, ok := executeRaceCandidateBatch(req, iter, candidates, deadline)
		if ok {
			return raceFallbackOutcome{result: result, executed: true, consumedToIndex: lastConsumedIndex}
		}
	}

	return raceFallbackOutcome{executed: true, consumedToIndex: lastConsumedIndex}
}

func buildRaceCandidateBatch(req *relayRequest, iter *balancer.Iterator, startIdx, maxConcurrency int) ([]raceCandidate, int, int) {
	if req == nil || iter == nil || startIdx >= iter.Len() {
		return nil, iter.Len(), iter.Index()
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	candidates := make([]raceCandidate, 0, maxConcurrency)
	lastConsumed := iter.Index()
	idx := startIdx
	for ; idx < iter.Len() && len(candidates) < maxConcurrency; idx++ {
		lastConsumed = idx
		item, ok := iter.CandidateAt(idx)
		if !ok {
			continue
		}
		channel, err := op.ChannelGet(item.ChannelID, req.c.Request.Context())
		if err != nil {
			iter.Skip(item.ChannelID, 0, "", fmt.Sprintf("race candidate channel unavailable: %v", err))
			continue
		}
		if !channel.Enabled {
			iter.Skip(channel.ID, 0, channel.Name, "race candidate channel disabled")
			continue
		}

		targetModel := strings.TrimSpace(item.ModelName)
		if targetModel == "" {
			targetModel = req.requestModel
		}
		requestFormat := req.requestCapabilityFor(channel, targetModel)
		if !channel.HasConfiguredKeyForRequest(targetModel, requestFormat) {
			iter.Skip(channel.ID, 0, channel.Name, fmt.Sprintf("race candidate stale route item: channel does not declare model or has no configured key for model %s and request format %s", targetModel, requestFormat))
			continue
		}

		usedKey := channel.GetChannelKeyForRequest(targetModel, requestFormat)
		if strings.TrimSpace(usedKey.ChannelKey) == "" {
			iter.Skip(channel.ID, 0, channel.Name, "race candidate has no available key")
			continue
		}
		if iter.SkipCircuitBreak(channel.ID, usedKey.ID, channel.Name) {
			continue
		}

		if req.internalRequest != nil {
			if req.internalRequest.IsEmbeddingRequest() && !outbound.IsEmbeddingChannelType(channel.Type) {
				iter.Skip(channel.ID, usedKey.ID, channel.Name, "channel type not compatible with embedding request")
				continue
			}
			if req.internalRequest.IsChatRequest() && !outbound.IsChatChannelType(channel.Type) {
				iter.Skip(channel.ID, usedKey.ID, channel.Name, "channel type not compatible with chat request")
				continue
			}
		}

		policy := op.ResolveRouteTargetPolicy(channel, usedKey, targetModel)
		if !allowsRacingByPolicy(policy) {
			iter.Skip(channel.ID, usedKey.ID, channel.Name, "route-target forbids racing")
			continue
		}

		releaseBudget, err := acquireRaceBudgets(req.c.Request.Context(), iter.GroupID(), channel.ID, usedKey.ID)
		if err != nil {
			iter.Skip(channel.ID, usedKey.ID, channel.Name, fmt.Sprintf("race candidate budget exhausted: %v", err))
			continue
		}

		outAdapter := outbound.GetForModel(channel.Type, targetModel)
		if outAdapter == nil {
			releaseBudget()
			iter.Skip(channel.ID, usedKey.ID, channel.Name, fmt.Sprintf("unsupported channel type: %d", channel.Type))
			continue
		}

		candidates = append(candidates, raceCandidate{
			order:         len(candidates),
			channel:       channel,
			usedKey:       usedKey,
			internalReq:   cloneInternalRequestForRace(req.internalRequest, targetModel),
			outAdapter:    outAdapter,
			releaseBudget: releaseBudget,
		})
	}

	return candidates, idx, lastConsumed
}

func executeRaceCandidateBatch(req *relayRequest, iter *balancer.Iterator, candidates []raceCandidate, deadline time.Time) (attemptResult, bool) {
	if req == nil || iter == nil || len(candidates) == 0 {
		return attemptResult{}, false
	}

	baseCtx := req.c.Request.Context()
	var raceCtx context.Context
	var cancel context.CancelFunc
	if !deadline.IsZero() {
		raceCtx, cancel = context.WithDeadline(baseCtx, deadline)
	} else {
		raceCtx, cancel = context.WithCancel(baseCtx)
	}
	defer cancel()

	outcomesCh := make(chan raceOutcome, len(candidates))
	var wg sync.WaitGroup
	for _, candidate := range candidates {
		candidate := candidate
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer candidate.releaseBudget()
			result := runRaceProbe(raceCtx, req, candidate.channel, candidate.usedKey, candidate.outAdapter, candidate.internalReq)
			outcomesCh <- raceOutcome{index: candidate.order, result: result}
		}()
	}

	go func() {
		wg.Wait()
		close(outcomesCh)
	}()

	completed := make(map[int]attemptResult, len(candidates))
	outcomes := make([]raceOutcome, 0, len(candidates))
	selectedIndex := -1
	for outcome := range outcomesCh {
		completed[outcome.index] = outcome.result
		outcomes = append(outcomes, outcome)

		if selectedIndex >= 0 {
			continue
		}
		for i := 0; i < len(candidates); i++ {
			result, ok := completed[i]
			if !ok {
				break
			}
			if result.Success {
				selectedIndex = i
				cancel()
				break
			}
		}
	}

	sort.Slice(outcomes, func(i, j int) bool {
		return outcomes[i].index < outcomes[j].index
	})

	for _, outcome := range outcomes {
		if outcome.index == selectedIndex {
			continue
		}
		result := outcome.result
		if result.Channel == nil {
			continue
		}
		if result.Success {
			continue
		}
		if errors.Is(result.Err, context.Canceled) || errors.Is(result.Err, context.DeadlineExceeded) {
			iter.Record(result.Channel.ID, result.UsedKey.ID, result.Channel.Name, result.ActualModel, dbmodel.AttemptCanceled, result.StatusCode, result.Duration, "race candidate canceled after winner selection")
			continue
		}
		if result.Err != nil {
			result.UsedKey.StatusCode = result.StatusCode
			result.UsedKey.LastUseTimeStamp = time.Now().Unix()
			if err := op.ChannelKeyUpdate(result.UsedKey); err != nil {
				log.Warnf("failed to update race failed key state: %v", err)
			}
			op.StatsChannelUpdate(result.Channel.ID, dbmodel.StatsMetrics{
				WaitTime:      result.Duration.Milliseconds(),
				RequestFailed: 1,
			})
			balancer.RecordFailure(result.Channel.ID, result.UsedKey.ID, req.requestModel)
			recordDynamicRouteLearning(req.c.Request.Context(), op.DynamicRouteLearningObservation{
				ChannelID: result.Channel.ID,
				KeyID:     result.UsedKey.ID,
				ModelName: result.ActualModel,
				Success:   false,
				Fallback:  true,
				LatencyMs: result.Duration.Milliseconds(),
				Message:   result.Err.Error(),
			})
			iter.Record(result.Channel.ID, result.UsedKey.ID, result.Channel.Name, result.ActualModel, dbmodel.AttemptFailed, result.StatusCode, result.Duration, result.Err.Error())
			recordProbeFailureEvent(result.Channel, result.UsedKey, result.ActualModel, result.StatusCode, result.Duration, result.Err.Error())
		}
	}

	if selectedIndex >= 0 {
		recordRaceAttemptOutcomes(iter, outcomes, selectedIndex)
		recordProbeSuccessOutcomes(outcomes, selectedIndex)
		winner := completed[selectedIndex]
		winner.AttemptRecorded = true
		return winner, true
	}

	return attemptResult{}, false
}

func waitForRetryDelay(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func recordDynamicRouteLearning(ctx context.Context, obs op.DynamicRouteLearningObservation) {
	if err := op.DynamicRouteLearningRecord(ctx, obs); err != nil {
		log.Warnf("failed to record dynamic route learning signal: %v", err)
	}
}
