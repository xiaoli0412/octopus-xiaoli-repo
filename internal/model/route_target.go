package model

import (
	"sort"
	"strings"
)

const (
	RouteTargetPolicyBasisChannelKeyInheritance = "channel_key_inheritance"
	RouteTargetPolicyBasisModelDefault          = "model_default"
	RouteTargetPolicyBasisExplicitOverride      = "route_target_override"
)

// RouteTargetOverride persists explicit route-target policy overrides at
// (channel, key, model) granularity.
type RouteTargetOverride struct {
	ID                    int         `json:"id" gorm:"primaryKey"`
	ChannelID             int         `json:"channel_id" gorm:"not null;uniqueIndex:idx_route_target_override"`
	ChannelKeyID          int         `json:"channel_key_id" gorm:"not null;uniqueIndex:idx_route_target_override"`
	ModelName             string      `json:"model_name" gorm:"not null;uniqueIndex:idx_route_target_override"`
	BillingMode           BillingMode `json:"billing_mode"`
	ProbePolicy           ProbePolicy `json:"probe_policy"`
	ProbeIntervalSeconds  int         `json:"probe_interval_seconds"`
	ProbeConcurrencyLimit int         `json:"probe_concurrency_limit"`
}

type RouteTargetOverrideUpsertRequest struct {
	ChannelID             int         `json:"channel_id" binding:"required"`
	ChannelKeyID          int         `json:"channel_key_id" binding:"required"`
	ModelName             string      `json:"model_name" binding:"required"`
	BillingMode           BillingMode `json:"billing_mode" binding:"required"`
	ProbePolicy           ProbePolicy `json:"probe_policy" binding:"required"`
	ProbeIntervalSeconds  int         `json:"probe_interval_seconds"`
	ProbeConcurrencyLimit int         `json:"probe_concurrency_limit"`
}

type RouteTargetOverrideDeleteRequest struct {
	ChannelID    int    `json:"channel_id" binding:"required"`
	ChannelKeyID int    `json:"channel_key_id" binding:"required"`
	ModelName    string `json:"model_name" binding:"required"`
}

// RouteTargetResolvedPolicy is the effective policy after applying
// explicit override > model default > channel/key inheritance.
type RouteTargetResolvedPolicy struct {
	ChannelID             int         `json:"channel_id"`
	ChannelKeyID          int         `json:"channel_key_id"`
	ModelName             string      `json:"model_name"`
	SourceType            string      `json:"source_type"`
	SourceTypeBasis       string      `json:"source_type_basis"`
	BillingMode           BillingMode `json:"billing_mode"`
	BillingModeBasis      string      `json:"billing_mode_basis"`
	ProbePolicy           ProbePolicy `json:"probe_policy"`
	ProbePolicyBasis      string      `json:"probe_policy_basis"`
	ProbeIntervalSeconds  int         `json:"probe_interval_seconds"`
	ProbeIntervalBasis    string      `json:"probe_interval_basis"`
	ProbeConcurrencyLimit int         `json:"probe_concurrency_limit"`
	ProbeConcurrencyBasis string      `json:"probe_concurrency_basis"`
}

func NormalizeRouteTargetModelName(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func (p RouteTargetResolvedPolicy) PolicyBasisSummary() string {
	bases := []string{
		strings.TrimSpace(p.SourceTypeBasis),
		strings.TrimSpace(p.BillingModeBasis),
		strings.TrimSpace(p.ProbePolicyBasis),
		strings.TrimSpace(p.ProbeIntervalBasis),
		strings.TrimSpace(p.ProbeConcurrencyBasis),
	}
	seen := make(map[string]struct{}, len(bases))
	out := make([]string, 0, len(bases))
	for _, basis := range bases {
		if basis == "" {
			continue
		}
		if _, ok := seen[basis]; ok {
			continue
		}
		seen[basis] = struct{}{}
		out = append(out, basis)
	}
	if len(out) == 0 {
		return ""
	}
	preferredOrder := map[string]int{
		RouteTargetPolicyBasisExplicitOverride:      0,
		RouteTargetPolicyBasisModelDefault:          1,
		RouteTargetPolicyBasisChannelKeyInheritance: 2,
	}
	sort.SliceStable(out, func(i, j int) bool {
		return preferredOrder[out[i]] < preferredOrder[out[j]]
	})
	return strings.Join(out, "+")
}
