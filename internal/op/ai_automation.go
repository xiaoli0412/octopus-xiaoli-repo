package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

var ErrAIAutomationDisabled = errors.New("ai automation is disabled")

const aiTaskListMaxOffset = 10000

func AIAutomationConfigGet(ctx context.Context) (model.AIAutomationConfig, error) {
	config, err := aiAutomationConfigGetRaw(ctx)
	if err != nil {
		return model.AIAutomationConfig{}, err
	}
	return RedactAIAutomationConfigForResponse(config), nil
}

func aiAutomationConfigGetRaw(ctx context.Context) (model.AIAutomationConfig, error) {
	manual := model.AIAutomationConfigValues{
		BaseURL:         settingStringOrDefault(model.SettingKeyAIAutomationBaseURL, model.DefaultAIAutomationBaseURL),
		APIKey:          settingStringOrDefault(model.SettingKeyAIAutomationAPIKey, ""),
		ChannelType:     settingStringOrDefault(model.SettingKeyAIAutomationChannelType, model.DefaultAIAutomationChannelType),
		Model:           settingStringOrDefault(model.SettingKeyAIAutomationModel, ""),
		UseLocalDefault: settingBoolOrDefault(model.SettingKeyAIAutomationUseLocalDefault, true),
	}
	activeID, _ := SettingGetInt(model.SettingKeyActiveAIProfileID)
	requestedSourceMode := settingStringOrDefault(model.SettingKeyConfigSourceMode, model.ConfigSourceModeManual)
	config := model.AIAutomationConfig{
		Enabled:                       settingBoolOrDefault(model.SettingKeyAIAutomationEnabled, false),
		BaseURL:                       manual.BaseURL,
		APIKey:                        manual.APIKey,
		ChannelType:                   manual.ChannelType,
		Model:                         manual.Model,
		UseLocalDefault:               manual.UseLocalDefault,
		DefaultSelectionPolicy:        model.AIAutomationDefaultSelectionPolicy,
		RequestedConfigSourceMode:     requestedSourceMode,
		ConfigSourceMode:              requestedSourceMode,
		RequestedActiveAIProfileID:    activeID,
		ActiveAIProfileID:             activeID,
		DynamicRoutingLearningEnabled: settingBoolOrDefault(model.SettingKeyDynamicRoutingLearningEnabled, false),
		ManualConfig:                  manual,
		EffectiveConfig:               manual,
	}
	if activeID > 0 {
		config.RequestedActiveAIProfile = aiAutomationProfileRefByID(activeID, ctx)
	}
	return applyActiveAIProfileConfig(config, ctx)
}

func AIAutomationConfigUpdate(req model.AIAutomationConfigUpdateRequest, ctx context.Context) (model.AIAutomationConfig, error) {
	_ = ctx
	updates := []model.Setting{}
	if req.Enabled != nil {
		updates = append(updates, model.Setting{Key: model.SettingKeyAIAutomationEnabled, Value: strconv.FormatBool(*req.Enabled)})
	}
	if req.BaseURL != nil {
		if err := validateAIAutomationBaseURL(strings.TrimSpace(*req.BaseURL)); err != nil {
			return model.AIAutomationConfig{}, err
		}
		updates = append(updates, model.Setting{Key: model.SettingKeyAIAutomationBaseURL, Value: strings.TrimSpace(*req.BaseURL)})
	}
	if req.APIKey != nil {
		updates = append(updates, model.Setting{Key: model.SettingKeyAIAutomationAPIKey, Value: strings.TrimSpace(*req.APIKey)})
	}
	if req.ChannelType != nil {
		updates = append(updates, model.Setting{Key: model.SettingKeyAIAutomationChannelType, Value: strings.TrimSpace(*req.ChannelType)})
	}
	if req.Model != nil {
		updates = append(updates, model.Setting{Key: model.SettingKeyAIAutomationModel, Value: strings.TrimSpace(*req.Model)})
	}
	if req.UseLocalDefault != nil {
		updates = append(updates, model.Setting{Key: model.SettingKeyAIAutomationUseLocalDefault, Value: strconv.FormatBool(*req.UseLocalDefault)})
	}
	for _, update := range updates {
		if err := update.Validate(); err != nil {
			return model.AIAutomationConfig{}, err
		}
	}
	for _, update := range updates {
		if err := SettingSetString(update.Key, update.Value); err != nil {
			return model.AIAutomationConfig{}, err
		}
	}
	return AIAutomationConfigGet(ctx)
}

func settingStringOrDefault(key model.SettingKey, fallback string) string {
	value, err := SettingGetString(key)
	if err != nil || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func settingBoolOrDefault(key model.SettingKey, fallback bool) bool {
	value, err := SettingGetBool(key)
	if err != nil {
		return fallback
	}
	return value
}

func ensureAIAutomationEnabled(ctx context.Context) error {
	_ = ctx
	if !settingBoolOrDefault(model.SettingKeyAIAutomationEnabled, false) {
		return ErrAIAutomationDisabled
	}
	return nil
}

func AIAutomationFetchModels(req model.AIModelsFetchRequest, ctx context.Context) (model.AIModelsFetchResult, error) {
	if err := ensureAIAutomationEnabled(ctx); err != nil {
		return model.AIModelsFetchResult{}, err
	}
	return aiAutomationFetchModels(req, ctx)
}

func splitModelNames(values ...string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			name := strings.TrimSpace(part)
			if name == "" {
				continue
			}
			key := strings.ToLower(name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, name)
		}
	}
	return out
}

func aiAutomationFreeLikely(info model.LLMInfo) bool {
	billingMode := model.NormalizeBillingMode(info.BillingMode)
	if billingMode == model.BillingModeFree {
		return true
	}
	if billingMode != model.BillingModeUnknown {
		return false
	}
	return info.Input == 0 && info.Output == 0 && info.CacheRead == 0 && info.CacheWrite == 0
}

func modelCandidateFromChannel(modelName string, channel *model.Channel, infoByName map[string]model.LLMInfo) model.AIModelCandidate {
	candidate := model.AIModelCandidate{
		Name:        strings.TrimSpace(modelName),
		Source:      model.AIAutomationModelSourceLocalCache,
		Available:   channel != nil && channel.Enabled,
		SuccessRate: 1,
	}
	if channel != nil {
		candidate.ChannelID = channel.ID
		candidate.ChannelName = channel.Name
		if len(channel.BaseUrls) > 0 {
			best := 0
			set := false
			for _, baseURL := range channel.BaseUrls {
				if baseURL.Delay <= 0 {
					continue
				}
				if !set || baseURL.Delay < best {
					best = baseURL.Delay
					set = true
				}
			}
			if set {
				candidate.AvgLatencyMs = float64(best)
			}
		}
		stats, ok := StatsChannelSnapshot(channel.ID)
		if ok {
			total := stats.RequestSuccess + stats.RequestFailed
			if total > 0 {
				candidate.SuccessRate = float64(stats.RequestSuccess) / float64(total)
				candidate.AvgLatencyMs = float64(stats.WaitTime) / float64(total)
			}
		}
	}
	if info, ok := infoByName[strings.ToLower(candidate.Name)]; ok {
		candidate.FreeLikely = aiAutomationFreeLikely(info)
	}
	return candidate
}

func modelCandidateFromInfo(info model.LLMInfo) model.AIModelCandidate {
	freeLikely := aiAutomationFreeLikely(info)
	return model.AIModelCandidate{
		Name:        strings.TrimSpace(info.Name),
		Source:      model.AIAutomationModelSourceLocalCache,
		Available:   true,
		FreeLikely:  freeLikely,
		SuccessRate: 1,
	}
}

func mergeModelCandidate(a, b model.AIModelCandidate) model.AIModelCandidate {
	if !a.Available && b.Available {
		a = b
	}
	a.FreeLikely = a.FreeLikely || b.FreeLikely
	if b.SuccessRate > a.SuccessRate {
		a.SuccessRate = b.SuccessRate
	}
	if a.AvgLatencyMs <= 0 || (b.AvgLatencyMs > 0 && b.AvgLatencyMs < a.AvgLatencyMs) {
		a.AvgLatencyMs = b.AvgLatencyMs
	}
	return a
}

func aiModelCandidateRank(candidate model.AIModelCandidate) float64 {
	score := candidate.SuccessRate * 100
	if candidate.FreeLikely {
		score += 50
	}
	if candidate.Available {
		score += 20
	}
	if candidate.AvgLatencyMs > 0 {
		score -= candidate.AvgLatencyMs / 1000
	}
	return score
}

func aiModelCandidateReason(candidate model.AIModelCandidate) string {
	reasons := []string{model.AIAutomationDefaultSelectionPolicy}
	if candidate.FreeLikely {
		reasons = append(reasons, "free_likely")
	}
	if candidate.SuccessRate >= 0.95 {
		reasons = append(reasons, "high_success_rate")
	}
	if candidate.AvgLatencyMs > 0 && candidate.AvgLatencyMs < 1500 {
		reasons = append(reasons, "low_latency")
	}
	return strings.Join(reasons, ",")
}

func AIPromptTemplateList(ctx context.Context) ([]model.AIPromptTemplate, error) {
	ensureBuiltinPromptTemplates(ctx)
	rows := []model.AIPromptTemplate{}
	if err := db.GetDB().WithContext(ctx).Order("source asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func AIPromptTemplateCreate(req model.AIPromptTemplateCreateRequest, ctx context.Context) (model.AIPromptTemplate, error) {
	if !model.IsValidAITaskType(req.TaskType) {
		return model.AIPromptTemplate{}, fmt.Errorf("invalid task type")
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Prompt) == "" {
		return model.AIPromptTemplate{}, fmt.Errorf("name and prompt are required")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	row := model.AIPromptTemplate{
		Name:            strings.TrimSpace(req.Name),
		Source:          model.AIPromptTemplateSourceCustom,
		TaskType:        model.NormalizeAITaskType(req.TaskType),
		Domain:          model.NormalizeAIProfileDomain(req.Domain),
		Prompt:          strings.TrimSpace(req.Prompt),
		WorkRequirement: strings.TrimSpace(req.WorkRequirement),
		Enabled:         enabled,
	}
	if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
		return model.AIPromptTemplate{}, err
	}
	return row, nil
}

func ensureBuiltinPromptTemplates(ctx context.Context) {
	builtins := []model.AIPromptTemplate{
		{Name: "智能分组", Source: model.AIPromptTemplateSourceBuiltin, TaskType: model.AIAutomationTaskTypeGroupSuggestion, Domain: model.AIProfileDomainGrouping, Prompt: "Analyze channels and models, then propose non-destructive grouping profiles.", WorkRequirement: "Do not overwrite manual groups or group_items.", Enabled: true},
		{Name: "渠道识别", Source: model.AIPromptTemplateSourceBuiltin, TaskType: model.AIAutomationTaskTypeChannelRecognition, Domain: model.AIProfileDomainChannelRecognition, Prompt: "Identify channel types, source types, and model coverage from the provided context.", WorkRequirement: "Return suggestions as an AI Profile only.", Enabled: true},
		{Name: "价格识别", Source: model.AIPromptTemplateSourceBuiltin, TaskType: model.AIAutomationTaskTypePriceRecognition, Domain: model.AIProfileDomainPriceRecognition, Prompt: "Infer possible pricing and billing modes from public metadata and local context.", WorkRequirement: "Never change llm_infos directly.", Enabled: true},
		{Name: "动态路由说明整理", Source: model.AIPromptTemplateSourceBuiltin, TaskType: model.AIAutomationTaskTypeDynamicRoutingDigest, Domain: model.AIProfileDomainDynamicRoutingDigest, Prompt: "Summarize dynamic routing behavior and learning signals for operator review.", WorkRequirement: "Dynamic routing learning affects runtime recommendations only.", Enabled: true},
	}
	for _, row := range builtins {
		var existing int64
		_ = db.GetDB().WithContext(ctx).Model(&model.AIPromptTemplate{}).Where("source = ? AND name = ?", row.Source, row.Name).Count(&existing).Error
		if existing == 0 {
			_ = db.GetDB().WithContext(ctx).Create(&row).Error
		}
	}
}

func AITaskCreate(req model.AITaskCreateRequest, ctx context.Context) (model.AITask, error) {
	if err := ensureAIAutomationEnabled(ctx); err != nil {
		return model.AITask{}, err
	}
	taskType := model.NormalizeAITaskType(req.Type)
	if !model.IsValidAITaskType(taskType) {
		return model.AITask{}, fmt.Errorf("invalid task type")
	}
	ids := make([]string, 0, len(req.PromptTemplateIDs))
	for _, id := range req.PromptTemplateIDs {
		if id > 0 {
			ids = append(ids, strconv.Itoa(id))
		}
	}
	currentConfig, err := aiAutomationConfigGetRaw(ctx)
	if err != nil {
		return model.AITask{}, err
	}
	useLocalDefault := currentConfig.UseLocalDefault
	configSnapshot := model.AIAutomationTaskConfig{
		BaseURL:         currentConfig.BaseURL,
		APIKey:          currentConfig.APIKey,
		ChannelType:     currentConfig.ChannelType,
		Model:           currentConfig.Model,
		UseLocalDefault: &useLocalDefault,
	}
	if req.ConfigSnapshot != nil {
		normalized := normalizeAITaskConfigSnapshot(*req.ConfigSnapshot)
		if !isEmptyAITaskConfigSnapshot(normalized) {
			configSnapshot = mergeAITaskConfigSnapshot(configSnapshot, normalized)
		}
	}
	configSnapshot = normalizeAITaskConfigSnapshot(configSnapshot)
	configSnapshotJSON := ""
	if !isEmptyAITaskConfigSnapshot(configSnapshot) {
		raw, err := json.Marshal(configSnapshot)
		if err != nil {
			return model.AITask{}, err
		}
		configSnapshotJSON = string(raw)
	}
	task := model.AITask{
		Type:               taskType,
		InputText:          strings.TrimSpace(req.InputText),
		ContextScope:       strings.TrimSpace(req.ContextScope),
		PromptTemplateIDs:  strings.Join(ids, ","),
		CustomPrompt:       strings.TrimSpace(req.CustomPrompt),
		Status:             model.AITaskStatusPending,
		Progress:           0,
		ConfigSnapshotJSON: configSnapshotJSON,
		ResumeToken:        fmt.Sprintf("ai-task-%d", time.Now().UnixNano()),
		ResumeState:        model.AITaskResumeStateCollectContext,
		ExecutorVersion:    aiTaskExecutorVersion,
	}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		return model.AITask{}, err
	}
	steps := defaultAITaskSteps(task.ID)
	if err := db.GetDB().WithContext(ctx).Create(&steps).Error; err != nil {
		return model.AITask{}, err
	}
	task.Steps = steps
	AITaskStartAsync(task.ID)
	return task, nil
}

func mergeAITaskConfigSnapshot(base, override model.AIAutomationTaskConfig) model.AIAutomationTaskConfig {
	if strings.TrimSpace(override.BaseURL) != "" {
		base.BaseURL = strings.TrimSpace(override.BaseURL)
	}
	if strings.TrimSpace(override.APIKey) != "" {
		base.APIKey = strings.TrimSpace(override.APIKey)
	}
	if strings.TrimSpace(override.ChannelType) != "" {
		base.ChannelType = strings.TrimSpace(override.ChannelType)
	}
	if strings.TrimSpace(override.Model) != "" {
		base.Model = strings.TrimSpace(override.Model)
	}
	if override.UseLocalDefault != nil {
		value := *override.UseLocalDefault
		base.UseLocalDefault = &value
	}
	if len(override.ToolKeys) > 0 {
		base.ToolKeys = append([]string{}, override.ToolKeys...)
	}
	return base
}

func normalizeAITaskConfigSnapshot(input model.AIAutomationTaskConfig) model.AIAutomationTaskConfig {
	output := model.AIAutomationTaskConfig{
		BaseURL:     strings.TrimSpace(input.BaseURL),
		APIKey:      strings.TrimSpace(input.APIKey),
		ChannelType: strings.TrimSpace(input.ChannelType),
		Model:       strings.TrimSpace(input.Model),
	}
	if len(input.ToolKeys) > 0 {
		seen := map[string]struct{}{}
		output.ToolKeys = make([]string, 0, len(input.ToolKeys))
		for _, rawKey := range input.ToolKeys {
			key := model.NormalizeAITaskToolKey(rawKey)
			if !model.IsValidAITaskToolKey(key) {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			output.ToolKeys = append(output.ToolKeys, key)
		}
	}
	if input.UseLocalDefault != nil {
		value := *input.UseLocalDefault
		output.UseLocalDefault = &value
	}
	return output
}

func isEmptyAITaskConfigSnapshot(input model.AIAutomationTaskConfig) bool {
	return strings.TrimSpace(input.BaseURL) == "" &&
		strings.TrimSpace(input.APIKey) == "" &&
		strings.TrimSpace(input.ChannelType) == "" &&
		strings.TrimSpace(input.Model) == "" &&
		len(input.ToolKeys) == 0 &&
		input.UseLocalDefault == nil
}

func loadAITask(id int, ctx context.Context) (model.AITask, error) {
	if id <= 0 {
		return model.AITask{}, fmt.Errorf("invalid task id")
	}
	var task model.AITask
	if err := db.GetDB().WithContext(ctx).Preload("Steps", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order asc")
	}).First(&task, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.AITask{}, fmt.Errorf("task not found")
		}
		return model.AITask{}, err
	}
	return task, nil
}

func AITaskGet(id int, ctx context.Context) (model.AITask, error) {
	task, err := loadAITask(id, ctx)
	if err != nil {
		return model.AITask{}, err
	}
	return redactAIAutomationTask(task), nil
}

func AITaskList(req model.AITaskListRequest, ctx context.Context) (model.AITaskListResult, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		return model.AITaskListResult{}, fmt.Errorf("invalid page_size")
	}
	if int64(page-1)*int64(pageSize) >= aiTaskListMaxOffset {
		return model.AITaskListResult{}, fmt.Errorf("invalid page")
	}
	query := db.GetDB().WithContext(ctx).Model(&model.AITask{})
	if status := strings.TrimSpace(req.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType := model.NormalizeAITaskType(req.Type); strings.TrimSpace(req.Type) != "" && model.IsValidAITaskType(taskType) {
		query = query.Where("type = ?", taskType)
	}
	if domain := model.NormalizeAIProfileDomain(req.ProfileDomain); strings.TrimSpace(req.ProfileDomain) != "" && model.IsValidAIProfileDomain(domain) {
		profileTaskIDs := []int{}
		if err := db.GetDB().WithContext(ctx).Model(&model.AIProfile{}).Where("domain = ? AND source_task_id IS NOT NULL", domain).Pluck("source_task_id", &profileTaskIDs).Error; err != nil {
			return model.AITaskListResult{}, err
		}
		if len(profileTaskIDs) == 0 {
			return model.AITaskListResult{Items: []model.AITask{}, Total: 0, Page: page, PageSize: pageSize}, nil
		}
		query = query.Where("id IN ?", profileTaskIDs)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("input_text LIKE ? OR custom_prompt LIKE ? OR result_summary LIKE ? OR error_message LIKE ? OR selected_model LIKE ?", like, like, like, like, like)
	}
	if !req.CreatedFrom.IsZero() {
		query = query.Where("created_at >= ?", req.CreatedFrom)
	}
	if !req.CreatedTo.IsZero() {
		query = query.Where("created_at <= ?", req.CreatedTo)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return model.AITaskListResult{}, err
	}
	items := []model.AITask{}
	if err := query.Preload("Steps", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order asc") }).
		Order("created_at desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return model.AITaskListResult{}, err
	}
	for i := range items {
		items[i] = redactAIAutomationTask(items[i])
	}
	return model.AITaskListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func AITaskRetry(id int, ctx context.Context) (model.AITask, error) {
	task, err := loadAITask(id, ctx)
	if err != nil {
		return model.AITask{}, err
	}
	if !isTerminalAITaskStatus(task.Status) {
		return model.AITask{}, fmt.Errorf("invalid task status for retry")
	}
	req := model.AITaskCreateRequest{
		Type:              task.Type,
		InputText:         task.InputText,
		ContextScope:      task.ContextScope,
		PromptTemplateIDs: parseAITaskPromptTemplateIDs(task.PromptTemplateIDs),
		CustomPrompt:      task.CustomPrompt,
	}
	if strings.TrimSpace(task.ConfigSnapshotJSON) != "" {
		var snapshot model.AIAutomationTaskConfig
		if err := json.Unmarshal([]byte(task.ConfigSnapshotJSON), &snapshot); err != nil {
			return model.AITask{}, fmt.Errorf("invalid task config snapshot")
		}
		normalized := normalizeAITaskConfigSnapshot(snapshot)
		if !isEmptyAITaskConfigSnapshot(normalized) {
			req.ConfigSnapshot = &normalized
		}
	}
	return AITaskCreate(req, ctx)
}

func parseAITaskPromptTemplateIDs(raw string) []int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	ids := make([]int, 0)
	seen := map[int]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func AITaskArtifacts(id int, ctx context.Context) (model.AITaskArtifacts, error) {
	task, err := loadAITask(id, ctx)
	if err != nil {
		return model.AITaskArtifacts{}, err
	}
	artifacts := model.AITaskArtifacts{
		TaskID:             task.ID,
		ConfigSnapshotJSON: task.ConfigSnapshotJSON,
		ContextPayloadJSON: task.ContextPayloadJSON,
		ResultJSON:         task.ResultJSON,
		PromptText:         task.PromptText,
		SelectedModel:      task.SelectedModel,
		ModelReason:        task.ModelReason,
		ResumeState:        task.ResumeState,
		Steps:              task.Steps,
	}
	if strings.TrimSpace(task.ConfigSnapshotJSON) != "" {

		var snapshot model.AIAutomationTaskConfig
		if err := json.Unmarshal([]byte(task.ConfigSnapshotJSON), &snapshot); err == nil {
			normalized := normalizeAITaskConfigSnapshot(snapshot)
			if !isEmptyAITaskConfigSnapshot(normalized) {
				artifacts.ConfigSnapshot = &normalized
			}
		}
	}
	artifacts.ContextPayload = decodeAITaskArtifactsPayload(task.ContextPayloadJSON)
	artifacts.ResultPayload = decodeAITaskArtifactsPayload(task.ResultJSON)
	return redactAIAutomationArtifacts(artifacts), nil
}
func AITaskCancel(id int, ctx context.Context) (model.AITask, error) {
	task, err := AITaskGet(id, ctx)
	if err != nil {
		return model.AITask{}, err
	}
	trackedForExecution := isAITaskExecutionTracked(id)
	if isTerminalAITaskStatus(task.Status) {
		if task.Status == model.AITaskStatusFailed && trackedForExecution {
			return forceAITaskCanceled(id, ctx)
		}
		return task, nil
	}
	now := time.Now()
	updates := map[string]any{"status": model.AITaskStatusCanceled, "finished_at": &now, "progress": 100}
	result := db.GetDB().WithContext(ctx).Model(&model.AITask{}).
		Where("id = ? AND status NOT IN ?", id, []string{model.AITaskStatusSucceeded, model.AITaskStatusFailed, model.AITaskStatusFailedUnrecoverable, model.AITaskStatusCanceled}).
		Updates(updates)
	if result.Error != nil {
		return model.AITask{}, result.Error
	}
	if result.RowsAffected == 0 {
		return AITaskGet(id, ctx)
	}
	cancelAITaskExecution(id)
	if !waitForAITaskStop(id, aiTaskCancelWaitTimeout) {
		return model.AITask{}, fmt.Errorf("timed out waiting for AI task %d to stop", id)
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", id).Update("progress", 100).Error; err != nil {
		return model.AITask{}, err
	}
	markAITaskCanceledSteps(id)
	return AITaskGet(id, ctx)
}

func forceAITaskCanceled(id int, ctx context.Context) (model.AITask, error) {
	now := time.Now()
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", id).Updates(map[string]any{
		"status":        model.AITaskStatusCanceled,
		"error_message": "",
		"finished_at":   &now,
		"progress":      100,
	}).Error; err != nil {
		return model.AITask{}, err
	}
	cancelAITaskExecution(id)
	if !waitForAITaskStop(id, aiTaskCancelWaitTimeout) {
		return model.AITask{}, fmt.Errorf("timed out waiting for AI task %d to stop", id)
	}
	markAITaskCanceledSteps(id)
	return AITaskGet(id, ctx)
}

func RecoverInterruptedAITasks(ctx context.Context) error {
	incompleteStatuses := []string{model.AITaskStatusPending, model.AITaskStatusRunning, model.AITaskStatusRecoverable}
	tasks := make([]model.AITask, 0)
	if err := db.GetDB().WithContext(ctx).Preload("Steps", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order asc")
	}).Where("status IN ?", incompleteStatuses).Find(&tasks).Error; err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		if task.Status == model.AITaskStatusCanceled || isTerminalAITaskStatus(task.Status) {
			continue
		}
		resumeState := normalizeAITaskResumeState(task.ResumeState, task)
		if err := validateAITaskRecoverable(task, resumeState); err != nil {
			message := trimTextWithSuffix(err.Error(), aiTaskResultSummaryLimit, "...")
			now := time.Now()
			if updateErr := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", task.ID).Updates(map[string]any{
				"status":        model.AITaskStatusFailedUnrecoverable,
				"resume_state":  resumeState,
				"error_message": message,
				"finished_at":   &now,
			}).Error; updateErr != nil {
				return updateErr
			}
			continue
		}
		if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", task.ID).Updates(map[string]any{
			"status":        model.AITaskStatusRecoverable,
			"resume_state":  resumeState,
			"error_message": "",
		}).Error; err != nil {
			return err
		}
		AITaskStartAsync(task.ID)
	}
	return nil
}

func isTerminalAITaskStatus(status string) bool {
	switch status {
	case model.AITaskStatusSucceeded, model.AITaskStatusFailed, model.AITaskStatusFailedUnrecoverable, model.AITaskStatusCanceled:
		return true
	default:
		return false
	}
}

func normalizeAITaskResumeState(resumeState string, task model.AITask) string {
	switch strings.TrimSpace(resumeState) {
	case model.AITaskResumeStateCollectContext,
		model.AITaskResumeStateSelectModel,
		model.AITaskResumeStateCallAI,
		model.AITaskResumeStateParse,
		model.AITaskResumeStateGenerateProfile,
		model.AITaskResumeStateSaveResult:
		return strings.TrimSpace(resumeState)
	}
	for _, step := range task.Steps {
		if step.Status == model.AITaskStepStatusRunning {
			switch step.StepKey {
			case "collect_context":
				return model.AITaskResumeStateCollectContext
			case "select_model":
				return model.AITaskResumeStateSelectModel
			case "call_ai":
				return model.AITaskResumeStateCallAI
			case "parse_output":
				return model.AITaskResumeStateParse
			case "generate_profile":
				return model.AITaskResumeStateGenerateProfile
			case "save_result":
				return model.AITaskResumeStateSaveResult
			}
		}
	}
	return model.AITaskResumeStateCollectContext
}

func decodeAITaskArtifactsPayload(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil
	}
	return payload
}

func validateAITaskRecoverable(task model.AITask, resumeState string) error {
	if task.Status == model.AITaskStatusCanceled || isTerminalAITaskStatus(task.Status) {
		return fmt.Errorf("task is already terminal")
	}
	if strings.TrimSpace(task.ConfigSnapshotJSON) == "" {
		return fmt.Errorf("missing config snapshot")
	}
	switch resumeState {
	case model.AITaskResumeStateCollectContext:
		return nil
	case model.AITaskResumeStateSelectModel:
		if strings.TrimSpace(task.ContextPayloadJSON) == "" || strings.TrimSpace(task.PromptText) == "" {
			return fmt.Errorf("missing context or prompt checkpoint")
		}
	case model.AITaskResumeStateCallAI:
		if strings.TrimSpace(task.ContextPayloadJSON) == "" || strings.TrimSpace(task.PromptText) == "" || strings.TrimSpace(task.SelectedModel) == "" {
			return fmt.Errorf("missing selected model checkpoint")
		}
	case model.AITaskResumeStateParse:
		if strings.TrimSpace(task.ResultJSON) == "" {
			return fmt.Errorf("missing saved AI response")
		}
	case model.AITaskResumeStateGenerateProfile:
		if _, ok := aiTaskStepCheckpoint(task, "parse_output"); !ok {
			return fmt.Errorf("missing parsed result checkpoint")
		}
	case model.AITaskResumeStateSaveResult:
		if _, ok := aiTaskStepCheckpoint(task, "parse_output"); !ok {
			return fmt.Errorf("missing parsed result checkpoint")
		}
	}
	return nil
}

func aiTaskStepCheckpoint(task model.AITask, stepKey string) (string, bool) {
	for _, step := range task.Steps {
		if step.StepKey == stepKey && strings.TrimSpace(step.OutputJSON) != "" {
			return step.OutputJSON, true
		}
	}
	return "", false
}

func defaultAITaskSteps(taskID int) []model.AITaskStep {
	steps := []struct{ key, name string }{
		{"collect_context", "收集上下文"},
		{"select_model", "选择模型"},
		{"call_ai", "调用 AI"},
		{"parse_output", "解析输出"},
		{"generate_profile", "生成方案"},
		{"save_result", "保存结果"},
	}
	out := make([]model.AITaskStep, 0, len(steps))
	for idx, step := range steps {
		out = append(out, model.AITaskStep{TaskID: taskID, StepKey: step.key, Name: step.name, Status: model.AITaskStepStatusPending, SortOrder: idx + 1})
	}
	return out
}

func AIProfileList(ctx context.Context) ([]model.AIProfile, error) {
	rows := []model.AIProfile{}
	if err := db.GetDB().WithContext(ctx).Order("updated_at desc, id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		if strings.TrimSpace(rows[i].MigrationStatus) == "" && isTypedAIProfileDomain(rows[i].Domain) {
			if aiProfileHasTypedPayload(rows[i], ctx) {
				rows[i].MigrationStatus = model.AIProfileMigrationStatusTypedBackfilled
			} else {
				rows[i].MigrationStatus = model.AIProfileMigrationStatusLegacyOnly
			}
		}
	}
	return rows, nil
}

func AIProfileGet(id int, ctx context.Context) (model.AIProfile, error) {
	if id <= 0 {
		return model.AIProfile{}, fmt.Errorf("invalid profile id")
	}
	var profile model.AIProfile
	if err := db.GetDB().WithContext(ctx).Preload("Versions", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("version desc")
	}).First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.AIProfile{}, fmt.Errorf("profile not found")
		}
		return model.AIProfile{}, err
	}
	return enrichAIProfileDetail(profile, ctx), nil
}

func AIProfileGetRedacted(id int, ctx context.Context) (model.AIProfile, error) {
	profile, err := AIProfileGet(id, ctx)
	if err != nil {
		return model.AIProfile{}, err
	}
	return redactAIProfile(profile), nil
}

func AIProfileActivate(id int, ctx context.Context) (model.AIProfile, error) {
	profile, err := AIProfileGet(id, ctx)
	if err != nil {
		return model.AIProfile{}, err
	}
	if profile.Status == model.AIProfileStatusInvalid || profile.Status == model.AIProfileStatusArchived {
		return model.AIProfile{}, fmt.Errorf("profile cannot be activated in current status")
	}
	if err := SettingSetString(model.SettingKeyConfigSourceMode, model.ConfigSourceModeAIProfile); err != nil {
		return model.AIProfile{}, err
	}
	if err := SettingSetString(model.SettingKeyActiveAIProfileID, strconv.Itoa(profile.ID)); err != nil {
		return model.AIProfile{}, err
	}
	if err := db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AIProfile{}).
			Where("status = ? AND id <> ?", model.AIProfileStatusActive, profile.ID).
			Update("status", model.AIProfileStatusReady).Error; err != nil {
			return err
		}
		return tx.Model(&model.AIProfile{}).
			Where("id = ?", profile.ID).
			Update("status", model.AIProfileStatusActive).Error
	}); err != nil {
		return model.AIProfile{}, err
	}
	return AIProfileGet(profile.ID, ctx)
}

func AIProfileCreate(profile model.AIProfile, contentJSON string, ctx context.Context) (model.AIProfile, error) {
	profile.Domain = model.NormalizeAIProfileDomain(profile.Domain)
	if !model.IsValidAIProfileDomain(profile.Domain) {
		return model.AIProfile{}, fmt.Errorf("invalid profile domain")
	}
	profile.Name = strings.TrimSpace(profile.Name)
	if profile.Name == "" {
		return model.AIProfile{}, fmt.Errorf("profile name is required")
	}
	if profile.Version <= 0 {
		profile.Version = 1
	}
	if profile.Status == "" {
		profile.Status = model.AIProfileStatusReady
	}
	if err := db.GetDB().WithContext(ctx).Create(&profile).Error; err != nil {
		return model.AIProfile{}, err
	}
	version := model.AIProfileVersion{ProfileID: profile.ID, Version: profile.Version, ContentJSON: contentJSON, Explanation: profile.Explanation}
	if err := db.GetDB().WithContext(ctx).Create(&version).Error; err != nil {
		return model.AIProfile{}, err
	}
	if err := persistAIProfileTypedPayload(profile, contentJSON, ctx); err != nil {
		_ = markAIProfileMigrationStatus(profile.ID, model.AIProfileMigrationStatusLegacyOnly, err.Error(), ctx)
	}
	return AIProfileGet(profile.ID, ctx)
}

func applyActiveAIProfileConfig(config model.AIAutomationConfig, ctx context.Context) (model.AIAutomationConfig, error) {
	config.BaseURL = config.ManualConfig.BaseURL
	config.APIKey = config.ManualConfig.APIKey
	config.ChannelType = config.ManualConfig.ChannelType
	config.Model = config.ManualConfig.Model
	config.UseLocalDefault = config.ManualConfig.UseLocalDefault
	config.EffectiveConfig = config.ManualConfig
	if config.RequestedConfigSourceMode != model.ConfigSourceModeAIProfile || config.RequestedActiveAIProfileID <= 0 {
		config.ActiveAIProfile = aiAutomationProfileRefByID(config.ActiveAIProfileID, ctx)
		return config, nil
	}
	profile, err := AIProfileGet(config.RequestedActiveAIProfileID, ctx)
	if err != nil {
		return fallbackManualAIAutomationConfig(config, "profile_missing"), nil
	}
	config.RequestedActiveAIProfile = aiAutomationProfileRef(profile)
	if profile.Status == model.AIProfileStatusInvalid || profile.Status == model.AIProfileStatusArchived {
		return fallbackManualAIAutomationConfig(config, "profile_unusable"), nil
	}
	parsed, hasTyped, err := extractAIAutomationConfigFromTypedProfile(profile, ctx)
	if err != nil {
		return fallbackManualAIAutomationConfig(config, "profile_invalid_typed_payload"), nil
	}
	if !hasTyped {
		if len(profile.Versions) == 0 || strings.TrimSpace(profile.Versions[0].ContentJSON) == "" {
			return fallbackManualAIAutomationConfig(config, "profile_missing_content"), nil
		}
		parsed, err = extractAIAutomationConfigFromProfile(profile.Versions[0].ContentJSON)
		if err != nil {
			return fallbackManualAIAutomationConfig(config, "profile_invalid_content"), nil
		}
	}
	if strings.TrimSpace(parsed.BaseURL) != "" {
		if err := validateAIAutomationBaseURL(parsed.BaseURL); err != nil {
			return fallbackManualAIAutomationConfig(config, "profile_forbidden_base_url"), nil
		}
	}
	effective := config.ManualConfig
	if parsed.BaseURL != "" {
		effective.BaseURL = parsed.BaseURL
	}
	if parsed.APIKey != "" {
		effective.APIKey = parsed.APIKey
	}
	if parsed.ChannelType != "" {
		effective.ChannelType = parsed.ChannelType
	}
	if parsed.Model != "" {
		effective.Model = parsed.Model
	}
	if parsed.UseLocalDefault != nil {
		effective.UseLocalDefault = *parsed.UseLocalDefault
	}
	config.BaseURL = effective.BaseURL
	config.APIKey = effective.APIKey
	config.ChannelType = effective.ChannelType
	config.Model = effective.Model
	config.UseLocalDefault = effective.UseLocalDefault
	config.EffectiveConfig = effective
	config.ActiveAIProfileID = profile.ID
	config.ActiveAIProfile = aiAutomationProfileRef(profile)
	return config, nil
}

func fallbackManualAIAutomationConfig(config model.AIAutomationConfig, reason string) model.AIAutomationConfig {
	config.ConfigSourceMode = model.ConfigSourceModeManual
	config.ActiveAIProfileID = 0
	config.ActiveAIProfile = nil
	config.SourceFallbackReason = reason
	config.BaseURL = config.ManualConfig.BaseURL
	config.APIKey = config.ManualConfig.APIKey
	config.ChannelType = config.ManualConfig.ChannelType
	config.Model = config.ManualConfig.Model
	config.UseLocalDefault = config.ManualConfig.UseLocalDefault
	config.EffectiveConfig = config.ManualConfig
	return config
}

func aiAutomationProfileRefByID(id int, ctx context.Context) *model.AIAutomationProfileRef {
	if id <= 0 {
		return nil
	}
	profile, err := AIProfileGet(id, ctx)
	if err != nil {
		return nil
	}
	return aiAutomationProfileRef(profile)
}

func aiAutomationProfileRef(profile model.AIProfile) *model.AIAutomationProfileRef {
	return &model.AIAutomationProfileRef{
		ID:          profile.ID,
		Name:        profile.Name,
		Domain:      profile.Domain,
		Version:     profile.Version,
		Status:      profile.Status,
		Confidence:  profile.Confidence,
		Explanation: profile.Explanation,
		UpdatedAt:   profile.UpdatedAt,
	}
}

func extractAIAutomationConfigFromProfile(contentJSON string) (model.AIAutomationTaskConfig, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(contentJSON), &payload); err != nil {
		return model.AIAutomationTaskConfig{}, err
	}
	configSections := []map[string]any{}
	if runtime, ok := payload["runtime"].(map[string]any); ok {
		configSections = append(configSections, runtime)
	}
	if config, ok := payload["config"].(map[string]any); ok {
		configSections = append(configSections, config)
	}
	for _, section := range configSections {
		parsed := model.AIAutomationTaskConfig{}
		if baseURL, ok := section["base_url"].(string); ok {
			parsed.BaseURL = strings.TrimSpace(baseURL)
		}
		if apiKey, ok := section["api_key"].(string); ok {
			parsed.APIKey = strings.TrimSpace(apiKey)
		}
		if channelType, ok := section["channel_type"].(string); ok {
			parsed.ChannelType = strings.TrimSpace(channelType)
		}
		if channelType, ok := section["ai_channel_type"].(string); ok && parsed.ChannelType == "" {
			parsed.ChannelType = strings.TrimSpace(channelType)
		}
		if modelName, ok := section["model"].(string); ok {
			parsed.Model = strings.TrimSpace(modelName)
		}
		if modelName, ok := section["ai_configured_model"].(string); ok && parsed.Model == "" {
			parsed.Model = strings.TrimSpace(modelName)
		}
		if useLocalDefault, ok := section["use_local_default"].(bool); ok {
			value := useLocalDefault
			parsed.UseLocalDefault = &value
		}
		if useLocalDefault, ok := section["ai_use_local_default"].(bool); ok && parsed.UseLocalDefault == nil {
			value := useLocalDefault
			parsed.UseLocalDefault = &value
		}
		if !isEmptyAITaskConfigSnapshot(parsed) {
			return parsed, nil
		}
	}
	return model.AIAutomationTaskConfig{}, fmt.Errorf("profile does not contain AI automation config")
}
