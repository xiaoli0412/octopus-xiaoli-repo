package model

import (
	"sync"
	"testing"
	"time"
)

func resetKeyRoundRobin() {
	keyRoundRobin = sync.Map{}
	fillPriorityPrimary = sync.Map{}
}

func keyIDs(keys []ChannelKey) []int {
	ids := make([]int, 0, len(keys))
	for _, key := range keys {
		ids = append(ids, key.ID)
	}
	return ids
}

func TestNormalizeChannelKeyAllowedModels(t *testing.T) {
	got := NormalizeChannelKeyAllowedModels(" gpt-4o, ,gpt-4o, claude-3-5-sonnet , claude-3-5-sonnet ")
	want := "claude-3-5-sonnet,gpt-4o"
	if got != want {
		t.Fatalf("NormalizeChannelKeyAllowedModels() = %q, want %q", got, want)
	}
}

func TestNormalizeChannelKeySourceType(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty becomes unknown", input: "", want: "unknown"},
		{name: "whitespace becomes unknown", input: "   ", want: "unknown"},
		{name: "trim and lowercase", input: "  Public/Free  ", want: "public/free"},
		{name: "free alias becomes public/free", input: "Free", want: "public/free"},
		{name: "public alias becomes public/free", input: "public", want: "public/free"},
		{name: "paid alias becomes paid/metered", input: "paid", want: "paid/metered"},
		{name: "metered paid alias becomes paid/metered", input: "metered/paid", want: "paid/metered"},
		{name: "private alias becomes private/internal", input: "private", want: "private/internal"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeChannelKeySourceType(tc.input)
			if got != tc.want {
				t.Fatalf("NormalizeChannelKeySourceType(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsValidChannelKeySourceType(t *testing.T) {
	validInputs := []string{"", "unknown", "public", "free", "public/free", "paid", "metered", "paid/metered", "metered/paid", "private", "internal", "private/internal", "internal/private"}
	for _, input := range validInputs {
		if !IsValidChannelKeySourceType(input) {
			t.Fatalf("IsValidChannelKeySourceType(%q) = false, want true", input)
		}
	}
	if IsValidChannelKeySourceType("enterprise") {
		t.Fatalf("IsValidChannelKeySourceType(%q) = true, want false", "enterprise")
	}
}

func TestEffectiveChannelKeySourceTypeFallsBackToUnknown(t *testing.T) {
	if got := EffectiveChannelKeySourceType("enterprise"); got != ChannelKeySourceTypeUnknown {
		t.Fatalf("EffectiveChannelKeySourceType(%q) = %q, want %q", "enterprise", got, ChannelKeySourceTypeUnknown)
	}
}

func TestChannelGetBaseUrlSelectsLowestDelay(t *testing.T) {
	channel := &Channel{
		BaseUrls: []BaseUrl{
			{URL: "https://slow.example.com", Delay: 120},
			{URL: "", Delay: 1},
			{URL: "https://fast.example.com", Delay: 10},
			{URL: "https://mid.example.com", Delay: 30},
		},
	}

	got := channel.GetBaseUrl()
	want := "https://fast.example.com"
	if got != want {
		t.Fatalf("GetBaseUrl() = %q, want %q", got, want)
	}
}

func TestChannelGetChannelKeyPrefersLowestCostAvailableKey(t *testing.T) {
	now := time.Now().Unix()
	channel := &Channel{
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", TotalCost: 9.9},
			{ID: 2, Enabled: true, ChannelKey: "k2", TotalCost: 1.5},
			{ID: 3, Enabled: true, ChannelKey: "k3", TotalCost: 0.5, StatusCode: 429, LastUseTimeStamp: now},
			{ID: 4, Enabled: false, ChannelKey: "k4", TotalCost: 0.1},
		},
	}

	got := channel.GetChannelKey()
	if got.ID != 2 {
		t.Fatalf("GetChannelKey() = %d, want 2", got.ID)
	}
}

func TestEligibleChannelKeysForModelClassifiedFiltersByModelAndCooldown(t *testing.T) {
	now := time.Now().Unix()
	channel := &Channel{
		ID:                1,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"},
			{ID: 2, Enabled: true, ChannelKey: "k2", AllowedModels: "claude-3-5-sonnet"},
			{ID: 3, Enabled: false, ChannelKey: "k3", AllowedModels: "gpt-4o"},
			{ID: 4, Enabled: true, ChannelKey: "k4", AllowedModels: "", StatusCode: 429, LastUseTimeStamp: now},
			{ID: 5, Enabled: true, ChannelKey: "k5", AllowedModels: ""},
			{ID: 6, Enabled: true, ChannelKey: "", AllowedModels: "gpt-4o"},
		},
	}

	keys := channel.EligibleChannelKeysForModel("gpt-4o")
	if len(keys) != 2 {
		t.Fatalf("EligibleChannelKeysForModel() len = %d, want 2", len(keys))
	}
	if keys[0].ID != 1 || keys[1].ID != 5 {
		t.Fatalf("EligibleChannelKeysForModel() IDs = [%d %d], want [1 5]", keys[0].ID, keys[1].ID)
	}
}

func TestEligibleChannelKeysForModelPooledHonorsPerKeyAllowedModels(t *testing.T) {
	channel := &Channel{
		ID:                2,
		Model:             "gpt-4o,claude-3-5-sonnet",
		KeyManagementMode: KeyManagementModePooled,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"},
			{ID: 2, Enabled: true, ChannelKey: "k2", AllowedModels: "claude-3-5-sonnet"},
			{ID: 3, Enabled: true, ChannelKey: "k3", AllowedModels: "gpt-4o,claude-3-5-sonnet"},
			{ID: 4, Enabled: false, ChannelKey: "k4", AllowedModels: "gpt-4o"},
		},
	}

	keys := channel.EligibleChannelKeysForModel("gpt-4o")
	got := keyIDs(keys)
	want := []int{1, 3}
	if len(got) != len(want) {
		t.Fatalf("EligibleChannelKeysForModel() IDs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("EligibleChannelKeysForModel() IDs = %v, want %v", got, want)
		}
	}
}

func TestEligibleChannelKeysForModelPooledUnsupportedModelReturnsEmpty(t *testing.T) {
	channel := &Channel{
		ID:                3,
		Model:             "gpt-4o,claude-3-5-sonnet",
		KeyManagementMode: KeyManagementModePooled,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"},
			{ID: 2, Enabled: true, ChannelKey: "k2", AllowedModels: "claude-3-5-sonnet"},
		},
	}

	keys := channel.EligibleChannelKeysForModel("gemini-2.5-pro")
	if len(keys) != 0 {
		t.Fatalf("EligibleChannelKeysForModel() IDs = %v, want []", keyIDs(keys))
	}
}

func TestEligibleChannelKeysForModelHonorsAllowedModelsInBothModes(t *testing.T) {
	classified := &Channel{
		ID:                31,
		Model:             "gpt-4o,claude-3-5-sonnet",
		KeyManagementMode: KeyManagementModeClassified,
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"},
			{ID: 2, Enabled: true, ChannelKey: "k2", AllowedModels: "claude-3-5-sonnet"},
		},
	}
	pooled := &Channel{
		ID:                32,
		Model:             "gpt-4o,claude-3-5-sonnet",
		KeyManagementMode: KeyManagementModePooled,
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"},
			{ID: 2, Enabled: true, ChannelKey: "k2", AllowedModels: "claude-3-5-sonnet"},
		},
	}

	classifiedIDs := keyIDs(classified.EligibleChannelKeysForModel("gpt-4o"))
	pooledIDs := keyIDs(pooled.EligibleChannelKeysForModel("gpt-4o"))
	if len(classifiedIDs) != 1 || classifiedIDs[0] != 1 {
		t.Fatalf("classified eligible IDs = %v, want [1]", classifiedIDs)
	}
	if len(pooledIDs) != 1 || pooledIDs[0] != 1 {
		t.Fatalf("pooled eligible IDs = %v, want [1]", pooledIDs)
	}
}

func TestSupportsModelFallsBackWhenChannelDeclaresNoModels(t *testing.T) {
	channel := &Channel{
		ID:                33,
		KeyManagementMode: KeyManagementModePooled,
		Keys:              []ChannelKey{{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"}},
	}

	if !channel.SupportsModel("gpt-4o") {
		t.Fatalf("SupportsModel(gpt-4o) = false, want true when channel declares no explicit models")
	}
}

func TestHasConfiguredKeyForModelPooledHonorsAllowedModelsWhenScoped(t *testing.T) {
	channel := &Channel{
		ID:                34,
		KeyManagementMode: KeyManagementModePooled,
		Model:             "gpt-4o,claude-3-5-sonnet",
		Keys: []ChannelKey{
			{ID: 1, Enabled: true, ChannelKey: "k1", AllowedModels: "gpt-4o"},
			{ID: 2, Enabled: true, ChannelKey: "k2", AllowedModels: "claude-3-5-sonnet"},
		},
	}

	if !channel.HasConfiguredKeyForModel("gpt-4o") {
		t.Fatalf("HasConfiguredKeyForModel(gpt-4o) = false, want true")
	}
	if channel.HasConfiguredKeyForModel("gemini-2.5-pro") {
		t.Fatalf("HasConfiguredKeyForModel(gemini-2.5-pro) = true, want false")
	}
}

func TestGetChannelKeyForModelFillPriorityUsesFirstEligibleKey(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                4,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
		},
	}

	first := channel.GetChannelKeyForModel("gpt-4o")
	second := channel.GetChannelKeyForModel("gpt-4o")
	if first.ID != 10 || second.ID != 10 {
		t.Fatalf("fill_priority should keep selecting first eligible key, got [%d %d]", first.ID, second.ID)
	}
}

func TestGetChannelKeyForModelFillPriorityKeepsRecoveredPrimaryAcrossRequests(t *testing.T) {
	resetKeyRoundRobin()
	now := time.Now().Unix()
	channel := &Channel{
		ID:                41,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o", StatusCode: 429, LastUseTimeStamp: now},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
			{ID: 30, Enabled: true, ChannelKey: "k30", AllowedModels: "gpt-4o"},
		},
	}

	degraded := channel.GetChannelKeyForModel("gpt-4o")
	if degraded.ID != 20 {
		t.Fatalf("GetChannelKeyForModel() during primary cooldown = %d, want 20", degraded.ID)
	}

	channel.Keys[0].LastUseTimeStamp = now - int64(6*time.Minute/time.Second)
	recovered := channel.GetChannelKeyForModel("gpt-4o")
	if recovered.ID != 10 {
		t.Fatalf("GetChannelKeyForModel() after primary recovery = %d, want 10", recovered.ID)
	}

	followup := channel.GetChannelKeyForModel("gpt-4o")
	if followup.ID != 10 {
		t.Fatalf("GetChannelKeyForModel() followup = %d, want remembered primary 10", followup.ID)
	}
}

func TestGetChannelKeyForModelPriorityOrderUsesFirstEligibleKeyInPhaseOne(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                5,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyPriority,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
		},
	}

	key := channel.GetChannelKeyForModel("gpt-4o")
	if key.ID != 10 {
		t.Fatalf("priority_order phase-1 should select first eligible key, got %d", key.ID)
	}
}

func TestGetChannelKeyForModelPriorityOrderResetsToFirstKeyOnNewRequest(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                10,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyPriority,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
			{ID: 30, Enabled: true, ChannelKey: "k30", AllowedModels: "gpt-4o"},
		},
	}

	excluded := map[int]struct{}{10: {}}
	fallback := channel.GetChannelKeyForModelExcept("gpt-4o", excluded)
	if fallback.ID != 20 {
		t.Fatalf("GetChannelKeyForModelExcept() = %d, want 20", fallback.ID)
	}

	fresh := channel.GetChannelKeyForModel("gpt-4o")
	if fresh.ID != 10 {
		t.Fatalf("GetChannelKeyForModel() on a new request = %d, want 10", fresh.ID)
	}
}

func TestGetChannelKeyForModelFillPriorityReturnsToPrimaryAfterCooldownRecovery(t *testing.T) {
	resetKeyRoundRobin()
	now := time.Now().Unix()

	channel := &Channel{
		ID:                11,
		KeyManagementMode: KeyManagementModePooled,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{
				ID:               10,
				Enabled:          true,
				ChannelKey:       "k10",
				AllowedModels:    "gpt-4o",
				StatusCode:       429,
				LastUseTimeStamp: now,
			},
			{
				ID:            20,
				Enabled:       true,
				ChannelKey:    "k20",
				AllowedModels: "gpt-4o",
			},
		},
	}

	degraded := channel.GetChannelKeyForModel("gpt-4o")
	if degraded.ID != 20 {
		t.Fatalf("GetChannelKeyForModel() during cooldown = %d, want 20", degraded.ID)
	}

	channel.Keys[0].LastUseTimeStamp = now - int64(6*time.Minute/time.Second)

	recovered := channel.GetChannelKeyForModel("gpt-4o")
	if recovered.ID != 10 {
		t.Fatalf("GetChannelKeyForModel() after cooldown recovery = %d, want 10", recovered.ID)
	}
}

func TestGetChannelKeyForModelRoundRobinRotatesEligibleKeys(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                6,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyRoundRobin,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
		},
	}

	first := channel.GetChannelKeyForModel("gpt-4o")
	second := channel.GetChannelKeyForModel("gpt-4o")
	third := channel.GetChannelKeyForModel("gpt-4o")

	got := []int{first.ID, second.ID, third.ID}
	want := []int{10, 20, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("round_robin rotation = %v, want %v", got, want)
		}
	}
}

func TestGetChannelKeyForModelExceptRoundRobinReturnsNextAvailableKeyInOrder(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                7,
		KeyManagementMode: KeyManagementModePooled,
		KeyRoutingPolicy:  KeyRoutingPolicyRoundRobin,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
			{ID: 30, Enabled: true, ChannelKey: "k30", AllowedModels: "gpt-4o"},
		},
	}

	first := channel.GetChannelKeyForModel("gpt-4o")
	if first.ID != 10 {
		t.Fatalf("GetChannelKeyForModel() = %d, want 10", first.ID)
	}

	excluded := map[int]struct{}{10: {}}
	next := channel.GetChannelKeyForModelExcept("gpt-4o", excluded)
	if next.ID != 20 {
		t.Fatalf("GetChannelKeyForModelExcept() = %d, want 20", next.ID)
	}
}

func TestGetChannelKeyForModelExceptClassifiedSkipsForeignModelKey(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                8,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyRoundRobin,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "claude-3-5-sonnet"},
			{ID: 30, Enabled: true, ChannelKey: "k30", AllowedModels: "gpt-4o"},
		},
	}

	first := channel.GetChannelKeyForModel("gpt-4o")
	if first.ID != 10 {
		t.Fatalf("GetChannelKeyForModel() = %d, want 10", first.ID)
	}

	excluded := map[int]struct{}{10: {}}
	next := channel.GetChannelKeyForModelExcept("gpt-4o", excluded)
	if next.ID != 30 {
		t.Fatalf("GetChannelKeyForModelExcept() = %d, want 30", next.ID)
	}
}

func TestNextEligibleChannelKeyAfterReturnsStrictNextKeyInOrder(t *testing.T) {
	channel := &Channel{
		ID:                12,
		KeyManagementMode: KeyManagementModePooled,
		KeyRoutingPolicy:  KeyRoutingPolicyRoundRobin,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
			{ID: 30, Enabled: true, ChannelKey: "k30", AllowedModels: "gpt-4o"},
		},
	}

	firstExcluded := map[int]struct{}{10: {}}
	next := channel.NextEligibleChannelKeyAfter("gpt-4o", 10, firstExcluded)
	if next.ID != 20 {
		t.Fatalf("NextEligibleChannelKeyAfter(..., afterKeyID=10, excluded={10}) = %d, want 20", next.ID)
	}

	secondExcluded := map[int]struct{}{
		10: {},
		20: {},
	}
	next = channel.NextEligibleChannelKeyAfter("gpt-4o", 10, secondExcluded)
	if next.ID != 30 {
		t.Fatalf("NextEligibleChannelKeyAfter(..., afterKeyID=10, excluded={10,20}) = %d, want 30", next.ID)
	}

	allExcluded := map[int]struct{}{
		10: {},
		20: {},
		30: {},
	}
	next = channel.NextEligibleChannelKeyAfter("gpt-4o", 10, allExcluded)
	if next.ID != 0 {
		t.Fatalf("NextEligibleChannelKeyAfter(..., afterKeyID=10, excluded={10,20,30}) = %d, want 0", next.ID)
	}
	if next.ChannelKey != "" {
		t.Fatalf("NextEligibleChannelKeyAfter(..., afterKeyID=10, excluded={10,20,30}) returned non-empty key %q", next.ChannelKey)
	}
}

func TestGetChannelKeyForModelExceptSkipsExcludedKeys(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:                9,
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyFillPriority,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "k10", AllowedModels: "gpt-4o"},
			{ID: 20, Enabled: true, ChannelKey: "k20", AllowedModels: "gpt-4o"},
		},
	}

	excluded := map[int]struct{}{10: {}}
	key := channel.GetChannelKeyForModelExcept("gpt-4o", excluded)
	if key.ID != 20 {
		t.Fatalf("GetChannelKeyForModelExcept() = %d, want 20", key.ID)
	}
}

func TestNormalizeChannelKeyRequestCapabilities(t *testing.T) {
	got := NormalizeChannelKeyRequestCapabilities(" OpenAI/Chat-Completions, gemini contents,openai_chat, anthropic/messages ")
	want := "anthropic_messages,gemini_contents,openai_chat"
	if got != want {
		t.Fatalf("NormalizeChannelKeyRequestCapabilities() = %q, want %q", got, want)
	}
}

func TestEligibleChannelKeysForRequestHonorsRequestCapabilities(t *testing.T) {
	channel := &Channel{
		ID:                91,
		Model:             "gpt-4o",
		KeyManagementMode: KeyManagementModePooled,
		KeyRoutingPolicy:  KeyRoutingPolicyRoundRobin,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "chat", AllowedModels: "gpt-4o", RequestCapabilities: "openai_chat"},
			{ID: 20, Enabled: true, ChannelKey: "gemini", AllowedModels: "gpt-4o", RequestCapabilities: "gemini_contents"},
			{ID: 30, Enabled: true, ChannelKey: "unrestricted", AllowedModels: "gpt-4o"},
		},
	}

	chatIDs := keyIDs(channel.EligibleChannelKeysForRequest("gpt-4o", "openai/chat_completions"))
	if len(chatIDs) != 2 || chatIDs[0] != 10 || chatIDs[1] != 30 {
		t.Fatalf("chat eligible IDs = %v, want [10 30]", chatIDs)
	}

	geminiIDs := keyIDs(channel.EligibleChannelKeysForRequest("gpt-4o", "gemini/contents"))
	if len(geminiIDs) != 2 || geminiIDs[0] != 20 || geminiIDs[1] != 30 {
		t.Fatalf("gemini eligible IDs = %v, want [20 30]", geminiIDs)
	}
}

func TestGetChannelKeyForRequestExceptKeepsProtocolFilter(t *testing.T) {
	resetKeyRoundRobin()
	channel := &Channel{
		ID:               92,
		Model:            "gpt-4o",
		KeyRoutingPolicy: KeyRoutingPolicyRoundRobin,
		Keys: []ChannelKey{
			{ID: 10, Enabled: true, ChannelKey: "chat-a", AllowedModels: "gpt-4o", RequestCapabilities: "openai_chat"},
			{ID: 20, Enabled: true, ChannelKey: "gemini", AllowedModels: "gpt-4o", RequestCapabilities: "gemini_contents"},
			{ID: 30, Enabled: true, ChannelKey: "chat-b", AllowedModels: "gpt-4o", RequestCapabilities: "openai_chat"},
		},
	}

	next := channel.GetChannelKeyForRequestExcept("gpt-4o", "openai_chat", map[int]struct{}{10: {}})
	if next.ID != 30 {
		t.Fatalf("GetChannelKeyForRequestExcept() = %d, want 30", next.ID)
	}
}
