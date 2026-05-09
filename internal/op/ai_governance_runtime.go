package op

import (
	"fmt"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func GovernanceRuntimePolicyGet() model.GovernanceRuntimePolicyView {
	strategy := settingStringOrDefault(model.SettingKeyAIRuntimeStrategy, "highest_success_rate")
	dispatchMode := settingStringOrDefault(model.SettingKeyAIRuntimeDispatchMode, "single_best")
	maxParallelRuns, err := SettingGetInt(model.SettingKeyAIRuntimeMaxParallelRuns)
	if err != nil || maxParallelRuns <= 0 {
		maxParallelRuns = 2
	}
	doubleReviewEnabled := settingBoolOrDefault(model.SettingKeyAIRuntimeDoubleReviewEnabled, false)
	fallbackToDeterministic := settingBoolOrDefault(model.SettingKeyAIRuntimeFallbackDeterministic, true)
	return model.GovernanceRuntimePolicyView{
		Strategy:                strategy,
		DispatchMode:            dispatchMode,
		MaxParallelRuns:         maxParallelRuns,
		DoubleReviewEnabled:     doubleReviewEnabled,
		FallbackToDeterministic: fallbackToDeterministic,
		DegradedToDeterministic: fallbackToDeterministic,
		Label:                   governanceRuntimePolicyLabel(strategy, dispatchMode, maxParallelRuns, doubleReviewEnabled),
	}
}

func GovernanceRuntimePolicyUpdate(policy model.GovernanceRuntimePolicyView) (model.GovernanceRuntimePolicyView, error) {
	if !isValidGovernanceRuntimeStrategy(policy.Strategy) {
		return model.GovernanceRuntimePolicyView{}, fmt.Errorf("invalid runtime strategy")
	}
	if !isValidGovernanceDispatchMode(policy.DispatchMode) {
		return model.GovernanceRuntimePolicyView{}, fmt.Errorf("invalid runtime dispatch mode")
	}
	if policy.MaxParallelRuns <= 0 {
		return model.GovernanceRuntimePolicyView{}, fmt.Errorf("invalid runtime max parallel runs")
	}
	if err := SettingSetString(model.SettingKeyAIRuntimeStrategy, policy.Strategy); err != nil {
		return model.GovernanceRuntimePolicyView{}, err
	}
	if err := SettingSetString(model.SettingKeyAIRuntimeDispatchMode, policy.DispatchMode); err != nil {
		return model.GovernanceRuntimePolicyView{}, err
	}
	if err := SettingSetInt(model.SettingKeyAIRuntimeMaxParallelRuns, policy.MaxParallelRuns); err != nil {
		return model.GovernanceRuntimePolicyView{}, err
	}
	if err := SettingSetString(model.SettingKeyAIRuntimeDoubleReviewEnabled, fmt.Sprintf("%t", policy.DoubleReviewEnabled)); err != nil {
		return model.GovernanceRuntimePolicyView{}, err
	}
	if err := SettingSetString(model.SettingKeyAIRuntimeFallbackDeterministic, fmt.Sprintf("%t", policy.FallbackToDeterministic)); err != nil {
		return model.GovernanceRuntimePolicyView{}, err
	}
	return GovernanceRuntimePolicyGet(), nil
}

func isValidGovernanceRuntimeStrategy(value string) bool {
	switch value {
	case "highest_success_rate", "balanced_latency", "cost_first":
		return true
	default:
		return false
	}
}

func isValidGovernanceDispatchMode(value string) bool {
	switch value {
	case "single_best", "bounded_parallel", "round_robin_review":
		return true
	default:
		return false
	}
}

func governanceRuntimePolicyLabel(strategy, dispatchMode string, maxParallelRuns int, doubleReviewEnabled bool) string {
	label := fmt.Sprintf("%s / %s / x%d", strategy, dispatchMode, maxParallelRuns)
	if doubleReviewEnabled {
		label += " / double-review"
	}
	return label
}
