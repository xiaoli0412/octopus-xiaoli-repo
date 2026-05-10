package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

var errGovernanceSessionStale = errors.New("governance session is stale")

func ErrGovernanceSessionStale() error {
	return errGovernanceSessionStale
}

func AIGovernanceOverviewGet(ctx context.Context) (model.AIGovernanceOverview, error) {
	config, err := AIAutomationConfigGet(ctx)
	if err != nil {
		return model.AIGovernanceOverview{}, err
	}
	managedGroupName := settingStringOrDefault(model.SettingKeyAIGovernanceManagedGroupName, "AI Governance Managed")
	strategyProfiles, err := StrategyProfileList(ctx)
	if err != nil {
		return model.AIGovernanceOverview{}, err
	}
	var activeStrategy *model.StrategyProfileSummary
	for i := range strategyProfiles {
		if strategyProfiles[i].IsActive {
			profile := strategyProfiles[i]
			activeStrategy = &profile
			break
		}
	}
	learning, err := AIGovernanceLearningSummaryGet(ctx)
	if err != nil {
		return model.AIGovernanceOverview{}, err
	}
	var recentSession *model.GovernanceSessionSummary
	var session model.GovernanceSession
	if err := db.GetDB().WithContext(ctx).Order("updated_at desc, id desc").First(&session).Error; err == nil {
		summary, convErr := governanceSessionSummaryFromRecord(session)
		if convErr != nil {
			return model.AIGovernanceOverview{}, convErr
		}
		recentSession = &summary
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.AIGovernanceOverview{}, err
	}
	return model.AIGovernanceOverview{
		Enabled:               config.Enabled,
		ExecutionSource:       governanceExecutionSourceFromConfig(config),
		RuntimePolicy:         GovernanceRuntimePolicyGet(),
		ManagedGroupName:      managedGroupName,
		Learning:              learning,
		ActiveStrategyProfile: activeStrategy,
		RecentSession:         recentSession,
	}, nil
}

func AIGovernanceLearningSummaryGet(ctx context.Context) (model.AIGovernanceLearningSummary, error) {
	result, err := DynamicRouteLearningList(ctx)
	if err != nil {
		return model.AIGovernanceLearningSummary{}, err
	}
	summary := model.AIGovernanceLearningSummary{Enabled: result.Enabled, SampleCount: len(result.States)}
	var top *model.DynamicRouteLearningState
	for i := range result.States {
		state := result.States[i]
		if top == nil || state.Score > top.Score {
			copyState := state
			top = &copyState
		}
	}
	if top != nil {
		summary.TopTarget = fmt.Sprintf("%s / ch#%d / key#%d", top.ModelName, top.ChannelID, top.ChannelKeyID)
		summary.LastSampleAt = &top.LastSampleAt
		summary.TopScore = top.Score
	}
	return summary, nil
}

func GovernanceSessionList(ctx context.Context) ([]model.GovernanceSessionSummary, error) {
	rows := make([]model.GovernanceSession, 0)
	if err := db.GetDB().WithContext(ctx).Order("updated_at desc, id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]model.GovernanceSessionSummary, 0, len(rows))
	for _, row := range rows {
		summary, err := governanceSessionSummaryFromRecord(row)
		if err != nil {
			return nil, err
		}
		items = append(items, summary)
	}
	return items, nil
}

func GovernanceSessionGet(id int, ctx context.Context) (model.GovernanceSessionDetail, error) {
	if id <= 0 {
		return model.GovernanceSessionDetail{}, fmt.Errorf("invalid governance session id")
	}
	var row model.GovernanceSession
	if err := db.GetDB().WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GovernanceSessionDetail{}, fmt.Errorf("governance session not found")
		}
		return model.GovernanceSessionDetail{}, err
	}
	plan, preview, snapshotSummary, err := parseGovernanceArtifacts(row)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	applyRuns, err := GovernanceApplyRunList(id, ctx)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	rollbackPoints, err := GovernanceRollbackPointList(id, ctx)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	summary, err := governanceSessionSummaryFromRecord(row)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	return model.GovernanceSessionDetail{
		GovernanceSessionSummary: summary,
		Plan:                     plan,
		Preview:                  preview,
		SnapshotChecksum:         row.SnapshotChecksum,
		ApplyRuns:                applyRuns,
		RollbackPoints:           rollbackPoints,
		SnapshotSummary:          snapshotSummary,
	}, nil
}

func GovernanceSessionCreate(req model.GovernanceSessionCreateRequest, ctx context.Context) (model.GovernanceSessionDetail, error) {
	if err := ensureAIAutomationEnabled(ctx); err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	goal := strings.TrimSpace(req.Goal)
	if goal == "" {
		return model.GovernanceSessionDetail{}, fmt.Errorf("goal is required")
	}
	presetID := strings.TrimSpace(req.ExpertPresetID)
	if presetID == "" {
		presetID = model.GovernanceExpertPresetBalanced
	}
	if _, ok := governanceExpertPresetByID(presetID); !ok {
		return model.GovernanceSessionDetail{}, fmt.Errorf("invalid expert preset")
	}
	snapshot, checksum, err := governanceBuildSnapshot(ctx)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	plan := governancePlanFromSnapshot(goal, presetID, snapshot)
	preview := governancePreviewFromPlan(plan)
	now := time.Now()
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	status := model.GovernanceSessionStatusReady
	if !preview.CanApply {
		status = model.GovernanceSessionStatusFailed
	}
	row := model.GovernanceSession{
		Goal:             goal,
		Scope:            model.GovernanceScopeRoutingGrouping,
		ExpertPresetID:   presetID,
		Status:           status,
		CurrentStage:     model.GovernanceStageCompleted,
		OperatorSummary:  plan.OperatorSummary,
		RiskSummary:      plan.RiskSummary,
		Confidence:       plan.Confidence,
		SnapshotChecksum: checksum,
		SnapshotJSON:     string(snapshotJSON),
		PlanJSON:         string(planJSON),
		PreviewJSON:      string(previewJSON),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	return GovernanceSessionGet(row.ID, ctx)
}

func GovernanceSessionReplan(id int, ctx context.Context) (model.GovernanceSessionDetail, error) {
	var row model.GovernanceSession
	if err := db.GetDB().WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.GovernanceSessionDetail{}, fmt.Errorf("governance session not found")
		}
		return model.GovernanceSessionDetail{}, err
	}
	snapshot, checksum, err := governanceBuildSnapshot(ctx)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	plan := governancePlanFromSnapshot(row.Goal, row.ExpertPresetID, snapshot)
	preview := governancePreviewFromPlan(plan)
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	previewJSON, err := json.Marshal(preview)
	if err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	status := model.GovernanceSessionStatusReady
	if !preview.CanApply {
		status = model.GovernanceSessionStatusFailed
	}
	updates := map[string]any{
		"status":            status,
		"current_stage":     model.GovernanceStageCompleted,
		"operator_summary":  plan.OperatorSummary,
		"risk_summary":      plan.RiskSummary,
		"confidence":        plan.Confidence,
		"snapshot_checksum": checksum,
		"snapshot_json":     string(snapshotJSON),
		"plan_json":         string(planJSON),
		"preview_json":      string(previewJSON),
		"error_message":     "",
		"updated_at":        time.Now(),
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.GovernanceSession{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return model.GovernanceSessionDetail{}, err
	}
	return GovernanceSessionGet(id, ctx)
}

func GovernanceApplyRunList(sessionID int, ctx context.Context) ([]model.GovernanceApplyRunView, error) {
	rows := make([]model.GovernanceApplyRun, 0)
	if err := db.GetDB().WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at desc, id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]model.GovernanceApplyRunView, 0, len(rows))
	for _, row := range rows {
		view, err := governanceApplyRunViewFromRecord(row)
		if err != nil {
			return nil, err
		}
		items = append(items, view)
	}
	return items, nil
}

func StrategyProfileList(ctx context.Context) ([]model.StrategyProfileSummary, error) {
	activeID, _ := SettingGetInt(model.SettingKeyActiveStrategyProfileID)
	rows := make([]model.StrategyProfile, 0)
	if err := db.GetDB().WithContext(ctx).Order("updated_at desc, id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]model.StrategyProfileSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, strategyProfileSummaryFromRecord(row, activeID))
	}
	return items, nil
}

func StrategyProfileCreate(req model.GovernanceStrategyProfileCreateRequest, ctx context.Context) (model.StrategyProfileSummary, error) {
	name := strings.TrimSpace(req.Name)
	if req.SessionID <= 0 {
		return model.StrategyProfileSummary{}, fmt.Errorf("session_id is required")
	}
	if name == "" {
		return model.StrategyProfileSummary{}, fmt.Errorf("name is required")
	}
	session, err := GovernanceSessionGet(req.SessionID, ctx)
	if err != nil {
		return model.StrategyProfileSummary{}, err
	}
	mutationsJSON, err := json.Marshal(session.Plan.Mutations)
	if err != nil {
		return model.StrategyProfileSummary{}, err
	}
	now := time.Now()
	row := model.StrategyProfile{
		Name:            name,
		Summary:         session.GovernanceSessionSummary.OperatorSummary,
		Status:          model.StrategyProfileStatusReady,
		SourceSessionID: &req.SessionID,
		MutationsJSON:   string(mutationsJSON),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
		return model.StrategyProfileSummary{}, err
	}
	activeID, _ := SettingGetInt(model.SettingKeyActiveStrategyProfileID)
	return strategyProfileSummaryFromRecord(row, activeID), nil
}

func StrategyProfileActivate(id int, ctx context.Context) (model.StrategyProfileSummary, error) {
	if id <= 0 {
		return model.StrategyProfileSummary{}, fmt.Errorf("invalid strategy profile id")
	}
	var row model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.StrategyProfileSummary{}, fmt.Errorf("strategy profile not found")
		}
		return model.StrategyProfileSummary{}, err
	}
	now := time.Now()
	err := db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.StrategyProfile{}).Where("status = ?", model.StrategyProfileStatusActive).Updates(map[string]any{"status": model.StrategyProfileStatusReady, "updated_at": now}).Error; err != nil {
			return err
		}
		result := tx.Model(&model.StrategyProfile{}).Where("id = ?", id).Updates(map[string]any{"status": model.StrategyProfileStatusActive, "activated_at": now, "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("strategy profile not found")
		}
		if err := tx.Model(&model.Setting{}).Where("key = ?", model.SettingKeyActiveStrategyProfileID).Update("value", fmt.Sprintf("%d", id)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return model.StrategyProfileSummary{}, err
	}
	if err := db.GetDB().WithContext(ctx).First(&row, id).Error; err != nil {
		return model.StrategyProfileSummary{}, err
	}
	settingCache.Set(model.SettingKeyActiveStrategyProfileID, fmt.Sprintf("%d", id))
	return strategyProfileSummaryFromRecord(row, id), nil
}

func AIGovernanceExpertPresetList() []model.ExpertPresetView {
	presets := governanceExpertPresets()
	return append([]model.ExpertPresetView(nil), presets...)
}

func governanceExecutionSourceFromConfig(config model.AIAutomationConfig) model.AIGovernanceExecutionSource {
	mode := config.ConfigSourceMode
	if strings.TrimSpace(mode) == "" {
		mode = model.ConfigSourceModeManual
	}
	label := "Manual AI endpoint"
	if mode == model.ConfigSourceModeAIProfile {
		label = "AI profile runtime source"
	}
	return model.AIGovernanceExecutionSource{
		Mode:            mode,
		BaseURL:         config.BaseURL,
		ChannelType:     config.ChannelType,
		Model:           config.Model,
		UseLocalDefault: config.UseLocalDefault,
		Label:           label,
	}
}

func governanceSessionSummaryFromRecord(row model.GovernanceSession) (model.GovernanceSessionSummary, error) {
	preview := model.GovernancePreviewView{}
	if strings.TrimSpace(row.PreviewJSON) != "" {
		if err := json.Unmarshal([]byte(row.PreviewJSON), &preview); err != nil {
			return model.GovernanceSessionSummary{}, err
		}
	}
	return model.GovernanceSessionSummary{ID: row.ID, Goal: row.Goal, Scope: row.Scope, ExpertPresetID: row.ExpertPresetID, Status: row.Status, CurrentStage: row.CurrentStage, OperatorSummary: row.OperatorSummary, RiskSummary: row.RiskSummary, Confidence: row.Confidence, MutationCount: preview.MutationCount, CanApply: preview.CanApply, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, AppliedAt: row.AppliedAt}, nil
}

func parseGovernanceArtifacts(row model.GovernanceSession) (model.GovernancePlanView, model.GovernancePreviewView, model.GovernanceSnapshotSummary, error) {
	plan := model.GovernancePlanView{}
	preview := model.GovernancePreviewView{}
	summary := model.GovernanceSnapshotSummary{}
	if strings.TrimSpace(row.PlanJSON) != "" {
		if err := json.Unmarshal([]byte(row.PlanJSON), &plan); err != nil {
			return plan, preview, summary, err
		}
	}
	if strings.TrimSpace(row.PreviewJSON) != "" {
		if err := json.Unmarshal([]byte(row.PreviewJSON), &preview); err != nil {
			return plan, preview, summary, err
		}
	}
	if strings.TrimSpace(row.SnapshotJSON) != "" {
		var snapshot governanceSnapshot
		if err := json.Unmarshal([]byte(row.SnapshotJSON), &snapshot); err != nil {
			return plan, preview, summary, err
		}
		summary = snapshot.SnapshotSummary
	}
	return plan, preview, summary, nil
}

func governanceApplyRunViewFromRecord(row model.GovernanceApplyRun) (model.GovernanceApplyRunView, error) {
	audit := model.GovernanceApplyAudit{}
	if strings.TrimSpace(row.AuditJSON) != "" {
		if err := json.Unmarshal([]byte(row.AuditJSON), &audit); err != nil {
			return model.GovernanceApplyRunView{}, err
		}
	}
	return model.GovernanceApplyRunView{ID: row.ID, SessionID: row.SessionID, Status: row.Status, ResultSummary: row.ResultSummary, ErrorMessage: row.ErrorMessage, Audit: audit, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func strategyProfileSummaryFromRecord(row model.StrategyProfile, activeID int) model.StrategyProfileSummary {
	return model.StrategyProfileSummary{ID: row.ID, Name: row.Name, Summary: row.Summary, Status: row.Status, SourceSessionID: row.SourceSessionID, ActivatedAt: row.ActivatedAt, IsActive: row.ID == activeID || row.Status == model.StrategyProfileStatusActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func governanceExpertPresets() []model.ExpertPresetView {
	return []model.ExpertPresetView{
		{ID: model.GovernanceExpertPresetBalanced, Name: "Balanced governance", Description: "Default routing and grouping governance with cleanup and ordering refresh.", ReviewDepth: "standard", CreateManagedGroup: true, SyncBindings: true, CleanupStale: true},
		{ID: model.GovernanceExpertPresetConservative, Name: "Conservative review", Description: "Favor minimal writes and avoid stale cleanup unless it is clearly invalid.", ReviewDepth: "light", CreateManagedGroup: true, SyncBindings: true, CleanupStale: false},
		{ID: model.GovernanceExpertPresetDeepReview, Name: "Deep review", Description: "Produce a denser review with explicit cleanup and drift notes for operators.", ReviewDepth: "deep", CreateManagedGroup: true, SyncBindings: true, CleanupStale: true},
	}
}

func governanceExpertPresetByID(id string) (model.ExpertPresetView, bool) {
	for _, preset := range governanceExpertPresets() {
		if preset.ID == id {
			return preset, true
		}
	}
	return model.ExpertPresetView{}, false
}
