package op

import (
	"fmt"
	"strings"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func governancePlanFromSnapshot(goal, presetID string, snapshot governanceSnapshot) model.GovernancePlanView {
	preset, _ := governanceExpertPresetByID(presetID)
	findings := make([]model.GovernanceFindingView, 0)
	decisions := make([]model.GovernanceDecisionView, 0)
	mutations := make([]model.GovernanceMutation, 0)
	domainPlans := make([]model.GovernanceDomainPlanView, 0, 4)
	riskNotes := make([]string, 0)
	routingFindings := make([]model.GovernanceFindingView, 0)
	routingMutations := make([]model.GovernanceMutation, 0)
	pricingFindings := make([]model.GovernanceFindingView, 0)
	pricingMutations := make([]model.GovernanceMutation, 0)
	dynamicFindings := make([]model.GovernanceFindingView, 0)
	dynamicMutations := make([]model.GovernanceMutation, 0)
	runtimeFindings := make([]model.GovernanceFindingView, 0)
	runtimeMutations := make([]model.GovernanceMutation, 0)
	managedGroup := findSnapshotGroupByName(snapshot.Groups, snapshot.ManagedGroupName)
	if managedGroup == nil {
		finding := model.GovernanceFindingView{Severity: "info", Title: "缺少托管治理组", Detail: "当前还没有托管治理组，将自动创建一个统一承载可治理候选。"}
		findings = append(findings, finding)
		routingFindings = append(routingFindings, finding)
		decisions = append(decisions, model.GovernanceDecisionView{Title: "创建托管治理组", Summary: fmt.Sprintf("创建 %s 作为 AI 自动化的统一治理组。", snapshot.ManagedGroupName)})
		mutation := model.GovernanceMutation{Type: model.GovernanceMutationTypeGroupUpsert, Summary: fmt.Sprintf("创建或更新分组 %s", snapshot.ManagedGroupName), GroupUpsert: &model.GovernanceGroupUpsertMutation{GroupName: snapshot.ManagedGroupName, Mode: model.GroupModeFailover}}
		mutations = append(mutations, mutation)
		routingMutations = append(routingMutations, mutation)
	}
	missingModelTargets := make(map[string]governanceSnapshotGroupItem)
	duplicateByModel := make(map[string]int)
	for _, group := range snapshot.Groups {
		for _, item := range group.Items {
			key := strings.ToLower(strings.TrimSpace(item.ModelName))
			duplicateByModel[key]++
			if !item.Valid {
				missingModelTargets[key] = item
			}
		}
	}
	for modelName, item := range missingModelTargets {
		finding := model.GovernanceFindingView{Severity: "warning", Title: fmt.Sprintf("分组项已失效：%s", modelName), Detail: fmt.Sprintf("分组 %s 中的渠道 #%d 已不再声明模型 %s，将移除这条失效映射。", item.GroupName, item.ChannelID, item.ModelName)}
		findings = append(findings, finding)
		routingFindings = append(routingFindings, finding)
		mutation := model.GovernanceMutation{Type: model.GovernanceMutationTypeGroupItemDetach, Summary: fmt.Sprintf("移除渠道 #%d 上的失效模型 %s", item.ChannelID, item.ModelName), GroupItemDetach: &model.GovernanceGroupItemMutation{GroupName: item.GroupName, ChannelID: item.ChannelID, ModelName: item.ModelName}}
		mutations = append(mutations, mutation)
		routingMutations = append(routingMutations, mutation)
	}
	curatedItems := make([]model.GovernanceGroupItemMutation, 0)
	priority := 1
	for _, channel := range snapshot.Channels {
		if !channel.Enabled {
			continue
		}
		for _, modelName := range channel.ConfiguredModels {
			if strings.TrimSpace(modelName) == "" {
				continue
			}
			if duplicateByModel[modelName] > 1 {
				finding := model.GovernanceFindingView{Severity: "warning", Title: fmt.Sprintf("模型分布散乱：%s", modelName), Detail: fmt.Sprintf("模型 %s 当前分散在 %d 个分组项中，将统一并入托管治理组排序。", modelName, duplicateByModel[modelName])}
				findings = append(findings, finding)
				routingFindings = append(routingFindings, finding)
			}
			curatedItems = append(curatedItems, model.GovernanceGroupItemMutation{GroupName: snapshot.ManagedGroupName, ChannelID: channel.ID, ModelName: modelName, Priority: priority, Weight: 1})
			priority++
		}
	}
	if len(curatedItems) == 0 {
		finding := model.GovernanceFindingView{Severity: "critical", Title: "没有可治理的路由候选", Detail: "当前没有启用且声明了模型的渠道，无法生成可应用的路由治理方案。"}
		findings = append(findings, finding)
		routingFindings = append(routingFindings, finding)
		riskNotes = append(riskNotes, "未找到可用的启用渠道与模型候选。")
	}
	if len(curatedItems) > 0 {
		decisions = append(decisions, model.GovernanceDecisionView{Title: "整理托管治理组", Summary: fmt.Sprintf("将启用渠道里的可治理模型统一归并到 %s，并刷新排序。", snapshot.ManagedGroupName)})
		for _, item := range curatedItems {
			mutation := model.GovernanceMutation{Type: model.GovernanceMutationTypeGroupItemAttach, Summary: fmt.Sprintf("挂载模型 %s 到渠道 #%d", item.ModelName, item.ChannelID), GroupItemAttach: &model.GovernanceGroupItemMutation{GroupName: item.GroupName, ChannelID: item.ChannelID, ModelName: item.ModelName, Priority: item.Priority, Weight: item.Weight}}
			mutations = append(mutations, mutation)
			routingMutations = append(routingMutations, mutation)
		}
		reorderMutation := model.GovernanceMutation{Type: model.GovernanceMutationTypeGroupItemReorder, Summary: fmt.Sprintf("刷新分组 %s 的排序", snapshot.ManagedGroupName), GroupItemReorder: &model.GovernanceGroupItemReorderMutation{GroupName: snapshot.ManagedGroupName, Items: curatedItems}}
		mutations = append(mutations, reorderMutation)
		routingMutations = append(routingMutations, reorderMutation)
	}
	for _, override := range snapshot.RouteTargetOverrides {
		if override.Issue == "" {
			continue
		}
		finding := model.GovernanceFindingView{Severity: "warning", Title: fmt.Sprintf("路由目标漂移：%s", override.ModelName), Detail: fmt.Sprintf("渠道 #%d / 密钥 #%d 上的模型 %s override 已脱离当前声明范围，将被移除。", override.ChannelID, override.ChannelKeyID, override.ModelName)}
		findings = append(findings, finding)
		routingFindings = append(routingFindings, finding)
		mutation := model.GovernanceMutation{Type: model.GovernanceMutationTypeRouteTargetOverrideDelete, Summary: fmt.Sprintf("删除渠道 #%d / 密钥 #%d 上的 %s override", override.ChannelID, override.ChannelKeyID, override.ModelName), RouteTargetDelete: &model.GovernanceRouteTargetOverrideMutation{ChannelID: override.ChannelID, ChannelKeyID: override.ChannelKeyID, ModelName: override.ModelName}}
		mutations = append(mutations, mutation)
		routingMutations = append(routingMutations, mutation)
	}
	for _, modelInfo := range snapshot.Models {
		if modelInfo.HasPrice {
			continue
		}
		finding := model.GovernanceFindingView{Severity: "warning", Title: fmt.Sprintf("模型缺少价格：%s", modelInfo.Name), Detail: fmt.Sprintf("模型 %s 已被声明，但当前没有可用价格元数据，AI 自动化会标记为价格缺口。", modelInfo.Name)}
		findings = append(findings, finding)
		pricingFindings = append(pricingFindings, finding)
	}
	if snapshot.DynamicRouting.Mode != "hybrid" {
		finding := model.GovernanceFindingView{Severity: "info", Title: "动态路由未使用默认模式", Detail: fmt.Sprintf("当前动态路由模式为 %s，AI 自动化会继续沿用现值，但会在总控台中纳入统一评估。", snapshot.DynamicRouting.Mode)}
		findings = append(findings, finding)
		dynamicFindings = append(dynamicFindings, finding)
	}
	runtimeMutation := model.GovernanceMutation{Type: model.GovernanceMutationTypeRuntimePolicySet, Summary: fmt.Sprintf("同步 AI 运行时策略为 %s", snapshot.RuntimePolicy.Label), RuntimePolicySet: &snapshot.RuntimePolicy}
	mutations = append(mutations, runtimeMutation)
	runtimeMutations = append(runtimeMutations, runtimeMutation)
	confidence := 0.88
	if len(curatedItems) == 0 {
		confidence = 0.34
	} else if len(missingModelTargets) > 0 || len(riskNotes) > 0 || snapshot.SnapshotSummary.MissingPrices > 0 {
		confidence = 0.71
	}
	operatorSummary := fmt.Sprintf("AI 自动化已评估 %d 个渠道，并生成 %d 条可控变更，覆盖分组路由、价格完整性、动态路由与 AI 运行时策略。", snapshot.SnapshotSummary.Channels, len(mutations))
	if strings.TrimSpace(goal) != "" {
		operatorSummary = fmt.Sprintf("%s 目标：%s", operatorSummary, goal)
	}
	if preset.ID == model.GovernanceExpertPresetDeepReview {
		riskNotes = append(riskNotes, "深度审阅模式会保留更多漂移对象并优先清理。")
	}
	riskSummary := "所有落地都需要显式点击应用。当前仅允许修改分组路由、价格记录、动态路由策略和 AI 运行时策略，不会直接改渠道地址与上游密钥。"
	if len(riskNotes) > 0 {
		riskSummary = fmt.Sprintf("%s %s", riskSummary, strings.Join(riskNotes, " "))
	}
	appendDomain := func(key, title, summary string, domainFindings []model.GovernanceFindingView, domainMutations []model.GovernanceMutation) {
		status := "ready"
		if len(domainMutations) == 0 && len(domainFindings) == 0 {
			status = "idle"
		}
		for _, finding := range domainFindings {
			if finding.Severity == "critical" {
				status = "blocked"
				break
			}
		}
		domainPlans = append(domainPlans, model.GovernanceDomainPlanView{Key: key, Title: title, Summary: summary, Status: status, FindingCount: len(domainFindings), MutationCount: len(domainMutations), Findings: domainFindings, Mutations: domainMutations})
	}
	appendDomain("routing_grouping", "分组与路由", "统一整理托管治理组、分组项与 route target。", routingFindings, routingMutations)
	appendDomain("pricing", "价格覆盖", "识别缺价模型并在可安全补全时生成价格写入。", pricingFindings, pricingMutations)
	appendDomain("dynamic_routing", "动态路由", "纳入模式、学习开关与竞速预算的统一治理。", dynamicFindings, dynamicMutations)
	appendDomain("runtime_policy", "AI 运行时", "统一 AI 调度策略、分发模式与降级路径。", runtimeFindings, runtimeMutations)
	return model.GovernancePlanView{Findings: findings, Decisions: decisions, Mutations: mutations, Domains: domainPlans, RiskSummary: riskSummary, Confidence: confidence, OperatorSummary: operatorSummary}
}

func governancePreviewFromPlan(plan model.GovernancePlanView) model.GovernancePreviewView {
	counts := model.GovernancePreviewImpactCounts{}
	for _, mutation := range plan.Mutations {
		switch mutation.Type {
		case model.GovernanceMutationTypeGroupUpsert:
			counts.Groups++
		case model.GovernanceMutationTypeGroupItemAttach, model.GovernanceMutationTypeGroupItemDetach, model.GovernanceMutationTypeGroupItemReorder:
			counts.Items++
		case model.GovernanceMutationTypeRouteTargetOverrideUpsert, model.GovernanceMutationTypeRouteTargetOverrideDelete:
			counts.Overrides++
		case model.GovernanceMutationTypeLLMPriceUpsert:
			counts.Overrides++
		case model.GovernanceMutationTypeDynamicRoutingSettingSet, model.GovernanceMutationTypeRuntimePolicySet:
			counts.Profiles++
		case model.GovernanceMutationTypeStrategyProfileActivate:
			counts.Profiles++
		}
	}
	summaryLines := []string{
		fmt.Sprintf("%d findings / %d decisions", len(plan.Findings), len(plan.Decisions)),
		fmt.Sprintf("%d 条 typed mutation 已就绪", len(plan.Mutations)),
	}
	riskNotes := []string{"应用必须显式确认，且全程事务化。", "写入前会再次校验 snapshot checksum。"}
	applyBlockers := make([]string, 0)
	canApply := len(plan.Mutations) > 0
	if !canApply {
		applyBlockers = append(applyBlockers, "当前没有生成可应用的 typed mutation。")
	}
	for _, finding := range plan.Findings {
		if finding.Severity == "critical" {
			applyBlockers = append(applyBlockers, finding.Detail)
			canApply = false
		}
	}
	return model.GovernancePreviewView{Headline: plan.OperatorSummary, SummaryLines: summaryLines, ImpactCounts: counts, RiskNotes: riskNotes, ApplyBlockers: applyBlockers, CanApply: canApply, MutationCount: len(plan.Mutations), Mutations: plan.Mutations}
}

func findSnapshotGroupByName(groups []governanceSnapshotGroup, name string) *governanceSnapshotGroup {
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i]
		}
	}
	return nil
}
