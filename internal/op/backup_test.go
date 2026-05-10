package op

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db/migrate"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	transformerOutbound "github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
)

func TestDBExportAllIncludesSecretsByDefault(t *testing.T) {
	ctx := setupOpTestDB(t)
	exportedUser := createTestUser(t, ctx, "backup-admin", "backup-secret")

	proxyURL := "https://octopus:secret@example.com:8443"
	channel := model.Channel{
		Name:    "safe-export-channel",
		Enabled: true,
		CustomHeader: []model.CustomHeader{
			{HeaderKey: "Authorization", HeaderValue: "Bearer upstream-secret"},
			{HeaderKey: "X-Workspace-ID", HeaderValue: "workspace-1"},
		},
		ChannelProxy: &proxyURL,
	}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	channelKey := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "upstream-key"}
	if err := db.GetDB().WithContext(ctx).Create(&channelKey).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	apiKey := model.APIKey{Name: "client", APIKey: "sk-octopus-client-secret", Enabled: true}
	if err := db.GetDB().WithContext(ctx).Create(&apiKey).Error; err != nil {
		t.Fatalf("create api key error = %v", err)
	}

	dump, err := DBExportAll(ctx, false, false, true)
	if err != nil {
		t.Fatalf("DBExportAll() error = %v", err)
	}
	if !dump.Manifest.ContainsSecrets {
		t.Fatalf("dump.Manifest.ContainsSecrets = false, want true")
	}
	var exportedBackupAdmin *model.User
	for i := range dump.Users {
		if dump.Users[i].Username == exportedUser.Username {
			exportedBackupAdmin = &dump.Users[i]
			break
		}
	}
	if exportedBackupAdmin == nil {
		t.Fatalf("exported users = %#v, want user %q present", dump.Users, exportedUser.Username)
	}
	if exportedBackupAdmin.Password == "" || exportedBackupAdmin.Password == "backup-secret" {
		t.Fatalf("exported user password = %q, want stored hashed password", exportedBackupAdmin.Password)
	}
	if dump.Manifest.Checksum == "" {
		t.Fatalf("dump.Manifest.Checksum = empty, want non-empty")
	}

	if len(dump.ChannelKeys) != 1 || dump.ChannelKeys[0].ChannelKey != "upstream-key" {
		t.Fatalf("exported channel keys = %#v, want original credential", dump.ChannelKeys)
	}
	if len(dump.APIKeys) != 1 || dump.APIKeys[0].APIKey != "sk-octopus-client-secret" {
		t.Fatalf("exported api keys = %#v, want original credential", dump.APIKeys)
	}
	if len(dump.Channels) != 1 {
		t.Fatalf("exported channels len = %d, want 1", len(dump.Channels))
	}
	if got := dump.Channels[0].CustomHeader[0].HeaderValue; got != "Bearer upstream-secret" {
		t.Fatalf("authorization header value = %q, want preserved secret", got)
	}
	if got := dump.Channels[0].CustomHeader[1].HeaderValue; got != "workspace-1" {
		t.Fatalf("non-secret header value = %q, want workspace-1", got)
	}
	if dump.Channels[0].ChannelProxy == nil || *dump.Channels[0].ChannelProxy != proxyURL {
		t.Fatalf("channel proxy = %#v, want preserved proxy url", dump.Channels[0].ChannelProxy)
	}
}

func TestDBExportAllRedactsUserPasswordsWhenSecretsExcluded(t *testing.T) {
	ctx := setupOpTestDB(t)
	exportedUser := createTestUser(t, ctx, "backup-admin", "backup-secret")

	proxyURL := "https://octopus:secret@example.com:8443"
	channel := model.Channel{
		Name:    "redacted-export-channel",
		Enabled: true,
		CustomHeader: []model.CustomHeader{
			{HeaderKey: "Authorization", HeaderValue: "Bearer upstream-secret"},
			{HeaderKey: "X-Workspace-ID", HeaderValue: "workspace-1"},
			{HeaderKey: "X-Trace-Label", HeaderValue: "blue"},
		},
		ChannelProxy: &proxyURL,
	}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	channelKey := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "upstream-key"}
	if err := db.GetDB().WithContext(ctx).Create(&channelKey).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	apiKey := model.APIKey{Name: "client", APIKey: "sk-octopus-client-secret", Enabled: true}
	if err := db.GetDB().WithContext(ctx).Create(&apiKey).Error; err != nil {
		t.Fatalf("create api key error = %v", err)
	}

	dump, err := DBExportAll(ctx, false, false, false)
	if err != nil {
		t.Fatalf("DBExportAll() error = %v", err)
	}
	if dump.Manifest.ContainsSecrets {
		t.Fatalf("dump.Manifest.ContainsSecrets = true, want false after redaction")
	}
	var exportedBackupAdmin *model.User
	for i := range dump.Users {
		if dump.Users[i].Username == exportedUser.Username {
			exportedBackupAdmin = &dump.Users[i]
			break
		}
	}
	if exportedBackupAdmin == nil {
		t.Fatalf("exported users = %#v, want user %q present", dump.Users, exportedUser.Username)
	}
	if exportedBackupAdmin.Password != "" {
		t.Fatalf("exported user password = %q, want empty after redaction", exportedBackupAdmin.Password)
	}
	if len(dump.ChannelKeys) != 1 || dump.ChannelKeys[0].ChannelKey != "" {
		t.Fatalf("redacted channel keys = %#v, want empty credential", dump.ChannelKeys)
	}
	if len(dump.APIKeys) != 1 || dump.APIKeys[0].APIKey != "" {
		t.Fatalf("redacted api keys = %#v, want empty credential", dump.APIKeys)
	}
	if len(dump.Channels) != 1 {
		t.Fatalf("redacted channels len = %d, want 1", len(dump.Channels))
	}
	if got := dump.Channels[0].CustomHeader[0].HeaderValue; got != "" {
		t.Fatalf("authorization header value = %q, want empty after redaction", got)
	}
	if got := dump.Channels[0].CustomHeader[1].HeaderValue; got != "workspace-1" {
		t.Fatalf("workspace header value = %q, want preserved non-secret value", got)
	}
	if got := dump.Channels[0].CustomHeader[2].HeaderValue; got != "blue" {
		t.Fatalf("trace header value = %q, want preserved non-secret value", got)
	}
	if dump.Channels[0].ChannelProxy == nil || *dump.Channels[0].ChannelProxy != "https://example.com:8443" {
		t.Fatalf("redacted channel proxy = %#v, want credentials stripped and endpoint preserved", dump.Channels[0].ChannelProxy)
	}
}

func TestDBExportAllExcludesInternalAuthTokenSecret(t *testing.T) {
	ctx := setupOpTestDB(t)
	setTestSetting(t, ctx, model.SettingKeyAuthTokenSecret, "internal-signing-secret")

	dump, err := DBExportAll(ctx, false, false, true)
	if err != nil {
		t.Fatalf("DBExportAll() error = %v", err)
	}
	for _, setting := range dump.Settings {
		if setting.Key == model.SettingKeyAuthTokenSecret {
			t.Fatalf("dump.Settings leaked internal auth token secret: %#v", dump.Settings)
		}
	}
}

func TestDBExportAllIncludesAIAutomationStateAndRedactsSecretsWhenRequested(t *testing.T) {
	ctx := setupOpTestDB(t)

	profile, err := AIProfileCreate(model.AIProfile{
		Domain:      model.AIProfileDomainGrouping,
		Name:        "backup-profile",
		Status:      model.AIProfileStatusReady,
		Confidence:  0.9,
		Explanation: "profile explanation",
	}, `{"config":{"base_url":"https://profile.example/v1","api_key":"profile-secret-key","channel_type":"openai","model":"gpt-4o"},"domain_payload":{"typed_config":{"api_key":"typed-secret-key"}}}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.AIPromptTemplate{Name: "backup template", Source: model.AIPromptTemplateSourceCustom, TaskType: model.AIAutomationTaskTypeGroupSuggestion, Domain: model.AIProfileDomainGrouping, Prompt: "prompt", Enabled: true}).Error; err != nil {
		t.Fatalf("create prompt template error = %v", err)
	}
	task := model.AITask{
		Type:               model.AIAutomationTaskTypeNaturalLanguage,
		InputText:          "backup task",
		Status:             model.AITaskStatusSucceeded,
		ConfigSnapshotJSON: `{"base_url":"https://task.example/v1","api_key":"task-secret-key","channel_type":"openai","model":"gpt-4o"}`,
		ContextPayloadJSON: `{"config":{"api_key":"context-secret-key"}}`,
		ResultJSON:         `{"summary":"ok","config":{"api_key":"result-secret-key"}}`,
		PromptText:         "task prompt",
		SelectedModel:      "gpt-4o",
		ResumeState:        model.AITaskResumeStateCompleted,
		ExecutorVersion:    "test",
	}
	if err := db.GetDB().WithContext(ctx).Create(&task).Error; err != nil {
		t.Fatalf("create ai task error = %v", err)
	}
	step := model.AITaskStep{TaskID: task.ID, StepKey: "call_ai", Name: "调用 AI", Status: model.AITaskStepStatusSucceeded, OutputJSON: `{"api_key":"step-secret-key"}`, SortOrder: 1}
	if err := db.GetDB().WithContext(ctx).Create(&step).Error; err != nil {
		t.Fatalf("create ai task step error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 1, ChannelKeyID: 2, ModelName: "gpt-4o", SuccessCount: 3, Score: 0.7, Confidence: 0.5}).Error; err != nil {
		t.Fatalf("create dynamic route learning state error = %v", err)
	}

	fullDump, err := DBExportAll(ctx, false, false, true)
	if err != nil {
		t.Fatalf("DBExportAll(includeSecrets=true) error = %v", err)
	}
	if len(fullDump.AITasks) != 1 || len(fullDump.AITaskSteps) != 1 || len(fullDump.AIPromptTemplates) != 1 || len(fullDump.AIProfiles) != 1 || len(fullDump.AIProfileVersions) != 1 || len(fullDump.AIGroupingProfiles) != 1 || len(fullDump.DynamicRouteLearningStates) != 1 {
		t.Fatalf("full dump missing ai automation state: tasks=%d steps=%d templates=%d profiles=%d versions=%d typed=%d learning=%d", len(fullDump.AITasks), len(fullDump.AITaskSteps), len(fullDump.AIPromptTemplates), len(fullDump.AIProfiles), len(fullDump.AIProfileVersions), len(fullDump.AIGroupingProfiles), len(fullDump.DynamicRouteLearningStates))
	}
	if !strings.Contains(fullDump.AITasks[0].ConfigSnapshotJSON, "task-secret-key") || !strings.Contains(fullDump.AITasks[0].ContextPayloadJSON, "context-secret-key") || !strings.Contains(fullDump.AITasks[0].ResultJSON, "result-secret-key") {
		t.Fatalf("full dump ai task secrets not preserved: %#v", fullDump.AITasks[0])
	}
	if !strings.Contains(fullDump.AIProfileVersions[0].ContentJSON, "profile-secret-key") {
		t.Fatalf("full dump ai profile version missing secret content: %s", fullDump.AIProfileVersions[0].ContentJSON)
	}
	if !strings.Contains(fullDump.AIGroupingProfiles[0].TypedPayloadJSON, "typed-secret-key") {
		t.Fatalf("full dump typed profile missing secret content: %s", fullDump.AIGroupingProfiles[0].TypedPayloadJSON)
	}

	redactedDump, err := DBExportAll(ctx, false, false, false)
	if err != nil {
		t.Fatalf("DBExportAll(includeSecrets=false) error = %v", err)
	}
	if redactedDump.Manifest.ContainsSecrets {
		t.Fatalf("redacted dump Manifest.ContainsSecrets = true, want false")
	}
	for _, raw := range []string{redactedDump.AITasks[0].ConfigSnapshotJSON, redactedDump.AITasks[0].ContextPayloadJSON, redactedDump.AITasks[0].ResultJSON, redactedDump.AIProfileVersions[0].ContentJSON, redactedDump.AIGroupingProfiles[0].TypedPayloadJSON} {
		if strings.Contains(raw, "secret-key") {
			t.Fatalf("redacted dump leaked ai secret payload: %s", raw)
		}
		if !strings.Contains(raw, aiAutomationRedactedSecret) {
			t.Fatalf("redacted dump payload = %s, want redaction marker", raw)
		}
	}
	if redactedDump.AIProfiles[0].ID != profile.ID {
		t.Fatalf("redacted ai profile id = %d, want %d", redactedDump.AIProfiles[0].ID, profile.ID)
	}
}

func TestDBExportAllIncludesGovernanceState(t *testing.T) {
	ctx := setupOpTestDB(t)

	now := time.Now().UTC().Truncate(time.Second)
	session := model.GovernanceSession{
		Goal:             "governance backup",
		Scope:            model.GovernanceScopeRoutingGrouping,
		ExpertPresetID:   model.GovernanceExpertPresetBalanced,
		Status:           model.GovernanceSessionStatusReady,
		CurrentStage:     model.GovernanceStageCompleted,
		OperatorSummary:  "session summary",
		RiskSummary:      "risk summary",
		Confidence:       0.88,
		SnapshotChecksum: "checksum-1",
		SnapshotJSON:     `{"channels":1}`,
		PlanJSON:         `{"mutations":[]}`,
		PreviewJSON:      `{"can_apply":true}`,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&session).Error; err != nil {
		t.Fatalf("create governance session error = %v", err)
	}
	applyRun := model.GovernanceApplyRun{
		SessionID:     session.ID,
		Status:        model.GovernanceApplyRunStatusSucceeded,
		ResultSummary: "Applied governance changes",
		AuditJSON:     `{"summary":"ok"}`,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&applyRun).Error; err != nil {
		t.Fatalf("create governance apply run error = %v", err)
	}
	rollbackPoint := model.GovernanceRollbackPoint{
		SessionID:        session.ID,
		ApplyRunID:       &applyRun.ID,
		SnapshotChecksum: "checksum-rollback",
		SnapshotJSON:     `{"channels":1}`,
		Summary:          "rollback point",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&rollbackPoint).Error; err != nil {
		t.Fatalf("create governance rollback point error = %v", err)
	}
	strategyProfile := model.StrategyProfile{
		Name:            "governance strategy",
		Summary:         "strategy summary",
		Status:          model.StrategyProfileStatusActive,
		SourceSessionID: &session.ID,
		MutationsJSON:   `[{"type":"group_upsert"}]`,
		CreatedAt:       now,
		UpdatedAt:       now,
		ActivatedAt:     &now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&strategyProfile).Error; err != nil {
		t.Fatalf("create strategy profile error = %v", err)
	}

	dump, err := DBExportAll(ctx, false, false, true)
	if err != nil {
		t.Fatalf("DBExportAll() error = %v", err)
	}
	if len(dump.GovernanceSessions) != 1 || dump.GovernanceSessions[0].Goal != session.Goal {
		t.Fatalf("governance_sessions = %#v, want exported session", dump.GovernanceSessions)
	}
	if len(dump.GovernanceApplyRuns) != 1 || dump.GovernanceApplyRuns[0].ResultSummary != applyRun.ResultSummary {
		t.Fatalf("governance_apply_runs = %#v, want exported apply run", dump.GovernanceApplyRuns)
	}
	if len(dump.GovernanceRollbackPoints) != 1 || dump.GovernanceRollbackPoints[0].Summary != rollbackPoint.Summary {
		t.Fatalf("governance_rollback_points = %#v, want exported rollback point", dump.GovernanceRollbackPoints)
	}
	if len(dump.StrategyProfiles) != 1 || dump.StrategyProfiles[0].Name != strategyProfile.Name {
		t.Fatalf("strategy_profiles = %#v, want exported strategy profile", dump.StrategyProfiles)
	}
}

func TestChannelUpdateRemovesStaleGroupItemsWhenModelsShrink(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := createConfiguredTestChannel(t, ctx, "channel-model-shrink", "gpt-4o,o1-mini", "custom-a")

	group := &model.Group{Name: "group-model-shrink", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	for _, item := range []model.GroupItem{
		{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1},
		{GroupID: group.ID, ChannelID: channel.ID, ModelName: "o1-mini", Priority: 2, Weight: 1},
		{GroupID: group.ID, ChannelID: channel.ID, ModelName: "custom-a", Priority: 3, Weight: 1},
	} {
		item := item
		if err := GroupItemAdd(&item, ctx); err != nil {
			t.Fatalf("GroupItemAdd() error = %v", err)
		}
	}

	updatedModel := "gpt-4o"
	updatedCustomModel := ""
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID:          channel.ID,
		Model:       &updatedModel,
		CustomModel: &updatedCustomModel,
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate() error = %v", err)
	}
	if updated.Model != "gpt-4o" || updated.CustomModel != "" {
		t.Fatalf("updated channel models = model:%q custom:%q, want only gpt-4o", updated.Model, updated.CustomModel)
	}

	refreshedGroup, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(refreshedGroup.Items) != 1 {
		t.Fatalf("group items len after channel model shrink = %d, want 1", len(refreshedGroup.Items))
	}
	if refreshedGroup.Items[0].ModelName != "gpt-4o" {
		t.Fatalf("remaining group item model = %q, want gpt-4o", refreshedGroup.Items[0].ModelName)
	}

	var storedItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Where("group_id = ?", group.ID).Order("priority asc").Find(&storedItems).Error; err != nil {
		t.Fatalf("query group items error = %v", err)
	}
	if len(storedItems) != 1 || storedItems[0].ModelName != "gpt-4o" {
		t.Fatalf("stored group items = %#v, want only gpt-4o", storedItems)
	}
}

func TestLLMDeleteRemovesReferencedGroupItems(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := createConfiguredTestChannel(t, ctx, "channel-llm-delete", "gpt-4o,o1-mini", "")
	group := &model.Group{Name: "group-llm-delete", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	for _, item := range []model.GroupItem{
		{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1},
		{GroupID: group.ID, ChannelID: channel.ID, ModelName: "o1-mini", Priority: 2, Weight: 1},
	} {
		item := item
		if err := GroupItemAdd(&item, ctx); err != nil {
			t.Fatalf("GroupItemAdd() error = %v", err)
		}
	}
	if err := LLMCreate(model.LLMInfo{Name: "gpt-4o", CanonicalName: "gpt-4o"}, ctx); err != nil {
		t.Fatalf("LLMCreate(gpt-4o) error = %v", err)
	}
	if err := LLMCreate(model.LLMInfo{Name: "o1-mini", CanonicalName: "o1-mini"}, ctx); err != nil {
		t.Fatalf("LLMCreate(o1-mini) error = %v", err)
	}

	if err := LLMDelete("o1-mini", ctx); err != nil {
		t.Fatalf("LLMDelete() error = %v", err)
	}

	refreshedGroup, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(refreshedGroup.Items) != 1 || refreshedGroup.Items[0].ModelName != "gpt-4o" {
		t.Fatalf("group items after LLMDelete = %#v, want only gpt-4o", refreshedGroup.Items)
	}

	if _, err := LLMGet("o1-mini"); err == nil {
		t.Fatalf("LLMGet(o1-mini) expected error after delete")
	}
}

func TestLLMBatchDeleteRemovesReferencedGroupItems(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := createConfiguredTestChannel(t, ctx, "channel-llm-batch-delete", "gpt-4o,o1-mini,claude-3-5-sonnet", "")
	group := &model.Group{Name: "group-llm-batch-delete", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	for idx, modelName := range []string{"gpt-4o", "o1-mini", "claude-3-5-sonnet"} {
		item := model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: modelName, Priority: idx + 1, Weight: 1}
		if err := GroupItemAdd(&item, ctx); err != nil {
			t.Fatalf("GroupItemAdd(%s) error = %v", modelName, err)
		}
		if err := LLMCreate(model.LLMInfo{Name: modelName, CanonicalName: modelName}, ctx); err != nil {
			t.Fatalf("LLMCreate(%s) error = %v", modelName, err)
		}
	}

	if err := LLMBatchDelete([]string{"o1-mini", "claude-3-5-sonnet", "o1-mini"}, ctx); err != nil {
		t.Fatalf("LLMBatchDelete() error = %v", err)
	}

	refreshedGroup, err := GroupGet(group.ID, ctx)
	if err != nil {
		t.Fatalf("GroupGet() error = %v", err)
	}
	if len(refreshedGroup.Items) != 1 || refreshedGroup.Items[0].ModelName != "gpt-4o" {
		t.Fatalf("group items after LLMBatchDelete = %#v, want only gpt-4o", refreshedGroup.Items)
	}
}

func TestDBImportIncrementalDryRunReportsRedactedCredentials(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels:    []model.Channel{{ID: 101, Name: "preview-channel", Enabled: true}},
		ChannelKeys: []model.ChannelKey{{ID: 201, ChannelID: 101, Enabled: true, ChannelKey: ""}},
		APIKeys:     []model.APIKey{{ID: 301, Name: "preview-client", APIKey: ""}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if !containsWarning(res.Warnings, snapshotCredentialsRedactedWarning) {
		t.Fatalf("warnings = %#v, want redacted credential warning", res.Warnings)
	}
	if !containsWarning(res.Compatibility.SkippedTargets, "channel_key:201 empty credential") {
		t.Fatalf("skipped targets = %#v, want empty channel key entry", res.Compatibility.SkippedTargets)
	}
	if !containsWarning(res.Compatibility.SkippedTargets, "api_key:301 empty credential") {
		t.Fatalf("skipped targets = %#v, want empty api key entry", res.Compatibility.SkippedTargets)
	}
	if got := res.Compatibility.Summary.CredentialRebindTargets; got != 2 {
		t.Fatalf("summary.credential_rebind_targets = %d, want 2", got)
	}
	if got := res.Compatibility.Summary.ChannelKeyRebindTargets; got != 1 {
		t.Fatalf("summary.channel_key_rebind_targets = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.APIKeyRebindTargets; got != 1 {
		t.Fatalf("summary.api_key_rebind_targets = %d, want 1", got)
	}
	if len(res.Compatibility.CredentialRebindTargets) != 2 {
		t.Fatalf("credential_rebind_targets = %#v, want 2", res.Compatibility.CredentialRebindTargets)
	}
	channelKeyTarget := res.Compatibility.CredentialRebindTargets[0]
	if channelKeyTarget.TargetType != "api_key" && channelKeyTarget.TargetType != "channel_key" {
		t.Fatalf("credential_rebind_target.target_type = %q, want api_key or channel_key", channelKeyTarget.TargetType)
	}
	foundChannelKeyRebind := false
	foundAPIKeyRebind := false
	for _, target := range res.Compatibility.CredentialRebindTargets {
		switch target.TargetType {
		case "channel_key":
			foundChannelKeyRebind = true
			if target.SnapshotID != 201 || target.ChannelName != "preview-channel" {
				t.Fatalf("channel key rebind target = %#v, want preview-channel / 201", target)
			}
			if !containsWarning(target.Warnings, "rebind required") {
				t.Fatalf("channel key rebind warnings = %#v, want rebind warning", target.Warnings)
			}
		case "api_key":
			foundAPIKeyRebind = true
			if target.SnapshotID != 301 || target.KeyName != "preview-client" {
				t.Fatalf("api key rebind target = %#v, want preview-client / 301", target)
			}
			if !containsWarning(target.Contexts, "api_key:preview-client") {
				t.Fatalf("api key rebind contexts = %#v, want api_key:preview-client", target.Contexts)
			}
		}
	}
	if !foundChannelKeyRebind || !foundAPIKeyRebind {
		t.Fatalf("credential_rebind_targets = %#v, want both channel_key and api_key targets", res.Compatibility.CredentialRebindTargets)
	}
}

func TestDBImportIncrementalSkipsEmptyCredentialsOnApply(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{ID: 111, Name: "apply-channel", Enabled: true}},
		ChannelKeys: []model.ChannelKey{
			{ID: 211, ChannelID: 111, Enabled: true, ChannelKey: ""},
			{ID: 212, ChannelID: 111, Enabled: true, ChannelKey: "restored-key"},
		},
		APIKeys: []model.APIKey{
			{ID: 311, Name: "blank-client", APIKey: "", Enabled: true},
			{ID: 312, Name: "restored-client", APIKey: "sk-octopus-restored", Enabled: true},
		},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["channel_keys"]; got != 1 {
		t.Fatalf("rows_affected[channel_keys] = %d, want 1", got)
	}
	if got := res.RowsAffected["api_keys"]; got != 1 {
		t.Fatalf("rows_affected[api_keys] = %d, want 1", got)
	}
	if !containsWarning(res.Warnings, "skipped 1 channel keys without credentials") {
		t.Fatalf("warnings = %#v, want skipped channel key warning", res.Warnings)
	}
	if !containsWarning(res.Warnings, "skipped 1 API keys without credentials") {
		t.Fatalf("warnings = %#v, want skipped api key warning", res.Warnings)
	}
	if got := res.Compatibility.Summary.CredentialRebindTargets; got != 2 {
		t.Fatalf("summary.credential_rebind_targets = %d, want 2", got)
	}
	if len(res.Compatibility.CredentialRebindTargets) != 2 {
		t.Fatalf("credential_rebind_targets = %#v, want 2", res.Compatibility.CredentialRebindTargets)
	}

	var channelKeys []model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&channelKeys).Error; err != nil {
		t.Fatalf("query channel keys error = %v", err)
	}
	if len(channelKeys) != 1 || channelKeys[0].ChannelKey != "restored-key" {
		t.Fatalf("stored channel keys = %#v, want only restored-key", channelKeys)
	}

	var apiKeys []model.APIKey
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&apiKeys).Error; err != nil {
		t.Fatalf("query api keys error = %v", err)
	}
	if len(apiKeys) != 1 || apiKeys[0].APIKey != "sk-octopus-restored" {
		t.Fatalf("stored api keys = %#v, want only sk-octopus-restored", apiKeys)
	}
}

func TestDBImportIncrementalRejectsInvalidSettingValue(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion: "v1",
			ExportSource:  "octopus",
		},
		Settings: []model.Setting{{Key: model.SettingKeyDynamicRoutingMode, Value: "broken-mode"}},
	}

	_, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err == nil || !strings.Contains(err.Error(), "invalid setting dynamic_routing_mode") {
		t.Fatalf("DBImportIncremental() error = %v, want invalid setting dynamic_routing_mode", err)
	}
}

func TestDBImportIncrementalRejectsUnknownSettingKey(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion: "v1",
			ExportSource:  "octopus",
		},
		Settings: []model.Setting{{Key: model.SettingKey("unknown_setting_key"), Value: "value"}},
	}

	_, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err == nil || !strings.Contains(err.Error(), "unknown setting key: unknown_setting_key") {
		t.Fatalf("DBImportIncremental() error = %v, want unknown setting key error", err)
	}
}

func TestDBImportIncrementalRejectsChannelProxyWithCredentials(t *testing.T) {
	ctx := setupOpTestDB(t)
	proxyURL := "https://user:pass@example.com:8443"

	_, err := DBImportIncremental(ctx, &model.DBDump{
		Version: dbDumpVersion,
		Channels: []model.Channel{{
			ID:           1,
			Name:         "import-channel-proxy-creds",
			Enabled:      true,
			Model:        "gpt-4o",
			ChannelProxy: &proxyURL,
		}},
	}, model.DBImportModeIncremental, true)
	if err == nil {
		t.Fatalf("DBImportIncremental() expected invalid channel proxy error")
	}
	if got, want := err.Error(), "invalid channel import-channel-proxy-creds proxy: channel_proxy must not include credentials"; got != want {
		t.Fatalf("DBImportIncremental() error = %q, want %q", got, want)
	}
}

func TestDBImportIncrementalDryRunReportsRedactedCredentialRouteContexts(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{
			ID:      901,
			Name:    "context-channel",
			Enabled: true,
			Model:   "gpt-4o",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            902,
			ChannelID:     901,
			Enabled:       true,
			ChannelKey:    "",
			AllowedModels: "gpt-4o",
			SourceType:    "paid/metered",
			Remark:        "primary-key",
		}},
		Groups:     []model.Group{{ID: 903, Name: "context-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 904, GroupID: 903, ChannelID: 901, ModelName: "gpt-4o", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if len(res.Compatibility.CredentialRebindTargets) != 1 {
		t.Fatalf("credential_rebind_targets = %#v, want 1", res.Compatibility.CredentialRebindTargets)
	}
	target := res.Compatibility.CredentialRebindTargets[0]
	if target.TargetType != "channel_key" || target.ChannelName != "context-channel" || target.KeyName != "primary-key" {
		t.Fatalf("credential_rebind_target = %#v, want channel key for context-channel/primary-key", target)
	}
	if !containsWarning(target.Models, "gpt-4o") {
		t.Fatalf("credential_rebind_target.models = %#v, want gpt-4o", target.Models)
	}
	if !containsWarning(target.AffectedGroups, "context-group") {
		t.Fatalf("credential_rebind_target.affected_groups = %#v, want context-group", target.AffectedGroups)
	}
	if !containsWarning(target.Contexts, "group:context-group") {
		t.Fatalf("credential_rebind_target.contexts = %#v, want group:context-group", target.Contexts)
	}
	if !containsWarning(target.Contexts, "channel:context-channel") {
		t.Fatalf("credential_rebind_target.contexts = %#v, want channel:context-channel", target.Contexts)
	}
}

func TestDBImportIncrementalDryRunBuildsStructuredCompatibility(t *testing.T) {
	ctx := setupOpTestDB(t)

	currentChannel := model.Channel{
		Name:     "preview-channel",
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "https://current.example.com/v1", Delay: 0}},
		Model:    "gpt-5.4-pro",
	}
	if err := db.GetDB().WithContext(ctx).Create(&currentChannel).Error; err != nil {
		t.Fatalf("create current channel error = %v", err)
	}
	currentGroup := model.Group{Name: "preview-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&currentGroup).Error; err != nil {
		t.Fatalf("create current group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: currentGroup.ID, ChannelID: currentChannel.ID, ModelName: "gpt-5.4-pro", Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create current group item error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-5.4-pro", CanonicalName: "gpt-5.4-pro"}).Error; err != nil {
		t.Fatalf("create llm info error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://current-gateway.example.com")

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v2",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{
			ID:       101,
			Name:     "preview-channel",
			Enabled:  true,
			BaseUrls: []model.BaseUrl{{URL: "https://snapshot.example.com/v1", Delay: 0}},
			Model:    "gpt54pro,missing-model",
		}},
		Groups:     []model.Group{{ID: 201, Name: "preview-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 301, GroupID: 201, ChannelID: 101, ModelName: "gpt54pro", Priority: 2, Weight: 1}},
		LLMInfos:   []model.LLMInfo{{Name: "gpt54pro", CanonicalName: "gpt-5.4-pro"}},
		Settings:   []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://snapshot-gateway.example.com"}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeSkip, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if res.Mode != model.DBImportModeSkip {
		t.Fatalf("mode = %q, want %q", res.Mode, model.DBImportModeSkip)
	}
	if !containsWarning(res.Compatibility.BaseURLMismatches, "preview-channel") {
		t.Fatalf("base_url_mismatches = %#v, want preview-channel mismatch", res.Compatibility.BaseURLMismatches)
	}
	if !containsWarning(res.Compatibility.SchemaMismatches, "snapshot schema:v2 differs") {
		t.Fatalf("schema_mismatches = %#v, want schema mismatch", res.Compatibility.SchemaMismatches)
	}
	if !containsWarning(res.Compatibility.AliasConflicts, "gpt54pro") {
		t.Fatalf("alias_conflicts = %#v, want gpt54pro alias conflict", res.Compatibility.AliasConflicts)
	}
	if got := res.Compatibility.Summary.AliasPreviewMaps; got == 0 {
		t.Fatalf("summary.alias_preview_mappings = %d, want > 0", got)
	}
	if len(res.Compatibility.AliasPreviewMappings) == 0 {
		t.Fatalf("alias_preview_mappings = %#v, want populated preview mappings", res.Compatibility.AliasPreviewMappings)
	}
	aliasPreview := res.Compatibility.AliasPreviewMappings[0]
	if aliasPreview.SnapshotModel != "gpt54pro" || aliasPreview.CurrentModel != "gpt-5.4-pro" {
		t.Fatalf("alias_preview_mapping = %#v, want gpt54pro -> gpt-5.4-pro", aliasPreview)
	}
	if len(aliasPreview.Contexts) == 0 {
		t.Fatalf("alias_preview_mapping.contexts = %#v, want populated contexts", aliasPreview)
	}
	if !containsWarning(res.Compatibility.RouteConflicts, "preview-group") {
		t.Fatalf("route_conflicts = %#v, want preview-group route conflict", res.Compatibility.RouteConflicts)
	}
	if !containsWarning(res.Compatibility.SkippedTargets, "setting:api_base_url existing row preserved by skip mode") {
		t.Fatalf("skipped_targets = %#v, want skip-mode setting preservation", res.Compatibility.SkippedTargets)
	}
	if got := res.Compatibility.Summary.BaseURLMismatches; got != 1 {
		t.Fatalf("summary.base_url_mismatches = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.SchemaMismatches; got != 1 {
		t.Fatalf("summary.schema_mismatches = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.RoutePreviewDiffs; got == 0 {
		t.Fatalf("summary.route_preview_diffs = %d, want > 0", got)
	}
	if len(res.Compatibility.RoutePreviewDiffs) == 0 {
		t.Fatalf("route_preview_diffs = %#v, want populated diffs", res.Compatibility.RoutePreviewDiffs)
	}
	diff := res.Compatibility.RoutePreviewDiffs[0]
	if diff.GroupName == "" || diff.Model == "" {
		t.Fatalf("route_preview_diff = %#v, want group/model", diff)
	}
	if len(diff.BeforeCandidates) == 0 {
		t.Fatalf("route_preview_diff candidates = %#v, want before populated", diff)
	}
	if len(diff.AfterCandidates) == 0 && len(diff.RemovedCandidates) == 0 {
		t.Fatalf("route_preview_diff candidates = %#v, want after or removed candidates", diff)
	}
	if len(diff.SkipReasons) == 0 {
		t.Fatalf("route_preview_diff skip_reasons = %#v, want structured reasons", diff)
	}
	if containsWarning(diff.SkipReasons, "missing_model") {
		t.Fatalf("route_preview_diff skip_reasons = %#v, do not want missing_model after alias remap", diff.SkipReasons)
	}
	previewCandidates := diff.AfterCandidates
	if len(previewCandidates) == 0 {
		previewCandidates = diff.RemovedCandidates
	}
	if len(previewCandidates) == 0 {
		t.Fatalf("route_preview_diff candidates = %#v, want alias-remapped candidate evidence", diff)
	}
	afterCandidate := previewCandidates[0]
	if !afterCandidate.Declared {
		t.Fatalf("preview candidate = %#v, want declared after alias remap", afterCandidate)
	}
	if containsWarning([]string{afterCandidate.Reason}, "missing_model") {
		t.Fatalf("preview candidate.reason = %q, do not want missing_model after alias remap", afterCandidate.Reason)
	}
	if !containsWarning(res.Compatibility.AliasPreviewMappings[0].Contexts, "channel:preview-channel") {
		t.Fatalf("alias_preview_mapping.contexts = %#v, want preview-channel context", res.Compatibility.AliasPreviewMappings[0].Contexts)
	}
}

func TestDBImportIncrementalDryRunRoutePreviewDiffDetectsResolvedKeyChanges(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-5.4-pro", CanonicalName: "gpt-5.4-pro"}).Error; err != nil {
		t.Fatalf("create current llm info error = %v", err)
	}

	currentChannel := model.Channel{
		Name:              "key-diff-channel",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModeClassified,
		BaseUrls:          []model.BaseUrl{{URL: "https://current.example.com/v1", Delay: 0}},
		Model:             "gpt-5.4-pro",
	}
	if err := db.GetDB().WithContext(ctx).Create(&currentChannel).Error; err != nil {
		t.Fatalf("create current channel error = %v", err)
	}
	legacyKey := model.ChannelKey{
		ChannelID:     currentChannel.ID,
		Enabled:       true,
		ChannelKey:    "legacy-key",
		SourceType:    "legacy/import",
		AllowedModels: "gpt54pro",
	}
	if err := db.GetDB().WithContext(ctx).Create(&legacyKey).Error; err != nil {
		t.Fatalf("create legacy channel key error = %v", err)
	}
	canonicalKey := model.ChannelKey{
		ChannelID:     currentChannel.ID,
		Enabled:       true,
		ChannelKey:    "canonical-key",
		SourceType:    "paid/metered",
		AllowedModels: "gpt-5.4-pro",
	}
	if err := db.GetDB().WithContext(ctx).Create(&canonicalKey).Error; err != nil {
		t.Fatalf("create canonical channel key error = %v", err)
	}

	currentGroup := model.Group{Name: "key-diff-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&currentGroup).Error; err != nil {
		t.Fatalf("create current group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: currentGroup.ID, ChannelID: currentChannel.ID, ModelName: "gpt54pro", Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create current group item error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{
			ID:                101,
			Name:              "key-diff-channel",
			Enabled:           true,
			KeyManagementMode: model.KeyManagementModeClassified,
			BaseUrls:          []model.BaseUrl{{URL: "https://snapshot.example.com/v1", Delay: 0}},
			Model:             "gpt54pro",
		}},
		Groups:     []model.Group{{ID: 201, Name: "key-diff-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 301, GroupID: 201, ChannelID: 101, ModelName: "gpt54pro", Priority: 1, Weight: 1}},
		LLMInfos:   []model.LLMInfo{{Name: "gpt54pro", CanonicalName: "gpt-5.4-pro"}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if got := res.Compatibility.Summary.RoutePreviewDiffs; got == 0 {
		t.Fatalf("summary.route_preview_diffs = %d, want > 0", got)
	}

	var diff *model.DBImportRoutePreviewDiff
	for i := range res.Compatibility.RoutePreviewDiffs {
		candidate := &res.Compatibility.RoutePreviewDiffs[i]
		if candidate.GroupName == "key-diff-group" && candidate.Model == "gpt54pro" {
			diff = candidate
			break
		}
	}
	if diff == nil {
		t.Fatalf("route_preview_diffs = %#v, want key-diff-group/gpt54pro diff", res.Compatibility.RoutePreviewDiffs)
	}
	if len(diff.BeforeCandidates) != 1 || len(diff.AfterCandidates) != 1 {
		t.Fatalf("route_preview_diff candidates = %#v, want exactly one before and one after candidate", diff)
	}

	beforeCandidate := diff.BeforeCandidates[0]
	afterCandidate := diff.AfterCandidates[0]
	if beforeCandidate.ResolvedModel != "gpt54pro" {
		t.Fatalf("before_candidate.resolved_model = %q, want gpt54pro", beforeCandidate.ResolvedModel)
	}
	if afterCandidate.ResolvedModel != "gpt-5.4-pro" {
		t.Fatalf("after_candidate.resolved_model = %q, want gpt-5.4-pro", afterCandidate.ResolvedModel)
	}
	if beforeCandidate.KeyID != legacyKey.ID {
		t.Fatalf("before_candidate.key_id = %d, want %d", beforeCandidate.KeyID, legacyKey.ID)
	}
	if afterCandidate.KeyID != canonicalKey.ID {
		t.Fatalf("after_candidate.key_id = %d, want %d", afterCandidate.KeyID, canonicalKey.ID)
	}
	if beforeCandidate.KeySourceType != "legacy/import" {
		t.Fatalf("before_candidate.key_source_type = %q, want legacy/import", beforeCandidate.KeySourceType)
	}
	if afterCandidate.KeySourceType != "paid/metered" {
		t.Fatalf("after_candidate.key_source_type = %q, want paid/metered", afterCandidate.KeySourceType)
	}
	if !containsWarning([]string{beforeCandidate.Reason}, "missing_model") {
		t.Fatalf("before_candidate.reason = %q, want missing_model evidence", beforeCandidate.Reason)
	}
	if !containsWarning([]string{afterCandidate.Reason}, "alias_remapped") {
		t.Fatalf("after_candidate.reason = %q, want alias_remapped evidence", afterCandidate.Reason)
	}
	if len(diff.RemovedCandidates) != 1 || diff.RemovedCandidates[0].KeyID != legacyKey.ID {
		t.Fatalf("removed_candidates = %#v, want legacy key candidate removed", diff.RemovedCandidates)
	}
	if len(diff.AddedCandidates) != 1 || diff.AddedCandidates[0].KeyID != canonicalKey.ID {
		t.Fatalf("added_candidates = %#v, want canonical key candidate added", diff.AddedCandidates)
	}
}

func TestDBImportIncrementalDryRunBuildsModelPolicyDiffs(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{
		Name:                  "gpt-5.4-pro",
		CanonicalName:         "gpt-5.4-pro",
		BillingMode:           model.BillingModePerToken,
		ProbePolicy:           model.ProbePolicyConcurrent,
		ProbeIntervalSeconds:  30,
		ProbeConcurrencyLimit: 3,
	}).Error; err != nil {
		t.Fatalf("create current llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion: "v1",
			ExportSource:  "octopus",
		},
		LLMInfos: []model.LLMInfo{{
			Name:                  "gpt54pro",
			CanonicalName:         "gpt-5.4-pro",
			BillingMode:           model.BillingModePerRequest,
			ProbePolicy:           model.ProbePolicyPassiveOnly,
			ProbeIntervalSeconds:  600,
			ProbeConcurrencyLimit: 1,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeSkip, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if got := res.Compatibility.Summary.ModelPolicyDiffs; got == 0 {
		t.Fatalf("summary.model_policy_diffs = %d, want > 0", got)
	}
	if len(res.Compatibility.ModelPolicyDiffs) == 0 {
		t.Fatalf("model_policy_diffs = %#v, want populated policy diffs", res.Compatibility.ModelPolicyDiffs)
	}
	policyDiff := res.Compatibility.ModelPolicyDiffs[0]
	if policyDiff.Model != "gpt54pro" || policyDiff.CurrentModel != "gpt-5.4-pro" {
		t.Fatalf("model_policy_diff = %#v, want gpt54pro -> gpt-5.4-pro", policyDiff)
	}
	if !containsWarning(policyDiff.ChangedFields, "billing_mode") {
		t.Fatalf("changed_fields = %#v, want billing_mode", policyDiff.ChangedFields)
	}
	if !containsWarning(policyDiff.ChangedFields, "probe_policy") {
		t.Fatalf("changed_fields = %#v, want probe_policy", policyDiff.ChangedFields)
	}
	if !containsWarning(policyDiff.ChangedFields, "probe_interval") {
		t.Fatalf("changed_fields = %#v, want probe_interval", policyDiff.ChangedFields)
	}
	if !containsWarning(policyDiff.ChangedFields, "probe_concurrency") {
		t.Fatalf("changed_fields = %#v, want probe_concurrency", policyDiff.ChangedFields)
	}
	if policyDiff.ImpactLevel != "high" {
		t.Fatalf("impact_level = %q, want high", policyDiff.ImpactLevel)
	}
	if len(policyDiff.Warnings) == 0 {
		t.Fatalf("warnings = %#v, want populated policy warnings", policyDiff.Warnings)
	}
}

func TestDBImportIncrementalSkipModePreservesExistingRows(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "shared-channel", Enabled: true, Model: "existing-model"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	existingGroup := model.Group{Name: "shared-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&existingGroup).Error; err != nil {
		t.Fatalf("create existing group error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://existing.example.com")
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "existing-model", CanonicalName: "existing-model"}).Error; err != nil {
		t.Fatalf("create existing llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels:    []model.Channel{{ID: 1001, Name: "shared-channel", Enabled: true, Model: "snapshot-model"}},
		Groups:      []model.Group{{ID: 1002, Name: "shared-group", Mode: model.GroupModeWeighted}},
		Settings:    []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://snapshot.example.com"}},
		LLMInfos:    []model.LLMInfo{{Name: "existing-model", CanonicalName: "existing-model", BillingMode: model.BillingModePerToken}},
		ChannelKeys: []model.ChannelKey{{ID: 1003, ChannelID: 1001, Enabled: true, ChannelKey: "sk-snapshot-upstream"}},
		GroupItems:  []model.GroupItem{{ID: 1004, GroupID: 1002, ChannelID: 1001, ModelName: "snapshot-model", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeSkip, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=skip, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["channels"]; got != 0 {
		t.Fatalf("rows_affected[channels] = %d, want 0", got)
	}
	if got := res.RowsAffected["groups"]; got != 0 {
		t.Fatalf("rows_affected[groups] = %d, want 0", got)
	}
	if got := res.RowsAffected["settings"]; got != 0 {
		t.Fatalf("rows_affected[settings] = %d, want 0", got)
	}
	if got := res.RowsAffected["llm_infos"]; got != 0 {
		t.Fatalf("rows_affected[llm_infos] = %d, want 0", got)
	}
	if got := res.RowsAffected["channel_keys"]; got != 0 {
		t.Fatalf("rows_affected[channel_keys] = %d, want 0", got)
	}
	if got := res.RowsAffected["group_items"]; got != 0 {
		t.Fatalf("rows_affected[group_items] = %d, want 0", got)
	}
	if !containsWarning(res.Warnings, "skip mode preserved existing channel:shared-channel") {
		t.Fatalf("warnings = %#v, want skip mode channel warning", res.Warnings)
	}
	if !containsWarning(res.Warnings, "skip mode preserved existing group:shared-group") {
		t.Fatalf("warnings = %#v, want skip mode group warning", res.Warnings)
	}

	var setting model.Setting
	if err := db.GetDB().WithContext(ctx).First(&setting, "key = ?", model.SettingKeyAPIBaseURL).Error; err != nil {
		t.Fatalf("query setting error = %v", err)
	}
	if setting.Value != "https://existing.example.com" {
		t.Fatalf("setting value = %q, want existing value", setting.Value)
	}

	var channel model.Channel
	if err := db.GetDB().WithContext(ctx).First(&channel, "name = ?", "shared-channel").Error; err != nil {
		t.Fatalf("query channel error = %v", err)
	}
	if channel.Model != "existing-model" {
		t.Fatalf("channel model = %q, want existing-model", channel.Model)
	}

	var channelKeys []model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Find(&channelKeys).Error; err != nil {
		t.Fatalf("query channel keys error = %v", err)
	}
	if len(channelKeys) != 0 {
		t.Fatalf("channel keys = %#v, want none imported in skip mode", channelKeys)
	}
}

func TestDBImportIncrementalMergeModeUpdatesExistingRouteItems(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := model.Channel{Name: "merge-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	group := model.Group{Name: "merge-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&group).Error; err != nil {
		t.Fatalf("create group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 9, Weight: 9}).Error; err != nil {
		t.Fatalf("create existing group item error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 1001, Name: "merge-channel", Enabled: true, Model: "gpt-4o"}},
		Groups:   []model.Group{{ID: 1002, Name: "merge-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{
			ID:        1003,
			GroupID:   1002,
			ChannelID: 1001,
			ModelName: "gpt-4o",
			Priority:  1,
			Weight:    2,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeMerge, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=merge, dryRun=false) error = %v", err)
	}
	if res.Mode != model.DBImportModeMerge {
		t.Fatalf("mode = %q, want %q", res.Mode, model.DBImportModeMerge)
	}
	if got := res.RowsAffected["group_items"]; got == 0 {
		t.Fatalf("rows_affected[group_items] = %d, want > 0", got)
	}

	var stored model.GroupItem
	if err := db.GetDB().WithContext(ctx).First(&stored, "group_id = ? AND channel_id = ? AND model_name = ?", group.ID, channel.ID, "gpt-4o").Error; err != nil {
		t.Fatalf("query stored group item error = %v", err)
	}
	if stored.Priority != 1 || stored.Weight != 2 {
		t.Fatalf("stored group item = %#v, want priority=1 weight=2 after merge upsert", stored)
	}
}

func TestDBImportIncrementalReplaceModeRebuildsChannelKeysAndGroupItems(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := model.Channel{Name: "replace-channel", Enabled: true, Model: "gpt-4o,o1-mini"}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	group := model.Group{Name: "replace-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&group).Error; err != nil {
		t.Fatalf("create group error = %v", err)
	}
	legacyKey := model.ChannelKey{ChannelID: channel.ID, Enabled: true, ChannelKey: "legacy-key", AllowedModels: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&legacyKey).Error; err != nil {
		t.Fatalf("create legacy key error = %v", err)
	}
	legacyItem := model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "o1-mini", Priority: 3, Weight: 1}
	if err := db.GetDB().WithContext(ctx).Create(&legacyItem).Error; err != nil {
		t.Fatalf("create legacy group item error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 2001, Name: "replace-channel", Enabled: true, Model: "gpt-4o"}},
		ChannelKeys: []model.ChannelKey{{
			ID:            2101,
			ChannelID:     2001,
			Enabled:       true,
			ChannelKey:    "snapshot-key",
			AllowedModels: "gpt-4o",
		}},
		Groups: []model.Group{{ID: 2201, Name: "replace-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{
			ID:        2301,
			GroupID:   2201,
			ChannelID: 2001,
			ModelName: "gpt-4o",
			Priority:  1,
			Weight:    5,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=false) error = %v", err)
	}
	if res.Mode != model.DBImportModeReplace {
		t.Fatalf("mode = %q, want %q", res.Mode, model.DBImportModeReplace)
	}
	if got := res.RowsAffected["replaced_channel_keys"]; got == 0 {
		t.Fatalf("rows_affected[replaced_channel_keys] = %d, want > 0", got)
	}
	if got := res.RowsAffected["replaced_group_items"]; got == 0 {
		t.Fatalf("rows_affected[replaced_group_items] = %d, want > 0", got)
	}

	var channelKeys []model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&channelKeys, "channel_id = ?", channel.ID).Error; err != nil {
		t.Fatalf("query channel keys error = %v", err)
	}
	if len(channelKeys) != 1 || channelKeys[0].ChannelKey != "snapshot-key" {
		t.Fatalf("channel keys = %#v, want only snapshot-key after replace", channelKeys)
	}

	var groupItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Order("priority asc").Find(&groupItems, "group_id = ?", group.ID).Error; err != nil {
		t.Fatalf("query group items error = %v", err)
	}
	if len(groupItems) != 1 || groupItems[0].ModelName != "gpt-4o" || groupItems[0].Priority != 1 || groupItems[0].Weight != 5 {
		t.Fatalf("group items = %#v, want only snapshot gpt-4o route after replace", groupItems)
	}
}

func TestDBImportIncrementalReplaceModePrunesChannelsAndGroupsMissingFromSnapshot(t *testing.T) {
	ctx := setupOpTestDB(t)

	keepChannel := model.Channel{Name: "keep-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&keepChannel).Error; err != nil {
		t.Fatalf("create keep channel error = %v", err)
	}
	staleChannel := model.Channel{Name: "stale-channel", Enabled: true, Model: "o1-mini"}
	if err := db.GetDB().WithContext(ctx).Create(&staleChannel).Error; err != nil {
		t.Fatalf("create stale channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.ChannelKey{ChannelID: staleChannel.ID, Enabled: true, ChannelKey: "stale-key", AllowedModels: "o1-mini"}).Error; err != nil {
		t.Fatalf("create stale channel key error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.RouteTargetOverride{ChannelID: staleChannel.ID, ChannelKeyID: 9999, ModelName: "o1-mini", BillingMode: model.BillingModePerRequest}).Error; err == nil {
		// no-op; this row likely fails on unique semantics with fake key id, so keep explicit key-based override creation below
	}
	staleChannelKey := model.ChannelKey{ChannelID: staleChannel.ID, Enabled: true, ChannelKey: "stale-route-key", AllowedModels: "o1-mini"}
	if err := db.GetDB().WithContext(ctx).Create(&staleChannelKey).Error; err != nil {
		t.Fatalf("create stale route channel key error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.RouteTargetOverride{ChannelID: staleChannel.ID, ChannelKeyID: staleChannelKey.ID, ModelName: "o1-mini", BillingMode: model.BillingModePerRequest}).Error; err != nil {
		t.Fatalf("create stale route target override error = %v", err)
	}

	keepGroup := model.Group{Name: "keep-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&keepGroup).Error; err != nil {
		t.Fatalf("create keep group error = %v", err)
	}
	staleGroup := model.Group{Name: "stale-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&staleGroup).Error; err != nil {
		t.Fatalf("create stale group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: staleGroup.ID, ChannelID: staleChannel.ID, ModelName: "o1-mini", Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create stale group item error = %v", err)
	}

	dump := &model.DBDump{
		Version:     dbDumpVersion,
		Manifest:    model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels:    []model.Channel{{ID: 2001, Name: "keep-channel", Enabled: true, Model: "gpt-4o"}},
		ChannelKeys: []model.ChannelKey{{ID: 2101, ChannelID: 2001, Enabled: true, ChannelKey: "keep-key", AllowedModels: "gpt-4o"}},
		Groups:      []model.Group{{ID: 2201, Name: "keep-group", Mode: model.GroupModeRoundRobin}},
		GroupItems:  []model.GroupItem{{ID: 2301, GroupID: 2201, ChannelID: 2001, ModelName: "gpt-4o", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["replaced_channels"]; got < 1 {
		t.Fatalf("rows_affected[replaced_channels] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["replaced_groups"]; got < 1 {
		t.Fatalf("rows_affected[replaced_groups] = %d, want >= 1", got)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels error = %v", err)
	}
	if len(channels) != 1 || channels[0].Name != "keep-channel" {
		t.Fatalf("channels = %#v, want only keep-channel", channels)
	}

	var groups []model.Group
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&groups).Error; err != nil {
		t.Fatalf("query groups error = %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "keep-group" {
		t.Fatalf("groups = %#v, want only keep-group", groups)
	}

	var staleChannelKeys []model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Where("channel_id = ?", staleChannel.ID).Find(&staleChannelKeys).Error; err != nil {
		t.Fatalf("query stale channel keys error = %v", err)
	}
	if len(staleChannelKeys) != 0 {
		t.Fatalf("stale channel keys = %#v, want none", staleChannelKeys)
	}

	var staleOverrides []model.RouteTargetOverride
	if err := db.GetDB().WithContext(ctx).Where("channel_id = ?", staleChannel.ID).Find(&staleOverrides).Error; err != nil {
		t.Fatalf("query stale route target overrides error = %v", err)
	}
	if len(staleOverrides) != 0 {
		t.Fatalf("stale route target overrides = %#v, want none", staleOverrides)
	}

	var staleGroupItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Where("group_id = ?", staleGroup.ID).Find(&staleGroupItems).Error; err != nil {
		t.Fatalf("query stale group items error = %v", err)
	}
	if len(staleGroupItems) != 0 {
		t.Fatalf("stale group items = %#v, want none", staleGroupItems)
	}
}

func TestDBImportIncrementalReplaceModePrunesSettingsMissingFromSnapshot(t *testing.T) {
	ctx := setupOpTestDB(t)

	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://existing.example.com")
	setTestSetting(t, ctx, model.SettingKeyProxyURL, "http://stale-proxy")

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://snapshot.example.com"}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["replaced_settings"]; got < 1 {
		t.Fatalf("rows_affected[replaced_settings] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["settings"]; got != 1 {
		t.Fatalf("rows_affected[settings] = %d, want 1", got)
	}

	var apiBaseURLSetting model.Setting
	if err := db.GetDB().WithContext(ctx).First(&apiBaseURLSetting, "key = ?", model.SettingKeyAPIBaseURL).Error; err != nil {
		t.Fatalf("query api_base_url setting error = %v", err)
	}
	if apiBaseURLSetting.Value != "https://snapshot.example.com" {
		t.Fatalf("api_base_url value = %q, want snapshot value", apiBaseURLSetting.Value)
	}
	var proxySetting model.Setting
	if err := db.GetDB().WithContext(ctx).First(&proxySetting, "key = ?", model.SettingKeyProxyURL).Error; err != nil {
		t.Fatalf("query proxy_url setting error = %v", err)
	}
	if proxySetting.Value != "" {
		t.Fatalf("proxy_url value = %q, want default empty after replace pruning", proxySetting.Value)
	}
}

func TestDBImportIncrementalReplaceModePrunesLLMInfosMissingFromSnapshot(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "existing-model", CanonicalName: "existing-model", BillingMode: model.BillingModePerToken}).Error; err != nil {
		t.Fatalf("create existing llm info error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "stale-model", CanonicalName: "stale-model", BillingMode: model.BillingModePerRequest}).Error; err != nil {
		t.Fatalf("create stale llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		LLMInfos: []model.LLMInfo{{Name: "existing-model", CanonicalName: "existing-model", BillingMode: model.BillingModeFlat}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["replaced_llm_infos"]; got < 1 {
		t.Fatalf("rows_affected[replaced_llm_infos] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["llm_infos"]; got != 1 {
		t.Fatalf("rows_affected[llm_infos] = %d, want 1", got)
	}

	var llmInfos []model.LLMInfo
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&llmInfos).Error; err != nil {
		t.Fatalf("query llm infos error = %v", err)
	}
	if len(llmInfos) != 1 || llmInfos[0].Name != "existing-model" {
		t.Fatalf("llm_infos = %#v, want only existing-model after replace", llmInfos)
	}
	if llmInfos[0].BillingMode != model.BillingModeFlat {
		t.Fatalf("llm_infos[0].billing_mode = %q, want flat from snapshot", llmInfos[0].BillingMode)
	}
}

func TestDBImportIncrementalMergeModeReusesExistingChannelKeyIDAndRemapsStatsAndLogs(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "identity-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	existingKey := model.ChannelKey{
		ChannelID:     existingChannel.ID,
		Enabled:       false,
		ChannelKey:    "shared-upstream-key",
		SourceType:    "legacy/import",
		AllowedModels: "legacy-model",
		Remark:        "local-only",
	}
	if err := db.GetDB().WithContext(ctx).Create(&existingKey).Error; err != nil {
		t.Fatalf("create existing channel key error = %v", err)
	}

	dump := &model.DBDump{
		Version:      dbDumpVersion,
		IncludeStats: true,
		IncludeLogs:  true,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:      501,
			Name:    "identity-channel",
			Enabled: true,
			Model:   "gpt-4o",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            601,
			ChannelID:     501,
			Enabled:       true,
			ChannelKey:    "shared-upstream-key",
			SourceType:    "PAID/METERED",
			AllowedModels: "gpt-4o",
			Remark:        "snapshot-managed",
		}},
		StatsChannel: []model.StatsChannel{{
			ChannelID: 999,
			StatsMetrics: model.StatsMetrics{
				InputToken:     10,
				RequestSuccess: 1,
			},
		}, {
			ChannelID: 501,
			StatsMetrics: model.StatsMetrics{
				InputToken:    22,
				OutputToken:   7,
				OutputCost:    0.12,
				RequestFailed: 1,
			},
		}},
		RelayLogs: []model.RelayLog{{
			ID:               9001,
			Time:             1710000000,
			RequestModelName: "gpt-4o",
			ChannelId:        501,
			ChannelName:      "identity-channel",
			ActualModelName:  "gpt-4o",
			Attempts: []model.ChannelAttempt{{
				ChannelID:    501,
				ChannelKeyID: 601,
				ChannelName:  "identity-channel",
				ModelName:    "gpt-4o",
				AttemptNum:   1,
				Status:       model.AttemptSuccess,
				Duration:     12,
			}, {
				ChannelID:    501,
				ChannelKeyID: 9999,
				ChannelName:  "identity-channel",
				ModelName:    "gpt-4o",
				AttemptNum:   2,
				Status:       model.AttemptSkipped,
				Duration:     1,
				Msg:          "stale snapshot key reference",
			}},
			TotalAttempts: 2,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeMerge, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=merge, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["channel_keys"]; got != 1 {
		t.Fatalf("rows_affected[channel_keys] = %d, want 1 updated/reused row", got)
	}
	if got := res.RowsAffected["stats_channel"]; got != 1 {
		t.Fatalf("rows_affected[stats_channel] = %d, want 1 resolvable remapped row", got)
	}
	if got := res.RowsAffected["relay_logs"]; got != 1 {
		t.Fatalf("rows_affected[relay_logs] = %d, want 1", got)
	}
	if !containsWarning(res.Warnings, "skipped stats_channel for snapshot channel_id=999 because it could not be resolved after import") {
		t.Fatalf("warnings = %#v, want unresolved stats_channel warning", res.Warnings)
	}
	if !containsWarning(res.Warnings, "reset 1 relay log attempt channel key references that could not be resolved after import") {
		t.Fatalf("warnings = %#v, want relay log attempt key reset warning", res.Warnings)
	}

	var storedKeys []model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&storedKeys, "channel_id = ?", existingChannel.ID).Error; err != nil {
		t.Fatalf("query channel keys error = %v", err)
	}
	if len(storedKeys) != 1 {
		t.Fatalf("stored channel keys = %#v, want exactly one reused row", storedKeys)
	}
	if storedKeys[0].ID != existingKey.ID {
		t.Fatalf("stored channel key id = %d, want existing local id %d", storedKeys[0].ID, existingKey.ID)
	}
	if storedKeys[0].ID == 601 {
		t.Fatalf("stored channel key unexpectedly reused snapshot id %d", storedKeys[0].ID)
	}
	if !storedKeys[0].Enabled || storedKeys[0].SourceType != "paid/metered" || storedKeys[0].AllowedModels != "gpt-4o" || storedKeys[0].Remark != "snapshot-managed" {
		t.Fatalf("stored channel key = %#v, want merged snapshot fields on reused local row", storedKeys[0])
	}

	var statsChannelRows []model.StatsChannel
	if err := db.GetDB().WithContext(ctx).Order("channel_id asc").Find(&statsChannelRows).Error; err != nil {
		t.Fatalf("query stats_channel error = %v", err)
	}
	if len(statsChannelRows) != 1 {
		t.Fatalf("stats_channel rows = %#v, want exactly one remapped row", statsChannelRows)
	}
	if statsChannelRows[0].ChannelID != existingChannel.ID {
		t.Fatalf("stats_channel.channel_id = %d, want local channel id %d", statsChannelRows[0].ChannelID, existingChannel.ID)
	}
	if statsChannelRows[0].InputToken != 22 || statsChannelRows[0].OutputToken != 7 || statsChannelRows[0].RequestFailed != 1 {
		t.Fatalf("stats_channel row = %#v, want remapped metrics from resolvable snapshot row", statsChannelRows[0])
	}

	var storedLog model.RelayLog
	if err := db.GetDB().WithContext(ctx).First(&storedLog, "id = ?", int64(9001)).Error; err != nil {
		t.Fatalf("query relay log error = %v", err)
	}
	if storedLog.ChannelId != existingChannel.ID {
		t.Fatalf("relay_log.channel_id = %d, want local channel id %d", storedLog.ChannelId, existingChannel.ID)
	}
	if len(storedLog.Attempts) != 2 {
		t.Fatalf("relay_log.attempts = %#v, want 2 attempts preserved", storedLog.Attempts)
	}
	if storedLog.Attempts[0].ChannelID != existingChannel.ID || storedLog.Attempts[0].ChannelKeyID != existingKey.ID {
		t.Fatalf("relay_log first attempt = %#v, want remapped local channel/key ids", storedLog.Attempts[0])
	}
	if storedLog.Attempts[1].ChannelID != existingChannel.ID || storedLog.Attempts[1].ChannelKeyID != 0 {
		t.Fatalf("relay_log second attempt = %#v, want local channel id and reset key id", storedLog.Attempts[1])
	}
}

func TestDBImportIncrementalMergeModeReusesExistingAPIKeyIDAndRemapsStats(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingAPIKey := model.APIKey{
		Name:            "local-client",
		APIKey:          "sk-shared-client",
		Enabled:         false,
		SupportedModels: "legacy-model",
	}
	if err := db.GetDB().WithContext(ctx).Create(&existingAPIKey).Error; err != nil {
		t.Fatalf("create existing api key error = %v", err)
	}

	dump := &model.DBDump{
		Version:      dbDumpVersion,
		IncludeStats: true,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		APIKeys: []model.APIKey{{
			ID:              701,
			Name:            "snapshot-client",
			APIKey:          "sk-shared-client",
			Enabled:         true,
			SupportedModels: "GPT-4O, O1-mini",
			MaxCost:         12.5,
		}},
		StatsAPIKey: []model.StatsAPIKey{{
			APIKeyID: 999,
			StatsMetrics: model.StatsMetrics{
				InputToken:    5,
				RequestFailed: 1,
			},
		}, {
			APIKeyID: 701,
			StatsMetrics: model.StatsMetrics{
				InputToken:     42,
				OutputToken:    8,
				RequestSuccess: 3,
				InputCost:      0.23,
			},
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeMerge, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=merge, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["api_keys"]; got != 1 {
		t.Fatalf("rows_affected[api_keys] = %d, want 1 updated/reused row", got)
	}
	if got := res.RowsAffected["stats_api_key"]; got != 1 {
		t.Fatalf("rows_affected[stats_api_key] = %d, want 1 resolvable remapped row", got)
	}
	if !containsWarning(res.Warnings, "skipped stats_api_key for snapshot api_key_id=999 because it could not be resolved after import") {
		t.Fatalf("warnings = %#v, want unresolved stats_api_key warning", res.Warnings)
	}

	var storedAPIKeys []model.APIKey
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&storedAPIKeys).Error; err != nil {
		t.Fatalf("query api keys error = %v", err)
	}
	if len(storedAPIKeys) != 1 {
		t.Fatalf("stored api keys = %#v, want exactly one reused row", storedAPIKeys)
	}
	if storedAPIKeys[0].ID != existingAPIKey.ID {
		t.Fatalf("stored api key id = %d, want existing local id %d", storedAPIKeys[0].ID, existingAPIKey.ID)
	}
	if storedAPIKeys[0].ID == 701 {
		t.Fatalf("stored api key unexpectedly reused snapshot id %d", storedAPIKeys[0].ID)
	}
	if storedAPIKeys[0].Name != "snapshot-client" || !storedAPIKeys[0].Enabled || storedAPIKeys[0].SupportedModels != "gpt-4o,o1-mini" || storedAPIKeys[0].MaxCost != 12.5 {
		t.Fatalf("stored api key = %#v, want merged snapshot fields on reused local row", storedAPIKeys[0])
	}

	var statsAPIKeyRows []model.StatsAPIKey
	if err := db.GetDB().WithContext(ctx).Order("api_key_id asc").Find(&statsAPIKeyRows).Error; err != nil {
		t.Fatalf("query stats_api_key error = %v", err)
	}
	if len(statsAPIKeyRows) != 1 {
		t.Fatalf("stats_api_key rows = %#v, want exactly one remapped row", statsAPIKeyRows)
	}
	if statsAPIKeyRows[0].APIKeyID != existingAPIKey.ID {
		t.Fatalf("stats_api_key.api_key_id = %d, want local api key id %d", statsAPIKeyRows[0].APIKeyID, existingAPIKey.ID)
	}
	if statsAPIKeyRows[0].InputToken != 42 || statsAPIKeyRows[0].OutputToken != 8 || statsAPIKeyRows[0].RequestSuccess != 3 {
		t.Fatalf("stats_api_key row = %#v, want remapped metrics from resolvable snapshot row", statsAPIKeyRows[0])
	}
}

func TestDBImportIncrementalDryRunReplacePreviewShowsPrunedObjects(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "keep-channel", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create keep channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "stale-channel", Enabled: true, Model: "o1-mini"}).Error; err != nil {
		t.Fatalf("create stale channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.Group{Name: "keep-group", Mode: model.GroupModeRoundRobin}).Error; err != nil {
		t.Fatalf("create keep group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.Group{Name: "stale-group", Mode: model.GroupModeRoundRobin}).Error; err != nil {
		t.Fatalf("create stale group error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://keep.example.com")
	setTestSetting(t, ctx, model.SettingKeyProxyURL, "http://stale-proxy")
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "keep-model", CanonicalName: "keep-model"}).Error; err != nil {
		t.Fatalf("create keep llm info error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "stale-model", CanonicalName: "stale-model"}).Error; err != nil {
		t.Fatalf("create stale llm info error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create keep api key error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "stale-client", APIKey: "sk-stale-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create stale api key error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 1, Name: "keep-channel", Enabled: true, Model: "gpt-4o"}},
		Groups:   []model.Group{{ID: 2, Name: "keep-group", Mode: model.GroupModeRoundRobin}},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://snapshot.example.com"}},
		LLMInfos: []model.LLMInfo{{Name: "keep-model", CanonicalName: "keep-model"}},
		APIKeys:  []model.APIKey{{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=true) error = %v", err)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedChannels, "stale-channel") {
		t.Fatalf("replace_pruned_channels = %#v, want stale-channel", res.Compatibility.ReplacePrunedChannels)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedGroups, "stale-group") {
		t.Fatalf("replace_pruned_groups = %#v, want stale-group", res.Compatibility.ReplacePrunedGroups)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedSettings, string(model.SettingKeyProxyURL)) {
		t.Fatalf("replace_pruned_settings = %#v, want proxy_url", res.Compatibility.ReplacePrunedSettings)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedLLMInfos, "stale-model") {
		t.Fatalf("replace_pruned_llm_infos = %#v, want stale-model", res.Compatibility.ReplacePrunedLLMInfos)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedAPIKeys, "stale-client") {
		t.Fatalf("replace_pruned_api_keys = %#v, want stale-client", res.Compatibility.ReplacePrunedAPIKeys)
	}
	if got := res.Compatibility.Summary.ReplacePrunedChannels; got != 1 {
		t.Fatalf("summary.replace_pruned_channels = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.ReplacePrunedGroups; got != 1 {
		t.Fatalf("summary.replace_pruned_groups = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.ReplacePrunedSettings; got == 0 {
		t.Fatalf("summary.replace_pruned_settings = %d, want > 0", got)
	}
	if got := res.Compatibility.Summary.ReplacePrunedLLMInfos; got != 1 {
		t.Fatalf("summary.replace_pruned_llm_infos = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.ReplacePrunedAPIKeys; got != 1 {
		t.Fatalf("summary.replace_pruned_api_keys = %d, want 1", got)
	}
}

func TestDBImportIncrementalDryRunReplacePreviewHonorsSelectiveScopes(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "stale-channel", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create stale channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.Group{Name: "stale-group", Mode: model.GroupModeRoundRobin}).Error; err != nil {
		t.Fatalf("create stale group error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyProxyURL, "http://stale-proxy")
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "stale-client", APIKey: "sk-stale-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create stale api key error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 1, Name: "incoming-channel", Enabled: true, Model: "gpt-4o"}},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://incoming.example.com"}},
		APIKeys:  []model.APIKey{{Name: "incoming-client", APIKey: "sk-incoming-client", Enabled: true}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeReplace, true, model.DBImportOptions{
		ImportScopes: &model.DBImportScopes{Settings: true},
	})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=replace, dryRun=true, import_scopes=settings) error = %v", err)
	}
	if len(res.Compatibility.ReplacePrunedChannels) != 0 {
		t.Fatalf("replace_pruned_channels = %#v, want none when routing scope disabled", res.Compatibility.ReplacePrunedChannels)
	}
	if len(res.Compatibility.ReplacePrunedGroups) != 0 {
		t.Fatalf("replace_pruned_groups = %#v, want none when routing scope disabled", res.Compatibility.ReplacePrunedGroups)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedSettings, string(model.SettingKeyProxyURL)) {
		t.Fatalf("replace_pruned_settings = %#v, want proxy_url", res.Compatibility.ReplacePrunedSettings)
	}
	if len(res.Compatibility.ReplacePrunedAPIKeys) != 0 {
		t.Fatalf("replace_pruned_api_keys = %#v, want none when api_keys scope disabled", res.Compatibility.ReplacePrunedAPIKeys)
	}
}

func TestDBImportIncrementalDryRunReplacePreviewShowsAPIKeyPruneWhenSecretsComeFromChannelConfig(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create keep api key error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "stale-client", APIKey: "sk-stale-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create stale api key error = %v", err)
	}
	proxy := "http://secret-proxy"
	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 1, Name: "incoming-channel", Enabled: true, Model: "gpt-4o", ChannelProxy: &proxy}},
		APIKeys:  []model.APIKey{{Name: "redacted-client", APIKey: "", Enabled: true}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=true) error = %v", err)
	}
	if !containsWarning(res.Compatibility.ReplacePrunedAPIKeys, "keep-client") || !containsWarning(res.Compatibility.ReplacePrunedAPIKeys, "stale-client") {
		t.Fatalf("replace_pruned_api_keys = %#v, want both current API keys because keep-set is empty", res.Compatibility.ReplacePrunedAPIKeys)
	}
}

func TestDBPreviewRollbackImportSnapshotRejectsEmptyImportScopes(t *testing.T) {
	ctx := setupOpTestDB(t)
	snapshotName := createImportSnapshotForRollbackScopeTest(t, ctx)

	preview, err := DBPreviewRollbackImportSnapshot(ctx, snapshotName, &model.DBImportScopes{})
	if err == nil {
		t.Fatalf("DBPreviewRollbackImportSnapshot() preview = %#v, want validation error", preview)
	}
	if err.Error() != "at least one import scope must be enabled" {
		t.Fatalf("DBPreviewRollbackImportSnapshot() error = %v, want at least one import scope must be enabled", err)
	}
}

func TestDBRollbackImportSnapshotRejectsEmptyImportScopes(t *testing.T) {
	ctx := setupOpTestDB(t)
	snapshotName := createImportSnapshotForRollbackScopeTest(t, ctx)

	result, err := DBRollbackImportSnapshot(ctx, snapshotName, &model.DBImportScopes{})
	if err == nil {
		t.Fatalf("DBRollbackImportSnapshot() result = %#v, want validation error", result)
	}
	if err.Error() != "at least one import scope must be enabled" {
		t.Fatalf("DBRollbackImportSnapshot() error = %v, want at least one import scope must be enabled", err)
	}
}

func createImportSnapshotForRollbackScopeTest(t *testing.T, ctx context.Context) string {
	t.Helper()

	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://rollback-scope.example.com"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}
	snapshots, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("DBListImportSnapshots() returned no snapshots")
	}
	return snapshots[0].SnapshotName
}

func TestDBImportIncrementalMapModeAppliesModelMappingsToRoutePreviewAndImport(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-4o", CanonicalName: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create llm info error = %v", err)
	}

	currentChannel := model.Channel{
		Name:     "mapped-channel",
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "https://mapped.example.com/v1", Delay: 1}},
		Model:    "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&currentChannel).Error; err != nil {
		t.Fatalf("create current channel error = %v", err)
	}
	currentKey := model.ChannelKey{ChannelID: currentChannel.ID, Enabled: true, ChannelKey: "mapped-key", AllowedModels: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&currentKey).Error; err != nil {
		t.Fatalf("create current channel key error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:      1201,
			Name:    "mapped-channel",
			Enabled: true,
			Model:   "legacy-model",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            1202,
			ChannelID:     1201,
			Enabled:       true,
			ChannelKey:    "mapped-key",
			AllowedModels: "legacy-model",
		}},
		Groups:     []model.Group{{ID: 1203, Name: "mapped-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 1204, GroupID: 1203, ChannelID: 1201, ModelName: "legacy-model", Priority: 1, Weight: 1}},
	}

	dryRunWithoutMap, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeMap, true, model.DBImportOptions{})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=map, dryRun=true) without mappings error = %v", err)
	}
	if len(dryRunWithoutMap.Compatibility.RoutePreviewDiffs) == 0 {
		t.Fatalf("route_preview_diffs = %#v, want diff without mappings", dryRunWithoutMap.Compatibility.RoutePreviewDiffs)
	}
	withoutCandidate := dryRunWithoutMap.Compatibility.RoutePreviewDiffs[0].AfterCandidates[0]
	if withoutCandidate.ResolvedModel != "legacy-model" {
		t.Fatalf("resolved_model without mappings = %q, want legacy-model", withoutCandidate.ResolvedModel)
	}
	if !containsWarning([]string{withoutCandidate.Reason}, "missing_model") {
		t.Fatalf("candidate reason without mappings = %q, want missing_model", withoutCandidate.Reason)
	}

	mapOptions := model.DBImportOptions{ModelMappings: map[string]string{"legacy-model": "gpt-4o"}}
	dryRunWithMap, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeMap, true, mapOptions)
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=map, dryRun=true) with mappings error = %v", err)
	}
	if dryRunWithMap.Mode != model.DBImportModeMap {
		t.Fatalf("mode = %q, want %q", dryRunWithMap.Mode, model.DBImportModeMap)
	}
	if len(dryRunWithMap.Compatibility.RoutePreviewDiffs) == 0 {
		t.Fatalf("route_preview_diffs = %#v, want diff with mappings", dryRunWithMap.Compatibility.RoutePreviewDiffs)
	}
	withCandidate := dryRunWithMap.Compatibility.RoutePreviewDiffs[0].AfterCandidates[0]
	if withCandidate.Model != "gpt-4o" || withCandidate.ResolvedModel != "gpt-4o" {
		t.Fatalf("candidate with mappings = %#v, want mapped model gpt-4o", withCandidate)
	}
	if withCandidate.KeyID != currentKey.ID {
		t.Fatalf("candidate key id with mappings = %d, want %d", withCandidate.KeyID, currentKey.ID)
	}
	if containsWarning([]string{withCandidate.Reason}, "missing_model") {
		t.Fatalf("candidate reason with mappings = %q, do not want missing_model after map", withCandidate.Reason)
	}
	if len(dryRunWithoutMap.Compatibility.ModelMappingPreviews) != 0 {
		t.Fatalf("model_mapping_previews without mappings = %#v, want empty", dryRunWithoutMap.Compatibility.ModelMappingPreviews)
	}
	if got := dryRunWithMap.Compatibility.Summary.ModelMappingPreviews; got != 1 {
		t.Fatalf("summary.model_mapping_previews = %d, want 1", got)
	}
	if got := dryRunWithMap.Compatibility.Summary.UsedModelMappings; got != 1 {
		t.Fatalf("summary.used_model_mappings = %d, want 1", got)
	}
	if got := dryRunWithMap.Compatibility.Summary.UnusedModelMappings; got != 0 {
		t.Fatalf("summary.unused_model_mappings = %d, want 0", got)
	}
	if got := dryRunWithMap.Compatibility.Summary.MissingMappingTargets; got != 0 {
		t.Fatalf("summary.missing_mapping_targets = %d, want 0", got)
	}
	if len(dryRunWithMap.Compatibility.ModelMappingPreviews) != 1 {
		t.Fatalf("model_mapping_previews = %#v, want one preview", dryRunWithMap.Compatibility.ModelMappingPreviews)
	}
	preview := dryRunWithMap.Compatibility.ModelMappingPreviews[0]
	if preview.SourceModel != "legacy-model" || preview.TargetModel != "gpt-4o" {
		t.Fatalf("model_mapping_preview = %#v, want legacy-model -> gpt-4o", preview)
	}
	if !preview.Used {
		t.Fatalf("model_mapping_preview.used = false, want true: %#v", preview)
	}
	if !preview.TargetExists {
		t.Fatalf("model_mapping_preview.target_exists = false, want true: %#v", preview)
	}
	if preview.UsageCount < 3 {
		t.Fatalf("model_mapping_preview.usage_count = %d, want >= 3", preview.UsageCount)
	}
	if !containsWarning(preview.Contexts, "channel:mapped-channel") {
		t.Fatalf("model_mapping_preview.contexts = %#v, want channel:mapped-channel", preview.Contexts)
	}
	if !containsWarning(preview.Contexts, "group:mapped-group") && !containsWarning(preview.Contexts, "group_route:mapped-group") {
		t.Fatalf("model_mapping_preview.contexts = %#v, want mapped-group context", preview.Contexts)
	}
	if !containsWarning(preview.TouchedFields, "channels.model") {
		t.Fatalf("model_mapping_preview.touched_fields = %#v, want channels.model", preview.TouchedFields)
	}
	if !containsWarning(preview.TouchedFields, "channel_keys.allowed_models") {
		t.Fatalf("model_mapping_preview.touched_fields = %#v, want channel_keys.allowed_models", preview.TouchedFields)
	}
	if !containsWarning(preview.TouchedFields, "group_items.model_name") {
		t.Fatalf("model_mapping_preview.touched_fields = %#v, want group_items.model_name", preview.TouchedFields)
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeMap, false, mapOptions)
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=map, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["group_items"]; got != 1 {
		t.Fatalf("rows_affected[group_items] = %d, want 1 imported mapped route", got)
	}

	var storedGroup model.Group
	if err := db.GetDB().WithContext(ctx).First(&storedGroup, "name = ?", "mapped-group").Error; err != nil {
		t.Fatalf("query mapped group error = %v", err)
	}
	var storedItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Order("priority asc").Find(&storedItems, "group_id = ?", storedGroup.ID).Error; err != nil {
		t.Fatalf("query mapped group items error = %v", err)
	}
	if len(storedItems) != 1 {
		t.Fatalf("group items = %#v, want one mapped item", storedItems)
	}
	if storedItems[0].ModelName != "gpt-4o" {
		t.Fatalf("stored group item model = %q, want gpt-4o after map", storedItems[0].ModelName)
	}
}

func TestDBImportIncrementalWithOptionsRejectsBlankModelMappings(t *testing.T) {
	ctx := setupOpTestDB(t)
	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
	}

	tests := []struct {
		name     string
		mappings map[string]string
	}{
		{name: "blank source", mappings: map[string]string{"   ": "gpt-4o"}},
		{name: "blank target", mappings: map[string]string{"legacy-model": "   "}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeMap, true, model.DBImportOptions{ModelMappings: tt.mappings})
			if err == nil || err.Error() != "invalid model_mappings" {
				t.Fatalf("DBImportIncrementalWithOptions() error = %v, want invalid model_mappings", err)
			}
		})
	}
}

func TestDBImportIncrementalMapModeReportsUnusedAndMissingMappingTargets(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-4o", CanonicalName: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:      2201,
			Name:    "preview-map-channel",
			Enabled: true,
			Model:   "legacy-model",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            2202,
			ChannelID:     2201,
			Enabled:       true,
			ChannelKey:    "preview-key",
			AllowedModels: "legacy-model",
		}},
		Groups:     []model.Group{{ID: 2203, Name: "preview-map-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 2204, GroupID: 2203, ChannelID: 2201, ModelName: "legacy-model", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeMap, true, model.DBImportOptions{ModelMappings: map[string]string{
		"legacy-model": "missing-target",
		"unused-model": "gpt-4o",
	}})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=map, dryRun=true) error = %v", err)
	}
	if got := res.Compatibility.Summary.ModelMappingPreviews; got != 2 {
		t.Fatalf("summary.model_mapping_previews = %d, want 2", got)
	}
	if got := res.Compatibility.Summary.UsedModelMappings; got != 1 {
		t.Fatalf("summary.used_model_mappings = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.UnusedModelMappings; got != 1 {
		t.Fatalf("summary.unused_model_mappings = %d, want 1", got)
	}
	if got := res.Compatibility.Summary.MissingMappingTargets; got != 1 {
		t.Fatalf("summary.missing_mapping_targets = %d, want 1", got)
	}
	if len(res.Compatibility.ModelMappingPreviews) != 2 {
		t.Fatalf("model_mapping_previews = %#v, want 2", res.Compatibility.ModelMappingPreviews)
	}
	previewsBySource := make(map[string]model.DBImportModelMappingPreview, len(res.Compatibility.ModelMappingPreviews))
	for _, preview := range res.Compatibility.ModelMappingPreviews {
		previewsBySource[preview.SourceModel] = preview
	}
	usedPreview, ok := previewsBySource["legacy-model"]
	if !ok {
		t.Fatalf("legacy-model preview missing from %#v", res.Compatibility.ModelMappingPreviews)
	}
	if !usedPreview.Used {
		t.Fatalf("legacy-model preview.used = false, want true: %#v", usedPreview)
	}
	if usedPreview.TargetExists {
		t.Fatalf("legacy-model preview.target_exists = true, want false: %#v", usedPreview)
	}
	if !containsWarning(usedPreview.Warnings, "mapped target not found") {
		t.Fatalf("legacy-model preview.warnings = %#v, want missing target warning", usedPreview.Warnings)
	}
	unusedPreview, ok := previewsBySource["unused-model"]
	if !ok {
		t.Fatalf("unused-model preview missing from %#v", res.Compatibility.ModelMappingPreviews)
	}
	if unusedPreview.Used {
		t.Fatalf("unused-model preview.used = true, want false: %#v", unusedPreview)
	}
	if !containsWarning(unusedPreview.Warnings, "not referenced") {
		t.Fatalf("unused-model preview.warnings = %#v, want unused mapping warning", unusedPreview.Warnings)
	}
}

func TestDBImportIncrementalDryRunReportsInvalidRouteTargetForUndeclaredModel(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-4o", CanonicalName: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version:     dbDumpVersion,
		Manifest:    model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels:    []model.Channel{{ID: 501, Name: "invalid-route-channel", Enabled: true, Model: "gpt-4o"}},
		ChannelKeys: []model.ChannelKey{{ID: 502, ChannelID: 501, Enabled: true, ChannelKey: "valid-key", AllowedModels: "gpt-4o"}},
		Groups:      []model.Group{{ID: 503, Name: "invalid-route-group", Mode: model.GroupModeRoundRobin}},
		GroupItems:  []model.GroupItem{{ID: 504, GroupID: 503, ChannelID: 501, ModelName: "o1-mini", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if got := res.Compatibility.Summary.InvalidRouteTargets; got != 1 {
		t.Fatalf("summary.invalid_route_targets = %d, want 1", got)
	}
	if len(res.Compatibility.InvalidRouteTargets) != 1 {
		t.Fatalf("invalid_route_targets = %#v, want 1", res.Compatibility.InvalidRouteTargets)
	}
	issue := res.Compatibility.InvalidRouteTargets[0]
	if issue.GroupName != "invalid-route-group" || issue.ChannelName != "invalid-route-channel" {
		t.Fatalf("invalid route issue = %#v, want group/channel names", issue)
	}
	if issue.Model != "o1-mini" || issue.IssueType != "invalid_route_target" {
		t.Fatalf("invalid route issue = %#v, want o1-mini invalid_route_target", issue)
	}
	if issue.Action != "undeclared_model" {
		t.Fatalf("invalid route issue.action = %q, want undeclared_model", issue.Action)
	}
	if !containsWarning([]string{issue.Reason}, "missing_model") {
		t.Fatalf("invalid route issue.reason = %q, want missing_model evidence", issue.Reason)
	}
	if !containsWarning(res.Compatibility.RoutePreviewWarnings, "invalid route targets: 1") {
		t.Fatalf("route_preview_warnings = %#v, want invalid route target warning", res.Compatibility.RoutePreviewWarnings)
	}
	if got := res.Compatibility.Summary.RoutePreviewWarnings; got != len(res.Compatibility.RoutePreviewWarnings) {
		t.Fatalf("summary.route_preview_warnings = %d, want %d", got, len(res.Compatibility.RoutePreviewWarnings))
	}
}

func TestDBImportIncrementalDryRunReportsInvalidRouteTargetForMissingKey(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-4o", CanonicalName: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version:     dbDumpVersion,
		Manifest:    model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: false},
		Channels:    []model.Channel{{ID: 601, Name: "missing-key-channel", Enabled: true, Model: "gpt-4o"}},
		ChannelKeys: []model.ChannelKey{{ID: 602, ChannelID: 601, Enabled: true, ChannelKey: "", AllowedModels: "gpt-4o"}},
		Groups:      []model.Group{{ID: 603, Name: "missing-key-group", Mode: model.GroupModeRoundRobin}},
		GroupItems:  []model.GroupItem{{ID: 604, GroupID: 603, ChannelID: 601, ModelName: "gpt-4o", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}
	if got := res.Compatibility.Summary.InvalidRouteTargets; got != 1 {
		t.Fatalf("summary.invalid_route_targets = %d, want 1", got)
	}
	if len(res.Compatibility.InvalidRouteTargets) != 1 {
		t.Fatalf("invalid_route_targets = %#v, want 1", res.Compatibility.InvalidRouteTargets)
	}
	issue := res.Compatibility.InvalidRouteTargets[0]
	if issue.GroupName != "missing-key-group" || issue.ChannelName != "missing-key-channel" {
		t.Fatalf("invalid route issue = %#v, want group/channel names", issue)
	}
	if issue.Model != "gpt-4o" || issue.Action != "missing_key" {
		t.Fatalf("invalid route issue = %#v, want gpt-4o missing_key", issue)
	}
	if !containsWarning(res.Compatibility.SkippedTargets, "channel_key:602 empty credential") {
		t.Fatalf("skipped_targets = %#v, want empty credential evidence", res.Compatibility.SkippedTargets)
	}
}

func TestDBImportIncrementalDryRunReportsSkippedRouteTargetPreviewInSkipMode(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "skip-preview-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.ChannelKey{ChannelID: existingChannel.ID, Enabled: true, ChannelKey: "existing-key", AllowedModels: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create existing channel key error = %v", err)
	}
	existingGroup := model.Group{Name: "skip-preview-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&existingGroup).Error; err != nil {
		t.Fatalf("create existing group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: existingGroup.ID, ChannelID: existingChannel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create existing group item error = %v", err)
	}

	dump := &model.DBDump{
		Version:     dbDumpVersion,
		Manifest:    model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels:    []model.Channel{{ID: 701, Name: "skip-preview-channel", Enabled: true, Model: "gpt-4o"}},
		ChannelKeys: []model.ChannelKey{{ID: 702, ChannelID: 701, Enabled: true, ChannelKey: "incoming-key", AllowedModels: "gpt-4o"}},
		Groups:      []model.Group{{ID: 703, Name: "skip-preview-group", Mode: model.GroupModeRoundRobin}},
		GroupItems:  []model.GroupItem{{ID: 704, GroupID: 703, ChannelID: 701, ModelName: "gpt-4o", Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeSkip, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=skip, dryRun=true) error = %v", err)
	}
	if got := res.Compatibility.Summary.SkippedRoutePreviews; got != 1 {
		t.Fatalf("summary.skipped_route_target_previews = %d, want 1", got)
	}
	if len(res.Compatibility.SkippedRoutePreviews) != 1 {
		t.Fatalf("skipped_route_target_previews = %#v, want 1", res.Compatibility.SkippedRoutePreviews)
	}
	issue := res.Compatibility.SkippedRoutePreviews[0]
	if issue.GroupName != "skip-preview-group" || issue.Model != "gpt-4o" {
		t.Fatalf("skipped preview issue = %#v, want skip-preview-group / gpt-4o", issue)
	}
	if issue.IssueType != "skipped_route_target_preview" || issue.Action != "skip_mode_preserved_existing_group" {
		t.Fatalf("skipped preview issue = %#v, want skipped_route_target_preview / skip_mode_preserved_existing_group", issue)
	}
	if !containsWarning([]string{issue.Reason}, "skip_mode_preserved_existing_group") {
		t.Fatalf("skipped preview issue.reason = %q, want skip_mode_preserved_existing_group", issue.Reason)
	}
	if !containsWarning(res.Compatibility.RoutePreviewWarnings, "skipped route target previews: 1") {
		t.Fatalf("route_preview_warnings = %#v, want skipped route target preview warning", res.Compatibility.RoutePreviewWarnings)
	}
	if got := res.Compatibility.Summary.RoutePreviewWarnings; got != len(res.Compatibility.RoutePreviewWarnings) {
		t.Fatalf("summary.route_preview_warnings = %d, want %d", got, len(res.Compatibility.RoutePreviewWarnings))
	}
}

func TestDBImportIncrementalApplySavesPreImportSnapshot(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "rollback-source-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 1301, Name: "rollback-target-channel", Enabled: true, Model: "gpt-4o"}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if res == nil {
		t.Fatalf("DBImportIncremental() = nil, want result")
	}

	snapshotDir, err := importSnapshotDir()
	if err != nil {
		t.Fatalf("importSnapshotDir() error = %v", err)
	}
	latestPath := filepath.Join(snapshotDir, importSnapshotLatestFilename)
	if _, err := os.Stat(latestPath); err != nil {
		t.Fatalf("latest snapshot metadata missing at %s: %v", latestPath, err)
	}
	metadata, storedDump, err := loadLatestImportSnapshot()
	if err != nil {
		t.Fatalf("loadLatestImportSnapshot() error = %v", err)
	}
	if metadata == nil || strings.TrimSpace(metadata.SnapshotPath) == "" {
		t.Fatalf("metadata = %#v, want populated snapshot path", metadata)
	}
	if _, err := os.Stat(metadata.SnapshotPath); err != nil {
		t.Fatalf("snapshot file missing at %s: %v", metadata.SnapshotPath, err)
	}
	if runtime.GOOS != "windows" {
		if info, err := os.Stat(snapshotDir); err != nil {
			t.Fatalf("stat snapshot dir error = %v", err)
		} else if got := info.Mode().Perm(); got != importSnapshotDirPerm {
			t.Fatalf("snapshot dir perm = %04o, want %04o", got, importSnapshotDirPerm)
		}
		if info, err := os.Stat(metadata.SnapshotPath); err != nil {
			t.Fatalf("stat snapshot file error = %v", err)
		} else if got := info.Mode().Perm(); got != importSnapshotFilePerm {
			t.Fatalf("snapshot file perm = %04o, want %04o", got, importSnapshotFilePerm)
		}
		if info, err := os.Stat(latestPath); err != nil {
			t.Fatalf("stat latest metadata error = %v", err)
		} else if got := info.Mode().Perm(); got != importSnapshotFilePerm {
			t.Fatalf("latest metadata perm = %04o, want %04o", got, importSnapshotFilePerm)
		}
	}
	if storedDump == nil || len(storedDump.Channels) == 0 {
		t.Fatalf("stored dump = %#v, want exported pre-import snapshot", storedDump)
	}
	if storedDump.Channels[0].Name != "rollback-source-channel" {
		t.Fatalf("stored snapshot first channel = %#v, want pre-import data", storedDump.Channels[0])
	}
}

func TestDBRollbackLatestImportSnapshotRestoresPreviousState(t *testing.T) {
	ctx := setupOpTestDB(t)
	if err := UserChangePassword("admin", "before-secret"); err != nil {
		t.Fatalf("UserChangePassword(before import) error = %v", err)
	}
	if err := UserVerify("admin", "before-secret"); err != nil {
		t.Fatalf("UserVerify(before import) error = %v", err)
	}

	importedUser := model.User{Username: "admin", Password: "after-secret"}
	if err := importedUser.HashPassword(); err != nil {
		t.Fatalf("HashPassword() imported user error = %v", err)
	}

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "before-rollback-channel", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create original channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 123, Status: migrate.MigrationRecordStatusSuccess}).Error; err != nil {
		t.Fatalf("create original migration record error = %v", err)
	}
	var migrationRecordsBeforeImport []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecordsBeforeImport).Error; err != nil {
		t.Fatalf("query migration records before import error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Users:            []model.User{{ID: 1501, Username: importedUser.Username, Password: importedUser.Password}},
		Channels:         []model.Channel{{ID: 1401, Name: "after-import-channel", Enabled: true, Model: "gpt-4o"}},
		MigrationRecords: []model.DBDumpMigrationRecord{{Version: 999, Status: int(migrate.MigrationRecordStatusFailed)}},
	}

	if _, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if err := UserVerify("admin", "after-secret"); err != nil {
		t.Fatalf("UserVerify(after import) error = %v", err)
	}
	if err := UserVerify("admin", "before-secret"); err == nil {
		t.Fatalf("UserVerify(before-secret) after import expected failure")
	}

	metadataBeforeRollback, storedDumpBeforeRollback, err := loadLatestImportSnapshot()
	if err != nil {
		t.Fatalf("loadLatestImportSnapshot() before rollback error = %v", err)
	}
	if metadataBeforeRollback == nil || strings.TrimSpace(metadataBeforeRollback.SnapshotPath) == "" {
		t.Fatalf("metadata before rollback = %#v, want populated snapshot metadata", metadataBeforeRollback)
	}
	if storedDumpBeforeRollback == nil || len(storedDumpBeforeRollback.Channels) != 1 || storedDumpBeforeRollback.Channels[0].Name != "before-rollback-channel" {
		t.Fatalf("stored dump before rollback = %#v, want pre-import snapshot", storedDumpBeforeRollback)
	}

	var channelsAfterImport []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channelsAfterImport).Error; err != nil {
		t.Fatalf("query channels after import error = %v", err)
	}
	if len(channelsAfterImport) != 2 {
		t.Fatalf("channels after import = %#v, want 2", channelsAfterImport)
	}

	rollbackRes, err := DBRollbackLatestImportSnapshot(ctx)
	if err != nil {
		t.Fatalf("DBRollbackLatestImportSnapshot() error = %v", err)
	}
	if rollbackRes == nil || rollbackRes.Result == nil {
		t.Fatalf("rollback result = %#v, want populated result", rollbackRes)
	}

	var channelsAfterRollback []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channelsAfterRollback).Error; err != nil {
		t.Fatalf("query channels after rollback error = %v", err)
	}
	if len(channelsAfterRollback) != 1 {
		t.Fatalf("channels after rollback = %#v, want 1 restored channel", channelsAfterRollback)
	}
	if channelsAfterRollback[0].Name != "before-rollback-channel" {
		t.Fatalf("channels after rollback = %#v, want pre-import state restored", channelsAfterRollback)
	}
	if err := UserVerify("admin", "before-secret"); err != nil {
		t.Fatalf("UserVerify(after rollback) error = %v", err)
	}
	if err := UserVerify("admin", "after-secret"); err == nil {
		t.Fatalf("UserVerify(after-secret) after rollback expected failure")
	}
	var migrationRecordsAfterRollback []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecordsAfterRollback).Error; err != nil {
		t.Fatalf("query migration records after rollback error = %v", err)
	}
	if len(migrationRecordsAfterRollback) != len(migrationRecordsBeforeImport) {
		t.Fatalf("migration records after rollback len = %d, want %d", len(migrationRecordsAfterRollback), len(migrationRecordsBeforeImport))
	}
	if !migrationRecordVersionsContain(migrationRecordsAfterRollback, 123) {
		t.Fatalf("migration records after rollback = %#v, want preserved version 123", migrationRecordsAfterRollback)
	}
	if migrationRecordVersionsContain(migrationRecordsAfterRollback, 999) {
		t.Fatalf("migration records after rollback = %#v, do not want imported version 999", migrationRecordsAfterRollback)
	}

	metadataAfterRollback, storedDumpAfterRollback, err := loadLatestImportSnapshot()
	if err != nil {
		t.Fatalf("loadLatestImportSnapshot() after rollback error = %v", err)
	}
	if metadataAfterRollback == nil {
		t.Fatalf("metadata after rollback = nil, want preserved snapshot metadata")
	}
	if metadataAfterRollback.SnapshotPath != metadataBeforeRollback.SnapshotPath {
		t.Fatalf("snapshot path after rollback = %q, want preserved %q", metadataAfterRollback.SnapshotPath, metadataBeforeRollback.SnapshotPath)
	}
	if metadataAfterRollback.SnapshotName != metadataBeforeRollback.SnapshotName {
		t.Fatalf("snapshot name after rollback = %q, want preserved %q", metadataAfterRollback.SnapshotName, metadataBeforeRollback.SnapshotName)
	}
	if storedDumpAfterRollback == nil || len(storedDumpAfterRollback.Channels) != 1 || storedDumpAfterRollback.Channels[0].Name != "before-rollback-channel" {
		t.Fatalf("stored dump after rollback = %#v, want preserved pre-import snapshot", storedDumpAfterRollback)
	}
	if len(storedDumpAfterRollback.MigrationRecords) != len(migrationRecordsBeforeImport) {
		t.Fatalf("stored dump after rollback migration record len = %d, want %d", len(storedDumpAfterRollback.MigrationRecords), len(migrationRecordsBeforeImport))
	}
	if !dumpMigrationRecordVersionsContain(storedDumpAfterRollback.MigrationRecords, 123) {
		t.Fatalf("stored dump after rollback migration records = %#v, want preserved version 123", storedDumpAfterRollback.MigrationRecords)
	}
	if dumpMigrationRecordVersionsContain(storedDumpAfterRollback.MigrationRecords, 999) {
		t.Fatalf("stored dump after rollback migration records = %#v, do not want imported version 999", storedDumpAfterRollback.MigrationRecords)
	}
}

func TestDBRollbackLatestImportSnapshotRestoresGovernanceState(t *testing.T) {
	ctx := setupOpTestDB(t)
	now := time.Now().UTC().Truncate(time.Second)

	originalSession := model.GovernanceSession{
		Goal:             "before rollback governance",
		Scope:            model.GovernanceScopeRoutingGrouping,
		ExpertPresetID:   model.GovernanceExpertPresetBalanced,
		Status:           model.GovernanceSessionStatusApplied,
		CurrentStage:     model.GovernanceStageCompleted,
		OperatorSummary:  "before summary",
		RiskSummary:      "before risk",
		Confidence:       0.75,
		SnapshotChecksum: "before-checksum",
		SnapshotJSON:     `{"before":true}`,
		PlanJSON:         `{"before":true}`,
		PreviewJSON:      `{"can_apply":false}`,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&originalSession).Error; err != nil {
		t.Fatalf("create original governance session error = %v", err)
	}
	originalApplyRun := model.GovernanceApplyRun{
		SessionID:     originalSession.ID,
		Status:        model.GovernanceApplyRunStatusSucceeded,
		ResultSummary: "before apply",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&originalApplyRun).Error; err != nil {
		t.Fatalf("create original governance apply run error = %v", err)
	}
	originalRollbackPoint := model.GovernanceRollbackPoint{
		SessionID:        originalSession.ID,
		ApplyRunID:       &originalApplyRun.ID,
		SnapshotChecksum: "before-rollback",
		SnapshotJSON:     `{"before":true}`,
		Summary:          "before rollback point",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&originalRollbackPoint).Error; err != nil {
		t.Fatalf("create original governance rollback point error = %v", err)
	}
	originalStrategyProfile := model.StrategyProfile{
		Name:          "before strategy",
		Summary:       "before strategy summary",
		Status:        model.StrategyProfileStatusActive,
		MutationsJSON: `[{"type":"group_upsert"}]`,
		CreatedAt:     now,
		UpdatedAt:     now,
		ActivatedAt:   &now,
	}
	if err := db.GetDB().WithContext(ctx).Create(&originalStrategyProfile).Error; err != nil {
		t.Fatalf("create original strategy profile error = %v", err)
	}

	importedApplyRunID := 9101
	importDump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		GovernanceSessions: []model.GovernanceSession{{
			ID:               9001,
			Goal:             "imported governance",
			Scope:            model.GovernanceScopeRoutingGrouping,
			ExpertPresetID:   model.GovernanceExpertPresetDeepReview,
			Status:           model.GovernanceSessionStatusReady,
			CurrentStage:     model.GovernanceStageCompleted,
			OperatorSummary:  "imported summary",
			RiskSummary:      "imported risk",
			Confidence:       0.91,
			SnapshotChecksum: "imported-checksum",
			SnapshotJSON:     `{"imported":true}`,
			PlanJSON:         `{"imported":true}`,
			PreviewJSON:      `{"can_apply":true}`,
			CreatedAt:        now.Add(time.Minute),
			UpdatedAt:        now.Add(time.Minute),
		}},
		GovernanceApplyRuns: []model.GovernanceApplyRun{{
			ID:            importedApplyRunID,
			SessionID:     9001,
			Status:        model.GovernanceApplyRunStatusRunning,
			ResultSummary: "imported apply",
			CreatedAt:     now.Add(time.Minute),
			UpdatedAt:     now.Add(time.Minute),
		}},
		GovernanceRollbackPoints: []model.GovernanceRollbackPoint{{
			ID:               9201,
			SessionID:        9001,
			ApplyRunID:       &importedApplyRunID,
			SnapshotChecksum: "imported-rollback",
			SnapshotJSON:     `{"imported":true}`,
			Summary:          "imported rollback point",
			CreatedAt:        now.Add(time.Minute),
			UpdatedAt:        now.Add(time.Minute),
		}},
		StrategyProfiles: []model.StrategyProfile{{
			ID:            9301,
			Name:          "imported strategy",
			Summary:       "imported strategy summary",
			Status:        model.StrategyProfileStatusReady,
			MutationsJSON: `[{"type":"runtime_policy_set"}]`,
			CreatedAt:     now.Add(time.Minute),
			UpdatedAt:     now.Add(time.Minute),
		}},
	}

	if _, err := DBImportIncremental(ctx, importDump, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental(import governance) error = %v", err)
	}

	rollbackRes, err := DBRollbackLatestImportSnapshot(ctx)
	if err != nil {
		t.Fatalf("DBRollbackLatestImportSnapshot() error = %v", err)
	}
	if rollbackRes == nil || rollbackRes.Result == nil {
		t.Fatalf("rollback result = %#v, want populated result", rollbackRes)
	}

	var sessions []model.GovernanceSession
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&sessions).Error; err != nil {
		t.Fatalf("query governance sessions after rollback error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].Goal != originalSession.Goal {
		t.Fatalf("governance sessions after rollback = %#v, want original state restored", sessions)
	}
	var applyRuns []model.GovernanceApplyRun
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&applyRuns).Error; err != nil {
		t.Fatalf("query governance apply runs after rollback error = %v", err)
	}
	if len(applyRuns) != 1 || applyRuns[0].ResultSummary != originalApplyRun.ResultSummary {
		t.Fatalf("governance apply runs after rollback = %#v, want original state restored", applyRuns)
	}
	var rollbackPoints []model.GovernanceRollbackPoint
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&rollbackPoints).Error; err != nil {
		t.Fatalf("query governance rollback points after rollback error = %v", err)
	}
	if len(rollbackPoints) != 1 || rollbackPoints[0].Summary != originalRollbackPoint.Summary {
		t.Fatalf("governance rollback points after rollback = %#v, want original state restored", rollbackPoints)
	}
	var strategyProfiles []model.StrategyProfile
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&strategyProfiles).Error; err != nil {
		t.Fatalf("query strategy profiles after rollback error = %v", err)
	}
	if len(strategyProfiles) != 1 || strategyProfiles[0].Name != originalStrategyProfile.Name {
		t.Fatalf("strategy profiles after rollback = %#v, want original state restored", strategyProfiles)
	}
}

func TestLoadLatestImportSnapshotRejectsMetadataPathOutsideSnapshotDir(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "metadata-path-check-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 2301, Name: "metadata-path-check-target", Enabled: true, Model: "gpt-4o"}},
	}

	if _, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}

	snapshotDir, err := importSnapshotDir()
	if err != nil {
		t.Fatalf("importSnapshotDir() error = %v", err)
	}
	outsidePath := filepath.Join(t.TempDir(), "outside-snapshot.json")
	if err := os.WriteFile(outsidePath, []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("write outside snapshot error = %v", err)
	}
	metadataPath := filepath.Join(snapshotDir, importSnapshotLatestFilename)
	metadataPayload := []byte(fmt.Sprintf(`{"snapshot_path":%q,"snapshot_name":"outside.json","imported_at":"2026-04-20T00:00:00Z"}`, filepath.ToSlash(outsidePath)))
	if err := os.WriteFile(metadataPath, metadataPayload, importSnapshotFilePerm); err != nil {
		t.Fatalf("write latest metadata error = %v", err)
	}

	_, _, err = loadLatestImportSnapshot()
	if err == nil {
		t.Fatalf("loadLatestImportSnapshot() error = nil, want path validation failure")
	}
	if !strings.Contains(err.Error(), "outside import snapshot directory") {
		t.Fatalf("loadLatestImportSnapshot() error = %v, want outside snapshot dir", err)
	}
}

func TestDBPreviewRollbackImportSnapshotRowsSummaryIncludesUsersAndMigrationRecords(t *testing.T) {
	ctx := setupOpTestDB(t)
	createTestUser(t, ctx, "preview-rollback-admin", "preview-secret")
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 77, Status: migrate.MigrationRecordStatusSuccess}).Error; err != nil {
		t.Fatalf("create migration record error = %v", err)
	}
	var migrationRecordsBeforeImport []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecordsBeforeImport).Error; err != nil {
		t.Fatalf("query migration records before import error = %v", err)
	}
	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 1, Name: "preview-rollback-target", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}
	snapshots, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatalf("snapshots = %#v, want at least one snapshot", snapshots)
	}
	preview, err := DBPreviewRollbackImportSnapshot(ctx, snapshots[0].SnapshotName, nil)
	if err != nil {
		t.Fatalf("DBPreviewRollbackImportSnapshot() error = %v", err)
	}
	if preview.RowsSummary["users"] != 2 {
		t.Fatalf("rows_summary[users] = %d, want 2", preview.RowsSummary["users"])
	}
	if preview.RowsSummary["migration_records"] != len(migrationRecordsBeforeImport) {
		t.Fatalf("rows_summary[migration_records] = %d, want %d", preview.RowsSummary["migration_records"], len(migrationRecordsBeforeImport))
	}
}

func migrationRecordVersionsContain(rows []migrate.MigrationRecord, version int) bool {
	for _, row := range rows {
		if row.Version == version {
			return true
		}
	}
	return false
}

func dumpMigrationRecordVersionsContain(rows []model.DBDumpMigrationRecord, version int) bool {
	for _, row := range rows {
		if row.Version == version {
			return true
		}
	}
	return false
}

func TestDBImportIncrementalSelectiveImportScopesOnlyApplyChosenDomains(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "existing-channel", Enabled: true, Model: "existing-model"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	existingChannelKey := model.ChannelKey{ChannelID: existingChannel.ID, Enabled: true, ChannelKey: "existing-upstream-key"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannelKey).Error; err != nil {
		t.Fatalf("create existing channel key error = %v", err)
	}
	existingGroup := model.Group{Name: "existing-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&existingGroup).Error; err != nil {
		t.Fatalf("create existing group error = %v", err)
	}
	existingGroupItem := model.GroupItem{GroupID: existingGroup.ID, ChannelID: existingChannel.ID, ModelName: "existing-model", Priority: 1, Weight: 1}
	if err := db.GetDB().WithContext(ctx).Create(&existingGroupItem).Error; err != nil {
		t.Fatalf("create existing group item error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "existing-model", CanonicalName: "existing-model"}).Error; err != nil {
		t.Fatalf("create existing llm info error = %v", err)
	}
	existingAPIKey := model.APIKey{Name: "existing-client", APIKey: "sk-existing-client", Enabled: true}
	if err := db.GetDB().WithContext(ctx).Create(&existingAPIKey).Error; err != nil {
		t.Fatalf("create existing api key error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://existing.example.com")
	if err := db.GetDB().WithContext(ctx).Create(&model.StatsTotal{ID: 1, StatsMetrics: model.StatsMetrics{InputToken: 7}}).Error; err != nil {
		t.Fatalf("create existing stats_total error = %v", err)
	}
	existingRelayLog := model.RelayLog{ID: 9101, Time: 1710000000, RequestModelName: "existing-model", ChannelId: existingChannel.ID, ChannelName: existingChannel.Name, ActualModelName: "existing-model", TotalAttempts: 1}
	if err := db.GetDB().WithContext(ctx).Create(&existingRelayLog).Error; err != nil {
		t.Fatalf("create existing relay log error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels:     []model.Channel{{ID: 2101, Name: "incoming-channel", Enabled: true, Model: "incoming-model"}},
		ChannelKeys:  []model.ChannelKey{{ID: 2102, ChannelID: 2101, Enabled: true, ChannelKey: "incoming-upstream-key"}},
		Groups:       []model.Group{{ID: 2103, Name: "incoming-group", Mode: model.GroupModeWeighted}},
		GroupItems:   []model.GroupItem{{ID: 2104, GroupID: 2103, ChannelID: 2101, ModelName: "incoming-model", Priority: 2, Weight: 3}},
		LLMInfos:     []model.LLMInfo{{Name: "incoming-model", CanonicalName: "incoming-model"}},
		APIKeys:      []model.APIKey{{ID: 2105, Name: "incoming-client", APIKey: "sk-incoming-client", Enabled: true}},
		Settings:     []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://incoming.example.com"}},
		IncludeStats: true,
		StatsTotal:   []model.StatsTotal{{ID: 1, StatsMetrics: model.StatsMetrics{InputToken: 99}}},
		IncludeLogs:  true,
		RelayLogs:    []model.RelayLog{{ID: 9102, Time: 1710000001, RequestModelName: "incoming-model", ChannelId: 2101, ChannelName: "incoming-channel", ActualModelName: "incoming-model", TotalAttempts: 1}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeIncremental, false, model.DBImportOptions{
		ImportScopes: &model.DBImportScopes{
			Settings: true,
		},
	})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., import_scopes=settings) error = %v", err)
	}
	if got := res.RowsAffected["settings"]; got != 1 {
		t.Fatalf("rows_affected[settings] = %d, want 1", got)
	}
	if got := res.RowsAffected["channels"]; got != 0 {
		t.Fatalf("rows_affected[channels] = %d, want 0 when routing scope disabled", got)
	}
	if got := res.RowsAffected["api_keys"]; got != 0 {
		t.Fatalf("rows_affected[api_keys] = %d, want 0 when api key scope disabled", got)
	}
	if got := res.RowsAffected["llm_infos"]; got != 0 {
		t.Fatalf("rows_affected[llm_infos] = %d, want 0 when model scope disabled", got)
	}
	if got := res.RowsAffected["relay_logs"]; got != 0 {
		t.Fatalf("rows_affected[relay_logs] = %d, want 0 when log scope disabled", got)
	}

	var settings []model.Setting
	if err := db.GetDB().WithContext(ctx).Find(&settings, "key = ?", model.SettingKeyAPIBaseURL).Error; err != nil {
		t.Fatalf("query settings error = %v", err)
	}
	if len(settings) != 1 || settings[0].Value != "https://incoming.example.com" {
		t.Fatalf("stored settings = %#v, want imported incoming setting only", settings)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels error = %v", err)
	}
	if len(channels) != 1 || channels[0].Name != "existing-channel" {
		t.Fatalf("stored channels = %#v, want existing routing untouched", channels)
	}

	var users []model.User
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&users).Error; err != nil {
		t.Fatalf("query users error = %v", err)
	}
	if len(users) != 1 || users[0].Username != "admin" {
		t.Fatalf("stored users = %#v, want default admin untouched when no user scope exists", users)
	}

	var migrationRecords []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecords).Error; err != nil {
		t.Fatalf("query migration records error = %v", err)
	}
	if len(migrationRecords) != 0 {
		t.Fatalf("migration_records = %#v, want untouched empty migration state when no migration scope exists", migrationRecords)
	}

	var channelKeys []model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&channelKeys).Error; err != nil {
		t.Fatalf("query channel keys error = %v", err)
	}
	if len(channelKeys) != 1 || channelKeys[0].ChannelKey != "existing-upstream-key" {
		t.Fatalf("stored channel keys = %#v, want existing routing keys untouched", channelKeys)
	}

	var groups []model.Group
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&groups).Error; err != nil {
		t.Fatalf("query groups error = %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "existing-group" {
		t.Fatalf("stored groups = %#v, want existing group untouched", groups)
	}

	var groupItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Order("priority asc").Find(&groupItems).Error; err != nil {
		t.Fatalf("query group items error = %v", err)
	}
	if len(groupItems) != 1 || groupItems[0].ModelName != "existing-model" {
		t.Fatalf("stored group items = %#v, want existing group routes untouched", groupItems)
	}

	var llmInfos []model.LLMInfo
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&llmInfos).Error; err != nil {
		t.Fatalf("query llm infos error = %v", err)
	}
	if len(llmInfos) != 1 || llmInfos[0].Name != "existing-model" {
		t.Fatalf("stored llm infos = %#v, want existing model metadata untouched", llmInfos)
	}

	var apiKeys []model.APIKey
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&apiKeys).Error; err != nil {
		t.Fatalf("query api keys error = %v", err)
	}
	if len(apiKeys) != 1 || apiKeys[0].Name != "existing-client" {
		t.Fatalf("stored api keys = %#v, want existing api keys untouched", apiKeys)
	}

	var statsTotal model.StatsTotal
	if err := db.GetDB().WithContext(ctx).First(&statsTotal, "id = ?", 1).Error; err != nil {
		t.Fatalf("query stats_total error = %v", err)
	}
	if statsTotal.InputToken != 7 {
		t.Fatalf("stats_total.input_token = %d, want original stats preserved", statsTotal.InputToken)
	}

	var relayLogs []model.RelayLog
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&relayLogs).Error; err != nil {
		t.Fatalf("query relay logs error = %v", err)
	}
	if len(relayLogs) != 1 || relayLogs[0].ID != existingRelayLog.ID {
		t.Fatalf("stored relay logs = %#v, want existing logs untouched", relayLogs)
	}
}

func TestDBImportIncrementalSelectiveReplaceDoesNotTouchMigrationRecords(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 101, Status: migrate.MigrationRecordStatusSuccess}).Error; err != nil {
		t.Fatalf("create existing migration record error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 202, Status: migrate.MigrationRecordStatusFailed}).Error; err != nil {
		t.Fatalf("create stale migration record error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Settings:         []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://incoming.example.com"}},
		MigrationRecords: []model.DBDumpMigrationRecord{{Version: 303, Status: int(migrate.MigrationRecordStatusSuccess)}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeReplace, false, model.DBImportOptions{
		ImportScopes: &model.DBImportScopes{Settings: true},
	})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=replace, import_scopes=settings) error = %v", err)
	}
	if got := res.RowsAffected["replaced_migration_records"]; got != 0 {
		t.Fatalf("rows_affected[replaced_migration_records] = %d, want 0 when migration scope disabled", got)
	}
	if got := res.RowsAffected["migration_records"]; got != 0 {
		t.Fatalf("rows_affected[migration_records] = %d, want 0 when migration scope disabled", got)
	}

	var migrationRecords []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecords).Error; err != nil {
		t.Fatalf("query migration records error = %v", err)
	}
	if len(migrationRecords) != 2 {
		t.Fatalf("migration_records = %#v, want existing records preserved", migrationRecords)
	}
	if migrationRecords[0].Version != 101 || migrationRecords[1].Version != 202 {
		t.Fatalf("migration_records = %#v, want preserved versions 101 and 202", migrationRecords)
	}
}

func TestDBImportIncrementalSelectiveSettingsReplaceOnlyTouchesSettingsScope(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "existing-channel", Enabled: true, Model: "existing-model"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://existing.example.com")
	setTestSetting(t, ctx, model.SettingKeyProxyURL, "http://stale-proxy")

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{ID: 2101, Name: "incoming-channel", Enabled: true, Model: "incoming-model"}},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://incoming.example.com"}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeReplace, false, model.DBImportOptions{
		ImportScopes: &model.DBImportScopes{Settings: true},
	})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=replace, import_scopes=settings) error = %v", err)
	}
	if got := res.RowsAffected["replaced_settings"]; got < 1 {
		t.Fatalf("rows_affected[replaced_settings] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["settings"]; got != 1 {
		t.Fatalf("rows_affected[settings] = %d, want 1", got)
	}
	if got := res.RowsAffected["channels"]; got != 0 {
		t.Fatalf("rows_affected[channels] = %d, want 0 when routing scope disabled", got)
	}

	var apiBaseURLSetting model.Setting
	if err := db.GetDB().WithContext(ctx).First(&apiBaseURLSetting, "key = ?", model.SettingKeyAPIBaseURL).Error; err != nil {
		t.Fatalf("query api_base_url setting error = %v", err)
	}
	if apiBaseURLSetting.Value != "https://incoming.example.com" {
		t.Fatalf("api_base_url value = %q, want incoming value", apiBaseURLSetting.Value)
	}
	var proxySetting model.Setting
	if err := db.GetDB().WithContext(ctx).First(&proxySetting, "key = ?", model.SettingKeyProxyURL).Error; err != nil {
		t.Fatalf("query proxy_url setting error = %v", err)
	}
	if proxySetting.Value != "" {
		t.Fatalf("proxy_url value = %q, want default empty after scoped replace", proxySetting.Value)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels error = %v", err)
	}
	if len(channels) != 1 || channels[0].Name != "existing-channel" {
		t.Fatalf("channels = %#v, want existing routing untouched", channels)
	}
}

func TestDBImportIncrementalSelectiveModelsReplaceOnlyTouchesModelScope(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "existing-channel", Enabled: true, Model: "existing-model"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "existing-model", CanonicalName: "existing-model", BillingMode: model.BillingModePerToken}).Error; err != nil {
		t.Fatalf("create existing llm info error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "stale-model", CanonicalName: "stale-model", BillingMode: model.BillingModePerRequest}).Error; err != nil {
		t.Fatalf("create stale llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		Channels: []model.Channel{{ID: 2101, Name: "incoming-channel", Enabled: true, Model: "incoming-model"}},
		LLMInfos: []model.LLMInfo{{Name: "existing-model", CanonicalName: "existing-model", BillingMode: model.BillingModeFlat}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeReplace, false, model.DBImportOptions{
		ImportScopes: &model.DBImportScopes{Models: true},
	})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=replace, import_scopes=models) error = %v", err)
	}
	if got := res.RowsAffected["replaced_llm_infos"]; got < 1 {
		t.Fatalf("rows_affected[replaced_llm_infos] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["llm_infos"]; got != 1 {
		t.Fatalf("rows_affected[llm_infos] = %d, want 1", got)
	}
	if got := res.RowsAffected["channels"]; got != 0 {
		t.Fatalf("rows_affected[channels] = %d, want 0 when routing scope disabled", got)
	}

	var llmInfos []model.LLMInfo
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&llmInfos).Error; err != nil {
		t.Fatalf("query llm infos error = %v", err)
	}
	if len(llmInfos) != 1 || llmInfos[0].Name != "existing-model" {
		t.Fatalf("llm_infos = %#v, want only existing-model after scoped replace", llmInfos)
	}
	if llmInfos[0].BillingMode != model.BillingModeFlat {
		t.Fatalf("llm_infos[0].billing_mode = %q, want flat from incoming snapshot", llmInfos[0].BillingMode)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels error = %v", err)
	}
	if len(channels) != 1 || channels[0].Name != "existing-channel" {
		t.Fatalf("channels = %#v, want existing routing untouched", channels)
	}
}

func TestDBImportIncrementalReplaceModePrunesAPIKeysWhenSnapshotContainsSecrets(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true, SupportedModels: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create keep api key error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "stale-client", APIKey: "sk-stale-client", Enabled: true, SupportedModels: "o1-mini"}).Error; err != nil {
		t.Fatalf("create stale api key error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		APIKeys: []model.APIKey{{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true, SupportedModels: "gpt-4o,o1-mini"}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["replaced_api_keys"]; got < 1 {
		t.Fatalf("rows_affected[replaced_api_keys] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["api_keys"]; got != 1 {
		t.Fatalf("rows_affected[api_keys] = %d, want 1", got)
	}

	var apiKeys []model.APIKey
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&apiKeys).Error; err != nil {
		t.Fatalf("query api keys error = %v", err)
	}
	if len(apiKeys) != 1 || apiKeys[0].APIKey != "sk-keep-client" {
		t.Fatalf("api_keys = %#v, want only keep-client after replace", apiKeys)
	}
	if apiKeys[0].SupportedModels != "gpt-4o,o1-mini" {
		t.Fatalf("api_keys[0].supported_models = %q, want snapshot value", apiKeys[0].SupportedModels)
	}
}

func TestDBImportIncrementalReplaceModeDoesNotPruneAPIKeysForRedactedSnapshot(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "existing-client", APIKey: "sk-existing-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create existing api key error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: false,
		},
		APIKeys: []model.APIKey{{Name: "redacted-client", APIKey: "", Enabled: true}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeReplace, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., mode=replace, dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["replaced_api_keys"]; got != 0 {
		t.Fatalf("rows_affected[replaced_api_keys] = %d, want 0 for redacted snapshot", got)
	}
	if !containsWarning(res.Warnings, snapshotCredentialsRedactedWarning) {
		t.Fatalf("warnings = %#v, want redacted credential warning", res.Warnings)
	}

	var apiKeys []model.APIKey
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&apiKeys).Error; err != nil {
		t.Fatalf("query api keys error = %v", err)
	}
	if len(apiKeys) != 1 || apiKeys[0].APIKey != "sk-existing-client" {
		t.Fatalf("api_keys = %#v, want existing key preserved for redacted snapshot", apiKeys)
	}
}

func TestDBImportIncrementalSelectiveAPIKeysReplaceOnlyTouchesAPIKeyScope(t *testing.T) {
	ctx := setupOpTestDB(t)

	existingChannel := model.Channel{Name: "existing-channel", Enabled: true, Model: "existing-model"}
	if err := db.GetDB().WithContext(ctx).Create(&existingChannel).Error; err != nil {
		t.Fatalf("create existing channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create keep api key error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.APIKey{Name: "stale-client", APIKey: "sk-stale-client", Enabled: true}).Error; err != nil {
		t.Fatalf("create stale api key error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{ID: 2101, Name: "incoming-channel", Enabled: true, Model: "incoming-model"}},
		APIKeys:  []model.APIKey{{Name: "keep-client", APIKey: "sk-keep-client", Enabled: true, SupportedModels: "gpt-4o"}},
	}

	res, err := DBImportIncrementalWithOptions(ctx, dump, model.DBImportModeReplace, false, model.DBImportOptions{
		ImportScopes: &model.DBImportScopes{APIKeys: true},
	})
	if err != nil {
		t.Fatalf("DBImportIncrementalWithOptions(..., mode=replace, import_scopes=api_keys) error = %v", err)
	}
	if got := res.RowsAffected["replaced_api_keys"]; got < 1 {
		t.Fatalf("rows_affected[replaced_api_keys] = %d, want >= 1", got)
	}
	if got := res.RowsAffected["api_keys"]; got != 1 {
		t.Fatalf("rows_affected[api_keys] = %d, want 1", got)
	}
	if got := res.RowsAffected["channels"]; got != 0 {
		t.Fatalf("rows_affected[channels] = %d, want 0 when routing scope disabled", got)
	}

	var apiKeys []model.APIKey
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&apiKeys).Error; err != nil {
		t.Fatalf("query api keys error = %v", err)
	}
	if len(apiKeys) != 1 || apiKeys[0].APIKey != "sk-keep-client" {
		t.Fatalf("api_keys = %#v, want only keep-client after scoped replace", apiKeys)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels error = %v", err)
	}
	if len(channels) != 1 || channels[0].Name != "existing-channel" {
		t.Fatalf("channels = %#v, want existing routing untouched", channels)
	}
}

func TestDBListImportSnapshotsReturnsLatestFirstAndMarksLatest(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "snapshot-source-a", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create source channel a error = %v", err)
	}
	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 3101, Name: "snapshot-target-a", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("first DBImportIncremental() error = %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "snapshot-source-b", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create source channel b error = %v", err)
	}
	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 3201, Name: "snapshot-target-b", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("second DBImportIncremental() error = %v", err)
	}

	items, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() error = %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("snapshot list = %#v, want at least 2 snapshots", items)
	}
	if !items[0].IsLatest {
		t.Fatalf("first snapshot = %#v, want latest marker", items[0])
	}
	if items[0].SnapshotName == items[1].SnapshotName {
		t.Fatalf("snapshot list names = %#v, want unique snapshot filenames", items)
	}
	if items[0].ImportedAt.Before(items[1].ImportedAt) {
		t.Fatalf("snapshot list = %#v, want newest snapshot first", items)
	}
	if items[0].SizeBytes <= 0 || items[1].SizeBytes <= 0 {
		t.Fatalf("snapshot list sizes = %#v, want positive file sizes", items)
	}
}

func TestDBRollbackImportSnapshotRestoresSpecifiedHistoricalSnapshot(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "history-before-a", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create history-before-a error = %v", err)
	}
	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 3301, Name: "history-import-a", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("first DBImportIncremental() error = %v", err)
	}
	historicalSnapshots, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() after first import error = %v", err)
	}
	if len(historicalSnapshots) == 0 {
		t.Fatalf("snapshot list after first import = %#v, want at least one snapshot", historicalSnapshots)
	}
	historicalSnapshot := historicalSnapshots[0]

	time.Sleep(5 * time.Millisecond)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "history-before-b", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create history-before-b error = %v", err)
	}
	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 3401, Name: "history-import-b", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("second DBImportIncremental() error = %v", err)
	}

	rollbackRes, err := DBRollbackImportSnapshot(ctx, historicalSnapshot.SnapshotName, nil)
	if err != nil {
		t.Fatalf("DBRollbackImportSnapshot() error = %v", err)
	}
	if rollbackRes == nil || rollbackRes.Result == nil {
		t.Fatalf("rollback result = %#v, want populated result", rollbackRes)
	}
	if rollbackRes.SnapshotName != historicalSnapshot.SnapshotName {
		t.Fatalf("rollback snapshot name = %q, want %q", rollbackRes.SnapshotName, historicalSnapshot.SnapshotName)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels after historical rollback error = %v", err)
	}
	channelNames := make([]string, 0, len(channels))
	for _, channel := range channels {
		channelNames = append(channelNames, channel.Name)
	}
	if !containsWarning(channelNames, "history-before-a") {
		t.Fatalf("channels after historical rollback = %#v, want history-before-a restored", channelNames)
	}
	if containsWarning(channelNames, "history-before-b") {
		t.Fatalf("channels after historical rollback = %#v, do not want newer snapshot state", channelNames)
	}
	if containsWarning(channelNames, "history-import-b") {
		t.Fatalf("channels after historical rollback = %#v, do not want second import rows", channelNames)
	}
}

func TestDBRollbackImportSnapshotCanRestoreOnlySelectedScopes(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := model.Channel{Name: "rollback-scope-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	setTestSetting(t, ctx, model.SettingKeyAPIBaseURL, "https://before.example.com")

	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Settings: []model.Setting{{Key: model.SettingKeyAPIBaseURL, Value: "https://snapshot.example.com"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}

	snapshots, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatalf("snapshots = %#v, want at least one snapshot", snapshots)
	}

	if err := db.GetDB().WithContext(ctx).Model(&model.Setting{}).Where("key = ?", model.SettingKeyAPIBaseURL).Update("value", "https://mutated.example.com").Error; err != nil {
		t.Fatalf("mutate setting error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "rollback-scope-new-channel", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create new channel error = %v", err)
	}

	rollbackRes, err := DBRollbackImportSnapshot(ctx, snapshots[0].SnapshotName, &model.DBImportScopes{Settings: true})
	if err != nil {
		t.Fatalf("DBRollbackImportSnapshot(..., settings-only) error = %v", err)
	}
	if rollbackRes == nil || rollbackRes.Result == nil {
		t.Fatalf("rollback result = %#v, want populated result", rollbackRes)
	}
	if rollbackRes.AppliedScopes == nil || !rollbackRes.AppliedScopes.Settings || rollbackRes.AppliedScopes.Routing {
		t.Fatalf("applied_scopes = %#v, want settings-only scope", rollbackRes.AppliedScopes)
	}

	var setting model.Setting
	if err := db.GetDB().WithContext(ctx).First(&setting, "key = ?", model.SettingKeyAPIBaseURL).Error; err != nil {
		t.Fatalf("query setting error = %v", err)
	}
	if setting.Value != "https://before.example.com" {
		t.Fatalf("setting.Value = %q, want restored before.example.com", setting.Value)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels error = %v", err)
	}
	if !containsWarning([]string{channels[0].Name, channels[1].Name}, "rollback-scope-new-channel") {
		t.Fatalf("channels = %#v, want routing rows untouched during settings-only rollback", channels)
	}
}

func TestDBRollbackImportSnapshotFullRestoreReplacesAIAutomationState(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.AIPromptTemplate{Name: "before-template", Source: model.AIPromptTemplateSourceCustom, TaskType: model.AIAutomationTaskTypeGroupSuggestion, Domain: model.AIProfileDomainGrouping, Prompt: "before", Enabled: true}).Error; err != nil {
		t.Fatalf("create before prompt template error = %v", err)
	}
	beforeProfile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "before-profile", Status: model.AIProfileStatusReady}, `{"config":{"base_url":"https://before.example/v1","api_key":"before-secret-key"}}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate(before) error = %v", err)
	}
	beforeTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "before task", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: `{"api_key":"before-task-secret"}`, PromptText: "before", SelectedModel: "gpt-4o", ResumeState: model.AITaskResumeStateCompleted, ExecutorVersion: "before"}
	if err := db.GetDB().WithContext(ctx).Create(&beforeTask).Error; err != nil {
		t.Fatalf("create before ai task error = %v", err)
	}
	beforeStep := model.AITaskStep{TaskID: beforeTask.ID, StepKey: "call_ai", Name: "调用 AI", Status: model.AITaskStepStatusSucceeded, SortOrder: 1}
	if err := db.GetDB().WithContext(ctx).Create(&beforeStep).Error; err != nil {
		t.Fatalf("create before ai task step error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 1, ChannelKeyID: 1, ModelName: "before-model", SuccessCount: 1, Score: 0.2}).Error; err != nil {
		t.Fatalf("create before dynamic learning state error = %v", err)
	}

	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 1, Name: "history-import-a", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}

	snapshots, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatalf("snapshot list = %#v, want at least one snapshot", snapshots)
	}
	historicalSnapshot := snapshots[0]

	if err := db.GetDB().WithContext(ctx).Create(&model.AIPromptTemplate{Name: "after-template", Source: model.AIPromptTemplateSourceCustom, TaskType: model.AIAutomationTaskTypeGroupSuggestion, Domain: model.AIProfileDomainGrouping, Prompt: "after", Enabled: true}).Error; err != nil {
		t.Fatalf("create after prompt template error = %v", err)
	}
	afterProfile, err := AIProfileCreate(model.AIProfile{Domain: model.AIProfileDomainGrouping, Name: "after-profile", Status: model.AIProfileStatusReady}, `{"config":{"base_url":"https://after.example/v1","api_key":"after-secret-key"}}`, ctx)
	if err != nil {
		t.Fatalf("AIProfileCreate(after) error = %v", err)
	}
	afterTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "after task", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: `{"api_key":"after-task-secret"}`, PromptText: "after", SelectedModel: "gpt-4o", ResumeState: model.AITaskResumeStateCompleted, ExecutorVersion: "after"}
	if err := db.GetDB().WithContext(ctx).Create(&afterTask).Error; err != nil {
		t.Fatalf("create after ai task error = %v", err)
	}
	afterStep := model.AITaskStep{TaskID: afterTask.ID, StepKey: "call_ai", Name: "调用 AI", Status: model.AITaskStepStatusSucceeded, SortOrder: 1}
	if err := db.GetDB().WithContext(ctx).Create(&afterStep).Error; err != nil {
		t.Fatalf("create after ai task step error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 2, ChannelKeyID: 2, ModelName: "after-model", SuccessCount: 5, Score: 0.9}).Error; err != nil {
		t.Fatalf("create after dynamic learning state error = %v", err)
	}

	rollbackRes, err := DBRollbackImportSnapshot(ctx, historicalSnapshot.SnapshotName, nil)
	if err != nil {
		t.Fatalf("DBRollbackImportSnapshot() error = %v", err)
	}
	if rollbackRes == nil || rollbackRes.Result == nil {
		t.Fatalf("rollback result = %#v, want nested result", rollbackRes)
	}

	var promptTemplates []model.AIPromptTemplate
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&promptTemplates).Error; err != nil {
		t.Fatalf("query prompt templates error = %v", err)
	}
	promptTemplateNames := make([]string, 0, len(promptTemplates))
	for _, row := range promptTemplates {
		promptTemplateNames = append(promptTemplateNames, row.Name)
	}
	sort.Strings(promptTemplateNames)
	if strings.Join(promptTemplateNames, ",") != "before-template" {
		t.Fatalf("prompt templates after rollback = %#v, want only before-template", promptTemplateNames)
	}

	profiles, err := AIProfileList(ctx)
	if err != nil {
		t.Fatalf("AIProfileList() error = %v", err)
	}
	profileNames := make([]string, 0, len(profiles))
	for _, row := range profiles {
		profileNames = append(profileNames, row.Name)
	}
	sort.Strings(profileNames)
	if strings.Join(profileNames, ",") != "before-profile" {
		t.Fatalf("profiles after rollback = %#v, want only before-profile", profileNames)
	}
	if profiles[0].ID != beforeProfile.ID {
		t.Fatalf("restored profile id = %d, want %d", profiles[0].ID, beforeProfile.ID)
	}

	var tasks []model.AITask
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&tasks).Error; err != nil {
		t.Fatalf("query ai tasks error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].InputText != "before task" {
		t.Fatalf("ai tasks after rollback = %#v, want only before task", tasks)
	}
	var steps []model.AITaskStep
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&steps).Error; err != nil {
		t.Fatalf("query ai task steps error = %v", err)
	}
	if len(steps) != 1 || steps[0].TaskID != beforeTask.ID {
		t.Fatalf("ai task steps after rollback = %#v, want only before task step", steps)
	}

	learning, err := DynamicRouteLearningList(ctx)
	if err != nil {
		t.Fatalf("DynamicRouteLearningList() error = %v", err)
	}
	if len(learning.States) != 1 || learning.States[0].ModelName != "before-model" {
		t.Fatalf("dynamic learning states after rollback = %#v, want only before-model", learning.States)
	}
	if afterProfile.ID == profiles[0].ID {
		t.Fatalf("after profile id %d unexpectedly survived rollback", afterProfile.ID)
	}
}

func TestDBRollbackImportSnapshotExplicitFullScopesRestoresUsersMigrationRecordsAndAIAutomationState(t *testing.T) {
	ctx := setupOpTestDB(t)

	beforeUser := model.User{Username: "before-admin", Password: "before-hash"}
	if err := db.GetDB().WithContext(ctx).Create(&beforeUser).Error; err != nil {
		t.Fatalf("create before user error = %v", err)
	}
	beforeTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "before explicit full scope task", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: `{"api_key":"before-secret"}`, PromptText: "before", SelectedModel: "gpt-4o", ResumeState: model.AITaskResumeStateCompleted, ExecutorVersion: "before"}
	if err := db.GetDB().WithContext(ctx).Create(&beforeTask).Error; err != nil {
		t.Fatalf("create before ai task error = %v", err)
	}
	beforeStep := model.AITaskStep{TaskID: beforeTask.ID, StepKey: "call_ai", Name: "调用 AI", Status: model.AITaskStepStatusSucceeded, SortOrder: 1}
	if err := db.GetDB().WithContext(ctx).Create(&beforeStep).Error; err != nil {
		t.Fatalf("create before ai task step error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 1, ChannelKeyID: 1, ModelName: "before-explicit-full-model", SuccessCount: 2, Score: 0.6}).Error; err != nil {
		t.Fatalf("create before dynamic learning state error = %v", err)
	}
	migrationRecordsBeforeImport := []migrate.MigrationRecord{{Version: 51, Status: migrate.MigrationRecordStatusSuccess}}
	for _, row := range migrationRecordsBeforeImport {
		if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
			t.Fatalf("create before migration record error = %v", err)
		}
	}

	if _, _, err := savePreImportSnapshot(ctx); err != nil {
		t.Fatalf("savePreImportSnapshot() historical snapshot error = %v", err)
	}
	historicalSnapshot, _, err := loadLatestImportSnapshot()
	if err != nil {
		t.Fatalf("loadLatestImportSnapshot() historical snapshot error = %v", err)
	}

	afterUser := model.User{Username: "after-admin", Password: "after-hash"}
	if err := db.GetDB().WithContext(ctx).Create(&afterUser).Error; err != nil {
		t.Fatalf("create after user error = %v", err)
	}
	afterTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "after explicit full scope task", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: `{"api_key":"after-secret"}`, PromptText: "after", SelectedModel: "gpt-4o", ResumeState: model.AITaskResumeStateCompleted, ExecutorVersion: "after"}
	if err := db.GetDB().WithContext(ctx).Create(&afterTask).Error; err != nil {
		t.Fatalf("create after ai task error = %v", err)
	}
	afterStep := model.AITaskStep{TaskID: afterTask.ID, StepKey: "call_ai", Name: "调用 AI", Status: model.AITaskStepStatusSucceeded, SortOrder: 1}
	if err := db.GetDB().WithContext(ctx).Create(&afterStep).Error; err != nil {
		t.Fatalf("create after ai task step error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 2, ChannelKeyID: 2, ModelName: "after-explicit-full-model", SuccessCount: 9, Score: 0.95}).Error; err != nil {
		t.Fatalf("create after dynamic learning state error = %v", err)
	}
	afterMigrationRecords := []migrate.MigrationRecord{{Version: 52, Status: migrate.MigrationRecordStatusSuccess}, {Version: 53, Status: migrate.MigrationRecordStatusSuccess}}
	for _, row := range afterMigrationRecords {
		if err := db.GetDB().WithContext(ctx).Create(&row).Error; err != nil {
			t.Fatalf("create after migration record error = %v", err)
		}
	}

	fullScopes := &model.DBImportScopes{Routing: true, Models: true, APIKeys: true, Settings: true, Stats: true, Logs: true}
	rollbackRes, err := DBRollbackImportSnapshot(ctx, historicalSnapshot.SnapshotName, fullScopes)
	if err != nil {
		t.Fatalf("DBRollbackImportSnapshot(explicit full scopes) error = %v", err)
	}
	if rollbackRes == nil || rollbackRes.Result == nil {
		t.Fatalf("rollback result = %#v, want nested result", rollbackRes)
	}
	if rollbackRes.AppliedScopes == nil || !rollbackRes.AppliedScopes.Routing || !rollbackRes.AppliedScopes.Models || !rollbackRes.AppliedScopes.APIKeys || !rollbackRes.AppliedScopes.Settings || !rollbackRes.AppliedScopes.Stats || !rollbackRes.AppliedScopes.Logs {
		t.Fatalf("applied_scopes = %#v, want explicit full scopes preserved", rollbackRes.AppliedScopes)
	}

	var users []model.User
	if err := db.GetDB().WithContext(ctx).Order("username asc").Find(&users).Error; err != nil {
		t.Fatalf("query users error = %v", err)
	}
	usernames := make([]string, 0, len(users))
	for _, row := range users {
		usernames = append(usernames, row.Username)
	}
	sort.Strings(usernames)
	if strings.Join(usernames, ",") != "admin,before-admin" {
		t.Fatalf("users after explicit full-scope rollback = %#v, want admin and before-admin", usernames)
	}

	var tasks []model.AITask
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&tasks).Error; err != nil {
		t.Fatalf("query ai tasks error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].InputText != beforeTask.InputText {
		t.Fatalf("ai tasks after explicit full-scope rollback = %#v, want only historical task", tasks)
	}

	var steps []model.AITaskStep
	if err := db.GetDB().WithContext(ctx).Order("id asc").Find(&steps).Error; err != nil {
		t.Fatalf("query ai task steps error = %v", err)
	}
	if len(steps) != 1 || steps[0].TaskID != beforeTask.ID {
		t.Fatalf("ai task steps after explicit full-scope rollback = %#v, want only historical step", steps)
	}

	learning, err := DynamicRouteLearningList(ctx)
	if err != nil {
		t.Fatalf("DynamicRouteLearningList() error = %v", err)
	}
	if len(learning.States) != 1 || learning.States[0].ModelName != "before-explicit-full-model" {
		t.Fatalf("dynamic learning states after explicit full-scope rollback = %#v, want only historical state", learning.States)
	}

	var migrationRecords []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecords).Error; err != nil {
		t.Fatalf("query migration records error = %v", err)
	}
	if len(migrationRecords) != len(migrationRecordsBeforeImport) || migrationRecords[0].Version != migrationRecordsBeforeImport[0].Version {
		t.Fatalf("migration records after explicit full-scope rollback = %#v, want %#v", migrationRecords, migrationRecordsBeforeImport)
	}
}

func TestDBRollbackLatestImportSnapshotKeepsExistingStateWhenRestoreFails(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "before-failed-rollback-channel", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create before channel error = %v", err)
	}
	if err := UserChangePassword("admin", "before-failed-rollback-secret"); err != nil {
		t.Fatalf("UserChangePassword(before failed rollback) error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 321, Status: migrate.MigrationRecordStatusSuccess}).Error; err != nil {
		t.Fatalf("create before migration record error = %v", err)
	}

	if _, _, err := savePreImportSnapshot(ctx); err != nil {
		t.Fatalf("savePreImportSnapshot() error = %v", err)
	}

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "after-failed-rollback-channel", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create after channel error = %v", err)
	}
	if err := UserChangePassword("before-failed-rollback-secret", "after-failed-rollback-secret"); err != nil {
		t.Fatalf("UserChangePassword(after failed rollback) error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 654, Status: migrate.MigrationRecordStatusFailed}).Error; err != nil {
		t.Fatalf("create after migration record error = %v", err)
	}

	latestMetadata, dump, err := loadLatestImportSnapshot()
	if err != nil {
		t.Fatalf("loadLatestImportSnapshot() error = %v", err)
	}
	if latestMetadata == nil || dump == nil {
		t.Fatalf("latest snapshot metadata=%#v dump=%#v, want populated snapshot", latestMetadata, dump)
	}
	dump.Settings = append(dump.Settings, model.Setting{Key: model.SettingKey("")})
	payload, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent(corrupted dump) error = %v", err)
	}
	if err := os.WriteFile(latestMetadata.SnapshotPath, payload, importSnapshotFilePerm); err != nil {
		t.Fatalf("os.WriteFile(corrupted snapshot) error = %v", err)
	}

	rollbackRes, err := DBRollbackLatestImportSnapshot(ctx)
	if err == nil {
		t.Fatalf("DBRollbackLatestImportSnapshot() result = %#v, want restore failure", rollbackRes)
	}
	if !strings.Contains(err.Error(), "unknown setting key") {
		t.Fatalf("DBRollbackLatestImportSnapshot() error = %v, want unknown setting key failure", err)
	}

	var channels []model.Channel
	if err := db.GetDB().WithContext(ctx).Order("name asc").Find(&channels).Error; err != nil {
		t.Fatalf("query channels after failed rollback error = %v", err)
	}
	channelNames := make([]string, 0, len(channels))
	for _, row := range channels {
		channelNames = append(channelNames, row.Name)
	}
	if strings.Join(channelNames, ",") != "after-failed-rollback-channel,before-failed-rollback-channel" {
		t.Fatalf("channels after failed rollback = %#v, want both pre-failure rows preserved", channelNames)
	}
	if err := UserVerify("admin", "after-failed-rollback-secret"); err != nil {
		t.Fatalf("UserVerify(after-failed-rollback-secret) error = %v", err)
	}
	if err := UserVerify("admin", "before-failed-rollback-secret"); err == nil {
		t.Fatalf("UserVerify(before-failed-rollback-secret) unexpectedly succeeded after failed rollback")
	}
	var migrationRecords []migrate.MigrationRecord
	if err := db.GetDB().WithContext(ctx).Order("version asc").Find(&migrationRecords).Error; err != nil {
		t.Fatalf("query migration records after failed rollback error = %v", err)
	}
	if !migrationRecordVersionsContain(migrationRecords, 321) || !migrationRecordVersionsContain(migrationRecords, 654) {
		t.Fatalf("migration records after failed rollback = %#v, want existing rows preserved", migrationRecords)
	}
}

func TestDBPreviewRollbackImportSnapshotExplicitFullScopesMatchesFullRestoreWarnings(t *testing.T) {
	ctx := setupOpTestDB(t)

	beforeUser := model.User{Username: "preview-before-admin", Password: "before-hash"}
	if err := db.GetDB().WithContext(ctx).Create(&beforeUser).Error; err != nil {
		t.Fatalf("create before user error = %v", err)
	}
	beforeTask := model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "preview before full task", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: `{"api_key":"preview-before-secret"}`, PromptText: "before", SelectedModel: "gpt-4o", ResumeState: model.AITaskResumeStateCompleted, ExecutorVersion: "before"}
	if err := db.GetDB().WithContext(ctx).Create(&beforeTask).Error; err != nil {
		t.Fatalf("create before ai task error = %v", err)
	}
	beforeStep := model.AITaskStep{TaskID: beforeTask.ID, StepKey: "call_ai", Name: "调用 AI", Status: model.AITaskStepStatusSucceeded, SortOrder: 1}
	if err := db.GetDB().WithContext(ctx).Create(&beforeStep).Error; err != nil {
		t.Fatalf("create before ai task step error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 1, ChannelKeyID: 1, ModelName: "preview-before-full-model", SuccessCount: 3, Score: 0.7}).Error; err != nil {
		t.Fatalf("create before dynamic learning state error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 81, Status: migrate.MigrationRecordStatusSuccess}).Error; err != nil {
		t.Fatalf("create before migration record error = %v", err)
	}

	if _, _, err := savePreImportSnapshot(ctx); err != nil {
		t.Fatalf("savePreImportSnapshot() historical snapshot error = %v", err)
	}
	historicalSnapshot, _, err := loadLatestImportSnapshot()
	if err != nil {
		t.Fatalf("loadLatestImportSnapshot() historical snapshot error = %v", err)
	}

	afterUser := model.User{Username: "preview-after-admin", Password: "after-hash"}
	if err := db.GetDB().WithContext(ctx).Create(&afterUser).Error; err != nil {
		t.Fatalf("create after user error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.AITask{Type: model.AIAutomationTaskTypeNaturalLanguage, InputText: "preview after full task", Status: model.AITaskStatusSucceeded, ConfigSnapshotJSON: `{"api_key":"preview-after-secret"}`, PromptText: "after", SelectedModel: "gpt-4o", ResumeState: model.AITaskResumeStateCompleted, ExecutorVersion: "after"}).Error; err != nil {
		t.Fatalf("create after ai task error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.DynamicRouteLearningState{ChannelID: 2, ChannelKeyID: 2, ModelName: "preview-after-full-model", SuccessCount: 9, Score: 0.95}).Error; err != nil {
		t.Fatalf("create after dynamic learning state error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&migrate.MigrationRecord{Version: 82, Status: migrate.MigrationRecordStatusSuccess}).Error; err != nil {
		t.Fatalf("create after migration record error = %v", err)
	}

	fullScopes := &model.DBImportScopes{Routing: true, Models: true, APIKeys: true, Settings: true, Stats: true, Logs: true}
	preview, err := DBPreviewRollbackImportSnapshot(ctx, historicalSnapshot.SnapshotName, fullScopes)
	if err != nil {
		t.Fatalf("DBPreviewRollbackImportSnapshot(explicit full scopes) error = %v", err)
	}
	if preview == nil {
		t.Fatal("preview = nil, want result")
	}
	if preview.AppliedScopes == nil || !preview.AppliedScopes.Routing || !preview.AppliedScopes.Models || !preview.AppliedScopes.APIKeys || !preview.AppliedScopes.Settings || !preview.AppliedScopes.Stats || !preview.AppliedScopes.Logs {
		t.Fatalf("applied_scopes = %#v, want explicit full scopes preserved", preview.AppliedScopes)
	}
	if got := preview.RowsSummary["users"]; got != 2 {
		t.Fatalf("rows_summary[users] = %d, want 2", got)
	}
	if got := preview.RowsSummary["migration_records"]; got != 1 {
		t.Fatalf("rows_summary[migration_records] = %d, want 1", got)
	}
	if got := preview.RowsSummary["ai_tasks"]; got != 1 {
		t.Fatalf("rows_summary[ai_tasks] = %d, want 1", got)
	}
	if got := preview.RowsSummary["dynamic_route_learning_states"]; got != 1 {
		t.Fatalf("rows_summary[dynamic_route_learning_states] = %d, want 1", got)
	}
	if !containsWarning(preview.PreviewWarnings, "restores admin users: 2") {
		t.Fatalf("preview.preview_warnings = %#v, want admin user restore warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "restores migration records: 1") {
		t.Fatalf("preview.preview_warnings = %#v, want migration record restore warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "restores AI automation state rows: 2") {
		t.Fatalf("preview.preview_warnings = %#v, want AI automation restore warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "restores dynamic route learning states: 1") {
		t.Fatalf("preview.preview_warnings = %#v, want dynamic learning restore warning", preview.PreviewWarnings)
	}
	if got := preview.Compatibility.Summary.ReplacePrunedChannels; got != 0 {
		t.Fatalf("preview.compatibility.summary.replace_pruned_channels = %d, want 0 for explicit full restore preview", got)
	}
}

func TestDBRollbackImportSnapshotRejectsPathTraversal(t *testing.T) {
	ctx := setupOpTestDB(t)

	_, err := DBRollbackImportSnapshot(ctx, "..\\outside.json", nil)
	if err == nil {
		t.Fatalf("DBRollbackImportSnapshot() error = nil, want invalid snapshot name error")
	}
	if !strings.Contains(err.Error(), "invalid snapshot_name") {
		t.Fatalf("DBRollbackImportSnapshot() error = %v, want invalid snapshot_name", err)
	}
}

func TestDBRollbackImportSnapshotRejectsSymlinkOutsideSnapshotDir(t *testing.T) {
	ctx := setupOpTestDB(t)

	snapshotDir, err := importSnapshotDir()
	if err != nil {
		t.Fatalf("importSnapshotDir() error = %v", err)
	}
	if err := os.MkdirAll(snapshotDir, importSnapshotDirPerm); err != nil {
		t.Fatalf("os.MkdirAll(snapshotDir) error = %v", err)
	}

	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "outside.json")
	if err := os.WriteFile(outsidePath, []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("os.WriteFile(outsidePath) error = %v", err)
	}

	linkName := "linked-outside.json"
	linkPath := filepath.Join(snapshotDir, linkName)
	if err := os.Symlink(outsidePath, linkPath); err != nil {
		if runtime.GOOS == "windows" {
			lowerErr := strings.ToLower(err.Error())
			if strings.Contains(lowerErr, "privilege") || strings.Contains(lowerErr, "not held") {
				t.Skipf("os.Symlink() requires elevated privilege on this Windows host: %v", err)
			}
		}
		t.Fatalf("os.Symlink() error = %v", err)
	}

	_, err = DBRollbackImportSnapshot(ctx, linkName, nil)
	if err == nil {
		t.Fatal("DBRollbackImportSnapshot() error = nil, want snapshot path outside import snapshot directory")
	}
	if !strings.Contains(err.Error(), "snapshot path is outside import snapshot directory") {
		t.Fatalf("DBRollbackImportSnapshot() error = %v, want outside snapshot directory", err)
	}
}

func TestDBPreviewRollbackImportSnapshotBuildsCompatibilityAndRowsSummary(t *testing.T) {
	ctx := setupOpTestDB(t)

	currentChannel := model.Channel{
		Name:     "preview-rollback-channel",
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "https://current-preview.example.com/v1", Delay: 0}},
		Model:    "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&currentChannel).Error; err != nil {
		t.Fatalf("create current channel error = %v", err)
	}
	currentGroup := model.Group{Name: "preview-rollback-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&currentGroup).Error; err != nil {
		t.Fatalf("create current group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: currentGroup.ID, ChannelID: currentChannel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create current group item error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{Name: "gpt-4o", CanonicalName: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create current llm info error = %v", err)
	}

	if err := db.GetDB().WithContext(ctx).Create(&model.Channel{Name: "preview-before-import", Enabled: true, Model: "gpt-4o"}).Error; err != nil {
		t.Fatalf("create pre-import channel error = %v", err)
	}
	if _, err := DBImportIncremental(ctx, &model.DBDump{
		Version:  dbDumpVersion,
		Manifest: model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{ID: 3501, Name: "preview-imported-channel", Enabled: true, Model: "gpt-4o"}},
	}, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}

	if err := db.GetDB().WithContext(ctx).Delete(&model.Channel{}, "name = ?", "preview-before-import").Error; err != nil {
		t.Fatalf("delete pre-import channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.LLMInfo{}, "name = ?", "gpt-4o").Error; err != nil {
		t.Fatalf("delete current llm info error = %v", err)
	}
	currentChannel.BaseUrls = []model.BaseUrl{{URL: "https://current-mutated.example.com/v1", Delay: 0}}
	currentChannel.Model = "gpt-4o,o1-mini"
	if err := db.GetDB().WithContext(ctx).Save(&currentChannel).Error; err != nil {
		t.Fatalf("mutate current channel error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: currentGroup.ID, ChannelID: currentChannel.ID, ModelName: "o1-mini", Priority: 2, Weight: 1}).Error; err != nil {
		t.Fatalf("create mutated group item error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.StatsTotal{ID: 1, StatsMetrics: model.StatsMetrics{InputToken: 42}}).Error; err != nil {
		t.Fatalf("create stats_total error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.RelayLog{ID: 9201, Time: 1710000010, RequestModelName: "gpt-4o", ChannelId: currentChannel.ID, ChannelName: currentChannel.Name, ActualModelName: "gpt-4o", TotalAttempts: 1}).Error; err != nil {
		t.Fatalf("create relay log error = %v", err)
	}

	snapshots, err := DBListImportSnapshots()
	if err != nil {
		t.Fatalf("DBListImportSnapshots() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatalf("snapshots = %#v, want at least one snapshot", snapshots)
	}

	preview, err := DBPreviewRollbackImportSnapshot(ctx, snapshots[0].SnapshotName, nil)
	if err != nil {
		t.Fatalf("DBPreviewRollbackImportSnapshot() error = %v", err)
	}
	if preview == nil {
		t.Fatalf("preview = nil, want result")
	}
	if preview.Manifest == nil || preview.Manifest.SchemaVersion != "v1" {
		t.Fatalf("preview.manifest = %#v, want schema v1", preview.Manifest)
	}
	if preview.RowsSummary == nil {
		t.Fatalf("preview.rows_summary = nil, want populated summary")
	}
	if got := preview.RowsSummary["channels"]; got == 0 {
		t.Fatalf("preview.rows_summary[channels] = %d, want > 0", got)
	}
	if preview.Compatibility == nil || preview.Compatibility.Summary == nil {
		t.Fatalf("preview.compatibility = %#v, want populated compatibility report", preview.Compatibility)
	}
	if got := preview.Compatibility.Summary.RoutePreviewDiffs; got == 0 {
		t.Fatalf("preview.compatibility.summary.route_preview_diffs = %d, want > 0", got)
	}
	if !containsWarning(preview.Compatibility.BaseURLMismatches, "preview-rollback-channel") {
		t.Fatalf("preview.compatibility.base_url_mismatches = %#v, want preview-rollback-channel", preview.Compatibility.BaseURLMismatches)
	}
	if !containsWarning(preview.Compatibility.MissingProviders, "preview-before-import") {
		t.Fatalf("preview.compatibility.missing_providers = %#v, want preview-before-import", preview.Compatibility.MissingProviders)
	}
	if !containsWarning(preview.Compatibility.MissingModels, "gpt-4o") {
		t.Fatalf("preview.compatibility.missing_models = %#v, want gpt-4o", preview.Compatibility.MissingModels)
	}
	if !containsWarning(preview.Compatibility.ReplacePrunedChannels, "preview-imported-channel") {
		t.Fatalf("preview.compatibility.replace_pruned_channels = %#v, want preview-imported-channel", preview.Compatibility.ReplacePrunedChannels)
	}
	if got := preview.Compatibility.Summary.ReplacePrunedChannels; got == 0 {
		t.Fatalf("preview.compatibility.summary.replace_pruned_channels = %d, want > 0", got)
	}
	if got := preview.Compatibility.Summary.ReplacePrunedGroups; got != 0 {
		t.Fatalf("preview.compatibility.summary.replace_pruned_groups = %d, want 0", got)
	}
	if !containsWarning(preview.PreviewWarnings, "route preview diffs") {
		t.Fatalf("preview.preview_warnings = %#v, want route preview warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "invalid route targets") && !containsWarning(preview.PreviewWarnings, "skipped route target previews") {
		t.Fatalf("preview.preview_warnings = %#v, want route preview compatibility warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "base URL mismatches") {
		t.Fatalf("preview.preview_warnings = %#v, want base URL mismatch warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "includes stats tables") {
		t.Fatalf("preview.preview_warnings = %#v, want stats warning", preview.PreviewWarnings)
	}
	if !containsWarning(preview.PreviewWarnings, "includes relay logs") {
		t.Fatalf("preview.preview_warnings = %#v, want relay logs warning", preview.PreviewWarnings)
	}
}

func TestDBImportIncrementalSkipsGroupItemsForUndeclaredChannelModels(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:      501,
			Name:    "import-channel",
			Enabled: true,
			Model:   "gpt-4o",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:         601,
			ChannelID:  501,
			Enabled:    true,
			ChannelKey: "import-key",
		}},
		Groups: []model.Group{{ID: 701, Name: "import-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{
			{ID: 801, GroupID: 701, ChannelID: 501, ModelName: "gpt-4o", Priority: 1, Weight: 1},
			{ID: 802, GroupID: 701, ChannelID: 501, ModelName: "claude-3-5-sonnet", Priority: 2, Weight: 1},
		},
		LLMInfos: []model.LLMInfo{{Name: "gpt-4o", CanonicalName: "gpt-4o"}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["group_items"]; got != 1 {
		t.Fatalf("rows_affected[group_items] = %d, want 1", got)
	}
	if !containsWarning(res.Warnings, "does not declare model:claude-3-5-sonnet") {
		t.Fatalf("warnings = %#v, want undeclared model warning", res.Warnings)
	}

	var groupItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Order("priority asc").Find(&groupItems).Error; err != nil {
		t.Fatalf("query group items error = %v", err)
	}
	if len(groupItems) != 1 || groupItems[0].ModelName != "gpt-4o" {
		t.Fatalf("group items after import = %#v, want only gpt-4o", groupItems)
	}

	if err := InitCache(); err != nil {
		t.Fatalf("InitCache() error = %v", err)
	}
	group, err := GroupGetMap("import-group", ctx)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	if len(group.Items) != 1 || group.Items[0].ModelName != "gpt-4o" {
		t.Fatalf("group cache items after import = %#v, want only gpt-4o", group.Items)
	}
}

func TestDBImportIncrementalApplyReportsPostImportValidation(t *testing.T) {
	ctx := setupOpTestDB(t)

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{
			{
				ID:       1001,
				Name:     "disabled-channel",
				Type:     transformerOutbound.OutboundTypeOpenAIChat,
				Enabled:  false,
				BaseUrls: []model.BaseUrl{{URL: "https://disabled.example.com/v1", Delay: 1}},
				Model:    "gpt-4o",
			},
			{
				ID:       1002,
				Name:     "no-key-channel",
				Type:     transformerOutbound.OutboundTypeOpenAIChat,
				Enabled:  true,
				BaseUrls: []model.BaseUrl{{URL: "https://nokey.example.com/v1", Delay: 1}},
				Model:    "gpt-4o",
			},
		},
		ChannelKeys: []model.ChannelKey{{
			ID:         1101,
			ChannelID:  1001,
			Enabled:    true,
			ChannelKey: "disabled-key",
		}},
		Groups: []model.Group{
			{ID: 1201, Name: "disabled-group", Mode: model.GroupModeRoundRobin},
			{ID: 1202, Name: "no-key-group", Mode: model.GroupModeRoundRobin},
			{ID: 1203, Name: "empty-group", Mode: model.GroupModeRoundRobin},
		},
		GroupItems: []model.GroupItem{
			{ID: 1301, GroupID: 1201, ChannelID: 1001, ModelName: "gpt-4o", Priority: 1, Weight: 1},
			{ID: 1302, GroupID: 1202, ChannelID: 1002, ModelName: "gpt-4o", Priority: 1, Weight: 1},
		},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if res.PostImportValidation == nil || res.PostImportValidation.Summary == nil {
		t.Fatalf("post_import_validation = %#v, want populated report", res.PostImportValidation)
	}
	if got := res.PostImportValidation.Summary.DegradedGroups; got != 3 {
		t.Fatalf("summary.degraded_groups = %d, want 3", got)
	}
	if got := res.PostImportValidation.Summary.EmptyGroups; got != 1 {
		t.Fatalf("summary.empty_groups = %d, want 1", got)
	}
	if got := res.PostImportValidation.Summary.DisabledChannels; got != 1 {
		t.Fatalf("summary.disabled_channels = %d, want 1", got)
	}
	if got := res.PostImportValidation.Summary.ChannelsWithoutKeys; got != 1 {
		t.Fatalf("summary.channels_without_keys = %d, want 1", got)
	}
	if got := res.PostImportValidation.Summary.StaleItemsRemoved; got != 0 {
		t.Fatalf("summary.stale_items_removed = %d, want 0", got)
	}
	if got := res.PostImportValidation.Summary.RouteWarnings; got != 1 {
		t.Fatalf("summary.route_warnings = %d, want 1", got)
	}
	if !containsWarning(res.Warnings, "post-import validation found 3 degraded groups") {
		t.Fatalf("warnings = %#v, want degraded groups warning", res.Warnings)
	}
	if !containsWarning(res.Warnings, "post-import validation found 1 route warnings") {
		t.Fatalf("warnings = %#v, want route warning summary", res.Warnings)
	}
	if !containsWarning(res.PostImportValidation.DegradedGroups, "disabled-group") || !containsWarning(res.PostImportValidation.DegradedGroups, "no-key-group") || !containsWarning(res.PostImportValidation.DegradedGroups, "empty-group") {
		t.Fatalf("degraded_groups = %#v, want disabled/no-key/empty groups", res.PostImportValidation.DegradedGroups)
	}
	if !containsWarning(res.PostImportValidation.EmptyGroups, "empty-group") {
		t.Fatalf("empty_groups = %#v, want empty-group", res.PostImportValidation.EmptyGroups)
	}
	if !containsWarning(res.PostImportValidation.DisabledChannels, "disabled-channel") {
		t.Fatalf("disabled_channels = %#v, want disabled-channel", res.PostImportValidation.DisabledChannels)
	}
	if !containsWarning(res.PostImportValidation.ChannelsWithoutKeys, "no-key-channel") {
		t.Fatalf("channels_without_keys = %#v, want no-key-channel", res.PostImportValidation.ChannelsWithoutKeys)
	}
	if !containsWarning(res.PostImportValidation.RouteWarnings, "disabled-group") {
		t.Fatalf("route_warnings = %#v, want disabled-group route mismatch", res.PostImportValidation.RouteWarnings)
	}
	if !containsWarning(res.PostImportValidation.RouteWarnings, "disabled-channel") {
		t.Fatalf("route_warnings = %#v, want disabled-channel route mismatch", res.PostImportValidation.RouteWarnings)
	}
}

func TestDBImportIncrementalApplyReportsRouteWarningsForMixedExistingRoutes(t *testing.T) {
	ctx := setupOpTestDB(t)

	legacyChannel := model.Channel{Name: "legacy-extra-channel", Enabled: true, Model: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&legacyChannel).Error; err != nil {
		t.Fatalf("create legacy channel error = %v", err)
	}
	legacyKey := model.ChannelKey{ChannelID: legacyChannel.ID, Enabled: true, ChannelKey: "legacy-extra-key", AllowedModels: "gpt-4o"}
	if err := db.GetDB().WithContext(ctx).Create(&legacyKey).Error; err != nil {
		t.Fatalf("create legacy key error = %v", err)
	}
	group := model.Group{Name: "mixed-route-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&group).Error; err != nil {
		t.Fatalf("create group error = %v", err)
	}
	legacyItem := model.GroupItem{GroupID: group.ID, ChannelID: legacyChannel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}
	if err := db.GetDB().WithContext(ctx).Create(&legacyItem).Error; err != nil {
		t.Fatalf("create legacy group item error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:      2001,
			Name:    "snapshot-primary-channel",
			Enabled: true,
			Model:   "gpt-4o",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            2101,
			ChannelID:     2001,
			Enabled:       true,
			ChannelKey:    "snapshot-primary-key",
			AllowedModels: "gpt-4o",
		}},
		Groups: []model.Group{{ID: 2201, Name: "mixed-route-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{
			ID:        2301,
			GroupID:   2201,
			ChannelID: 2001,
			ModelName: "gpt-4o",
			Priority:  1,
			Weight:    1,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if res.PostImportValidation == nil || res.PostImportValidation.Summary == nil {
		t.Fatalf("post_import_validation = %#v, want populated report", res.PostImportValidation)
	}
	if got := res.PostImportValidation.Summary.RouteWarnings; got != 1 {
		t.Fatalf("summary.route_warnings = %d, want 1", got)
	}
	if !containsWarning(res.Warnings, "post-import validation found 1 route warnings") {
		t.Fatalf("warnings = %#v, want route warning summary", res.Warnings)
	}
	if !containsWarning(res.PostImportValidation.RouteWarnings, "mixed-route-group") {
		t.Fatalf("route_warnings = %#v, want mixed-route-group context", res.PostImportValidation.RouteWarnings)
	}
	if !containsWarning(res.PostImportValidation.RouteWarnings, "legacy-extra-channel") {
		t.Fatalf("route_warnings = %#v, want legacy-extra-channel evidence", res.PostImportValidation.RouteWarnings)
	}
	if !containsWarning(res.PostImportValidation.RouteWarnings, "snapshot-primary-channel") {
		t.Fatalf("route_warnings = %#v, want snapshot-primary-channel evidence", res.PostImportValidation.RouteWarnings)
	}
	if !containsWarning(res.PostImportValidation.RouteWarnings, "candidate_count:1->2") {
		t.Fatalf("route_warnings = %#v, want candidate count drift", res.PostImportValidation.RouteWarnings)
	}
}

func TestDBImportIncrementalApplyReportsPriceRuleAndAliasValidation(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{
		Name:                  "gpt-5.4-pro",
		CanonicalName:         "gpt-5.4-pro",
		BillingMode:           model.BillingModePerToken,
		ProbePolicy:           model.ProbePolicyConcurrent,
		ProbeIntervalSeconds:  30,
		ProbeConcurrencyLimit: 3,
	}).Error; err != nil {
		t.Fatalf("create current llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:      2001,
			Name:    "alias-check-channel",
			Enabled: true,
			Model:   "gpt54pro",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            2101,
			ChannelID:     2001,
			Enabled:       true,
			ChannelKey:    "alias-check-key",
			AllowedModels: "gpt54pro",
		}},
		APIKeys: []model.APIKey{{
			ID:              2201,
			Name:            "alias-client",
			APIKey:          "sk-alias-client",
			Enabled:         true,
			SupportedModels: "gpt54pro",
		}},
		Groups:     []model.Group{{ID: 2301, Name: "alias-check-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 2401, GroupID: 2301, ChannelID: 2001, ModelName: "gpt54pro", Priority: 1, Weight: 1}},
		LLMInfos: []model.LLMInfo{{
			Name:                  "gpt54pro",
			CanonicalName:         "gpt-5.4-pro",
			BillingMode:           model.BillingModePerRequest,
			ProbePolicy:           model.ProbePolicyPassiveOnly,
			ProbeIntervalSeconds:  600,
			ProbeConcurrencyLimit: 1,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if res.PostImportValidation == nil || res.PostImportValidation.Summary == nil {
		t.Fatalf("post_import_validation = %#v, want populated report", res.PostImportValidation)
	}
	if got := res.PostImportValidation.Summary.PriceRuleWarnings; got == 0 {
		t.Fatalf("summary.price_rule_warnings = %d, want > 0", got)
	}
	if got := res.PostImportValidation.Summary.AliasMappings; got == 0 {
		t.Fatalf("summary.alias_mappings = %d, want > 0", got)
	}
	if got := res.PostImportValidation.Summary.AliasWarnings; got == 0 {
		t.Fatalf("summary.alias_warnings = %d, want > 0", got)
	}
	if !containsWarning(res.Warnings, "price rule warnings") {
		t.Fatalf("warnings = %#v, want price rule validation warning", res.Warnings)
	}
	if !containsWarning(res.Warnings, "alias warnings") {
		t.Fatalf("warnings = %#v, want alias validation warning", res.Warnings)
	}
	if !containsWarning(res.PostImportValidation.PriceRuleWarnings, "billing_mode changed") {
		t.Fatalf("price_rule_warnings = %#v, want billing_mode changed warning", res.PostImportValidation.PriceRuleWarnings)
	}
	if !containsWarning(res.PostImportValidation.PriceRuleWarnings, "probe_policy changed") {
		t.Fatalf("price_rule_warnings = %#v, want probe_policy changed warning", res.PostImportValidation.PriceRuleWarnings)
	}
	if !containsWarning(res.PostImportValidation.PriceRuleWarnings, "concurrent probe/race") {
		t.Fatalf("price_rule_warnings = %#v, want paid-like concurrency warning", res.PostImportValidation.PriceRuleWarnings)
	}
	if !containsWarning(res.PostImportValidation.AliasMappings, "remapped to current alias:gpt-5.4-pro") {
		t.Fatalf("alias_mappings = %#v, want canonical alias remap", res.PostImportValidation.AliasMappings)
	}
	if !containsWarning(res.PostImportValidation.AliasMappings, "group route model:gpt54pro can map to current alias:gpt-5.4-pro") {
		t.Fatalf("alias_mappings = %#v, want group route alias mapping", res.PostImportValidation.AliasMappings)
	}
	if !containsWarning(res.PostImportValidation.AliasWarnings, "channel:alias-check-channel model:gpt54pro resolves to current alias:gpt-5.4-pro") {
		t.Fatalf("alias_warnings = %#v, want channel alias warning", res.PostImportValidation.AliasWarnings)
	}
	if !containsWarning(res.PostImportValidation.AliasWarnings, "channel_key:2101 model:gpt54pro resolves to current alias:gpt-5.4-pro") {
		t.Fatalf("alias_warnings = %#v, want channel key alias warning", res.PostImportValidation.AliasWarnings)
	}
	if !containsWarning(res.PostImportValidation.AliasWarnings, "api_key:alias-client model:gpt54pro resolves to current alias:gpt-5.4-pro") {
		t.Fatalf("alias_warnings = %#v, want api key alias warning", res.PostImportValidation.AliasWarnings)
	}
}

func TestDBImportIncrementalApplyCleansHistoricalStaleGroupItems(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := &model.Channel{
		Name:     "historical-channel",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: "https://historical.example.com/v1", Delay: 1}},
		Model:    "gpt-4o",
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	if _, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: channel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "historical-key",
			SourceType:    "public/free",
			AllowedModels: "gpt-4o",
		}},
	}, ctx); err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}

	group := &model.Group{Name: "historical-group", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(group, ctx); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}
	validItem := model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}
	if err := GroupItemAdd(&validItem, ctx); err != nil {
		t.Fatalf("GroupItemAdd(valid) error = %v", err)
	}
	staleItem := model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "claude-3-5-sonnet", Priority: 2, Weight: 1}
	if err := db.GetDB().WithContext(ctx).Create(&staleItem).Error; err != nil {
		t.Fatalf("create stale group item error = %v", err)
	}
	if err := groupRefreshCacheByID(group.ID, ctx); err != nil {
		t.Fatalf("groupRefreshCacheByID() error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion: "v1",
			ExportSource:  "octopus",
		},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if got := res.RowsAffected["cleaned_group_items"]; got != 1 {
		t.Fatalf("rows_affected[cleaned_group_items] = %d, want 1", got)
	}
	if !containsWarning(res.Warnings, "cleaned 1 stale group items after import") {
		t.Fatalf("warnings = %#v, want stale cleanup warning", res.Warnings)
	}
	if res.PostImportValidation == nil || res.PostImportValidation.Summary == nil {
		t.Fatalf("post_import_validation = %#v, want populated report", res.PostImportValidation)
	}
	if got := res.PostImportValidation.Summary.StaleItemsRemoved; got != 1 {
		t.Fatalf("summary.stale_items_removed = %d, want 1", got)
	}
	if got := res.PostImportValidation.Summary.DegradedGroups; got != 0 {
		t.Fatalf("summary.degraded_groups = %d, want 0", got)
	}
	if got := res.PostImportValidation.Summary.RouteWarnings; got != 0 {
		t.Fatalf("summary.route_warnings = %d, want 0", got)
	}
	if !containsWarning(res.PostImportValidation.StaleItemsRemoved, "claude-3-5-sonnet") {
		t.Fatalf("stale_items_removed = %#v, want stale claude item message", res.PostImportValidation.StaleItemsRemoved)
	}

	var storedItems []model.GroupItem
	if err := db.GetDB().WithContext(ctx).Where("group_id = ?", group.ID).Order("priority asc").Find(&storedItems).Error; err != nil {
		t.Fatalf("query group items error = %v", err)
	}
	if len(storedItems) != 1 || storedItems[0].ModelName != "gpt-4o" {
		t.Fatalf("stored group items after cleanup = %#v, want only gpt-4o", storedItems)
	}

	refreshedGroup, err := GroupGetMap(group.Name, ctx)
	if err != nil {
		t.Fatalf("GroupGetMap() error = %v", err)
	}
	if len(refreshedGroup.Items) != 1 || refreshedGroup.Items[0].ModelName != "gpt-4o" {
		t.Fatalf("group cache items after cleanup = %#v, want only gpt-4o", refreshedGroup.Items)
	}
}

func TestCheckChannelModelHealthTreats200And429AsReachable(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelName := "gpt-4o"

	statusCode := http.StatusOK
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	channel := &model.Channel{
		Name:     "health-200-channel",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:    modelName,
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: channel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "health-key",
			SourceType:    "public/free",
			AllowedModels: modelName,
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}

	result := CheckChannelModelHealth(ctx, updated, modelName)
	if !result.Passed || result.Skipped {
		t.Fatalf("200 result = %#v, want passed", result)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("200 result status code = %d, want %d", result.StatusCode, http.StatusOK)
	}

	statusCode = http.StatusTooManyRequests
	result = CheckChannelModelHealth(ctx, updated, modelName)
	if !result.Passed || result.Skipped || !result.RateLimited {
		t.Fatalf("429 result = %#v, want passed + rate limited", result)
	}
	if result.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("429 result status code = %d, want %d", result.StatusCode, http.StatusTooManyRequests)
	}
}

func TestCheckChannelModelHealthSkipsDisabledAndNoKey(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelName := "gpt-4o"
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	disabledChannel := &model.Channel{
		Name:     "health-disabled-channel",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:    modelName,
	}
	if err := ChannelCreate(disabledChannel, ctx); err != nil {
		t.Fatalf("ChannelCreate(disabled) error = %v", err)
	}
	if err := ChannelEnabled(disabledChannel.ID, false, ctx); err != nil {
		t.Fatalf("ChannelEnabled(false) error = %v", err)
	}
	disabledChannel, err := ChannelGet(disabledChannel.ID, ctx)
	if err != nil {
		t.Fatalf("ChannelGet(disabled) error = %v", err)
	}
	disabledResult := CheckChannelModelHealth(ctx, disabledChannel, modelName)
	if !disabledResult.Skipped || disabledResult.Passed {
		t.Fatalf("disabled result = %#v, want skipped", disabledResult)
	}
	if !strings.Contains(disabledResult.Error, "is disabled") {
		t.Fatalf("disabled error = %q, want disabled message", disabledResult.Error)
	}

	noKeyChannel := &model.Channel{
		Name:     "health-no-key-channel",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:    modelName,
	}
	if err := ChannelCreate(noKeyChannel, ctx); err != nil {
		t.Fatalf("ChannelCreate(no-key) error = %v", err)
	}
	noKeyResult := CheckChannelModelHealth(ctx, noKeyChannel, modelName)
	if !noKeyResult.Skipped || noKeyResult.Passed {
		t.Fatalf("no-key result = %#v, want skipped", noKeyResult)
	}
	if !strings.Contains(noKeyResult.Error, "no available key") {
		t.Fatalf("no-key error = %q, want no available key", noKeyResult.Error)
	}
}

func TestCheckChannelModelHealthAppliesCustomHeader(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelName := "gpt-4o"
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		if got := r.Header.Get("X-Workspace"); got != "team-alpha" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"missing workspace header"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer server.Close()

	channel := &model.Channel{
		Name:         "health-custom-header-channel",
		Type:         transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:      true,
		BaseUrls:     []model.BaseUrl{{URL: server.URL, Delay: 1}},
		Model:        modelName,
		CustomHeader: []model.CustomHeader{{HeaderKey: "X-Workspace", HeaderValue: "team-alpha"}},
	}
	if err := ChannelCreate(channel, ctx); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: channel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "health-key",
			SourceType:    "public/free",
			AllowedModels: modelName,
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("ChannelUpdate(keys) error = %v", err)
	}

	result := CheckChannelModelHealth(ctx, updated, modelName)
	if !result.Passed || result.Skipped {
		t.Fatalf("custom-header result = %#v, want passed", result)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("custom-header result status code = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestDBImportIncrementalApplyAddsPostImportHealthCheck(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelName := "gpt-4o"

	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer server.Close()

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:       1501,
			Name:     "health-import-channel",
			Type:     transformerOutbound.OutboundTypeOpenAIChat,
			Enabled:  true,
			BaseUrls: []model.BaseUrl{{URL: server.URL, Delay: 1}},
			Model:    modelName,
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:         1601,
			ChannelID:  1501,
			Enabled:    true,
			ChannelKey: "import-health-key",
		}},
		Groups:     []model.Group{{ID: 1701, Name: "health-import-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 1801, GroupID: 1701, ChannelID: 1501, ModelName: modelName, Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if res.PostImportValidation == nil || res.PostImportValidation.HealthCheck == nil || res.PostImportValidation.HealthCheck.Summary == nil {
		t.Fatalf("post_import_validation.health_check = %#v, want populated report", res.PostImportValidation)
	}
	if got := res.PostImportValidation.HealthCheck.Summary.Targets; got != 1 {
		t.Fatalf("health_check.summary.targets = %d, want 1", got)
	}
	if got := res.PostImportValidation.HealthCheck.Summary.Passed; got != 1 {
		t.Logf("health check report = %#v", res.PostImportValidation.HealthCheck)
		t.Fatalf("health_check.summary.passed = %d, want 1", got)
	}
	if len(res.PostImportValidation.HealthCheck.Checks) != 1 {
		t.Fatalf("health_check.checks len = %d, want 1", len(res.PostImportValidation.HealthCheck.Checks))
	}
	check := res.PostImportValidation.HealthCheck.Checks[0]
	if check.GroupName != "health-import-group" || check.ChannelName != "health-import-channel" || check.Model != modelName || !check.Passed {
		t.Fatalf("health check = %#v, want passing import target", check)
	}
}

func TestDBImportIncrementalApplyHealthCheckTargetsOnlyImportedRoutes(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelName := "gpt-4o"

	historicalServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"historical"}`))
	}))
	defer historicalServer.Close()

	historicalChannel := &model.Channel{
		Name:     "historical-health-channel",
		Type:     transformerOutbound.OutboundTypeOpenAIChat,
		Enabled:  true,
		BaseUrls: []model.BaseUrl{{URL: historicalServer.URL, Delay: 1}},
		Model:    modelName,
	}
	if err := ChannelCreate(historicalChannel, ctx); err != nil {
		t.Fatalf("ChannelCreate(historical) error = %v", err)
	}
	if _, err := ChannelUpdate(&model.ChannelUpdateRequest{
		ID: historicalChannel.ID,
		KeysToAdd: []model.ChannelKeyAddRequest{{
			Enabled:       true,
			ChannelKey:    "historical-health-key",
			SourceType:    "public/free",
			AllowedModels: modelName,
		}},
	}, ctx); err != nil {
		t.Fatalf("ChannelUpdate(historical keys) error = %v", err)
	}
	historicalGroup := &model.Group{Name: "historical-health-group", Mode: model.GroupModeRoundRobin}
	if err := GroupCreate(historicalGroup, ctx); err != nil {
		t.Fatalf("GroupCreate(historical) error = %v", err)
	}
	historicalItem := model.GroupItem{GroupID: historicalGroup.ID, ChannelID: historicalChannel.ID, ModelName: modelName, Priority: 1, Weight: 1}
	if err := GroupItemAdd(&historicalItem, ctx); err != nil {
		t.Fatalf("GroupItemAdd(historical) error = %v", err)
	}

	importServer := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"imported"}`))
	}))
	defer importServer.Close()

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus",
			ContainsSecrets: true,
		},
		Channels: []model.Channel{{
			ID:       2501,
			Name:     "imported-health-channel",
			Type:     transformerOutbound.OutboundTypeOpenAIChat,
			Enabled:  true,
			BaseUrls: []model.BaseUrl{{URL: importServer.URL, Delay: 1}},
			Model:    modelName,
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:         2601,
			ChannelID:  2501,
			Enabled:    true,
			ChannelKey: "imported-health-key",
		}},
		Groups:     []model.Group{{ID: 2701, Name: "imported-health-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 2801, GroupID: 2701, ChannelID: 2501, ModelName: modelName, Priority: 1, Weight: 1}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=false) error = %v", err)
	}
	if res.PostImportValidation == nil || res.PostImportValidation.HealthCheck == nil || res.PostImportValidation.HealthCheck.Summary == nil {
		t.Fatalf("post_import_validation.health_check = %#v, want populated report", res.PostImportValidation)
	}
	if got := res.PostImportValidation.HealthCheck.Summary.Targets; got != 1 {
		t.Fatalf("health_check.summary.targets = %d, want 1", got)
	}
	if got := res.PostImportValidation.HealthCheck.Summary.TargetGroups; got != 1 {
		t.Fatalf("health_check.summary.target_groups = %d, want 1", got)
	}
	if len(res.PostImportValidation.HealthCheck.Checks) != 1 {
		t.Fatalf("health_check.checks len = %d, want 1", len(res.PostImportValidation.HealthCheck.Checks))
	}
	check := res.PostImportValidation.HealthCheck.Checks[0]
	if check.GroupName != "imported-health-group" || check.ChannelName != "imported-health-channel" || check.Model != modelName {
		t.Fatalf("health check = %#v, want imported target only", check)
	}
	if strings.Contains(check.GroupName, "historical") || strings.Contains(check.ChannelName, "historical") {
		t.Fatalf("health check unexpectedly included historical route = %#v", check)
	}
}

func TestRunImportHealthChecksHonorsConcurrencyLimit(t *testing.T) {
	ctx := setupOpTestDB(t)
	modelName := "gpt-4o"

	var inFlight int64
	var maxInFlight int64
	server := newIPv4TestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		current := atomic.AddInt64(&inFlight, 1)
		for {
			previous := atomic.LoadInt64(&maxInFlight)
			if current <= previous || atomic.CompareAndSwapInt64(&maxInFlight, previous, current) {
				break
			}
		}
		time.Sleep(120 * time.Millisecond)
		atomic.AddInt64(&inFlight, -1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ok"}`))
	}))
	defer server.Close()

	targets := make([]ChannelModelHealthCheckTarget, 0, 5)
	for i := 0; i < 5; i++ {
		channel := &model.Channel{
			Name:     "concurrency-channel-" + string(rune('a'+i)),
			Type:     transformerOutbound.OutboundTypeOpenAIChat,
			Enabled:  true,
			BaseUrls: []model.BaseUrl{{URL: server.URL, Delay: 1}},
			Model:    modelName,
		}
		if err := ChannelCreate(channel, ctx); err != nil {
			t.Fatalf("ChannelCreate(%d) error = %v", i, err)
		}
		updated, err := ChannelUpdate(&model.ChannelUpdateRequest{
			ID: channel.ID,
			KeysToAdd: []model.ChannelKeyAddRequest{{
				Enabled:       true,
				ChannelKey:    "concurrency-key-" + string(rune('a'+i)),
				SourceType:    "public/free",
				AllowedModels: modelName,
			}},
		}, ctx)
		if err != nil {
			t.Fatalf("ChannelUpdate(%d) error = %v", i, err)
		}
		targets = append(targets, ChannelModelHealthCheckTarget{
			GroupName: "concurrency-group",
			ChannelID: updated.ID,
			ModelName: modelName,
		})
	}

	report := RunImportHealthChecks(ctx, targets)
	if report == nil || report.Summary == nil {
		t.Fatalf("RunImportHealthChecks() = %#v, want populated report", report)
	}
	if got := report.Summary.Targets; got != len(targets) {
		t.Fatalf("summary.targets = %d, want %d", got, len(targets))
	}
	if got := report.Summary.Passed; got != len(targets) {
		t.Fatalf("summary.passed = %d, want %d", got, len(targets))
	}
	if max := atomic.LoadInt64(&maxInFlight); max > importHealthCheckMaxConcurrency {
		t.Fatalf("max concurrent requests = %d, want <= %d", max, importHealthCheckMaxConcurrency)
	}
	for idx := 1; idx < len(report.Checks); idx++ {
		prev, cur := report.Checks[idx-1], report.Checks[idx]
		if prev.GroupName > cur.GroupName || (prev.GroupName == cur.GroupName && prev.ChannelName > cur.ChannelName) || (prev.GroupName == cur.GroupName && prev.ChannelName == cur.ChannelName && prev.Model > cur.Model) {
			t.Fatalf("checks not sorted deterministically: %#v then %#v", prev, cur)
		}
	}
}

func TestDBImportIncrementalDryRunRoutePreviewDiffDetectsRouteTargetOverridePolicyChanges(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.LLMInfo{
		Name:                  "gpt-4o",
		CanonicalName:         "gpt-4o",
		BillingMode:           model.BillingModePerToken,
		ProbePolicy:           model.ProbePolicySequential,
		ProbeIntervalSeconds:  300,
		ProbeConcurrencyLimit: 2,
	}).Error; err != nil {
		t.Fatalf("create current llm info error = %v", err)
	}

	channel := model.Channel{
		Name:              "override-preview-channel",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModeClassified,
		BaseUrls:          []model.BaseUrl{{URL: "https://current.example.com/v1", Delay: 0}},
		Model:             "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	key := model.ChannelKey{
		ChannelID:     channel.ID,
		Enabled:       true,
		ChannelKey:    "override-preview-key",
		SourceType:    "paid/metered",
		AllowedModels: "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&key).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	group := model.Group{Name: "override-preview-group", Mode: model.GroupModeRoundRobin}
	if err := db.GetDB().WithContext(ctx).Create(&group).Error; err != nil {
		t.Fatalf("create group error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.GroupItem{GroupID: group.ID, ChannelID: channel.ID, ModelName: "gpt-4o", Priority: 1, Weight: 1}).Error; err != nil {
		t.Fatalf("create group item error = %v", err)
	}

	dump := &model.DBDump{
		Version: dbDumpVersion,
		Manifest: model.DBDumpManifest{
			SchemaVersion: "v1",
			ExportSource:  "octopus",
		},
		Channels: []model.Channel{{
			ID:                101,
			Name:              "override-preview-channel",
			Enabled:           true,
			KeyManagementMode: model.KeyManagementModeClassified,
			BaseUrls:          []model.BaseUrl{{URL: "https://snapshot.example.com/v1", Delay: 0}},
			Model:             "gpt-4o",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            201,
			ChannelID:     101,
			Enabled:       true,
			ChannelKey:    "override-preview-key",
			SourceType:    "paid/metered",
			AllowedModels: "gpt-4o",
		}},
		RouteTargetOverrides: []model.RouteTargetOverride{{
			ID:                    301,
			ChannelID:             101,
			ChannelKeyID:          201,
			ModelName:             "gpt-4o",
			BillingMode:           model.BillingModePerRequest,
			ProbePolicy:           model.ProbePolicyConcurrent,
			ProbeIntervalSeconds:  30,
			ProbeConcurrencyLimit: 4,
		}},
		Groups:     []model.Group{{ID: 401, Name: "override-preview-group", Mode: model.GroupModeRoundRobin}},
		GroupItems: []model.GroupItem{{ID: 501, GroupID: 401, ChannelID: 101, ModelName: "gpt-4o", Priority: 1, Weight: 1}},
		LLMInfos: []model.LLMInfo{{
			Name:                  "gpt-4o",
			CanonicalName:         "gpt-4o",
			BillingMode:           model.BillingModePerToken,
			ProbePolicy:           model.ProbePolicySequential,
			ProbeIntervalSeconds:  300,
			ProbeConcurrencyLimit: 2,
		}},
	}

	res, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, true)
	if err != nil {
		t.Fatalf("DBImportIncremental(..., dryRun=true) error = %v", err)
	}

	var diff *model.DBImportRoutePreviewDiff
	for i := range res.Compatibility.RoutePreviewDiffs {
		candidate := &res.Compatibility.RoutePreviewDiffs[i]
		if candidate.GroupName == "override-preview-group" && candidate.Model == "gpt-4o" {
			diff = candidate
			break
		}
	}
	if diff == nil {
		t.Fatalf("route_preview_diffs = %#v, want override-preview-group/gpt-4o diff", res.Compatibility.RoutePreviewDiffs)
	}
	if len(diff.BeforeCandidates) != 1 || len(diff.AfterCandidates) != 1 {
		t.Fatalf("route_preview_diff candidates = %#v, want exactly one before and one after candidate", diff)
	}

	beforeCandidate := diff.BeforeCandidates[0]
	afterCandidate := diff.AfterCandidates[0]
	if beforeCandidate.KeyID != key.ID || afterCandidate.KeyID != key.ID {
		t.Fatalf("candidate key ids = before:%d after:%d, want both %d", beforeCandidate.KeyID, afterCandidate.KeyID, key.ID)
	}
	if beforeCandidate.BillingMode != string(model.BillingModePerToken) {
		t.Fatalf("before_candidate.billing_mode = %q, want per_token", beforeCandidate.BillingMode)
	}
	if afterCandidate.BillingMode != string(model.BillingModePerRequest) {
		t.Fatalf("after_candidate.billing_mode = %q, want per_request", afterCandidate.BillingMode)
	}
	if beforeCandidate.ProbePolicy != string(model.ProbePolicySequential) {
		t.Fatalf("before_candidate.probe_policy = %q, want sequential", beforeCandidate.ProbePolicy)
	}
	if afterCandidate.ProbePolicy != string(model.ProbePolicyConcurrent) {
		t.Fatalf("after_candidate.probe_policy = %q, want concurrent", afterCandidate.ProbePolicy)
	}
	if beforeCandidate.ProbeIntervalSeconds != 300 || afterCandidate.ProbeIntervalSeconds != 30 {
		t.Fatalf("candidate probe intervals = before:%d after:%d, want 300 then 30", beforeCandidate.ProbeIntervalSeconds, afterCandidate.ProbeIntervalSeconds)
	}
	if beforeCandidate.ProbeConcurrencyLimit != 2 || afterCandidate.ProbeConcurrencyLimit != 4 {
		t.Fatalf("candidate probe concurrency = before:%d after:%d, want 2 then 4", beforeCandidate.ProbeConcurrencyLimit, afterCandidate.ProbeConcurrencyLimit)
	}
	if beforeCandidate.BillingModeBasis != model.RouteTargetPolicyBasisModelDefault {
		t.Fatalf("before_candidate.billing_mode_basis = %q, want model_default", beforeCandidate.BillingModeBasis)
	}
	if afterCandidate.BillingModeBasis != model.RouteTargetPolicyBasisExplicitOverride {
		t.Fatalf("after_candidate.billing_mode_basis = %q, want route_target_override", afterCandidate.BillingModeBasis)
	}
	if afterCandidate.ProbePolicyBasis != model.RouteTargetPolicyBasisExplicitOverride || afterCandidate.ProbeIntervalBasis != model.RouteTargetPolicyBasisExplicitOverride || afterCandidate.ProbeConcurrencyBasis != model.RouteTargetPolicyBasisExplicitOverride {
		t.Fatalf("after_candidate override bases = %#v, want route_target_override for all policy fields", afterCandidate)
	}
	if !containsWarning(diff.SkipReasons, "route_candidates_changed") {
		t.Fatalf("route_preview_diff skip_reasons = %#v, want route_candidates_changed", diff.SkipReasons)
	}
	if len(diff.RemovedCandidates) != 1 || len(diff.AddedCandidates) != 1 {
		t.Fatalf("route_preview_diff delta candidates = %#v, want one removed and one added candidate", diff)
	}
}

func TestDBImportIncrementalLegacyDumpPreservesRicherExistingFields(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := model.Channel{
		Name:              "legacy-channel",
		Enabled:           true,
		KeyManagementMode: model.KeyManagementModeClassified,
		KeyRoutingPolicy:  model.KeyRoutingPolicyFillPriority,
		BaseUrls:          []model.BaseUrl{{URL: "https://current.example.com/v1", Delay: 1}},
		Model:             "gpt-4o",
	}
	if err := db.GetDB().WithContext(ctx).Create(&channel).Error; err != nil {
		t.Fatalf("create channel error = %v", err)
	}
	channelKey := model.ChannelKey{
		ChannelID:     channel.ID,
		Enabled:       true,
		ChannelKey:    "legacy-upstream-key",
		SourceType:    model.ChannelKeySourceTypePaidMetered,
		AllowedModels: "gpt-4o",
		Remark:        "paid-key",
	}
	if err := db.GetDB().WithContext(ctx).Create(&channelKey).Error; err != nil {
		t.Fatalf("create channel key error = %v", err)
	}
	group := model.Group{
		Name:              "legacy-group",
		Mode:              model.GroupModeFailover,
		RetryRounds:       3,
		RetryDelayMs:      250,
		FailoverWindowSec: 360,
		RaceAfterFails:    2,
		RaceConcurrency:   4,
	}
	if err := db.GetDB().WithContext(ctx).Create(&group).Error; err != nil {
		t.Fatalf("create group error = %v", err)
	}
	llmInfo := model.LLMInfo{
		Name:                  "gpt-4o",
		CanonicalName:         "gpt-4o",
		BillingMode:           model.BillingModePerToken,
		ProbePolicy:           model.ProbePolicyConcurrent,
		ProbeIntervalSeconds:  180,
		ProbeConcurrencyLimit: 3,
	}
	if err := db.GetDB().WithContext(ctx).Create(&llmInfo).Error; err != nil {
		t.Fatalf("create llm info error = %v", err)
	}

	dump := &model.DBDump{
		Version:      dbDumpVersion,
		ExportedAt:   time.Now().UTC(),
		IncludeStats: false,
		Manifest: model.DBDumpManifest{
			SchemaVersion:   "v1",
			ExportSource:    "octopus-legacy",
			ContainsSecrets: true,
		},
		LegacyHints: &model.DBDumpLegacyHints{
			Legacy: true,
			ChannelsByName: map[string]model.DBDumpLegacyChannelHint{
				"legacy-channel": {MissingKeyManagementMode: true, MissingKeyRoutingPolicy: true},
			},
			ChannelKeysBySnapshotID: map[int]model.DBDumpLegacyChannelKeyHint{
				201: {MissingSourceType: true, MissingAllowedModels: true},
			},
			GroupsByName: map[string]model.DBDumpLegacyGroupHint{
				"legacy-group": {
					MissingRetryRounds:       true,
					MissingRetryDelayMs:      true,
					MissingFailoverWindowSec: true,
					MissingRaceAfterFails:    true,
					MissingRaceConcurrency:   true,
				},
			},
			LLMInfosByName: map[string]model.DBDumpLegacyLLMInfoHint{
				"gpt-4o": {
					MissingCanonicalName:         true,
					MissingBillingMode:           true,
					MissingProbePolicy:           true,
					MissingProbeIntervalSeconds:  true,
					MissingProbeConcurrencyLimit: true,
				},
			},
		},
		Channels: []model.Channel{{
			ID:       101,
			Name:     "legacy-channel",
			Enabled:  true,
			BaseUrls: []model.BaseUrl{{URL: "https://snapshot.example.com/v1", Delay: 2}},
			Model:    "gpt-4o",
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:         201,
			ChannelID:  101,
			Enabled:    true,
			ChannelKey: "legacy-upstream-key",
			Remark:     "legacy-imported",
		}},
		Groups: []model.Group{{
			ID:   301,
			Name: "legacy-group",
			Mode: model.GroupModeFailover,
		}},
		LLMInfos: []model.LLMInfo{{
			Name:     "gpt-4o",
			LLMPrice: model.LLMPrice{Input: 1.23, Output: 4.56},
		}},
	}

	if _, err := DBImportIncremental(ctx, dump, model.DBImportModeIncremental, false); err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}

	var storedChannel model.Channel
	if err := db.GetDB().WithContext(ctx).Where("name = ?", "legacy-channel").First(&storedChannel).Error; err != nil {
		t.Fatalf("query channel error = %v", err)
	}
	if storedChannel.KeyManagementMode != model.KeyManagementModeClassified {
		t.Fatalf("stored channel key_management_mode = %q, want classified", storedChannel.KeyManagementMode)
	}
	if storedChannel.KeyRoutingPolicy != model.KeyRoutingPolicyFillPriority {
		t.Fatalf("stored channel key_routing_policy = %q, want fill_priority", storedChannel.KeyRoutingPolicy)
	}

	var storedChannelKey model.ChannelKey
	if err := db.GetDB().WithContext(ctx).Where("channel_id = ? AND channel_key = ?", storedChannel.ID, "legacy-upstream-key").First(&storedChannelKey).Error; err != nil {
		t.Fatalf("query channel key error = %v", err)
	}
	if storedChannelKey.SourceType != model.ChannelKeySourceTypePaidMetered {
		t.Fatalf("stored channel key source_type = %q, want paid/metered", storedChannelKey.SourceType)
	}
	if storedChannelKey.AllowedModels != "gpt-4o" {
		t.Fatalf("stored channel key allowed_models = %q, want gpt-4o", storedChannelKey.AllowedModels)
	}

	var storedGroup model.Group
	if err := db.GetDB().WithContext(ctx).Where("name = ?", "legacy-group").First(&storedGroup).Error; err != nil {
		t.Fatalf("query group error = %v", err)
	}
	if storedGroup.RetryRounds != 3 || storedGroup.RetryDelayMs != 250 || storedGroup.FailoverWindowSec != 360 || storedGroup.RaceAfterFails != 2 || storedGroup.RaceConcurrency != 4 {
		t.Fatalf("stored group runtime fields = %#v, want preserved richer values", storedGroup)
	}

	var storedLLM model.LLMInfo
	if err := db.GetDB().WithContext(ctx).Where("name = ?", "gpt-4o").First(&storedLLM).Error; err != nil {
		t.Fatalf("query llm info error = %v", err)
	}
	if storedLLM.CanonicalName != "gpt-4o" {
		t.Fatalf("stored llm canonical_name = %q, want gpt-4o", storedLLM.CanonicalName)
	}
	if storedLLM.BillingMode != model.BillingModePerToken {
		t.Fatalf("stored llm billing_mode = %q, want per_token", storedLLM.BillingMode)
	}
	if storedLLM.ProbePolicy != model.ProbePolicyConcurrent || storedLLM.ProbeIntervalSeconds != 180 || storedLLM.ProbeConcurrencyLimit != 3 {
		t.Fatalf("stored llm probe fields = %#v, want preserved richer values", storedLLM)
	}
	if storedLLM.Input != 1.23 || storedLLM.Output != 4.56 {
		t.Fatalf("stored llm pricing = %#v, want imported prices preserved", storedLLM)
	}
}

func TestExportDumpLegacyViewDropsNewerFields(t *testing.T) {
	dump := &model.DBDump{
		Version:      dbDumpVersion,
		ExportedAt:   time.Now().UTC(),
		IncludeStats: true,
		Manifest:     model.DBDumpManifest{SchemaVersion: "v1", ExportSource: "octopus", ContainsSecrets: true},
		Channels: []model.Channel{{
			ID:                1,
			Name:              "legacy-export-channel",
			KeyManagementMode: model.KeyManagementModeClassified,
			KeyRoutingPolicy:  model.KeyRoutingPolicyPriority,
		}},
		ChannelKeys: []model.ChannelKey{{
			ID:            2,
			ChannelID:     1,
			ChannelKey:    "sk-test",
			SourceType:    model.ChannelKeySourceTypePrivateInternal,
			AllowedModels: "gpt-4o",
		}},
		Groups: []model.Group{{
			ID:                3,
			Name:              "legacy-export-group",
			RetryRounds:       2,
			RetryDelayMs:      100,
			FailoverWindowSec: 300,
			RaceAfterFails:    2,
			RaceConcurrency:   3,
		}},
		LLMInfos: []model.LLMInfo{{
			Name:                  "gpt-4o",
			CanonicalName:         "gpt-4o",
			BillingMode:           model.BillingModePerToken,
			ProbePolicy:           model.ProbePolicyConcurrent,
			ProbeIntervalSeconds:  60,
			ProbeConcurrencyLimit: 2,
		}},
		APIKeys: []model.APIKey{{ID: 4, Name: "client", APIKey: "sk-client", SupportedModels: "gpt-4o"}},
	}

	legacyView := ExportDumpLegacyView(dump)
	if legacyView == nil {
		t.Fatalf("ExportDumpLegacyView() = nil, want view")
	}
	if legacyView.Channels[0].Name != "legacy-export-channel" || legacyView.Channels[0].Type != dump.Channels[0].Type {
		t.Fatalf("legacy channel fields = %#v, want core legacy channel data preserved", legacyView.Channels[0])
	}
	if legacyView.ChannelKeys[0].ChannelKey != "sk-test" || legacyView.ChannelKeys[0].ChannelID != 1 {
		t.Fatalf("legacy channel key fields = %#v, want core legacy key data preserved", legacyView.ChannelKeys[0])
	}
	if legacyView.Groups[0].Name != "legacy-export-group" || legacyView.Groups[0].FirstTokenTimeOut != 0 || legacyView.Groups[0].SessionKeepTime != 0 {
		t.Fatalf("legacy group fields = %#v, want legacy group shape", legacyView.Groups[0])
	}
	if legacyView.LLMInfos[0].Name != "gpt-4o" || legacyView.LLMInfos[0].Input != 0 || legacyView.LLMInfos[0].Output != 0 {
		t.Fatalf("legacy llm info fields = %#v, want legacy llm info shape", legacyView.LLMInfos[0])
	}
	if legacyView.APIKeys[0].Name != "client" || legacyView.APIKeys[0].APIKey != "sk-client" {
		t.Fatalf("legacy api key fields = %#v, want legacy api key shape", legacyView.APIKeys[0])
	}
}

func containsWarning(items []string, want string) bool {
	for _, item := range items {
		if strings.Contains(item, want) {
			return true
		}
	}
	return false
}
