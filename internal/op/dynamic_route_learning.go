package op

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm/clause"
)

type DynamicRouteLearningObservation struct {
	ChannelID  int
	KeyID      int
	ModelName  string
	Success    bool
	Fallback   bool
	RaceWinner bool
	LatencyMs  int64
	Message    string
}

func DynamicRouteLearningList(ctx context.Context) (model.DynamicRouteLearningListResult, error) {
	rows := []model.DynamicRouteLearningState{}
	if err := db.GetDB().WithContext(ctx).Order("score desc, updated_at desc").Find(&rows).Error; err != nil {
		return model.DynamicRouteLearningListResult{}, err
	}
	enabled := settingBoolOrDefault(model.SettingKeyDynamicRoutingLearningEnabled, false)
	return model.DynamicRouteLearningListResult{Enabled: enabled, States: rows}, nil
}

func DynamicRouteLearningReset(ctx context.Context) error {
	return db.GetDB().WithContext(ctx).Where("1 = 1").Delete(&model.DynamicRouteLearningState{}).Error
}

func DynamicRouteLearningGet(channelID, keyID int, modelName string) (model.DynamicRouteLearningState, bool) {
	if channelID <= 0 || keyID <= 0 || strings.TrimSpace(modelName) == "" {
		return model.DynamicRouteLearningState{}, false
	}
	var state model.DynamicRouteLearningState
	err := db.GetDB().Where("channel_id = ? AND channel_key_id = ? AND model_name = ?", channelID, keyID, normalizeLearningModelName(modelName)).First(&state).Error
	if err != nil {
		return model.DynamicRouteLearningState{}, false
	}
	return state, true
}

func DynamicRouteLearningRecord(ctx context.Context, obs DynamicRouteLearningObservation) error {
	if !settingBoolOrDefault(model.SettingKeyDynamicRoutingLearningEnabled, false) {
		return nil
	}
	if obs.ChannelID <= 0 || obs.KeyID <= 0 || strings.TrimSpace(obs.ModelName) == "" {
		return fmt.Errorf("invalid dynamic route learning observation")
	}
	nowUnix := time.Now().Unix()
	state, ok := DynamicRouteLearningGet(obs.ChannelID, obs.KeyID, obs.ModelName)
	if !ok {
		state = model.DynamicRouteLearningState{ChannelID: obs.ChannelID, ChannelKeyID: obs.KeyID, ModelName: normalizeLearningModelName(obs.ModelName)}
	}
	if obs.Success {
		state.SuccessCount++
	} else {
		state.FailureCount++
	}
	if obs.Fallback {
		state.FallbackCount++
	}
	if obs.RaceWinner {
		state.RaceWinnerCount++
	}
	if obs.LatencyMs > 0 {
		if state.LatencyMsEWMA <= 0 {
			state.LatencyMsEWMA = float64(obs.LatencyMs)
		} else {
			state.LatencyMsEWMA = state.LatencyMsEWMA*0.8 + float64(obs.LatencyMs)*0.2
		}
	}
	state.LastSampleAt = nowUnix
	state.Score, state.Confidence = dynamicRouteLearningScore(state)
	state.RecentSamples = trimLearningSamples(state.RecentSamples, obs, nowUnix)
	return db.GetDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "channel_key_id"}, {Name: "model_name"}},
		UpdateAll: true,
	}).Create(&state).Error
}

func normalizeLearningModelName(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func dynamicRouteLearningScore(state model.DynamicRouteLearningState) (float64, float64) {
	total := state.SuccessCount + state.FailureCount
	if total <= 0 {
		return 0, 0
	}
	successRate := float64(state.SuccessCount) / float64(total)
	score := successRate*2 - 1
	if state.LatencyMsEWMA > 0 {
		score -= state.LatencyMsEWMA / 10000
	}
	score -= float64(state.FallbackCount) * 0.05
	score += float64(state.RaceWinnerCount) * 0.08
	confidence := float64(total) / 20
	if confidence > 1 {
		confidence = 1
	}
	return score, confidence
}

func trimLearningSamples(existing string, obs DynamicRouteLearningObservation, ts int64) string {
	status := "fail"
	if obs.Success {
		status = "success"
	}
	sample := fmt.Sprintf("%d:%s:%d", ts, status, obs.LatencyMs)
	parts := []string{}
	if strings.TrimSpace(existing) != "" {
		parts = strings.Split(existing, "\n")
	}
	parts = append(parts, sample)
	if len(parts) > 20 {
		parts = parts[len(parts)-20:]
	}
	return strings.Join(parts, "\n")
}
