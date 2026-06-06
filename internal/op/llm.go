package op

import (
	"context"
	"fmt"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/cache"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xstrings"
)

var llmModelCache = cache.New[string, model.LLMInfo](16)

func normalizeLLMPolicyFields(info model.LLMInfo) (model.LLMInfo, error) {
	info.BillingMode = model.NormalizeBillingMode(info.BillingMode)
	if !model.IsValidBillingMode(info.BillingMode) {
		return info, fmt.Errorf("invalid billing mode: %q", info.BillingMode)
	}
	info.ProbePolicy = model.NormalizeProbePolicy(info.ProbePolicy)
	if !model.IsValidProbePolicy(info.ProbePolicy) {
		return info, fmt.Errorf("invalid probe policy: %q", info.ProbePolicy)
	}
	info.CachePolicy = model.NormalizeCachePolicy(info.CachePolicy)
	if !model.IsValidCachePolicy(info.CachePolicy) {
		return info, fmt.Errorf("invalid cache policy: %q", info.CachePolicy)
	}
	if info.ProbeIntervalSeconds <= 0 {
		info.ProbeIntervalSeconds = 3600
	}
	if info.ProbeConcurrencyLimit <= 0 {
		info.ProbeConcurrencyLimit = 1
	}
	return info, nil
}

func LLMList(ctx context.Context) ([]model.LLMInfo, error) {
	models := llmModelCache.Values()
	return models, nil
}

func LLMUpdate(model model.LLMInfo, ctx context.Context) error {
	model.Name = strings.ToLower(strings.TrimSpace(model.Name))
	model.CanonicalName = strings.ToLower(strings.TrimSpace(model.CanonicalName))
	if model.CanonicalName == "" {
		model.CanonicalName = llmname.CanonicalModelName(model.Name)
	}
	var err error
	model, err = normalizeLLMPolicyFields(model)
	if err != nil {
		return err
	}
	_, ok := llmModelCache.Get(model.Name)
	if !ok {
		return fmt.Errorf("model not found")
	}
	if err := db.GetDB().WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	llmModelCache.Set(model.Name, model)
	return nil
}

func LLMDelete(modelName string, ctx context.Context) error {
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	_, ok := llmModelCache.Get(modelName)
	if !ok {
		return fmt.Errorf("model not found")
	}
	if err := deleteGroupItemsByModelNames([]string{modelName}, ctx); err != nil {
		return err
	}
	if err := RouteTargetOverrideDeleteByModels([]string{modelName}, ctx); err != nil {
		return err
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.LLMInfo{Name: modelName}).Error; err != nil {
		return err
	}
	llmModelCache.Del(modelName)
	return nil
}

func LLMBatchDelete(modelNames []string, ctx context.Context) error {
	if len(modelNames) == 0 {
		return nil
	}
	normalizedNames := make([]string, 0, len(modelNames))
	seen := make(map[string]struct{}, len(modelNames))
	for _, modelName := range modelNames {
		name := strings.ToLower(strings.TrimSpace(modelName))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalizedNames = append(normalizedNames, name)
	}
	if len(normalizedNames) == 0 {
		return nil
	}
	if err := deleteGroupItemsByModelNames(normalizedNames, ctx); err != nil {
		return err
	}
	if err := RouteTargetOverrideDeleteByModels(normalizedNames, ctx); err != nil {
		return err
	}
	if err := db.GetDB().WithContext(ctx).Where("name IN ?", normalizedNames).Delete(&model.LLMInfo{}).Error; err != nil {
		return err
	}
	llmModelCache.Del(normalizedNames...)
	return nil
}

func LLMCreate(model model.LLMInfo, ctx context.Context) error {
	model.Name = strings.ToLower(strings.TrimSpace(model.Name))
	model.CanonicalName = strings.ToLower(strings.TrimSpace(model.CanonicalName))
	if model.CanonicalName == "" {
		model.CanonicalName = llmname.CanonicalModelName(model.Name)
	}
	var err error
	model, err = normalizeLLMPolicyFields(model)
	if err != nil {
		return err
	}
	_, ok := llmModelCache.Get(model.Name)
	if ok {
		return fmt.Errorf("model already exists")
	}
	if err := db.GetDB().WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	llmModelCache.Set(model.Name, model)
	return nil
}

func LLMBatchCreate(llmInfos []model.LLMInfo, ctx context.Context) error {
	if len(llmInfos) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(llmInfos))
	newLLMInfos := make([]model.LLMInfo, 0, len(llmInfos))
	for _, llmInfo := range llmInfos {
		llmInfo.Name = strings.ToLower(llmInfo.Name)
		llmInfo.CanonicalName = strings.ToLower(strings.TrimSpace(llmInfo.CanonicalName))
		if llmInfo.CanonicalName == "" {
			llmInfo.CanonicalName = llmname.CanonicalModelName(llmInfo.Name)
		}
		var err error
		llmInfo, err = normalizeLLMPolicyFields(llmInfo)
		if err != nil {
			return err
		}
		if _, ok := seen[llmInfo.Name]; ok {
			continue
		}
		if _, ok := llmModelCache.Get(llmInfo.Name); ok {
			continue
		}
		seen[llmInfo.Name] = struct{}{}
		newLLMInfos = append(newLLMInfos, llmInfo)
	}
	if len(newLLMInfos) == 0 {
		return nil
	}
	if err := db.GetDB().WithContext(ctx).Create(&newLLMInfos).Error; err != nil {
		return err
	}
	for _, llmInfo := range newLLMInfos {
		llmModelCache.Set(llmInfo.Name, llmInfo)
	}
	return nil
}

func LLMGet(name string) (model.LLMInfo, error) {
	info, ok := llmModelCache.Get(name)
	if !ok {
		return model.LLMInfo{}, fmt.Errorf("model not found")
	}
	return info, nil
}

func LLMGetByCanonical(canonicalName string) (model.LLMInfo, error) {
	canonicalName = strings.ToLower(strings.TrimSpace(canonicalName))
	if canonicalName == "" {
		return model.LLMInfo{}, fmt.Errorf("model not found")
	}
	for _, info := range llmModelCache.GetAll() {
		if strings.ToLower(strings.TrimSpace(info.CanonicalName)) == canonicalName {
			return info, nil
		}
	}
	return model.LLMInfo{}, fmt.Errorf("model not found")
}

func llmRefreshCache(ctx context.Context) error {
	models := []model.LLMInfo{}
	if err := db.GetDB().WithContext(ctx).Find(&models).Error; err != nil {
		return err
	}
	llmModelCache.Clear()
	for _, info := range models {
		llmModelCache.Set(info.Name, info)
	}
	return nil
}

func deleteGroupItemsByModelNames(modelNames []string, ctx context.Context) error {
	if len(modelNames) == 0 {
		return nil
	}
	channels, err := ChannelList(ctx)
	if err != nil {
		return err
	}
	keys := make([]model.GroupIDAndLLMName, 0, len(channels)*len(modelNames))
	for _, channel := range channels {
		channelModels := xstrings.SplitTrimCompact(",", channel.Model, channel.CustomModel)
		channelModelSet := make(map[string]struct{}, len(channelModels))
		for _, name := range channelModels {
			channelModelSet[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
		}
		for _, modelName := range modelNames {
			if _, ok := channelModelSet[modelName]; ok {
				keys = append(keys, model.GroupIDAndLLMName{ChannelID: channel.ID, ModelName: modelName})
			}
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return GroupItemBatchDelByChannelAndModels(keys, ctx)
}
