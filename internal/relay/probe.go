package relay

import (
	"time"

	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/price"
	transformerModel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

func estimateProbeCosts(internalResp *transformerModel.InternalLLMResponse, actualModel string) (float64, float64) {
	if internalResp == nil || internalResp.Usage == nil {
		return 0, 0
	}

	modelPrice := price.GetLLMPrice(actualModel)
	if modelPrice == nil {
		return 0, 0
	}

	usage := internalResp.Usage
	cachedTokens := int64(0)
	if usage.PromptTokensDetails != nil {
		cachedTokens = usage.PromptTokensDetails.CachedTokens
	}
	nonCachedPromptTokens := usage.PromptTokens - cachedTokens
	if nonCachedPromptTokens < 0 {
		nonCachedPromptTokens = 0
	}

	if usage.AnthropicUsage {
		inputCost := (float64(cachedTokens)*modelPrice.CacheRead +
			float64(usage.PromptTokens)*modelPrice.Input +
			float64(usage.CacheCreationInputTokens)*modelPrice.CacheWrite) * 1e-6
		outputCost := float64(usage.CompletionTokens) * modelPrice.Output * 1e-6
		return inputCost, outputCost
	}

	inputCost := (float64(cachedTokens)*modelPrice.CacheRead + float64(nonCachedPromptTokens)*modelPrice.Input) * 1e-6
	outputCost := float64(usage.CompletionTokens) * modelPrice.Output * 1e-6
	return inputCost, outputCost
}

func applyProbeUsageToChannelKey(channel *dbmodel.Channel, usedKey dbmodel.ChannelKey, actualModel string, internalResp *transformerModel.InternalLLMResponse, statusCode int) (float64, float64) {
	inputCost, outputCost := estimateProbeCosts(internalResp, actualModel)
	if channel == nil || usedKey.ID == 0 {
		return inputCost, outputCost
	}

	usedKey.StatusCode = statusCode
	usedKey.LastUseTimeStamp = time.Now().Unix()
	usedKey.TotalCost += inputCost + outputCost
	if err := op.ChannelKeyUpdate(usedKey); err != nil {
		log.Warnf("failed to update probe success key state: %v", err)
	}

	return inputCost, outputCost
}

func recordProbeFailureEvent(channel *dbmodel.Channel, usedKey dbmodel.ChannelKey, actualModel string, statusCode int, duration time.Duration, message string) {
	if channel == nil {
		return
	}
	op.ProbeEventAdd(dbmodel.ProbeEvent{
		Time:         time.Now().Unix(),
		Status:       dbmodel.ProbeEventFailed,
		ChannelID:    channel.ID,
		ChannelKeyID: usedKey.ID,
		ChannelName:  channel.Name,
		ModelName:    actualModel,
		Duration:     int(duration.Milliseconds()),
		StatusCode:   statusCode,
		Message:      message,
	})
}

func recordProbeSuccessOutcomes(outcomes []raceOutcome, selectedIndex int) {
	for _, outcome := range outcomes {
		result := outcome.result
		if !result.Success || result.Channel == nil {
			continue
		}

		selected := outcome.index == selectedIndex
		inputCost, outputCost := 0.0, 0.0
		message := "race probe succeeded"
		status := dbmodel.ProbeEventSuccess
		if selected {
			message = "race fallback winner"
			status = dbmodel.ProbeEventSelected
		} else {
			inputCost, outputCost = applyProbeUsageToChannelKey(result.Channel, result.UsedKey, result.ActualModel, result.InternalResponse, result.StatusCode)
		}

		op.ProbeEventAdd(dbmodel.ProbeEvent{
			Time:                time.Now().Unix(),
			Status:              status,
			ChannelID:           result.Channel.ID,
			ChannelKeyID:        result.UsedKey.ID,
			ChannelName:         result.Channel.Name,
			ModelName:           result.ActualModel,
			Duration:            int(result.Duration.Milliseconds()),
			StatusCode:          result.StatusCode,
			Message:             message,
			EstimatedInputCost:  inputCost,
			EstimatedOutputCost: outputCost,
			PromotedToResponse:  selected,
		})
	}
}

func recordRaceAttemptOutcomes(iter raceAttemptRecorder, outcomes []raceOutcome, selectedIndex int) {
	if iter == nil {
		return
	}
	for _, outcome := range outcomes {
		result := outcome.result
		if !result.Success || result.Channel == nil {
			continue
		}
		if outcome.index == selectedIndex {
			iter.Record(result.Channel.ID, result.UsedKey.ID, result.Channel.Name, result.ActualModel, dbmodel.AttemptSuccess, result.StatusCode, result.Duration, "race fallback winner")
			continue
		}
		iter.Record(result.Channel.ID, result.UsedKey.ID, result.Channel.Name, result.ActualModel, dbmodel.AttemptCanceled, result.StatusCode, result.Duration, "race candidate succeeded but was canceled after a higher-priority winner was selected")
	}
}
