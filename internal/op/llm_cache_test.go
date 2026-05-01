package op

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestLLMRefreshCacheRemovesDeletedModelFromCache(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := LLMCreate(model.LLMInfo{Name: "gpt-refresh-delete"}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Delete(&model.LLMInfo{Name: "gpt-refresh-delete"}).Error; err != nil {
		t.Fatalf("delete llm error = %v", err)
	}

	if err := llmRefreshCache(ctx); err != nil {
		t.Fatalf("llmRefreshCache() error = %v", err)
	}
	if _, err := LLMGet("gpt-refresh-delete"); err == nil {
		t.Fatalf("LLMGet() expected deleted model to be absent after full refresh")
	}
}

func TestLLMCreateNormalizesPolicyFields(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := LLMCreate(model.LLMInfo{
		Name:                  "gpt-policy-normalize",
		BillingMode:           model.BillingMode(" Per_Token "),
		ProbePolicy:           model.ProbePolicy(" Concurrent "),
		ProbeIntervalSeconds:  0,
		ProbeConcurrencyLimit: 0,
	}, ctx); err != nil {
		t.Fatalf("LLMCreate() error = %v", err)
	}

	info, err := LLMGet("gpt-policy-normalize")
	if err != nil {
		t.Fatalf("LLMGet() error = %v", err)
	}
	if info.BillingMode != model.BillingModePerToken {
		t.Fatalf("billing_mode = %q, want %q", info.BillingMode, model.BillingModePerToken)
	}
	if info.ProbePolicy != model.ProbePolicyConcurrent {
		t.Fatalf("probe_policy = %q, want %q", info.ProbePolicy, model.ProbePolicyConcurrent)
	}
	if info.ProbeIntervalSeconds != 3600 {
		t.Fatalf("probe_interval_seconds = %d, want 3600", info.ProbeIntervalSeconds)
	}
	if info.ProbeConcurrencyLimit != 1 {
		t.Fatalf("probe_concurrency_limit = %d, want 1", info.ProbeConcurrencyLimit)
	}
}

func TestLLMCreateRejectsInvalidPolicyFields(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := LLMCreate(model.LLMInfo{
		Name:        "gpt-invalid-policy",
		BillingMode: model.BillingMode("dynamic_price"),
	}, ctx); err == nil {
		t.Fatalf("LLMCreate() error = nil, want invalid billing mode error")
	}
}
