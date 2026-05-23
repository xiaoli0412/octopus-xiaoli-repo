package op

import (
	"context"
	"testing"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestOpsRecordRelayAndQueryOverview(t *testing.T) {
	ctx := setupOpTestDB(t)

	channel := createConfiguredTestChannel(t, ctx, "ops-channel", "gpt-4o", "")
	channelKey := channel.Keys[0]

	apiKey := &model.APIKey{Name: "ops-client", APIKey: "sk-octopus-client", Enabled: true}
	if err := APIKeyCreate(apiKey, ctx); err != nil {
		t.Fatalf("APIKeyCreate() error = %v", err)
	}

	at := time.Now().Add(-30 * time.Minute)
	if err := OpsRecordRelay(ctx, OpsRelayEvent{
		Time:            at,
		APIKeyID:        apiKey.ID,
		ClientIP:        "203.0.113.10",
		RequestModel:    "gpt-4o",
		ActualModel:     "gpt-4o",
		Success:         true,
		DurationMS:      120,
		InputTokens:     100,
		OutputTokens:    40,
		CacheReadTokens: 25,
		Attempts: []model.ChannelAttempt{
			{
				ChannelID:    channel.ID,
				ChannelKeyID: channelKey.ID,
				ChannelName:  channel.Name,
				ModelName:    "gpt-4o",
				Status:       model.AttemptSuccess,
				Duration:     120,
			},
		},
	}); err != nil {
		t.Fatalf("OpsRecordRelay(success) error = %v", err)
	}

	if err := OpsRecordRelay(ctx, OpsRelayEvent{
		Time:         at.Add(5 * time.Minute),
		APIKeyID:     apiKey.ID,
		ClientIP:     "203.0.113.10",
		RequestModel: "gpt-4o",
		ActualModel:  "gpt-4o",
		Success:      false,
		DurationMS:   240,
		Attempts: []model.ChannelAttempt{
			{
				ChannelID:    channel.ID,
				ChannelKeyID: channelKey.ID,
				ChannelName:  channel.Name,
				ModelName:    "gpt-4o",
				Status:       model.AttemptFailed,
				Duration:     240,
			},
		},
	}); err != nil {
		t.Fatalf("OpsRecordRelay(failed) error = %v", err)
	}

	overview, err := OpsOverviewGet(ctx)
	if err != nil {
		t.Fatalf("OpsOverviewGet() error = %v", err)
	}
	if overview.Total.SuccessCount != 1 || overview.Total.FailureCount != 1 {
		t.Fatalf("overview total = %#v, want success=1 failed=1", overview.Total)
	}
	if overview.Total.CacheHitCount != 1 {
		t.Fatalf("overview total cache_hit_count = %d, want 1", overview.Total.CacheHitCount)
	}
	if len(overview.TopModels) == 0 || overview.TopModels[0].EntityKey != "gpt-4o" {
		t.Fatalf("overview top models = %#v, want gpt-4o first", overview.TopModels)
	}
	if len(overview.TopIPs) == 0 || overview.TopIPs[0].EntityKey != "203.0.113.10" {
		t.Fatalf("overview top ips = %#v, want 203.0.113.10 first", overview.TopIPs)
	}

	channelKeys, err := OpsEntityList(ctx, model.OpsScopeChannelKey, 10)
	if err != nil {
		t.Fatalf("OpsEntityList(channel_key) error = %v", err)
	}
	if len(channelKeys) != 1 || channelKeys[0].SuccessCount != 1 || channelKeys[0].FailureCount != 1 {
		t.Fatalf("channel key summaries = %#v, want one summary with success=1 failed=1", channelKeys)
	}

	series, err := OpsEntitySeries(ctx, model.OpsScopeAPIKey, entityKeyFromInt(apiKey.ID))
	if err != nil {
		t.Fatalf("OpsEntitySeries(api_key) error = %v", err)
	}
	if len(series) == 0 {
		t.Fatal("OpsEntitySeries(api_key) returned no points")
	}
	var successTotal, failureTotal int64
	for _, point := range series {
		successTotal += point.SuccessCount
		failureTotal += point.FailureCount
	}
	if successTotal != 1 || failureTotal != 1 {
		t.Fatalf("OpsEntitySeries(api_key) = %#v, want totals success=1 failed=1", series)
	}
}

func TestOpsRecentDetailsFiltersByIPAndChannelKey(t *testing.T) {
	ctx := setupOpTestDB(t)

	if err := db.GetDB().WithContext(ctx).Create(&model.RelayLog{
		ID:               101,
		Time:             time.Now().Add(-15 * time.Minute).Unix(),
		ClientIP:         "198.51.100.10",
		RequestModelName: "gpt-4o",
		ActualModelName:  "gpt-4o",
		ChannelId:        1,
		ChannelName:      "primary",
		UseTime:          90,
		InputTokens:      20,
		OutputTokens:     10,
		Attempts: []model.ChannelAttempt{
			{ChannelID: 1, ChannelKeyID: 11, ChannelName: "primary", ModelName: "gpt-4o", Status: model.AttemptSuccess, Duration: 90},
		},
	}).Error; err != nil {
		t.Fatalf("create relay log #101 error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&model.RelayLog{
		ID:               102,
		Time:             time.Now().Add(-10 * time.Minute).Unix(),
		ClientIP:         "198.51.100.11",
		RequestModelName: "claude-3-7-sonnet",
		ActualModelName:  "claude-3-7-sonnet",
		ChannelId:        2,
		ChannelName:      "secondary",
		UseTime:          140,
		Error:            "upstream failed",
		Attempts: []model.ChannelAttempt{
			{ChannelID: 2, ChannelKeyID: 22, ChannelName: "secondary", ModelName: "claude-3-7-sonnet", Status: model.AttemptFailed, Duration: 140},
		},
	}).Error; err != nil {
		t.Fatalf("create relay log #102 error = %v", err)
	}

	byIP, err := OpsRecentDetails(ctx, model.OpsScopeIP, "198.51.100.10", 10)
	if err != nil {
		t.Fatalf("OpsRecentDetails(ip) error = %v", err)
	}
	if len(byIP) != 1 || byIP[0].ClientIP != "198.51.100.10" {
		t.Fatalf("details by ip = %#v, want one entry for 198.51.100.10", byIP)
	}

	byChannelKey, err := OpsRecentDetails(ctx, model.OpsScopeChannelKey, "22", 10)
	if err != nil {
		t.Fatalf("OpsRecentDetails(channel_key) error = %v", err)
	}
	if len(byChannelKey) != 1 || byChannelKey[0].ChannelKeyID != 22 || byChannelKey[0].Success {
		t.Fatalf("details by channel key = %#v, want failed entry for key 22", byChannelKey)
	}
}

func TestOpsCleanupDeletesExpiredBuckets(t *testing.T) {
	ctx := setupOpTestDB(t)

	expired := model.OpsMetricBucket{
		Scope:        model.OpsScopeOverall,
		EntityKey:    model.OpsEntityOverall,
		EntityLabel:  "Overall",
		BucketStart:  time.Now().Add(-13 * time.Hour).Unix(),
		SuccessCount: 1,
	}
	fresh := model.OpsMetricBucket{
		Scope:        model.OpsScopeOverall,
		EntityKey:    model.OpsEntityOverall,
		EntityLabel:  "Overall",
		BucketStart:  time.Now().Add(-1 * time.Hour).Unix(),
		SuccessCount: 2,
	}
	if err := db.GetDB().WithContext(ctx).Create(&expired).Error; err != nil {
		t.Fatalf("create expired bucket error = %v", err)
	}
	if err := db.GetDB().WithContext(ctx).Create(&fresh).Error; err != nil {
		t.Fatalf("create fresh bucket error = %v", err)
	}

	if err := OpsCleanup(ctx); err != nil {
		t.Fatalf("OpsCleanup() error = %v", err)
	}

	var buckets []model.OpsMetricBucket
	if err := db.GetDB().WithContext(context.Background()).Find(&buckets).Error; err != nil {
		t.Fatalf("query buckets error = %v", err)
	}
	if len(buckets) != 1 || buckets[0].BucketStart != fresh.BucketStart {
		t.Fatalf("remaining buckets = %#v, want only fresh bucket", buckets)
	}
}
