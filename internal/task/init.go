package task

import (
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/price"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const (
	TaskPriceUpdate               = "price_update"
	TaskStatsSave                 = string(model.SettingKeyStatsSaveInterval)
	TaskRelayLogSave              = "relay_log_save"
	TaskSyncLLM                   = "sync_llm"
	TaskCleanLLM                  = "clean_llm"
	TaskBaseUrlDelay              = "base_url_delay"
	TaskDynamicRoutingSummaryScan = "dynamic_routing_summary_scan"
	TaskOpsCleanup                = "ops_cleanup"
)

func Init() {
	initTasks(func(key model.SettingKey) (int, error) {
		return op.SettingGetInt(key)
	})
}

func initTasks(getInt func(model.SettingKey) (int, error)) {
	if priceUpdateIntervalHours, err := getInt(model.SettingKeyModelInfoUpdateInterval); err != nil {
		log.Errorf("failed to get model info update interval: %v", err)
	} else {
		priceUpdateInterval := time.Duration(priceUpdateIntervalHours) * time.Hour
		// 注册价格更新任务
		Register(string(model.SettingKeyModelInfoUpdateInterval), priceUpdateInterval, true, func() {
			ctx, cancel := taskContext()
			defer cancel()
			if err := price.UpdateLLMPrice(ctx); err != nil {
				log.Warnf("failed to update price info: %v", err)
			}
		})
	}

	// 注册基础URL延迟任务
	Register(TaskBaseUrlDelay, 1*time.Hour, true, ChannelBaseUrlDelayTask)

	// Register the daily dynamic-routing summary scan used by the settings/dashboard surface.
	Register(TaskDynamicRoutingSummaryScan, dynamicRoutingSummaryScanInterval, true, DynamicRoutingSummaryScanTask)
	Register(TaskOpsCleanup, 12*time.Hour, true, func() {
		ctx, cancel := taskContext()
		defer cancel()
		if err := op.OpsCleanup(ctx); err != nil {
			log.Warnf("ops cleanup task failed: %v", err)
		}
	})

	// 注册LLM同步任务
	if syncLLMIntervalHours, err := getInt(model.SettingKeySyncLLMInterval); err != nil {
		log.Warnf("failed to get sync LLM interval: %v", err)
	} else {
		syncLLMInterval := time.Duration(syncLLMIntervalHours) * time.Hour
		Register(string(model.SettingKeySyncLLMInterval), syncLLMInterval, true, SyncModelsTask)
	}

	// 注册统计保存任务
	if statsSaveIntervalMinutes, err := getInt(model.SettingKeyStatsSaveInterval); err != nil {
		log.Warnf("failed to get stats save interval: %v", err)
	} else {
		statsSaveInterval := time.Duration(statsSaveIntervalMinutes) * time.Minute
		Register(TaskStatsSave, statsSaveInterval, false, op.StatsSaveDBTask)
	}
	// 注册中继日志保存任务
	Register(TaskRelayLogSave, 10*time.Minute, false, func() {
		ctx, cancel := taskContext()
		defer cancel()
		if err := op.RelayLogSaveDBTask(ctx); err != nil {
			log.Warnf("relay log save db task failed: %v", err)
		}
	})
}
