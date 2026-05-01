package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	tmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
	authropicoutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound/authropic"
	geminioutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound/gemini"
	openaioutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound/openai"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"gorm.io/gorm"
)

const (
	aiTaskExecutorVersion        = "ai-task-executor-v2"
	aiTaskProgressCollectContext = 12
	aiTaskProgressSelectModel    = 30
	aiTaskProgressCallAI         = 58
	aiTaskProgressParseOutput    = 74
	aiTaskProgressGenerate       = 90
	aiTaskProgressSave           = 100
	aiTaskHTTPTimeout            = 45 * time.Second
	aiTaskResultSummaryLimit     = 320
	aiTaskErrorBodyLimit         = 8192
)

var aiTaskStartGroup sync.Map
var aiTaskCancelFuncs sync.Map
var aiTaskCancelWaitTimeout = 5 * time.Second
var aiTaskCancelPollInterval = 25 * time.Millisecond

type aiTaskExecutionState struct {
	Task                  model.AITask
	Config                model.AIAutomationConfig
	PromptTemplates       []model.AIPromptTemplate
	ToolKeys              map[string]struct{}
	ContextPayload        aiTaskContextPayload
	ContextText           string
	ModelName             string
	ModelReason           string
	PromptText            string
	RawOutput             string
	ResultSummary         string
	ResultJSON            string
	ResultPayload         map[string]any
	DomainPayload         map[string]any
	Profile               model.AIProfile
	ProfileContent        string
	ProtectedActionResult aiTaskProtectedActionResult
}

type aiTaskResultCheckpoint struct {
	ResultSummary string         `json:"result_summary"`
	ResultJSON    string         `json:"result_json"`
	ResultPayload map[string]any `json:"result_payload"`
}

type aiTaskProfileCheckpoint struct {
	Profile        model.AIProfile `json:"profile"`
	ProfileContent string          `json:"profile_content"`
}

type aiTaskProtectedActionResult struct {
	SnapshotAuthorized bool
	SnapshotExecuted   bool
	SnapshotPath       string
	SnapshotName       string
	SnapshotReason     string
	ActivateAuthorized bool
	ActivateExecuted   bool
	ActivatedProfileID int
	ActivateReason     string
}

func aiTaskToolKeySet(config model.AIAutomationTaskConfig) map[string]struct{} {
	toolKeys := map[string]struct{}{}
	for _, rawKey := range config.ToolKeys {
		key := model.NormalizeAITaskToolKey(rawKey)
		if !model.IsValidAITaskToolKey(key) {
			continue
		}
		toolKeys[key] = struct{}{}
	}
	if len(toolKeys) == 0 {
		toolKeys[model.AITaskToolKeyChannelInventory] = struct{}{}
		toolKeys[model.AITaskToolKeyGroupTopology] = struct{}{}
		toolKeys[model.AITaskToolKeyModelCatalog] = struct{}{}
		toolKeys[model.AITaskToolKeyLearningState] = struct{}{}
		toolKeys[model.AITaskToolKeyProfileWrite] = struct{}{}
		toolKeys[model.AITaskToolKeySnapshotGuard] = struct{}{}
	}
	return toolKeys
}

func hasAITaskTool(toolKeys map[string]struct{}, key string) bool {
	_, ok := toolKeys[key]
	return ok
}

func sortedAITaskToolKeys(toolKeys map[string]struct{}) []string {
	out := make([]string, 0, len(toolKeys))
	for key := range toolKeys {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func aiTaskProfileWritable(toolKeys map[string]struct{}) bool {
	return hasAITaskTool(toolKeys, model.AITaskToolKeyProfileWrite)
}

func aiTaskProtectedActions(toolKeys map[string]struct{}) []string {
	out := []string{}
	if hasAITaskTool(toolKeys, model.AITaskToolKeySnapshotGuard) {
		out = append(out, model.AITaskToolKeySnapshotGuard)
	}
	if hasAITaskTool(toolKeys, model.AITaskToolKeyProfileActivate) {
		out = append(out, model.AITaskToolKeyProfileActivate)
	}
	return out
}

type aiTaskContextSummary struct {
	Channels        int `json:"channels"`
	Groups          int `json:"groups"`
	LLMs            int `json:"llms"`
	RouteOverrides  int `json:"route_target_overrides"`
	DynamicLearning int `json:"dynamic_learning_states"`
}

type aiTaskContextPayload struct {
	Task                 map[string]any       `json:"task"`
	Runtime              map[string]any       `json:"runtime"`
	Rules                []string             `json:"rules"`
	Channels             []map[string]any     `json:"channels,omitempty"`
	Groups               []map[string]any     `json:"groups,omitempty"`
	Models               []map[string]any     `json:"models,omitempty"`
	RouteTargetOverrides []map[string]any     `json:"route_target_overrides,omitempty"`
	DynamicLearning      []map[string]any     `json:"dynamic_learning,omitempty"`
	Counts               aiTaskContextSummary `json:"counts"`
}

func AITaskStartAsync(taskID int) {
	if taskID <= 0 {
		return
	}
	_, loaded := aiTaskStartGroup.LoadOrStore(taskID, struct{}{})
	if loaded {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	storeAITaskCancel(taskID, cancel)
	go func() {
		defer deleteAITaskCancel(taskID)
		defer aiTaskStartGroup.Delete(taskID)
		defer cancel()
		if err := executeAITask(ctx, taskID); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Warnf("ai task execute failed (task=%d): %v", taskID, err)
		}
	}()
}

func executeAITask(ctx context.Context, taskID int) error {
	task, err := loadAITask(taskID, ctx)
	if err != nil {
		return err
	}
	if task.Status == model.AITaskStatusSucceeded || task.Status == model.AITaskStatusFailed || task.Status == model.AITaskStatusCanceled {
		return nil
	}
	state := &aiTaskExecutionState{Task: task}
	if err := markAITaskRunning(taskID, ctx); err != nil {
		return err
	}

	steps := []struct {
		key      string
		progress int
		run      func(context.Context, *aiTaskExecutionState) error
	}{
		{key: "collect_context", progress: aiTaskProgressCollectContext, run: executeAITaskCollectContext},
		{key: "select_model", progress: aiTaskProgressSelectModel, run: executeAITaskSelectModel},
		{key: "call_ai", progress: aiTaskProgressCallAI, run: executeAITaskCallAI},
		{key: "parse_output", progress: aiTaskProgressParseOutput, run: executeAITaskParseOutput},
		{key: "generate_profile", progress: aiTaskProgressGenerate, run: executeAITaskGenerateProfile},
		{key: "save_result", progress: aiTaskProgressSave, run: executeAITaskSaveResult},
	}
	startIndex := aiTaskResumeStartIndex(task.ResumeState, steps)
	if err := hydrateAITaskExecutionState(ctx, state, task.ResumeState); err != nil {
		return finishAITaskUnrecoverable(taskID, err, ctx)
	}
	if err := resetAITaskStepsForResume(taskID, steps[startIndex].key, ctx); err != nil {
		return finishAITaskUnrecoverable(taskID, err, ctx)
	}

	for idx := startIndex; idx < len(steps); idx++ {
		step := steps[idx]
		if step.key != "save_result" {
			if err := ensureAITaskRunnable(taskID, ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					cleanupAITaskCancellation(taskID, step.key)
					return nil
				}
				return finishAITaskFailure(taskID, err, ctx)
			}
		}
		if err := startAITaskStep(taskID, step.key, ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				cleanupAITaskCancellation(taskID, step.key)
				return nil
			}
			return finishAITaskFailure(taskID, err, ctx)
		}
		if err := step.run(ctx, state); err != nil {
			if errors.Is(err, context.Canceled) {
				cleanupAITaskCancellation(taskID, step.key)
				return nil
			}
			_ = failAITaskStep(taskID, step.key, err.Error(), ctx)
			_ = markAITaskStepsAfter(taskID, step.key, model.AITaskStepStatusSkipped, "blocked by previous failure", ctx)
			return finishAITaskFailure(taskID, err, ctx)
		}
		if step.key != "save_result" {
			if err := ensureAITaskRunnable(taskID, ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					cleanupAITaskCancellation(taskID, step.key)
					return nil
				}
				return finishAITaskFailure(taskID, err, ctx)
			}
		}
		if err := completeAITaskStep(taskID, step.key, step.progress, ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				cleanupAITaskCancellation(taskID, step.key)
				return nil
			}
			return finishAITaskFailure(taskID, err, ctx)
		}
	}
	return nil
}

func executeAITaskCollectContext(ctx context.Context, state *aiTaskExecutionState) error {
	if err := ensureAIAutomationEnabled(ctx); err != nil {
		return err
	}
	config, err := aiAutomationConfigGetRaw(ctx)
	if err != nil {
		return err
	}
	if snapshot, ok := loadAITaskConfigSnapshot(state.Task); ok {
		config = applyAITaskConfigSnapshot(config, snapshot)
	}
	promptTemplates, err := listAITaskPromptTemplates(state.Task, ctx)
	if err != nil {
		return err
	}
	state.Config = config
	if snapshot, ok := loadAITaskConfigSnapshot(state.Task); ok {
		state.ToolKeys = aiTaskToolKeySet(snapshot)
	} else {
		state.ToolKeys = aiTaskToolKeySet(model.AIAutomationTaskConfig{})
	}
	state.PromptTemplates = promptTemplates
	state.ContextPayload, err = buildAITaskContextPayload(state.Task, config, state.ToolKeys, ctx)
	if err != nil {
		return err
	}
	state.ContextText, err = marshalAITaskContextPayload(state.ContextPayload)
	if err != nil {
		return err
	}
	state.PromptText = buildAITaskPromptText(state.Task, promptTemplates)
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", state.Task.ID).Updates(map[string]any{
		"context_payload_json": state.ContextText,
		"prompt_text":          state.PromptText,
		"resume_state":         model.AITaskResumeStateSelectModel,
	}).Error; err != nil {
		return err
	}
	state.Task.ContextPayloadJSON = state.ContextText
	state.Task.PromptText = state.PromptText
	if err := updateAITaskStepCheckpoint(state.Task.ID, "collect_context", map[string]any{"counts": state.ContextPayload.Counts, "tool_keys": sortedAITaskToolKeys(state.ToolKeys)}, model.AITaskResumeStateSelectModel, ctx); err != nil {
		return err
	}
	return updateAITaskStepMessage(state.Task.ID, "collect_context", fmt.Sprintf("assembled non-destructive operator context with tools: %s", strings.Join(sortedAITaskToolKeys(state.ToolKeys), ",")), ctx)
}

func executeAITaskSelectModel(ctx context.Context, state *aiTaskExecutionState) error {
	if strings.TrimSpace(state.Config.Model) != "" {
		state.ModelName = strings.TrimSpace(state.Config.Model)
		state.ModelReason = "configured_default_model"
		if err := persistAITaskSelectedModel(state.Task.ID, state.ModelName, state.ModelReason, ctx); err != nil {
			return err
		}
		return updateAITaskStepMessage(state.Task.ID, "select_model", "using configured AI automation model", ctx)
	}
	result, err := aiAutomationFetchModels(model.AIModelsFetchRequest{
		BaseURL:         state.Config.BaseURL,
		APIKey:          state.Config.APIKey,
		ChannelType:     state.Config.ChannelType,
		UseLocalDefault: &state.Config.UseLocalDefault,
	}, ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(result.SelectedName) == "" {
		return fmt.Errorf("no AI model candidates available")
	}
	state.ModelName = strings.TrimSpace(result.SelectedName)
	state.ModelReason = result.Source
	if err := persistAITaskSelectedModel(state.Task.ID, state.ModelName, state.ModelReason, ctx); err != nil {
		return err
	}
	return updateAITaskStepMessage(state.Task.ID, "select_model", fmt.Sprintf("selected model %s via %s", state.ModelName, result.Source), ctx)
}

func executeAITaskCallAI(ctx context.Context, state *aiTaskExecutionState) error {
	if err := updateAITaskResumeState(state.Task.ID, model.AITaskResumeStateCallAI, ctx); err != nil {
		return err
	}
	rawOutput, err := callAIAutomationModel(ctx, state.Config, state.ModelName, state.PromptText, state.ContextText, state.Task)
	if err != nil {
		return err
	}
	state.RawOutput = strings.TrimSpace(rawOutput)
	if state.RawOutput == "" {
		return fmt.Errorf("AI returned empty output")
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", state.Task.ID).Updates(map[string]any{
		"result_json":  state.RawOutput,
		"resume_state": model.AITaskResumeStateParse,
	}).Error; err != nil {
		return err
	}
	state.Task.ResultJSON = state.RawOutput
	if err := updateAITaskStepCheckpoint(state.Task.ID, "call_ai", map[string]any{"raw_output": state.RawOutput, "model": state.ModelName}, model.AITaskResumeStateParse, ctx); err != nil {
		return err
	}
	return updateAITaskStepMessage(state.Task.ID, "call_ai", fmt.Sprintf("received %d characters from AI", len(state.RawOutput)), ctx)
}

func executeAITaskParseOutput(ctx context.Context, state *aiTaskExecutionState) error {
	state.ResultSummary = deriveAITaskResultSummary(state.RawOutput)
	state.DomainPayload = buildAITaskDomainPayload(state)
	result := map[string]any{
		"source":            "ai_task",
		"task_id":           state.Task.ID,
		"task_type":         state.Task.Type,
		"input_text":        state.Task.InputText,
		"context_scope":     state.Task.ContextScope,
		"model":             state.ModelName,
		"model_reason":      state.ModelReason,
		"raw_output":        state.RawOutput,
		"summary":           state.ResultSummary,
		"tool_keys":         sortedAITaskToolKeys(state.ToolKeys),
		"profile_writable":  aiTaskProfileWritable(state.ToolKeys),
		"protected_actions": aiTaskProtectedActions(state.ToolKeys),
		"config": map[string]any{
			"base_url":          strings.TrimSpace(state.Config.BaseURL),
			"channel_type":      strings.TrimSpace(state.Config.ChannelType),
			"model":             strings.TrimSpace(state.ModelName),
			"use_local_default": state.Config.UseLocalDefault,
		},
		"domain_payload":         state.DomainPayload,
		"non_destructive":        true,
		"tool_execution":         buildAITaskToolExecution(state, false),
		"tool_execution_summary": buildAITaskToolExecutionSummary(state, false),
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	state.ResultPayload = result
	state.ResultJSON = string(raw)
	checkpoint := aiTaskResultCheckpoint{ResultSummary: state.ResultSummary, ResultJSON: state.ResultJSON, ResultPayload: state.ResultPayload}
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", state.Task.ID).Updates(map[string]any{
		"result_summary": state.ResultSummary,
		"result_json":    state.ResultJSON,
		"resume_state":   model.AITaskResumeStateGenerateProfile,
	}).Error; err != nil {
		return err
	}
	state.Task.ResultJSON = state.ResultJSON
	state.Task.ResultSummary = state.ResultSummary
	if err := updateAITaskStepCheckpoint(state.Task.ID, "parse_output", checkpoint, model.AITaskResumeStateGenerateProfile, ctx); err != nil {
		return err
	}
	return updateAITaskStepMessage(state.Task.ID, "parse_output", "AI output parsed into non-destructive result payload", ctx)
}

func executeAITaskGenerateProfile(ctx context.Context, state *aiTaskExecutionState) error {
	if !aiTaskProfileWritable(state.ToolKeys) {
		state.Profile = model.AIProfile{}
		state.ProfileContent = ""
		if err := updateAITaskResumeState(state.Task.ID, model.AITaskResumeStateSaveResult, ctx); err != nil {
			return err
		}
		return updateAITaskStepMessage(state.Task.ID, "generate_profile", "profile generation skipped because profile_write tool is disabled", ctx)
	}
	domain := aiProfileDomainFromTaskType(state.Task.Type)
	profileName := fmt.Sprintf("AI %s #%d", strings.ReplaceAll(domain, "_", " "), state.Task.ID)
	state.Profile = model.AIProfile{Domain: domain, Name: profileName, Version: 1, Status: model.AIProfileStatusReady, Confidence: aiTaskConfidenceFromOutput(state.RawOutput), Explanation: state.ResultSummary, SourceTaskID: &state.Task.ID}
	content := map[string]any{
		"source":            "ai_task",
		"task_id":           state.Task.ID,
		"task_type":         state.Task.Type,
		"domain":            domain,
		"input_text":        state.Task.InputText,
		"context_scope":     state.Task.ContextScope,
		"model":             state.ModelName,
		"model_reason":      state.ModelReason,
		"summary":           state.ResultSummary,
		"raw_output":        state.RawOutput,
		"tool_keys":         sortedAITaskToolKeys(state.ToolKeys),
		"protected_actions": aiTaskProtectedActions(state.ToolKeys),
		"config": map[string]any{
			"base_url":          strings.TrimSpace(state.Config.BaseURL),
			"channel_type":      strings.TrimSpace(state.Config.ChannelType),
			"model":             strings.TrimSpace(state.ModelName),
			"use_local_default": state.Config.UseLocalDefault,
		},
		"domain_payload": state.DomainPayload,
		"runtime": map[string]any{
			"config_source_mode":   state.Config.ConfigSourceMode,
			"active_ai_profile_id": state.Config.ActiveAIProfileID,
			"ai_channel_type":      strings.TrimSpace(state.Config.ChannelType),
			"ai_configured_model":  strings.TrimSpace(state.ModelName),
			"ai_use_local_default": state.Config.UseLocalDefault,
		},
		"non_destructive":        true,
		"tool_execution":         buildAITaskToolExecution(state, false),
		"tool_execution_summary": buildAITaskToolExecutionSummary(state, false),
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return err
	}
	state.ProfileContent = string(contentJSON)
	checkpoint := aiTaskProfileCheckpoint{Profile: state.Profile, ProfileContent: state.ProfileContent}
	if err := updateAITaskStepCheckpoint(state.Task.ID, "generate_profile", checkpoint, model.AITaskResumeStateSaveResult, ctx); err != nil {
		return err
	}
	if err := updateAITaskResumeState(state.Task.ID, model.AITaskResumeStateSaveResult, ctx); err != nil {
		return err
	}
	return updateAITaskStepMessage(state.Task.ID, "generate_profile", fmt.Sprintf("prepared %s profile draft", domain), ctx)
}

func executeAITaskSaveResult(ctx context.Context, state *aiTaskExecutionState) error {
	resultProfileID := any(nil)
	profileSaved := false
	if aiTaskProfileWritable(state.ToolKeys) {
		profile, err := AIProfileCreate(state.Profile, state.ProfileContent, ctx)
		if err != nil {
			return err
		}
		state.Profile = profile
		resultProfileID = profile.ID
		profileSaved = true
	}
	state.ProtectedActionResult = executeAITaskProtectedActions(ctx, state, profileSaved)
	if state.ResultPayload != nil {
		state.ResultPayload["tool_execution"] = buildAITaskToolExecution(state, profileSaved)
		state.ResultPayload["tool_execution_summary"] = buildAITaskToolExecutionSummary(state, profileSaved)
		raw, err := json.Marshal(state.ResultPayload)
		if err != nil {
			return err
		}
		state.ResultJSON = string(raw)
	}
	now := time.Now()
	return db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", state.Task.ID).Updates(map[string]any{
		"status":            model.AITaskStatusSucceeded,
		"progress":          aiTaskProgressSave,
		"result_profile_id": resultProfileID,
		"result_summary":    state.ResultSummary,
		"result_json":       state.ResultJSON,
		"error_message":     "",
		"resume_state":      model.AITaskResumeStateCompleted,
		"finished_at":       &now,
	}).Error
}

func persistAITaskSelectedModel(taskID int, modelName, reason string, ctx context.Context) error {
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", taskID).Updates(map[string]any{
		"selected_model": strings.TrimSpace(modelName),
		"model_reason":   strings.TrimSpace(reason),
		"resume_state":   model.AITaskResumeStateCallAI,
	}).Error; err != nil {
		return err
	}
	return updateAITaskStepCheckpoint(taskID, "select_model", map[string]any{"model": strings.TrimSpace(modelName), "reason": strings.TrimSpace(reason)}, model.AITaskResumeStateCallAI, ctx)
}

func executeAITaskProtectedActions(ctx context.Context, state *aiTaskExecutionState, profileSaved bool) aiTaskProtectedActionResult {
	result := aiTaskProtectedActionResult{
		SnapshotAuthorized: hasAITaskTool(state.ToolKeys, model.AITaskToolKeySnapshotGuard),
		ActivateAuthorized: hasAITaskTool(state.ToolKeys, model.AITaskToolKeyProfileActivate),
	}

	if result.SnapshotAuthorized {
		path, name, err := createAITaskSnapshot(ctx, state.Task.ID)
		if err != nil {
			result.SnapshotReason = err.Error()
		} else {
			result.SnapshotExecuted = true
			result.SnapshotPath = path
			result.SnapshotName = name
			result.SnapshotReason = "snapshot_saved"
		}
	}

	if result.ActivateAuthorized {
		if !profileSaved || state.Profile.ID <= 0 {
			result.ActivateReason = "profile_not_saved"
		} else if _, err := AIProfileActivate(state.Profile.ID, ctx); err != nil {
			result.ActivateReason = err.Error()
		} else {
			result.ActivateExecuted = true
			result.ActivatedProfileID = state.Profile.ID
			result.ActivateReason = "profile_activated"
		}
	}

	return result
}

func createAITaskSnapshot(ctx context.Context, taskID int) (string, string, error) {
	dump, err := DBExportAll(ctx, true, true, true)
	if err != nil {
		return "", "", err
	}
	dump.Manifest.ExportSource = "octopus-ai-task"
	payload, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return "", "", err
	}
	dir, err := aiTaskSnapshotDir()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dir, importSnapshotDirPerm); err != nil {
		return "", "", err
	}
	if err := os.Chmod(dir, importSnapshotDirPerm); err != nil {
		return "", "", err
	}
	name := buildAITaskSnapshotFilename(taskID, time.Now().UTC())
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, payload, importSnapshotFilePerm); err != nil {
		return "", "", err
	}
	if err := os.Chmod(path, importSnapshotFilePerm); err != nil {
		return "", "", err
	}
	return path, name, nil
}

func aiTaskSnapshotDir() (string, error) {
	baseDir, err := importSnapshotDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "ai-task-snapshots"), nil
}

func buildAITaskSnapshotFilename(taskID int, ts time.Time) string {
	utc := ts.UTC()
	return fmt.Sprintf("ai-task-%d-%s-%09d.json", taskID, utc.Format("20060102T150405Z"), utc.Nanosecond())
}

func loadAITaskConfigSnapshot(task model.AITask) (model.AIAutomationTaskConfig, bool) {
	if strings.TrimSpace(task.ConfigSnapshotJSON) == "" {
		return model.AIAutomationTaskConfig{}, false
	}
	var config model.AIAutomationTaskConfig
	if err := json.Unmarshal([]byte(task.ConfigSnapshotJSON), &config); err != nil {
		return model.AIAutomationTaskConfig{}, false
	}
	config = normalizeAITaskConfigSnapshot(config)
	if isEmptyAITaskConfigSnapshot(config) {
		return model.AIAutomationTaskConfig{}, false
	}
	return config, true
}

func storeAITaskCancel(taskID int, cancel context.CancelFunc) {
	if taskID <= 0 || cancel == nil {
		return
	}
	aiTaskCancelFuncs.Store(taskID, cancel)
}

func deleteAITaskCancel(taskID int) {
	if taskID <= 0 {
		return
	}
	aiTaskCancelFuncs.Delete(taskID)
}

func cancelAITaskExecution(taskID int) {
	if taskID <= 0 {
		return
	}
	raw, ok := aiTaskCancelFuncs.Load(taskID)
	if !ok {
		return
	}
	cancel, ok := raw.(context.CancelFunc)
	if !ok || cancel == nil {
		return
	}
	cancel()
}

func CancelAllAITasks() error {
	taskIDs := map[int]struct{}{}
	aiTaskCancelFuncs.Range(func(key, value any) bool {
		if id, ok := key.(int); ok && id > 0 {
			taskIDs[id] = struct{}{}
		}
		cancel, ok := value.(context.CancelFunc)
		if ok && cancel != nil {
			cancel()
		}
		return true
	})
	aiTaskStartGroup.Range(func(key, value any) bool {
		if id, ok := key.(int); ok && id > 0 {
			taskIDs[id] = struct{}{}
		}
		return true
	})
	timedOut := make([]string, 0)
	for taskID := range taskIDs {
		if !waitForAITaskStop(taskID, aiTaskCancelWaitTimeout) {
			timedOut = append(timedOut, strconv.Itoa(taskID))
		}
	}
	if len(timedOut) > 0 {
		sort.Strings(timedOut)
		return fmt.Errorf("timed out waiting for AI tasks to stop: %s", strings.Join(timedOut, ","))
	}
	return nil
}

func waitForAITaskStop(taskID int, timeout time.Duration) bool {
	if taskID <= 0 {
		return true
	}
	deadline := time.Now().Add(timeout)
	for {
		if _, ok := aiTaskStartGroup.Load(taskID); !ok {
			return true
		}
		if timeout > 0 && time.Now().After(deadline) {
			return false
		}
		time.Sleep(aiTaskCancelPollInterval)
	}
}

func cleanupAITaskCancellation(taskID int, stepKey string) {
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = failAITaskStep(taskID, stepKey, "task canceled", writeCtx)
	_ = markAITaskStepsAfter(taskID, stepKey, model.AITaskStepStatusSkipped, "task canceled", writeCtx)
}

func markAITaskCanceledSteps(taskID int) {
	if taskID <= 0 {
		return
	}
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	now := time.Now()
	_ = db.GetDB().WithContext(writeCtx).Model(&model.AITaskStep{}).Where("task_id = ? AND status = ?", taskID, model.AITaskStepStatusRunning).Updates(map[string]any{
		"status":      model.AITaskStepStatusFailed,
		"message":     "task canceled",
		"finished_at": &now,
	}).Error
	_ = db.GetDB().WithContext(writeCtx).Model(&model.AITaskStep{}).Where("task_id = ? AND status = ?", taskID, model.AITaskStepStatusPending).Updates(map[string]any{
		"status":      model.AITaskStepStatusSkipped,
		"message":     "task canceled",
		"finished_at": &now,
	}).Error
}

func applyAITaskConfigSnapshot(config model.AIAutomationConfig, snapshot model.AIAutomationTaskConfig) model.AIAutomationConfig {
	if strings.TrimSpace(snapshot.BaseURL) != "" {
		config.BaseURL = strings.TrimSpace(snapshot.BaseURL)
	}
	if strings.TrimSpace(snapshot.APIKey) != "" {
		config.APIKey = strings.TrimSpace(snapshot.APIKey)
	}
	if strings.TrimSpace(snapshot.ChannelType) != "" {
		config.ChannelType = strings.TrimSpace(snapshot.ChannelType)
	}
	if strings.TrimSpace(snapshot.Model) != "" {
		config.Model = strings.TrimSpace(snapshot.Model)
	}
	if snapshot.UseLocalDefault != nil {
		config.UseLocalDefault = *snapshot.UseLocalDefault
	}
	return config
}

func listAITaskPromptTemplates(task model.AITask, ctx context.Context) ([]model.AIPromptTemplate, error) {
	if strings.TrimSpace(task.PromptTemplateIDs) == "" {
		return []model.AIPromptTemplate{}, nil
	}
	rows, err := AIPromptTemplateList(ctx)
	if err != nil {
		return nil, err
	}
	idOrder := strings.Split(task.PromptTemplateIDs, ",")
	selected := make([]model.AIPromptTemplate, 0, len(idOrder))
	for _, rawID := range idOrder {
		id, err := strconv.Atoi(strings.TrimSpace(rawID))
		if err != nil || id <= 0 {
			continue
		}
		for _, row := range rows {
			if row.ID == id && row.Enabled {
				selected = append(selected, row)
				break
			}
		}
	}
	return selected, nil
}

func buildAITaskContextText(task model.AITask, config model.AIAutomationConfig, toolKeys map[string]struct{}, ctx context.Context) (string, error) {
	payload, err := buildAITaskContextPayload(task, config, toolKeys, ctx)
	if err != nil {
		return "", err
	}
	return marshalAITaskContextPayload(payload)
}

func marshalAITaskContextPayload(payload aiTaskContextPayload) (string, error) {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func buildAITaskContextPayload(task model.AITask, config model.AIAutomationConfig, toolKeys map[string]struct{}, ctx context.Context) (aiTaskContextPayload, error) {
	payload := aiTaskContextPayload{
		Task: map[string]any{
			"id":               task.ID,
			"type":             task.Type,
			"input_text":       strings.TrimSpace(task.InputText),
			"context_scope":    strings.TrimSpace(task.ContextScope),
			"custom_prompt":    strings.TrimSpace(task.CustomPrompt),
			"prompt_templates": task.PromptTemplateIDs,
			"tool_keys":        sortedAITaskToolKeys(toolKeys),
		},
		Runtime: map[string]any{
			"config_source_mode":               config.ConfigSourceMode,
			"active_ai_profile_id":             config.ActiveAIProfileID,
			"dynamic_routing_learning_enabled": config.DynamicRoutingLearningEnabled,
			"ai_channel_type":                  strings.TrimSpace(config.ChannelType),
			"ai_use_local_default":             config.UseLocalDefault,
			"ai_configured_model":              strings.TrimSpace(config.Model),
		},
		Rules: []string{
			"non_destructive_profile_only",
			"never_overwrite_manual_channels_or_groups",
			"never_mutate_group_items_llm_infos_or_route_target_overrides",
			"dynamic_routing_learning_is_runtime_only",
			"manual_and_ai_profiles_must_coexist",
		},
	}
	payload.Runtime["profile_write_enabled"] = aiTaskProfileWritable(toolKeys)
	payload.Runtime["protected_actions"] = aiTaskProtectedActions(toolKeys)

	if hasAITaskTool(toolKeys, model.AITaskToolKeyChannelInventory) {
		channels, err := ChannelList(ctx)
		if err != nil {
			return aiTaskContextPayload{}, err
		}
		payload.Channels = summarizeAITaskChannels(channels)
	}

	if hasAITaskTool(toolKeys, model.AITaskToolKeyGroupTopology) {
		groups, err := GroupList(ctx)
		if err != nil {
			return aiTaskContextPayload{}, err
		}
		payload.Groups = summarizeAITaskGroups(groups)
	}

	if hasAITaskTool(toolKeys, model.AITaskToolKeyPriceCatalog) || hasAITaskTool(toolKeys, model.AITaskToolKeyModelCatalog) {
		llms, err := LLMList(ctx)
		if err != nil {
			return aiTaskContextPayload{}, err
		}
		payload.Models = summarizeAITaskLLMs(llms)
	}

	if hasAITaskTool(toolKeys, model.AITaskToolKeyRouteOverrides) {
		overrides, err := RouteTargetOverrideList(ctx)
		if err != nil {
			return aiTaskContextPayload{}, err
		}
		payload.RouteTargetOverrides = summarizeAITaskRouteTargetOverrides(overrides)
	}

	if hasAITaskTool(toolKeys, model.AITaskToolKeyLearningState) {
		learning, err := DynamicRouteLearningList(ctx)
		if err != nil {
			return aiTaskContextPayload{}, err
		}
		payload.DynamicLearning = summarizeAITaskDynamicLearning(learning.States)
	}

	payload.Counts = aiTaskContextSummary{
		Channels:        len(payload.Channels),
		Groups:          len(payload.Groups),
		LLMs:            len(payload.Models),
		RouteOverrides:  len(payload.RouteTargetOverrides),
		DynamicLearning: len(payload.DynamicLearning),
	}
	return payload, nil
}

func buildAITaskToolExecution(state *aiTaskExecutionState, profileSaved bool) map[string]any {
	contextCounts := map[string]int{
		"channels":               state.ContextPayload.Counts.Channels,
		"groups":                 state.ContextPayload.Counts.Groups,
		"models":                 state.ContextPayload.Counts.LLMs,
		"route_target_overrides": state.ContextPayload.Counts.RouteOverrides,
		"dynamic_learning":       state.ContextPayload.Counts.DynamicLearning,
	}
	readTools := []map[string]any{}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyChannelInventory) {
		readTools = append(readTools, map[string]any{"key": model.AITaskToolKeyChannelInventory, "target": "channels", "count": state.ContextPayload.Counts.Channels, "executed": true})
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyGroupTopology) {
		readTools = append(readTools, map[string]any{"key": model.AITaskToolKeyGroupTopology, "target": "groups", "count": state.ContextPayload.Counts.Groups, "executed": true})
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyPriceCatalog) {
		readTools = append(readTools, map[string]any{"key": model.AITaskToolKeyPriceCatalog, "target": "models", "count": state.ContextPayload.Counts.LLMs, "executed": true})
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyModelCatalog) {
		readTools = append(readTools, map[string]any{"key": model.AITaskToolKeyModelCatalog, "target": "models", "count": state.ContextPayload.Counts.LLMs, "executed": true})
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyRouteOverrides) {
		readTools = append(readTools, map[string]any{"key": model.AITaskToolKeyRouteOverrides, "target": "route_target_overrides", "count": state.ContextPayload.Counts.RouteOverrides, "executed": true})
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyLearningState) {
		readTools = append(readTools, map[string]any{"key": model.AITaskToolKeyLearningState, "target": "dynamic_learning", "count": state.ContextPayload.Counts.DynamicLearning, "executed": true})
	}
	protectedActions := []map[string]any{}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeySnapshotGuard) {
		action := map[string]any{"key": model.AITaskToolKeySnapshotGuard, "authorized": true, "executed": state.ProtectedActionResult.SnapshotExecuted}
		if strings.TrimSpace(state.ProtectedActionResult.SnapshotReason) != "" {
			action["reason"] = state.ProtectedActionResult.SnapshotReason
		}
		if strings.TrimSpace(state.ProtectedActionResult.SnapshotPath) != "" {
			action["snapshot_path"] = state.ProtectedActionResult.SnapshotPath
		}
		if strings.TrimSpace(state.ProtectedActionResult.SnapshotName) != "" {
			action["snapshot_name"] = state.ProtectedActionResult.SnapshotName
		}
		protectedActions = append(protectedActions, action)
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyProfileActivate) {
		action := map[string]any{"key": model.AITaskToolKeyProfileActivate, "authorized": true, "executed": state.ProtectedActionResult.ActivateExecuted}
		if strings.TrimSpace(state.ProtectedActionResult.ActivateReason) != "" {
			action["reason"] = state.ProtectedActionResult.ActivateReason
		}
		if state.ProtectedActionResult.ActivatedProfileID > 0 {
			action["profile_id"] = state.ProtectedActionResult.ActivatedProfileID
		}
		protectedActions = append(protectedActions, action)
	}
	return map[string]any{
		"non_destructive":   true,
		"enabled_tool_keys": sortedAITaskToolKeys(state.ToolKeys),
		"read_tools":        readTools,
		"context_counts":    contextCounts,
		"writes":            []map[string]any{buildAITaskProfileWriteAction(state, profileSaved)},
		"protected_actions": protectedActions,
		"notes": []string{
			"manual configuration remains untouched",
			"dynamic routing learning remains runtime only",
		},
	}
}

func buildAITaskProfileWriteAction(state *aiTaskExecutionState, profileSaved bool) map[string]any {
	prepared := strings.TrimSpace(state.Profile.Name) != ""
	if !aiTaskProfileWritable(state.ToolKeys) {
		return map[string]any{
			"key":        model.AITaskToolKeyProfileWrite,
			"authorized": false,
			"prepared":   false,
			"executed":   false,
			"saved":      false,
			"reason":     "tool_disabled",
		}
	}
	action := map[string]any{
		"key":             model.AITaskToolKeyProfileWrite,
		"authorized":      true,
		"prepared":        prepared,
		"executed":        profileSaved,
		"saved":           profileSaved,
		"profile_domain":  state.Profile.Domain,
		"profile_version": state.Profile.Version,
	}
	if profileSaved {
		action["profile_id"] = state.Profile.ID
		action["reason"] = "profile_saved"
	} else if prepared {
		action["reason"] = "profile_draft_prepared"
	} else {
		action["reason"] = "not_executed_yet"
	}
	return action
}

func buildAITaskToolExecutionSummary(state *aiTaskExecutionState, profileSaved bool) map[string]any {
	readToolCount := 0
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyChannelInventory) {
		readToolCount++
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyGroupTopology) {
		readToolCount++
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyPriceCatalog) {
		readToolCount++
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyModelCatalog) {
		readToolCount++
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyRouteOverrides) {
		readToolCount++
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyLearningState) {
		readToolCount++
	}
	protectedAuthorizedCount := 0
	protectedExecutedCount := 0
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeySnapshotGuard) {
		protectedAuthorizedCount++
		if state.ProtectedActionResult.SnapshotExecuted {
			protectedExecutedCount++
		}
	}
	if hasAITaskTool(state.ToolKeys, model.AITaskToolKeyProfileActivate) {
		protectedAuthorizedCount++
		if state.ProtectedActionResult.ActivateExecuted {
			protectedExecutedCount++
		}
	}
	return map[string]any{
		"read_tool_count":               readToolCount,
		"context_object_count":          state.ContextPayload.Counts.Channels + state.ContextPayload.Counts.Groups + state.ContextPayload.Counts.LLMs + state.ContextPayload.Counts.RouteOverrides + state.ContextPayload.Counts.DynamicLearning,
		"profile_write_authorized":      aiTaskProfileWritable(state.ToolKeys),
		"profile_write_prepared":        strings.TrimSpace(state.Profile.Name) != "",
		"profile_write_executed":        profileSaved,
		"protected_actions_authorized":  protectedAuthorizedCount,
		"protected_actions_executed":    protectedExecutedCount,
		"manual_config_mutated":         false,
		"runtime_only_dynamic_learning": true,
	}
}

func summarizeAITaskChannels(channels []model.Channel) []map[string]any {
	sort.SliceStable(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })
	out := make([]map[string]any, 0, len(channels))
	for _, channel := range channels {
		baseURLs := make([]map[string]any, 0, len(channel.BaseUrls))
		for _, baseURL := range channel.BaseUrls {
			baseURLs = append(baseURLs, map[string]any{"url": strings.TrimSpace(baseURL.URL), "delay_ms": baseURL.Delay})
		}
		keys := make([]map[string]any, 0, len(channel.Keys))
		for _, key := range channel.Keys {
			keys = append(keys, map[string]any{
				"id":             key.ID,
				"enabled":        key.Enabled,
				"source_type":    model.EffectiveChannelKeySourceType(key.SourceType),
				"allowed_models": splitModelNames(key.AllowedModels),
				"remark":         strings.TrimSpace(key.Remark),
			})
		}
		out = append(out, map[string]any{
			"id":                  channel.ID,
			"name":                strings.TrimSpace(channel.Name),
			"type":                int(channel.Type),
			"enabled":             channel.Enabled,
			"key_management_mode": channel.KeyManagementMode,
			"key_routing_policy":  channel.KeyRoutingPolicy,
			"models":              splitModelNames(channel.Model, channel.CustomModel),
			"base_urls":           baseURLs,
			"keys":                keys,
			"match_regex":         strings.TrimSpace(loString(channel.MatchRegex)),
		})
	}
	return out
}

func summarizeAITaskGroups(groups []model.Group) []map[string]any {
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })
	out := make([]map[string]any, 0, len(groups))
	for _, group := range groups {
		items := make([]map[string]any, 0, len(group.Items))
		sort.SliceStable(group.Items, func(i, j int) bool {
			if group.Items[i].Priority == group.Items[j].Priority {
				return group.Items[i].ID < group.Items[j].ID
			}
			return group.Items[i].Priority < group.Items[j].Priority
		})
		for _, item := range group.Items {
			items = append(items, map[string]any{
				"id":         item.ID,
				"channel_id": item.ChannelID,
				"model_name": item.ModelName,
				"priority":   item.Priority,
				"weight":     item.Weight,
			})
		}
		out = append(out, map[string]any{
			"id":                  group.ID,
			"name":                strings.TrimSpace(group.Name),
			"mode":                group.Mode,
			"match_regex":         strings.TrimSpace(group.MatchRegex),
			"first_token_timeout": group.FirstTokenTimeOut,
			"session_keep_time":   group.SessionKeepTime,
			"retry_rounds":        group.RetryRounds,
			"retry_delay_ms":      group.RetryDelayMs,
			"failover_window_sec": group.FailoverWindowSec,
			"race_after_fails":    group.RaceAfterFails,
			"race_concurrency":    group.RaceConcurrency,
			"items":               items,
		})
	}
	return out
}

func summarizeAITaskLLMs(llms []model.LLMInfo) []map[string]any {
	sort.SliceStable(llms, func(i, j int) bool { return llms[i].Name < llms[j].Name })
	out := make([]map[string]any, 0, len(llms))
	for _, info := range llms {
		out = append(out, map[string]any{
			"name":                    info.Name,
			"canonical_name":          info.CanonicalName,
			"billing_mode":            info.BillingMode,
			"probe_policy":            info.ProbePolicy,
			"probe_interval_seconds":  info.ProbeIntervalSeconds,
			"probe_concurrency_limit": info.ProbeConcurrencyLimit,
			"input_price":             info.Input,
			"output_price":            info.Output,
		})
	}
	return out
}

func summarizeAITaskRouteTargetOverrides(rows []model.RouteTargetOverride) []map[string]any {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].ChannelID != rows[j].ChannelID {
			return rows[i].ChannelID < rows[j].ChannelID
		}
		if rows[i].ChannelKeyID != rows[j].ChannelKeyID {
			return rows[i].ChannelKeyID < rows[j].ChannelKeyID
		}
		return rows[i].ModelName < rows[j].ModelName
	})
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"channel_id":              row.ChannelID,
			"channel_key_id":          row.ChannelKeyID,
			"model_name":              row.ModelName,
			"billing_mode":            row.BillingMode,
			"probe_policy":            row.ProbePolicy,
			"probe_interval_seconds":  row.ProbeIntervalSeconds,
			"probe_concurrency_limit": row.ProbeConcurrencyLimit,
		})
	}
	return out
}

func summarizeAITaskDynamicLearning(states []model.DynamicRouteLearningState) []map[string]any {
	sort.SliceStable(states, func(i, j int) bool {
		if states[i].Score != states[j].Score {
			return states[i].Score > states[j].Score
		}
		return states[i].ID < states[j].ID
	})
	out := make([]map[string]any, 0, len(states))
	for _, state := range states {
		out = append(out, map[string]any{
			"channel_id":        state.ChannelID,
			"channel_key_id":    state.ChannelKeyID,
			"model_name":        state.ModelName,
			"success_count":     state.SuccessCount,
			"failure_count":     state.FailureCount,
			"fallback_count":    state.FallbackCount,
			"race_winner_count": state.RaceWinnerCount,
			"latency_ms_ewma":   state.LatencyMsEWMA,
			"score":             state.Score,
			"confidence":        state.Confidence,
			"last_sample_at":    state.LastSampleAt,
		})
	}
	return out
}

func loString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func callAIAutomationModel(ctx context.Context, config model.AIAutomationConfig, modelName, promptText, contextText string, task model.AITask) (string, error) {
	baseURL, err := resolveAIAutomationBaseURL(config.BaseURL, config.UseLocalDefault)
	if err != nil {
		return "", err
	}
	channelType := strings.ToLower(strings.TrimSpace(config.ChannelType))
	if channelType == "" {
		channelType = model.DefaultAIAutomationChannelType
	}
	requestCtx, cancel := context.WithTimeout(ctx, aiTaskHTTPTimeout)
	defer cancel()

	llmReq := &tmodel.InternalLLMRequest{
		Model: modelName,
		Messages: []tmodel.Message{
			{Role: "system", Content: tmodel.MessageContent{Content: strPtr(promptText)}},
			{Role: "user", Content: tmodel.MessageContent{Content: strPtr("Context summary:\n" + contextText)}},
		},
		Temperature: strFloatPtr(0.2),
		Metadata:    map[string]string{"task_id": fmt.Sprintf("%d", task.ID), "task_type": task.Type},
	}
	httpReq, err := buildAIAutomationOutboundRequest(requestCtx, channelType, llmReq, baseURL, strings.TrimSpace(config.APIKey))
	if err != nil {
		return "", err
	}
	httpReq.Close = true
	httpClient, err := newHealthHTTPClientNoProxy()
	if err != nil {
		return "", err
	}
	httpClient.Timeout = aiTaskHTTPTimeout
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(httpResp.Body, aiTaskErrorBodyLimit))
		return "", fmt.Errorf("AI endpoint returned status %d: %s", httpResp.StatusCode, strings.TrimSpace(string(body)))
	}
	internalResp, err := parseAIAutomationOutboundResponse(requestCtx, channelType, httpResp)
	if err != nil {
		return "", err
	}
	content := extractAIAutomationText(internalResp)
	if content == "" {
		return "", fmt.Errorf("AI endpoint returned empty message content")
	}
	return content, nil
}

func buildAIAutomationOutboundRequest(ctx context.Context, channelType string, req *tmodel.InternalLLMRequest, baseURL, apiKey string) (*http.Request, error) {
	var outbound tmodel.Outbound
	switch channelType {
	case "anthropic":
		outbound = &authropicoutbound.MessageOutbound{}
	case "gemini":
		outbound = &geminioutbound.MessagesOutbound{}
	case "openai", "openai-compatible", "":
		outbound = &openaioutbound.ChatOutbound{}
	default:
		return nil, fmt.Errorf("unsupported AI automation channel type: %s", channelType)
	}
	httpReq, err := outbound.TransformRequest(ctx, req, baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	if channelType == "gemini" {
		applyAIAutomationGeminiCompatKey(httpReq, apiKey)
	}
	return httpReq, nil
}

func parseAIAutomationOutboundResponse(ctx context.Context, channelType string, resp *http.Response) (*tmodel.InternalLLMResponse, error) {
	var outbound tmodel.Outbound
	switch channelType {
	case "anthropic":
		outbound = &authropicoutbound.MessageOutbound{}
	case "gemini":
		outbound = &geminioutbound.MessagesOutbound{}
	default:
		outbound = &openaioutbound.ChatOutbound{}
	}
	return outbound.TransformResponse(ctx, resp)
}

func applyAIAutomationGeminiCompatKey(req *http.Request, apiKey string) {
	if req == nil || strings.TrimSpace(apiKey) == "" || req.URL == nil {
		return
	}
	queryValues, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return
	}
	if strings.TrimSpace(queryValues.Get("key")) == "" {
		queryValues.Set("key", strings.TrimSpace(apiKey))
		req.URL.RawQuery = queryValues.Encode()
	}
}

func extractAIAutomationText(resp *tmodel.InternalLLMResponse) string {
	if resp == nil {
		return ""
	}
	parts := []string{}
	for _, choice := range resp.Choices {
		if choice.Message != nil {
			if text := extractMessageContent(choice.Message); text != "" {
				parts = append(parts, text)
			}
		}
		if choice.Delta != nil {
			if text := extractMessageContent(choice.Delta); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func extractMessageContent(message *tmodel.Message) string {
	if message == nil {
		return ""
	}
	parts := []string{}
	if message.Content.Content != nil {
		parts = append(parts, strings.TrimSpace(*message.Content.Content))
	}
	for _, item := range message.Content.MultipleContent {
		if item.Type == "text" && item.Text != nil {
			parts = append(parts, strings.TrimSpace(*item.Text))
		}
	}
	if reasoning := strings.TrimSpace(message.GetReasoningContent()); reasoning != "" {
		parts = append(parts, reasoning)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func strFloatPtr(v float64) *float64 {
	return &v
}

func deriveAITaskResultSummary(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "AI task completed without a textual summary."
	}
	lines := strings.FieldsFunc(trimmed, func(r rune) bool { return r == '\n' || r == '\r' })
	for _, line := range lines {
		clean := strings.TrimSpace(strings.TrimLeft(line, "-1234567890. "))
		if clean != "" {
			return trimTextWithSuffix(clean, aiTaskResultSummaryLimit, "...")
		}
	}
	return trimTextWithSuffix(trimmed, aiTaskResultSummaryLimit, "...")
}

func aiTaskConfidenceFromOutput(raw string) float64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0.35
	}
	if len(trimmed) > 2000 {
		return 0.92
	}
	if len(trimmed) > 800 {
		return 0.84
	}
	if len(trimmed) > 300 {
		return 0.72
	}
	return 0.6
}

func aiProfileDomainFromTaskType(taskType string) string {
	switch model.NormalizeAITaskType(taskType) {
	case model.AIAutomationTaskTypeChannelRecognition:
		return model.AIProfileDomainChannelRecognition
	case model.AIAutomationTaskTypePriceRecognition:
		return model.AIProfileDomainPriceRecognition
	case model.AIAutomationTaskTypeModelClassification:
		return model.AIProfileDomainModelClassification
	case model.AIAutomationTaskTypeConfigHealthCheck:
		return model.AIProfileDomainConfigHealthCheck
	case model.AIAutomationTaskTypeDynamicRoutingDigest:
		return model.AIProfileDomainDynamicRoutingDigest
	case model.AIAutomationTaskTypeNaturalLanguage, model.AIAutomationTaskTypeGroupSuggestion:
		fallthrough
	default:
		return model.AIProfileDomainGrouping
	}
}
func buildAITaskDomainPayload(state *aiTaskExecutionState) map[string]any {
	domain := aiProfileDomainFromTaskType(state.Task.Type)
	payload := map[string]any{
		"summary": state.ResultSummary,
		"findings": []any{
			map[string]any{"type": "ai_output", "description": trimTextWithSuffix(strings.TrimSpace(state.RawOutput), aiTaskResultSummaryLimit, "...")},
		},
		"recommendations": []any{
			map[string]any{"type": "manual_review", "description": "Review this AI-generated plan before activation."},
		},
		"risks": []any{
			map[string]any{"type": "manual_config_protection", "description": "Generated output must not overwrite manual configuration tables."},
		},
		"typed_config": map[string]any{
			"base_url":          strings.TrimSpace(state.Config.BaseURL),
			"channel_type":      strings.TrimSpace(state.Config.ChannelType),
			"model":             strings.TrimSpace(state.ModelName),
			"use_local_default": state.Config.UseLocalDefault,
		},
	}
	switch domain {
	case model.AIProfileDomainGrouping:
		payload["grouping_suggestions"] = buildAITaskGroupingSuggestions(state)
		payload["candidate_channel_model_mappings"] = buildAITaskGroupingMappings(state)
		payload["conflicts"] = buildAITaskGroupingConflicts(state)
	case model.AIProfileDomainChannelRecognition:
		payload["channel_type"] = inferAITaskPrimaryChannelType(state)
		payload["source_type"] = "channel_inventory"
		payload["model_coverage"] = buildAITaskChannelCoverage(state)
		payload["evidence"] = buildAITaskChannelEvidence(state)
	case model.AIProfileDomainPriceRecognition:
		payload["billing_mode"] = inferAITaskPrimaryBillingMode(state)
		payload["price_items"] = buildAITaskPriceItems(state)
		payload["currency"] = "usd"
		payload["unit"] = "per_1m_tokens"
		payload["missing_items"] = []any{}
	case model.AIProfileDomainModelClassification:
		payload["canonical_name"] = inferAITaskPrimaryCanonicalModel(state)
		payload["aliases"] = buildAITaskModelAliases(state)
		payload["classification"] = buildAITaskModelClassification(state)
		payload["source_type"] = "model_catalog"
		payload["route_hints"] = buildAITaskRouteHints(state)
	case model.AIProfileDomainConfigHealthCheck:
		payload["issues"] = buildAITaskConfigHealthIssues(state)
		payload["severity"] = inferAITaskConfigHealthSeverity(state)
		payload["suggested_actions"] = buildAITaskConfigHealthActions(state)
		payload["blocking_activation"] = false
	}
	return payload
}

func buildAITaskGroupingSuggestions(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.Groups)+1)
	for _, group := range state.ContextPayload.Groups {
		items := asAITaskMapSlice(group["items"])
		out = append(out, map[string]any{
			"group_name": stringFromAny(group["name"]),
			"mode":       stringFromAny(group["mode"]),
			"item_count": len(items),
			"suggestion": fmt.Sprintf("Review routing priorities inside group %s.", stringFromAny(group["name"])),
		})
	}
	if len(out) == 0 {
		groupName := strings.TrimSpace(state.Task.InputText)
		if groupName == "" {
			groupName = "ai-generated-default"
		}
		out = append(out, map[string]any{
			"group_name": groupName,
			"mode":       "weighted",
			"item_count": 0,
			"suggestion": "Create an initial AI-generated grouping draft and review it against manual routing rules.",
		})
	}
	return out
}

func buildAITaskGroupingMappings(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.Channels))
	for _, channel := range state.ContextPayload.Channels {
		out = append(out, map[string]any{
			"channel_name": stringFromAny(channel["name"]),
			"models":       asAITaskStringSlice(channel["models"]),
			"enabled":      channel["enabled"],
		})
	}
	return out
}

func buildAITaskGroupingConflicts(state *aiTaskExecutionState) []any {
	out := []any{}
	for _, group := range state.ContextPayload.Groups {
		if len(asAITaskMapSlice(group["items"])) == 0 {
			out = append(out, map[string]any{
				"group_name":  stringFromAny(group["name"]),
				"type":        "empty_group",
				"description": "Group has no routed channel items.",
			})
		}
	}
	return out
}

func inferAITaskPrimaryChannelType(state *aiTaskExecutionState) string {
	if len(state.ContextPayload.Channels) == 0 {
		return strings.TrimSpace(state.Config.ChannelType)
	}
	return channelTypeNameFromValue(state.ContextPayload.Channels[0]["type"])
}

func buildAITaskChannelCoverage(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.Channels))
	for _, channel := range state.ContextPayload.Channels {
		out = append(out, map[string]any{
			"channel_name": stringFromAny(channel["name"]),
			"models":       asAITaskStringSlice(channel["models"]),
		})
	}
	return out
}

func buildAITaskChannelEvidence(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.Channels))
	for _, channel := range state.ContextPayload.Channels {
		out = append(out, map[string]any{
			"channel_name": stringFromAny(channel["name"]),
			"base_urls":    asAITaskMapSlice(channel["base_urls"]),
			"key_count":    len(asAITaskMapSlice(channel["keys"])),
		})
	}
	return out
}

func inferAITaskPrimaryBillingMode(state *aiTaskExecutionState) string {
	for _, info := range state.ContextPayload.Models {
		if mode := strings.TrimSpace(stringFromAny(info["billing_mode"])); mode != "" {
			return mode
		}
	}
	return "unknown"
}

func buildAITaskPriceItems(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.Models))
	for _, info := range state.ContextPayload.Models {
		out = append(out, map[string]any{
			"model":        stringFromAny(info["name"]),
			"billing_mode": stringFromAny(info["billing_mode"]),
			"input_price":  info["input_price"],
			"output_price": info["output_price"],
		})
	}
	return out
}

func inferAITaskPrimaryCanonicalModel(state *aiTaskExecutionState) string {
	for _, info := range state.ContextPayload.Models {
		if canonical := strings.TrimSpace(stringFromAny(info["canonical_name"])); canonical != "" {
			return canonical
		}
		if name := strings.TrimSpace(stringFromAny(info["name"])); name != "" {
			return name
		}
	}
	return strings.TrimSpace(state.ModelName)
}

func buildAITaskModelAliases(state *aiTaskExecutionState) []any {
	seen := map[string]struct{}{}
	out := []any{}
	for _, info := range state.ContextPayload.Models {
		for _, candidate := range []string{strings.TrimSpace(stringFromAny(info["name"])), strings.TrimSpace(stringFromAny(info["canonical_name"]))} {
			if candidate == "" {
				continue
			}
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			out = append(out, candidate)
		}
	}
	return out
}

func buildAITaskModelClassification(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.Models))
	for _, info := range state.ContextPayload.Models {
		out = append(out, map[string]any{
			"model":        stringFromAny(info["name"]),
			"billing_mode": stringFromAny(info["billing_mode"]),
			"probe_policy": stringFromAny(info["probe_policy"]),
		})
	}
	return out
}

func buildAITaskRouteHints(state *aiTaskExecutionState) []any {
	out := make([]any, 0, len(state.ContextPayload.RouteTargetOverrides))
	for _, row := range state.ContextPayload.RouteTargetOverrides {
		out = append(out, map[string]any{
			"model_name":   stringFromAny(row["model_name"]),
			"billing_mode": stringFromAny(row["billing_mode"]),
			"probe_policy": stringFromAny(row["probe_policy"]),
		})
	}
	return out
}

func buildAITaskConfigHealthIssues(state *aiTaskExecutionState) []any {
	out := []any{}
	if len(state.ContextPayload.Channels) == 0 {
		out = append(out, map[string]any{"type": "channel_missing", "description": "No enabled channels found in current context."})
	}
	if len(state.ContextPayload.Groups) == 0 {
		out = append(out, map[string]any{"type": "group_missing", "description": "No groups found in current context."})
	}
	for _, group := range state.ContextPayload.Groups {
		if len(asAITaskMapSlice(group["items"])) == 0 {
			out = append(out, map[string]any{"type": "empty_group", "group_name": stringFromAny(group["name"]), "description": "Group has no routed channel items."})
		}
	}
	return out
}

func inferAITaskConfigHealthSeverity(state *aiTaskExecutionState) string {
	if len(buildAITaskConfigHealthIssues(state)) > 0 {
		return "medium"
	}
	return "low"
}

func buildAITaskConfigHealthActions(state *aiTaskExecutionState) []any {
	out := []any{}
	if len(state.ContextPayload.Channels) == 0 {
		out = append(out, map[string]any{"type": "add_channel", "description": "Create at least one enabled channel before activating AI-generated plans."})
	}
	if len(state.ContextPayload.Groups) == 0 {
		out = append(out, map[string]any{"type": "create_group", "description": "Create routing groups so generated plans can be reviewed against live topology."})
	}
	if len(out) == 0 {
		out = append(out, map[string]any{"type": "manual_review", "description": "Review generated suggestions and keep manual configuration as the fallback baseline."})
	}
	return out
}

func asAITaskMapSlice(value any) []map[string]any {
	if rows, ok := value.([]map[string]any); ok {
		return rows
	}
	generic, ok := value.([]any)
	if !ok {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(generic))
	for _, item := range generic {
		if row, ok := item.(map[string]any); ok {
			out = append(out, row)
		}
	}
	return out
}

func asAITaskStringSlice(value any) []string {
	if rows, ok := value.([]string); ok {
		return rows
	}
	generic, ok := value.([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(generic))
	for _, item := range generic {
		text := strings.TrimSpace(stringFromAny(item))
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func channelTypeNameFromValue(value any) string {
	switch fmt.Sprint(value) {
	case "1":
		return "openai"
	case "2":
		return "anthropic"
	case "3":
		return "gemini"
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func buildAITaskPromptText(task model.AITask, promptTemplates []model.AIPromptTemplate) string {
	parts := []string{
		"You are helping the Octopus operator organize AI automation outputs.",
		"Return a concise, operator-friendly result that can be saved as a non-destructive AI Profile.",
		"Do not overwrite manual configuration. Do not mutate channels, groups, group_items, llm_infos, or route_target_overrides.",
		fmt.Sprintf("Task type: %s", task.Type),
	}
	toolKeys := sortedAITaskToolKeys(aiTaskToolKeySet(loadAITaskConfigSnapshotOrDefault(task)))
	parts = append(parts, fmt.Sprintf("Enabled tool keys: %s", strings.Join(toolKeys, ", ")))
	if !aiTaskProfileWritable(aiTaskToolKeySet(loadAITaskConfigSnapshotOrDefault(task))) {
		parts = append(parts, "Do not create an AI Profile in this run because profile_write is disabled.")
	}
	if strings.TrimSpace(task.InputText) != "" {
		parts = append(parts, fmt.Sprintf("Operator request: %s", task.InputText))
	}
	if strings.TrimSpace(task.ContextScope) != "" {
		parts = append(parts, fmt.Sprintf("Context scope: %s", task.ContextScope))
	}
	if strings.TrimSpace(task.CustomPrompt) != "" {
		parts = append(parts, fmt.Sprintf("Custom operator requirement: %s", strings.TrimSpace(task.CustomPrompt)))
	}
	if len(promptTemplates) > 0 {
		parts = append(parts, "Selected prompt templates:")
		for _, template := range promptTemplates {
			line := fmt.Sprintf("- [%s] %s", template.Source, strings.TrimSpace(template.Name))
			if prompt := strings.TrimSpace(template.Prompt); prompt != "" {
				line += ": " + prompt
			}
			parts = append(parts, line)
			if requirement := strings.TrimSpace(template.WorkRequirement); requirement != "" {
				parts = append(parts, fmt.Sprintf("  Work requirement: %s", requirement))
			}
		}
	}
	parts = append(parts, "Structure the answer with: 1) short summary, 2) suggested AI Profile content, 3) caveats and fallback notes.")
	return strings.Join(parts, "\n")
}

func loadAITaskConfigSnapshotOrDefault(task model.AITask) model.AIAutomationTaskConfig {
	if snapshot, ok := loadAITaskConfigSnapshot(task); ok {
		return snapshot
	}
	return model.AIAutomationTaskConfig{}
}

func aiTaskResumeStartIndex(resumeState string, steps []struct {
	key      string
	progress int
	run      func(context.Context, *aiTaskExecutionState) error
}) int {
	switch strings.TrimSpace(resumeState) {
	case model.AITaskResumeStateSelectModel:
		return aiTaskStepIndex(steps, "select_model")
	case model.AITaskResumeStateCallAI:
		return aiTaskStepIndex(steps, "call_ai")
	case model.AITaskResumeStateParse:
		return aiTaskStepIndex(steps, "parse_output")
	case model.AITaskResumeStateGenerateProfile:
		return aiTaskStepIndex(steps, "generate_profile")
	case model.AITaskResumeStateSaveResult:
		return aiTaskStepIndex(steps, "save_result")
	default:
		return aiTaskStepIndex(steps, "collect_context")
	}
}

func aiTaskStepIndex(steps []struct {
	key      string
	progress int
	run      func(context.Context, *aiTaskExecutionState) error
}, key string) int {
	for idx, step := range steps {
		if step.key == key {
			return idx
		}
	}
	return 0
}

func hydrateAITaskExecutionState(ctx context.Context, state *aiTaskExecutionState, resumeState string) error {
	if strings.TrimSpace(resumeState) == "" || resumeState == model.AITaskResumeStateCollectContext {
		return nil
	}
	if err := hydrateAITaskCollectedState(ctx, state); err != nil {
		return err
	}
	if resumeState == model.AITaskResumeStateSelectModel {
		return nil
	}
	if err := hydrateAITaskSelectedModel(state); err != nil {
		return err
	}
	if resumeState == model.AITaskResumeStateCallAI {
		return nil
	}
	if strings.TrimSpace(state.Task.ResultJSON) == "" {
		return fmt.Errorf("missing saved AI response for resume state %s", resumeState)
	}
	if resumeState == model.AITaskResumeStateParse {
		state.RawOutput = state.Task.ResultJSON
		return nil
	}
	if err := hydrateAITaskParsedResult(state); err != nil {
		return err
	}
	if resumeState == model.AITaskResumeStateGenerateProfile {
		return nil
	}
	if err := hydrateAITaskProfileDraft(state); err != nil {
		return err
	}
	return nil
}

func hydrateAITaskCollectedState(ctx context.Context, state *aiTaskExecutionState) error {
	config, err := aiAutomationConfigGetRaw(ctx)
	if err != nil {
		return err
	}
	snapshot, ok := loadAITaskConfigSnapshot(state.Task)
	if !ok {
		return fmt.Errorf("missing config snapshot")
	}
	config = applyAITaskConfigSnapshot(config, snapshot)
	state.Config = config
	state.ToolKeys = aiTaskToolKeySet(snapshot)
	promptTemplates, err := listAITaskPromptTemplates(state.Task, ctx)
	if err != nil {
		return err
	}
	state.PromptTemplates = promptTemplates
	if strings.TrimSpace(state.Task.ContextPayloadJSON) == "" {
		return fmt.Errorf("missing context payload")
	}
	if err := json.Unmarshal([]byte(state.Task.ContextPayloadJSON), &state.ContextPayload); err != nil {
		return err
	}
	state.ContextText = state.Task.ContextPayloadJSON
	state.PromptText = state.Task.PromptText
	if strings.TrimSpace(state.PromptText) == "" {
		return fmt.Errorf("missing prompt text")
	}
	return nil
}

func hydrateAITaskSelectedModel(state *aiTaskExecutionState) error {
	state.ModelName = strings.TrimSpace(state.Task.SelectedModel)
	state.ModelReason = strings.TrimSpace(state.Task.ModelReason)
	if state.ModelName == "" {
		return fmt.Errorf("missing selected model")
	}
	return nil
}

func hydrateAITaskParsedResult(state *aiTaskExecutionState) error {
	step, ok := aiTaskStepByKey(state.Task, "parse_output")
	if !ok || strings.TrimSpace(step.OutputJSON) == "" {
		return fmt.Errorf("missing parsed result checkpoint")
	}
	var checkpoint aiTaskResultCheckpoint
	if err := json.Unmarshal([]byte(step.OutputJSON), &checkpoint); err != nil {
		return err
	}
	if strings.TrimSpace(checkpoint.ResultJSON) == "" {
		return fmt.Errorf("parsed result checkpoint is empty")
	}
	state.ResultSummary = checkpoint.ResultSummary
	state.ResultJSON = checkpoint.ResultJSON
	state.ResultPayload = checkpoint.ResultPayload
	state.RawOutput = state.Task.ResultJSON
	return nil
}

func hydrateAITaskProfileDraft(state *aiTaskExecutionState) error {
	step, ok := aiTaskStepByKey(state.Task, "generate_profile")
	if !ok || strings.TrimSpace(step.OutputJSON) == "" {
		if aiTaskProfileWritable(state.ToolKeys) {
			return fmt.Errorf("missing profile draft checkpoint")
		}
		return nil
	}
	var checkpoint aiTaskProfileCheckpoint
	if err := json.Unmarshal([]byte(step.OutputJSON), &checkpoint); err != nil {
		return err
	}
	state.Profile = checkpoint.Profile
	state.ProfileContent = checkpoint.ProfileContent
	return nil
}

func aiTaskStepByKey(task model.AITask, key string) (model.AITaskStep, bool) {
	for _, step := range task.Steps {
		if step.StepKey == key {
			return step, true
		}
	}
	return model.AITaskStep{}, false
}

func resetAITaskStepsForResume(taskID int, stepKey string, ctx context.Context) error {
	if stepKey == "" {
		return nil
	}
	task, err := loadAITask(taskID, ctx)
	if err != nil {
		return err
	}
	currentOrder := 0
	for _, step := range task.Steps {
		if step.StepKey == stepKey {
			currentOrder = step.SortOrder
			break
		}
	}
	if currentOrder == 0 {
		return nil
	}
	return db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).
		Where("task_id = ? AND sort_order >= ?", taskID, currentOrder).
		Updates(map[string]any{"status": model.AITaskStepStatusPending, "message": "", "started_at": nil, "finished_at": nil}).Error
}

func updateAITaskResumeState(taskID int, resumeState string, ctx context.Context) error {
	now := time.Now()
	return db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", taskID).Updates(map[string]any{
		"resume_state":      resumeState,
		"last_heartbeat_at": &now,
	}).Error
}

func updateAITaskStepCheckpoint(taskID int, stepKey string, output any, checkpointState string, ctx context.Context) error {
	raw := ""
	if output != nil {
		payload, err := json.Marshal(output)
		if err != nil {
			return err
		}
		raw = string(payload)
	}
	return db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).Where("task_id = ? AND step_key = ?", taskID, stepKey).Updates(map[string]any{
		"output_json":      raw,
		"checkpoint_state": checkpointState,
	}).Error
}

func finishAITaskUnrecoverable(taskID int, cause error, ctx context.Context) error {
	message := trimTextWithSuffix(cause.Error(), aiTaskResultSummaryLimit, "...")
	now := time.Now()
	if err := db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", taskID).Updates(map[string]any{
		"status":        model.AITaskStatusFailedUnrecoverable,
		"error_message": message,
		"finished_at":   &now,
	}).Error; err != nil {
		return err
	}
	return cause
}

func markAITaskRunning(taskID int, ctx context.Context) error {
	now := time.Now()
	result := db.GetDB().WithContext(ctx).Model(&model.AITask{}).
		Where("id = ? AND status IN ?", taskID, []string{model.AITaskStatusPending, model.AITaskStatusRunning, model.AITaskStatusRecoverable}).
		Updates(map[string]any{"status": model.AITaskStatusRunning, "started_at": &now, "progress": 1, "executor_version": aiTaskExecutorVersion, "attempt_count": gorm.Expr("attempt_count + 1"), "last_heartbeat_at": &now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ensureAITaskRunnable(taskID, ctx)
	}
	return nil
}

func ensureAITaskRunnable(taskID int, ctx context.Context) error {
	task, err := loadAITask(taskID, ctx)
	if err != nil {
		return err
	}
	if task.Status == model.AITaskStatusCanceled {
		return context.Canceled
	}
	if task.Status == model.AITaskStatusFailed || task.Status == model.AITaskStatusSucceeded {
		return fmt.Errorf("task is already finished")
	}
	return nil
}

func startAITaskStep(taskID int, stepKey string, ctx context.Context) error {
	now := time.Now()
	result := db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).
		Where("task_id = ? AND step_key = ? AND status = ?", taskID, stepKey, model.AITaskStepStatusPending).
		Updates(map[string]any{"status": model.AITaskStepStatusRunning, "started_at": &now, "message": ""})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ensureAITaskRunnable(taskID, ctx)
	}
	return nil
}

func completeAITaskStep(taskID int, stepKey string, progress int, ctx context.Context) error {
	now := time.Now()
	if err := db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).Where("task_id = ? AND step_key = ?", taskID, stepKey).Updates(map[string]any{"status": model.AITaskStepStatusSucceeded, "finished_at": &now}).Error; err != nil {
		return err
	}
	return db.GetDB().WithContext(ctx).Model(&model.AITask{}).Where("id = ?", taskID).Update("progress", progress).Error
}

func failAITaskStep(taskID int, stepKey, message string, ctx context.Context) error {
	now := time.Now()
	return db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).Where("task_id = ? AND step_key = ?", taskID, stepKey).Updates(map[string]any{"status": model.AITaskStepStatusFailed, "message": trimTextWithSuffix(message, aiTaskResultSummaryLimit, "..."), "finished_at": &now}).Error
}

func updateAITaskStepMessage(taskID int, stepKey, message string, ctx context.Context) error {
	return db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).Where("task_id = ? AND step_key = ?", taskID, stepKey).Update("message", trimTextWithSuffix(message, aiTaskResultSummaryLimit, "...")).Error
}

func markAITaskStepsAfter(taskID int, stepKey, status, message string, ctx context.Context) error {
	task, err := loadAITask(taskID, ctx)
	if err != nil {
		return err
	}
	currentOrder := 0
	for _, step := range task.Steps {
		if step.StepKey == stepKey {
			currentOrder = step.SortOrder
			break
		}
	}
	if currentOrder == 0 {
		return nil
	}
	now := time.Now()
	return db.GetDB().WithContext(ctx).Model(&model.AITaskStep{}).Where("task_id = ? AND sort_order > ?", taskID, currentOrder).Updates(map[string]any{"status": status, "message": message, "finished_at": &now}).Error
}

func finishAITaskFailure(taskID int, cause error, ctx context.Context) error {
	_ = ctx
	message := trimTextWithSuffix(cause.Error(), aiTaskResultSummaryLimit, "...")
	now := time.Now()
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.GetDB().WithContext(writeCtx).Model(&model.AITask{}).Where("id = ? AND status <> ?", taskID, model.AITaskStatusCanceled).Updates(map[string]any{"status": model.AITaskStatusFailed, "error_message": message, "finished_at": &now}).Error; err != nil {
		return err
	}
	return cause
}

func trimTextWithSuffix(input string, limit int, suffix string) string {
	trimmed := strings.TrimSpace(input)
	if limit <= 0 || len(trimmed) <= limit {
		return trimmed
	}
	if len(suffix) >= limit {
		return trimmed[:limit]
	}
	return strings.TrimSpace(trimmed[:limit-len(suffix)]) + suffix
}

func strPtr(v string) *string {
	return &v
}
