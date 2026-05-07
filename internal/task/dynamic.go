package task

import (
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const dynamicRoutingSummaryScanInterval = 24 * time.Hour

const dynamicRoutingSummaryMessageHealthDisabledScanSkipped = "health_disabled_scan_skipped"

type DynamicRoutingSummaryScanSummary struct {
	LastRunAt        time.Time `json:"last_run_at"`
	LastSuccessAt    time.Time `json:"last_success_at"`
	LastStatus       string    `json:"last_status,omitempty"`
	LastMessage      string    `json:"last_message,omitempty"`
	CurrentMode      string    `json:"current_mode,omitempty"`
	EffectiveMode    string    `json:"effective_mode,omitempty"`
	Decision         string    `json:"decision,omitempty"`
	DecisionReason   string    `json:"decision_reason,omitempty"`
	HealthEnabled    bool      `json:"health_enabled"`
	ChannelCount     int       `json:"channel_count"`
	EnabledChannels  int       `json:"enabled_channels"`
	GroupCount       int       `json:"group_count"`
	FailoverGroups   int       `json:"failover_groups"`
	FreePublicKeys   int       `json:"free_public_keys"`
	PaidMeteredKeys  int       `json:"paid_metered_keys"`
	PrivateInnerKeys int       `json:"private_internal_keys"`
	UnknownKeys      int       `json:"unknown_keys"`
	Basis            string    `json:"basis,omitempty"`
}

var (
	dynamicRoutingSummaryScanSummary   DynamicRoutingSummaryScanSummary
	dynamicRoutingSummaryScanSummaryMu sync.RWMutex
)

func DynamicRoutingSummaryScanTask() {
	log.Debugf("dynamic routing summary scan task started")
	startTime := time.Now()
	defer func() {
		log.Debugf("dynamic routing summary scan task finished, elapsed: %s", time.Since(startTime))
	}()

	summary := DynamicRoutingSummaryScanSummary{
		LastRunAt:  startTime,
		LastStatus: "ok",
		Basis:      "daily_summary_scan_no_runtime_mutation",
	}
	modeSnapshot := relay.DynamicRoutingSummarySnapshotForTask()
	summary.CurrentMode = modeSnapshot.CurrentMode
	summary.HealthEnabled = modeSnapshot.HealthEnabled
	summary.EffectiveMode = modeSnapshot.EffectiveMode
	summary.Decision = modeSnapshot.Decision
	summary.DecisionReason = modeSnapshot.DecisionReason

	healthEnabled, err := op.SettingGetBool(model.SettingKeyDynamicRoutingHealthEnabled)
	if err != nil {
		summary.LastStatus = "error"
		summary.LastMessage = err.Error()
		setDynamicRoutingSummaryScanSummary(summary)
		log.Warnf("dynamic routing summary scan task failed to read health switch: %v", err)
		return
	}
	summary.HealthEnabled = healthEnabled
	if !healthEnabled {
		summary.LastSuccessAt = startTime
		summary.LastStatus = "skipped"
		summary.LastMessage = dynamicRoutingSummaryMessageHealthDisabledScanSkipped
		setDynamicRoutingSummaryScanSummary(summary)
		return
	}

	ctx, cancel := taskContextWithTimeout(5 * time.Minute)
	defer cancel()

	channels, err := op.ChannelList(ctx)
	if err != nil {
		summary.LastStatus = "error"
		summary.LastMessage = err.Error()
		setDynamicRoutingSummaryScanSummary(summary)
		log.Warnf("dynamic routing summary scan task failed to list channels: %v", err)
		return
	}
	groups, err := op.GroupList(ctx)
	if err != nil {
		summary.LastStatus = "error"
		summary.LastMessage = err.Error()
		setDynamicRoutingSummaryScanSummary(summary)
		log.Warnf("dynamic routing summary scan task failed to list groups: %v", err)
		return
	}

	summary.ChannelCount = len(channels)
	summary.GroupCount = len(groups)
	for _, channel := range channels {
		if channel.Enabled {
			summary.EnabledChannels++
		}
		for _, key := range channel.Keys {
			switch model.EffectiveChannelKeySourceType(key.SourceType) {
			case model.ChannelKeySourceTypePublicFree:
				summary.FreePublicKeys++
			case model.ChannelKeySourceTypePaidMetered:
				summary.PaidMeteredKeys++
			case model.ChannelKeySourceTypePrivateInternal:
				summary.PrivateInnerKeys++
			default:
				summary.UnknownKeys++
			}
		}
	}
	for _, group := range groups {
		if group.Mode == model.GroupModeFailover {
			summary.FailoverGroups++
		}
	}

	summary.LastSuccessAt = startTime
	setDynamicRoutingSummaryScanSummary(summary)
}

func GetDynamicRoutingSummaryScanSummary() DynamicRoutingSummaryScanSummary {
	dynamicRoutingSummaryScanSummaryMu.RLock()
	defer dynamicRoutingSummaryScanSummaryMu.RUnlock()
	return dynamicRoutingSummaryScanSummary
}

func setDynamicRoutingSummaryScanSummary(summary DynamicRoutingSummaryScanSummary) {
	dynamicRoutingSummaryScanSummaryMu.Lock()
	defer dynamicRoutingSummaryScanSummaryMu.Unlock()
	dynamicRoutingSummaryScanSummary = summary
}
