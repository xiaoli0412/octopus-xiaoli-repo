package model

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/outbound"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/xurl"
)

type AutoGroupType int

const (
	AutoGroupTypeNone  AutoGroupType = 0 //不自动分组
	AutoGroupTypeFuzzy AutoGroupType = 1 //模糊匹配
	AutoGroupTypeExact AutoGroupType = 2 //准确匹配
	AutoGroupTypeRegex AutoGroupType = 3 //正则匹配
)

type KeyManagementMode string

const (
	KeyManagementModeClassified KeyManagementMode = "classified"
	KeyManagementModePooled     KeyManagementMode = "pooled"
)

type KeyRoutingPolicy string

const (
	KeyRoutingPolicyRoundRobin   KeyRoutingPolicy = "round_robin"
	KeyRoutingPolicyFillPriority KeyRoutingPolicy = "fill_priority"
	KeyRoutingPolicyPriority     KeyRoutingPolicy = "priority_order"
)

type Channel struct {
	ID                int                   `json:"id" gorm:"primaryKey"`
	Name              string                `json:"name" gorm:"unique;not null"`
	Type              outbound.OutboundType `json:"type"`
	Enabled           bool                  `json:"enabled" gorm:"default:true"`
	KeyManagementMode KeyManagementMode     `json:"key_management_mode" gorm:"default:'pooled'"`
	KeyRoutingPolicy  KeyRoutingPolicy      `json:"key_routing_policy" gorm:"default:'round_robin'"`
	BaseUrls          []BaseUrl             `json:"base_urls" gorm:"serializer:json"`
	Keys              []ChannelKey          `json:"keys" gorm:"foreignKey:ChannelID"`
	Model             string                `json:"model"`
	CustomModel       string                `json:"custom_model"`
	Proxy             bool                  `json:"proxy" gorm:"default:false"`
	AutoSync          bool                  `json:"auto_sync" gorm:"default:false"`
	AutoGroup         AutoGroupType         `json:"auto_group" gorm:"default:0"`
	CustomHeader      []CustomHeader        `json:"custom_header" gorm:"serializer:json"`
	ParamOverride     *string               `json:"param_override"`
	ChannelProxy      *string               `json:"channel_proxy"`
	Stats             *StatsChannel         `json:"stats,omitempty" gorm:"foreignKey:ChannelID"`
	MatchRegex        *string               `json:"match_regex"`
}

type BaseUrl struct {
	URL   string `json:"url"`
	Delay int    `json:"delay"`
}

type CustomHeader struct {
	HeaderKey   string `json:"header_key"`
	HeaderValue string `json:"header_value"`
}

type ChannelKey struct {
	ID               int     `json:"id" gorm:"primaryKey"`
	ChannelID        int     `json:"channel_id"`
	Enabled          bool    `json:"enabled" gorm:"default:true"`
	ChannelKey       string  `json:"channel_key"`
	SourceType       string  `json:"source_type"`
	StatusCode       int     `json:"status_code"`
	LastUseTimeStamp int64   `json:"last_use_time_stamp"`
	TotalCost        float64 `json:"total_cost"`
	Remark           string  `json:"remark"`
	// AllowedModels limits what models this key can serve.
	// Empty means "all models" (backward compatible).
	AllowedModels       string `json:"allowed_models"`
	RequestCapabilities string `json:"request_capabilities"`
}

const (
	ChannelKeySourceTypeUnknown         = "unknown"
	ChannelKeySourceTypePublicFree      = "public/free"
	ChannelKeySourceTypePaidMetered     = "paid/metered"
	ChannelKeySourceTypePrivateInternal = "private/internal"
)

const (
	RequestCapabilityOpenAIChat        = "openai_chat"
	RequestCapabilityOpenAIResponses   = "openai_responses"
	RequestCapabilityAnthropicMessages = "anthropic_messages"
	RequestCapabilityOpenAIEmbeddings  = "openai_embeddings"
	RequestCapabilityGeminiContents    = "gemini_contents"
)

func NormalizeChannelKeySourceType(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "" {
		return ChannelKeySourceTypeUnknown
	}
	switch v {
	case "public", "free", "public/free", "free/public":
		return ChannelKeySourceTypePublicFree
	case "paid", "metered", "paid/metered", "metered/paid":
		return ChannelKeySourceTypePaidMetered
	case "private", "internal", "private/internal", "internal/private":
		return ChannelKeySourceTypePrivateInternal
	default:
		return v
	}
}

func EffectiveChannelKeySourceType(input string) string {
	normalized := NormalizeChannelKeySourceType(input)
	switch normalized {
	case ChannelKeySourceTypeUnknown, ChannelKeySourceTypePublicFree, ChannelKeySourceTypePaidMetered, ChannelKeySourceTypePrivateInternal:
		return normalized
	default:
		return ChannelKeySourceTypeUnknown
	}
}

func IsValidChannelKeySourceType(input string) bool {
	switch NormalizeChannelKeySourceType(input) {
	case ChannelKeySourceTypeUnknown, ChannelKeySourceTypePublicFree, ChannelKeySourceTypePaidMetered, ChannelKeySourceTypePrivateInternal:
		return true
	default:
		return false
	}
}

func NormalizeKeyManagementMode(input KeyManagementMode) KeyManagementMode {
	v := KeyManagementMode(strings.ToLower(strings.TrimSpace(string(input))))
	if v == "" {
		return KeyManagementModePooled
	}
	return v
}

func IsValidKeyManagementMode(input KeyManagementMode) bool {
	switch NormalizeKeyManagementMode(input) {
	case KeyManagementModeClassified, KeyManagementModePooled:
		return true
	default:
		return false
	}
}

func NormalizeKeyRoutingPolicy(input KeyRoutingPolicy) KeyRoutingPolicy {
	v := KeyRoutingPolicy(strings.ToLower(strings.TrimSpace(string(input))))
	if v == "" {
		return KeyRoutingPolicyRoundRobin
	}
	return v
}

func IsValidKeyRoutingPolicy(input KeyRoutingPolicy) bool {
	switch NormalizeKeyRoutingPolicy(input) {
	case KeyRoutingPolicyRoundRobin, KeyRoutingPolicyFillPriority, KeyRoutingPolicyPriority:
		return true
	default:
		return false
	}
}

// ChannelUpdateRequest 渠道更新请求 - 仅包含变更的数据
type ChannelUpdateRequest struct {
	ID                int                    `json:"id" binding:"required"`
	Name              *string                `json:"name,omitempty"`
	Type              *outbound.OutboundType `json:"type,omitempty"`
	Enabled           *bool                  `json:"enabled,omitempty"`
	KeyManagementMode *KeyManagementMode     `json:"key_management_mode,omitempty"`
	KeyRoutingPolicy  *KeyRoutingPolicy      `json:"key_routing_policy,omitempty"`
	BaseUrls          *[]BaseUrl             `json:"base_urls,omitempty"`
	Model             *string                `json:"model,omitempty"`
	CustomModel       *string                `json:"custom_model,omitempty"`
	Proxy             *bool                  `json:"proxy,omitempty"`
	AutoSync          *bool                  `json:"auto_sync,omitempty"`
	AutoGroup         *AutoGroupType         `json:"auto_group,omitempty"`
	CustomHeader      *[]CustomHeader        `json:"custom_header,omitempty"`
	ChannelProxy      *string                `json:"channel_proxy,omitempty"`
	ParamOverride     *string                `json:"param_override,omitempty"`
	MatchRegex        *string                `json:"match_regex,omitempty"`

	KeysToAdd    []ChannelKeyAddRequest    `json:"keys_to_add,omitempty"`
	KeysToUpdate []ChannelKeyUpdateRequest `json:"keys_to_update,omitempty"`
	KeysToDelete []int                     `json:"keys_to_delete,omitempty"`
}

type ChannelKeyAddRequest struct {
	Enabled             bool   `json:"enabled"`
	ChannelKey          string `json:"channel_key" binding:"required"`
	SourceType          string `json:"source_type"`
	Remark              string `json:"remark"`
	AllowedModels       string `json:"allowed_models"`
	RequestCapabilities string `json:"request_capabilities"`
}

type ChannelKeyUpdateRequest struct {
	ID                  int     `json:"id" binding:"required"`
	Enabled             *bool   `json:"enabled,omitempty"`
	ChannelKey          *string `json:"channel_key,omitempty"`
	SourceType          *string `json:"source_type,omitempty"`
	Remark              *string `json:"remark,omitempty"`
	AllowedModels       *string `json:"allowed_models,omitempty"`
	RequestCapabilities *string `json:"request_capabilities,omitempty"`
}

// keyRoundRobin maintains per-channel+model rotation counters.
// key: "channelID:modelName" value: *uint64
// In-memory only; restarts reset rotation order (acceptable).
var keyRoundRobin sync.Map
var fillPriorityPrimary sync.Map

func fillPriorityStateKey(channelID int, modelName string) string {
	return fmt.Sprintf("fill:%d:%s", channelID, modelName)
}

func (c *Channel) primaryFillPriorityKeyID(modelName string, ordered []ChannelKey) int {
	if c == nil || len(ordered) == 0 {
		return 0
	}
	stateKey := fillPriorityStateKey(c.ID, modelName)
	if remembered, ok := fillPriorityPrimary.Load(stateKey); ok {
		if keyID, ok := remembered.(int); ok {
			return keyID
		}
	}
	fillPriorityPrimary.Store(stateKey, ordered[0].ID)
	return ordered[0].ID
}

func (c *Channel) rememberPrimaryFillPriorityKey(modelName string, keyID int) {
	if c == nil || keyID == 0 {
		return
	}
	fillPriorityPrimary.Store(fillPriorityStateKey(c.ID, modelName), keyID)
}

func keyAllowsModel(k ChannelKey, modelName string) bool {
	if modelName == "" {
		return true
	}
	allowed := strings.TrimSpace(k.AllowedModels)
	if allowed == "" {
		return true
	}
	start := 0
	for start <= len(allowed) {
		end := strings.IndexByte(allowed[start:], ',')
		if end < 0 {
			end = len(allowed)
		} else {
			end += start
		}
		if strings.TrimSpace(allowed[start:end]) == modelName {
			return true
		}
		if end == len(allowed) {
			break
		}
		start = end + 1
	}
	return false
}

func NormalizeRequestCapability(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, "/", "_")
	v = strings.ReplaceAll(v, " ", "_")
	switch v {
	case "", "unknown":
		return ""
	case "openai_chat", "openai_chat_completions", "chat", "chat_completions":
		return RequestCapabilityOpenAIChat
	case "openai_responses", "openai_response", "responses", "response":
		return RequestCapabilityOpenAIResponses
	case "anthropic_messages", "anthropic_message", "messages", "message":
		return RequestCapabilityAnthropicMessages
	case "openai_embeddings", "openai_embedding", "embeddings", "embedding":
		return RequestCapabilityOpenAIEmbeddings
	case "gemini_contents", "gemini_content", "generate_content", "generate_contents":
		return RequestCapabilityGeminiContents
	default:
		return v
	}
}

func RequestCapabilityForOutbound(channelType outbound.OutboundType, modelName string) string {
	switch channelType {
	case outbound.OutboundTypeOpenAIChat, outbound.OutboundTypeGithubCopilot:
		return RequestCapabilityOpenAIChat
	case outbound.OutboundTypeOpenAIResponse, outbound.OutboundTypeVolcengine:
		return RequestCapabilityOpenAIResponses
	case outbound.OutboundTypeOpenAIEmbedding:
		return RequestCapabilityOpenAIEmbeddings
	case outbound.OutboundTypeAnthropic:
		return RequestCapabilityAnthropicMessages
	case outbound.OutboundTypeGemini, outbound.OutboundTypeAntigravity:
		return RequestCapabilityGeminiContents
	case outbound.OutboundTypeZen:
		lower := strings.ToLower(strings.TrimSpace(modelName))
		switch {
		case strings.HasPrefix(lower, "claude-"):
			return RequestCapabilityAnthropicMessages
		case strings.HasPrefix(lower, "gpt-"):
			return RequestCapabilityOpenAIResponses
		case strings.HasPrefix(lower, "gemini-"):
			return RequestCapabilityGeminiContents
		default:
			return RequestCapabilityOpenAIChat
		}
	default:
		return ""
	}
}

func (c *Channel) RequestCapabilityForModel(modelName string) string {
	if c == nil {
		return ""
	}
	return RequestCapabilityForOutbound(c.Type, modelName)
}

func normalizeRequestCapabilities(input string) string {
	parts := strings.Split(input, ",")
	uniq := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		capability := NormalizeRequestCapability(part)
		if capability == "" {
			continue
		}
		if _, ok := uniq[capability]; ok {
			continue
		}
		uniq[capability] = struct{}{}
		out = append(out, capability)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func ChannelKeyRequestCapabilitiesList(input string) []string {
	normalized := normalizeRequestCapabilities(input)
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, ",")
}

func keyAllowsRequestCapability(k ChannelKey, requestFormat string) bool {
	requestCapability := NormalizeRequestCapability(requestFormat)
	if requestCapability == "" {
		return true
	}
	allowed := ChannelKeyRequestCapabilitiesList(k.RequestCapabilities)
	if len(allowed) == 0 {
		return true
	}
	for _, capability := range allowed {
		if capability == requestCapability {
			return true
		}
	}
	return false
}

func (c *Channel) pooledModelAllowed(modelName string) bool {
	if c == nil {
		return false
	}
	if modelName == "" {
		return true
	}
	if strings.TrimSpace(c.Model) != "" || strings.TrimSpace(c.CustomModel) != "" {
		return c.SupportsModel(modelName)
	}
	hasScopedModels := false
	for _, key := range c.Keys {
		if strings.TrimSpace(key.AllowedModels) == "" {
			continue
		}
		hasScopedModels = true
		if keyAllowsModel(key, modelName) {
			return true
		}
	}
	return !hasScopedModels
}

func normalizeAllowedModels(input string) string {
	parts := strings.Split(input, ",")
	uniq := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		m := strings.TrimSpace(p)
		if m == "" {
			continue
		}
		if _, ok := uniq[m]; ok {
			continue
		}
		uniq[m] = struct{}{}
		out = append(out, m)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func ChannelKeyAllowedModelsList(input string) []string {
	normalized := normalizeAllowedModels(input)
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, ",")
}

// ChannelFetchModelRequest is used by /channel/fetch-model (not persisted).
type ChannelFetchModelRequest struct {
	Type         outbound.OutboundType `json:"type"`
	BaseURL      string                `json:"base_url" binding:"required"`
	Key          string                `json:"key" binding:"required"`
	Proxy        bool                  `json:"proxy"`
	ChannelProxy *string               `json:"channel_proxy,omitempty"`
	MatchRegex   *string               `json:"match_regex,omitempty"`
	CustomHeader []CustomHeader        `json:"custom_header,omitempty"`
}

func NormalizeChannelProxy(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func ValidateChannelProxy(raw *string) error {
	normalized := NormalizeChannelProxy(raw)
	if normalized == nil {
		return nil
	}
	return xurl.ValidateProxyURL(*normalized, "channel_proxy")
}

func (c *Channel) GetBaseUrl() string {
	if c == nil || len(c.BaseUrls) == 0 {
		return ""
	}

	bestURL := ""
	bestDelay := 0
	bestSet := false

	for _, bu := range c.BaseUrls {
		if bu.URL == "" {
			continue
		}
		if !bestSet || bu.Delay < bestDelay {
			bestURL = bu.URL
			bestDelay = bu.Delay
			bestSet = true
		}
	}

	return bestURL
}

func (c *Channel) SupportsModel(modelName string) bool {
	if c == nil {
		return false
	}
	normalizedModel := strings.ToLower(strings.TrimSpace(modelName))
	if normalizedModel == "" {
		return true
	}
	if strings.TrimSpace(c.Model) == "" && strings.TrimSpace(c.CustomModel) == "" {
		// Backward-compatible fallback: some historical channel rows relied on key/model routing
		// without keeping an explicit declared model list on the channel itself.
		return true
	}
	for _, declared := range strings.Split(c.Model+","+c.CustomModel, ",") {
		if strings.ToLower(strings.TrimSpace(declared)) == normalizedModel {
			return true
		}
	}
	return false
}

func (c *Channel) GetChannelKey() ChannelKey {
	if c == nil || len(c.Keys) == 0 {
		return ChannelKey{}
	}

	nowSec := time.Now().Unix()

	best := ChannelKey{}
	bestCost := 0.0
	bestSet := false

	for _, k := range c.Keys {
		if !k.Enabled || k.ChannelKey == "" {
			continue
		}
		if k.StatusCode == 429 && k.LastUseTimeStamp > 0 {
			if nowSec-k.LastUseTimeStamp < int64(5*time.Minute/time.Second) {
				continue
			}
		}
		if !bestSet || k.TotalCost < bestCost {
			best = k
			bestCost = k.TotalCost
			bestSet = true
		}
	}

	if !bestSet {
		return ChannelKey{}
	}
	return best
}

// GetChannelKeyForModel picks an eligible key for the given model.
// classified:
//   - AllowedModels defines which key can serve which model.
//
// pooled:
//   - Keys are treated as a shared pool for the channel's model set.
//
// routing policies:
//   - round_robin: rotate through ordered eligible keys.
//   - fill_priority: in phase-1, prefer the first eligible key and only
//     move forward when the current request excludes it.
//   - priority_order: in phase-1, start every request from the first
//     eligible key and fall through the ordered list only via excluded keys.
//
// 429 cooldown logic is always applied AFTER eligibility filtering.
func (c *Channel) GetChannelKeyForModel(modelName string) ChannelKey {
	return c.GetChannelKeyForModelExcept(modelName, nil)
}

func (c *Channel) GetChannelKeyForRequest(modelName string, requestFormat string) ChannelKey {
	return c.GetChannelKeyForRequestExcept(modelName, requestFormat, nil)
}

func (c *Channel) NextEligibleChannelKeyAfter(modelName string, afterKeyID int, excluded map[int]struct{}) ChannelKey {
	return c.NextEligibleChannelKeyAfterRequest(modelName, "", afterKeyID, excluded)
}

func (c *Channel) NextEligibleChannelKeyAfterRequest(modelName string, requestFormat string, afterKeyID int, excluded map[int]struct{}) ChannelKey {
	if c == nil || len(c.Keys) == 0 {
		return ChannelKey{}
	}
	ordered := c.EligibleChannelKeysForRequest(modelName, requestFormat)
	if len(ordered) == 0 {
		return ChannelKey{}
	}

	startIdx := 0
	if afterKeyID != 0 {
		for idx, key := range ordered {
			if key.ID == afterKeyID {
				startIdx = idx + 1
				break
			}
		}
	}

	for i := 0; i < len(ordered); i++ {
		idx := (startIdx + i) % len(ordered)
		if excluded != nil {
			if _, skip := excluded[ordered[idx].ID]; skip {
				continue
			}
		}
		return ordered[idx]
	}

	return ChannelKey{}
}

func (c *Channel) GetChannelKeyForModelExcept(modelName string, excluded map[int]struct{}) ChannelKey {
	return c.getChannelKeyForRequestExcept(modelName, "", excluded)
}

func (c *Channel) GetChannelKeyForRequestExcept(modelName string, requestFormat string, excluded map[int]struct{}) ChannelKey {
	return c.getChannelKeyForRequestExcept(modelName, requestFormat, excluded)
}

func (c *Channel) getChannelKeyForRequestExcept(modelName string, requestFormat string, excluded map[int]struct{}) ChannelKey {
	if c == nil || len(c.Keys) == 0 {
		return ChannelKey{}
	}
	ordered := c.EligibleChannelKeysForRequest(modelName, requestFormat)
	if len(ordered) == 0 {
		return ChannelKey{}
	}

	switch c.KeyRoutingPolicy {
	case KeyRoutingPolicyFillPriority:
		if len(excluded) == 0 && len(ordered) > 0 {
			c.rememberPrimaryFillPriorityKey(modelName, ordered[0].ID)
		}
		primaryID := c.primaryFillPriorityKeyID(modelName, ordered)
		var firstAvailable ChannelKey
		for _, key := range ordered {
			if excluded != nil {
				if _, skip := excluded[key.ID]; skip {
					continue
				}
			}
			if key.ID == primaryID {
				return key
			}
			if firstAvailable.ID == 0 {
				firstAvailable = key
			}
		}
		if firstAvailable.ID == 0 {
			return ChannelKey{}
		}
		c.rememberPrimaryFillPriorityKey(modelName, firstAvailable.ID)
		return firstAvailable
	case KeyRoutingPolicyPriority:
		for _, key := range ordered {
			if excluded != nil {
				if _, skip := excluded[key.ID]; skip {
					continue
				}
			}
			return key
		}
		// Phase-1 priority_order starts from the first eligible key on each request
		// and falls through the ordered list only via excluded keys from this request.
		return ChannelKey{}
	case KeyRoutingPolicyRoundRobin, "":
		// continue below
	default:
		// Unknown policy falls back to round-robin for compatibility.
	}

	if len(excluded) > 0 {
		return c.NextEligibleChannelKeyAfterRequest(modelName, requestFormat, 0, excluded)
	}

	if len(ordered) == 1 {
		return ordered[0]
	}

	// Per-channel+model RR counter.
	rrKey := fmt.Sprintf("%d:%s", c.ID, modelName)
	ptrAny, _ := keyRoundRobin.LoadOrStore(rrKey, new(uint64))
	ptr := ptrAny.(*uint64)
	startIdx := int(atomic.AddUint64(ptr, 1)-1) % len(ordered)

	for i := 0; i < len(ordered); i++ {
		idx := (startIdx + i) % len(ordered)
		return ordered[idx]
	}

	return ChannelKey{}
}

func (c *Channel) orderEligibleKeys(eligible []ChannelKey) []ChannelKey {
	if len(eligible) == 0 {
		return nil
	}
	ordered := make([]ChannelKey, len(eligible))
	copy(ordered, eligible)
	return ordered
}

func (c *Channel) EligibleChannelKeysForModel(modelName string) []ChannelKey {
	return c.EligibleChannelKeysForRequest(modelName, "")
}

func (c *Channel) EligibleChannelKeysForRequest(modelName string, requestFormat string) []ChannelKey {
	if c == nil || len(c.Keys) == 0 {
		return nil
	}
	nowSec := time.Now().Unix()
	eligible := make([]ChannelKey, 0, len(c.Keys))
	for _, k := range c.Keys {
		if !k.Enabled || k.ChannelKey == "" {
			continue
		}
		if !c.KeyCanServeRequest(k, modelName, requestFormat) {
			continue
		}
		if k.StatusCode == 429 && k.LastUseTimeStamp > 0 {
			if nowSec-k.LastUseTimeStamp < int64(5*time.Minute/time.Second) {
				continue
			}
		}
		eligible = append(eligible, k)
	}
	return eligible
}

func (c *Channel) HasConfiguredKeyForModel(modelName string) bool {
	return c.HasConfiguredKeyForRequest(modelName, "")
}

func (c *Channel) HasConfiguredKeyForRequest(modelName string, requestFormat string) bool {
	if c == nil || len(c.Keys) == 0 {
		return false
	}
	for _, key := range c.Keys {
		if !key.Enabled || strings.TrimSpace(key.ChannelKey) == "" {
			continue
		}
		if c.KeyCanServeRequest(key, modelName, requestFormat) {
			return true
		}
	}
	return false
}

func (c *Channel) KeyCanServeModel(k ChannelKey, modelName string) bool {
	return c.KeyCanServeRequest(k, modelName, "")
}

func (c *Channel) KeyCanServeRequest(k ChannelKey, modelName string, requestFormat string) bool {
	if c == nil {
		return false
	}
	if !keyAllowsRequestCapability(k, requestFormat) {
		return false
	}
	if strings.TrimSpace(k.AllowedModels) != "" {
		return keyAllowsModel(k, modelName)
	}
	if strings.TrimSpace(c.Model) != "" || strings.TrimSpace(c.CustomModel) != "" {
		return c.SupportsModel(modelName)
	}
	return true
}

func NormalizeChannelKeyAllowedModels(input string) string {
	return normalizeAllowedModels(input)
}

func NormalizeChannelKeyRequestCapabilities(input string) string {
	return normalizeRequestCapabilities(input)
}
