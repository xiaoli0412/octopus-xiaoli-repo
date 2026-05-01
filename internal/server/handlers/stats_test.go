package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/task"
)

func TestGetStatsDynamicRoutingSummaryReturnsCurrentTaskSnapshot(t *testing.T) {
	setupHandlerTest(t)

	channel := &model.Channel{
		Name:    "stats-summary-channel",
		Enabled: true,
		Keys: []model.ChannelKey{
			{Enabled: true, ChannelKey: "free-key", SourceType: "public/free"},
			{Enabled: true, ChannelKey: "paid-key", SourceType: "paid"},
			{Enabled: true, ChannelKey: "private-key", SourceType: "private/internal"},
		},
	}
	if err := op.ChannelCreate(channel, t.Context()); err != nil {
		t.Fatalf("ChannelCreate() error = %v", err)
	}
	group := &model.Group{Name: "stats-summary-group", Mode: model.GroupModeFailover}
	if err := op.GroupCreate(group, t.Context()); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	task.DynamicRoutingSummaryScanTask()

	recorder := performJSONHandlerRequest(t, http.MethodGet, "/api/v1/stats/dynamic-routing-summary", nil, getStatsDynamicRoutingSummary)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	res := decodeHandlerResponse(t, recorder)
	var got task.DynamicRoutingSummaryScanSummary
	if err := json.Unmarshal(res.Data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, body = %s", err, recorder.Body.String())
	}

	if got.LastStatus != "ok" {
		t.Fatalf("LastStatus = %q, want ok", got.LastStatus)
	}
	if got.ChannelCount != 1 || got.EnabledChannels != 1 {
		t.Fatalf("channel counts = %#v, want 1 channel and 1 enabled channel", got)
	}
	if got.GroupCount != 1 || got.FailoverGroups != 1 {
		t.Fatalf("group counts = %#v, want 1 group and 1 failover group", got)
	}
	if got.FreePublicKeys != 1 || got.PaidMeteredKeys != 1 || got.PrivateInnerKeys != 1 || got.UnknownKeys != 0 {
		t.Fatalf("key summary = %#v, want one free, paid, and private key", got)
	}
	if got.Basis != "daily_summary_scan_no_runtime_mutation" {
		t.Fatalf("Basis = %q, want daily summary basis", got.Basis)
	}
	if got.CurrentMode != "hybrid" {
		t.Fatalf("CurrentMode = %q, want hybrid", got.CurrentMode)
	}
	if got.EffectiveMode == "" || got.Decision == "" {
		t.Fatalf("EffectiveMode/Decision should be populated: %#v", got)
	}
	if got.LastRunAt.IsZero() || got.LastSuccessAt.IsZero() {
		t.Fatalf("timestamps should be populated: %#v", got)
	}
}
