package op

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	opsRetentionWindow = 12 * time.Hour
	opsBucketSize      = 5 * time.Minute
	opsDetailScanLimit = 5000
)

type OpsRelayEvent struct {
	Time             time.Time
	APIKeyID         int
	ClientIP         string
	RequestModel     string
	ActualModel      string
	Success          bool
	DurationMS       int64
	InputTokens      int64
	OutputTokens     int64
	CacheReadTokens  int64
	CacheWriteTokens int64
	Attempts         []model.ChannelAttempt
}

type opsBucketDelta struct {
	SuccessCount      int64
	FailureCount      int64
	SkippedCount      int64
	WaitTime          int64
	InputToken        int64
	OutputToken       int64
	CacheReadToken    int64
	CacheWriteToken   int64
	CacheHitCount     int64
	CacheWriteCount   int64
	CacheSuccessCount int64
}

func entityKeyFromInt(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func opsBucketStart(at time.Time) int64 {
	return at.Unix() / int64(opsBucketSize/time.Second) * int64(opsBucketSize/time.Second)
}

func opsRetentionCutoff(now time.Time) int64 {
	return now.Add(-opsRetentionWindow).Unix()
}

func opsCanonicalModelKey(requestModel, actualModel string) string {
	candidate := strings.TrimSpace(actualModel)
	if candidate == "" {
		candidate = strings.TrimSpace(requestModel)
	}
	if candidate == "" {
		return ""
	}
	return llmname.CanonicalModelName(candidate)
}

func opsModelLabel(requestModel, actualModel string) string {
	if strings.TrimSpace(actualModel) != "" {
		return strings.TrimSpace(actualModel)
	}
	return strings.TrimSpace(requestModel)
}

func opsRequestDelta(event OpsRelayEvent) opsBucketDelta {
	delta := opsBucketDelta{
		WaitTime:        event.DurationMS,
		InputToken:      event.InputTokens,
		OutputToken:     event.OutputTokens,
		CacheReadToken:  event.CacheReadTokens,
		CacheWriteToken: event.CacheWriteTokens,
	}
	if event.Success {
		delta.SuccessCount = 1
	} else {
		delta.FailureCount = 1
	}
	if event.CacheReadTokens > 0 {
		delta.CacheHitCount = 1
	}
	if event.CacheWriteTokens > 0 {
		delta.CacheWriteCount = 1
	}
	if event.CacheReadTokens > 0 || event.CacheWriteTokens > 0 {
		delta.CacheSuccessCount = 1
	}
	return delta
}

func opsAttemptDelta(attempt model.ChannelAttempt) opsBucketDelta {
	delta := opsBucketDelta{
		WaitTime: int64(attempt.Duration),
	}
	switch attempt.Status {
	case model.AttemptSuccess:
		delta.SuccessCount = 1
	case model.AttemptFailed:
		delta.FailureCount = 1
	default:
		delta.SkippedCount = 1
	}
	return delta
}

func opsApplyCacheToAttemptDelta(delta opsBucketDelta, event OpsRelayEvent) opsBucketDelta {
	delta.InputToken = event.InputTokens
	delta.OutputToken = event.OutputTokens
	delta.CacheReadToken = event.CacheReadTokens
	delta.CacheWriteToken = event.CacheWriteTokens
	if event.CacheReadTokens > 0 {
		delta.CacheHitCount = 1
	}
	if event.CacheWriteTokens > 0 {
		delta.CacheWriteCount = 1
	}
	if event.CacheReadTokens > 0 || event.CacheWriteTokens > 0 {
		delta.CacheSuccessCount = 1
	}
	return delta
}

func opsBucketEntityLabel(scope, entityKey, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	switch scope {
	case model.OpsScopeOverall:
		return "Overall"
	case model.OpsScopeChannel:
		return "Channel " + entityKey
	case model.OpsScopeChannelKey:
		return "Channel Key " + entityKey
	case model.OpsScopeAPIKey:
		return "API Key " + entityKey
	default:
		return entityKey
	}
}

func OpsRecordRelay(ctx context.Context, event OpsRelayEvent) error {
	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	bucketStart := opsBucketStart(event.Time)
	requestDelta := opsRequestDelta(event)
	modelKey := opsCanonicalModelKey(event.RequestModel, event.ActualModel)
	modelLabel := opsModelLabel(event.RequestModel, event.ActualModel)

	return db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsertOpsBucket(tx, model.OpsScopeOverall, model.OpsEntityOverall, "Overall", bucketStart, requestDelta); err != nil {
			return err
		}
		if modelKey != "" {
			if err := upsertOpsBucket(tx, model.OpsScopeModel, modelKey, modelLabel, bucketStart, requestDelta); err != nil {
				return err
			}
		}
		if event.APIKeyID > 0 {
			if err := upsertOpsBucket(tx, model.OpsScopeAPIKey, entityKeyFromInt(event.APIKeyID), opsAPIKeyLabel(event.APIKeyID), bucketStart, requestDelta); err != nil {
				return err
			}
		}
		if strings.TrimSpace(event.ClientIP) != "" {
			if err := upsertOpsBucket(tx, model.OpsScopeIP, strings.TrimSpace(event.ClientIP), strings.TrimSpace(event.ClientIP), bucketStart, requestDelta); err != nil {
				return err
			}
		}

		successChannelID, successChannelKeyID := 0, 0
		for i := len(event.Attempts) - 1; i >= 0; i-- {
			if event.Attempts[i].Status == model.AttemptSuccess {
				successChannelID = event.Attempts[i].ChannelID
				successChannelKeyID = event.Attempts[i].ChannelKeyID
				break
			}
		}
		for _, attempt := range event.Attempts {
			if attempt.ChannelID <= 0 {
				continue
			}
			channelDelta := opsAttemptDelta(attempt)
			if attempt.ChannelID == successChannelID && attempt.Status == model.AttemptSuccess {
				channelDelta = opsApplyCacheToAttemptDelta(channelDelta, event)
			}
			if err := upsertOpsBucket(tx, model.OpsScopeChannel, entityKeyFromInt(attempt.ChannelID), opsBucketEntityLabel(model.OpsScopeChannel, entityKeyFromInt(attempt.ChannelID), attempt.ChannelName), bucketStart, channelDelta); err != nil {
				return err
			}
			if attempt.ChannelKeyID > 0 {
				keyDelta := opsAttemptDelta(attempt)
				if attempt.ChannelKeyID == successChannelKeyID && attempt.Status == model.AttemptSuccess {
					keyDelta = opsApplyCacheToAttemptDelta(keyDelta, event)
				}
				if err := upsertOpsBucket(tx, model.OpsScopeChannelKey, entityKeyFromInt(attempt.ChannelKeyID), opsChannelKeyLabel(attempt.ChannelKeyID, attempt.ChannelName), bucketStart, keyDelta); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func upsertOpsBucket(tx *gorm.DB, scope, entityKey, entityLabel string, bucketStart int64, delta opsBucketDelta) error {
	if strings.TrimSpace(scope) == "" || strings.TrimSpace(entityKey) == "" {
		return nil
	}
	row := model.OpsMetricBucket{
		Scope:             scope,
		EntityKey:         strings.TrimSpace(entityKey),
		EntityLabel:       opsBucketEntityLabel(scope, entityKey, entityLabel),
		BucketStart:       bucketStart,
		SuccessCount:      delta.SuccessCount,
		FailureCount:      delta.FailureCount,
		SkippedCount:      delta.SkippedCount,
		WaitTime:          delta.WaitTime,
		InputToken:        delta.InputToken,
		OutputToken:       delta.OutputToken,
		CacheReadToken:    delta.CacheReadToken,
		CacheWriteToken:   delta.CacheWriteToken,
		CacheHitCount:     delta.CacheHitCount,
		CacheWriteCount:   delta.CacheWriteCount,
		CacheSuccessCount: delta.CacheSuccessCount,
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "scope"}, {Name: "entity_key"}, {Name: "bucket_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"entity_label":        row.EntityLabel,
			"success_count":       gorm.Expr("success_count + ?", row.SuccessCount),
			"failure_count":       gorm.Expr("failure_count + ?", row.FailureCount),
			"skipped_count":       gorm.Expr("skipped_count + ?", row.SkippedCount),
			"wait_time":           gorm.Expr("wait_time + ?", row.WaitTime),
			"input_token":         gorm.Expr("input_token + ?", row.InputToken),
			"output_token":        gorm.Expr("output_token + ?", row.OutputToken),
			"cache_read_token":    gorm.Expr("cache_read_token + ?", row.CacheReadToken),
			"cache_write_token":   gorm.Expr("cache_write_token + ?", row.CacheWriteToken),
			"cache_hit_count":     gorm.Expr("cache_hit_count + ?", row.CacheHitCount),
			"cache_write_count":   gorm.Expr("cache_write_count + ?", row.CacheWriteCount),
			"cache_success_count": gorm.Expr("cache_success_count + ?", row.CacheSuccessCount),
		}),
	}).Create(&row).Error
}

func opsAPIKeyLabel(apiKeyID int) string {
	if apiKeyID <= 0 {
		return ""
	}
	if apiKeyObj, ok := apiKeyCache.Get(apiKeyID); ok && strings.TrimSpace(apiKeyObj.Name) != "" {
		return apiKeyObj.Name
	}
	return "API Key " + entityKeyFromInt(apiKeyID)
}

func opsChannelKeyLabel(channelKeyID int, channelName string) string {
	label := strings.TrimSpace(channelName)
	if cached, ok := channelKeyCache.Get(channelKeyID); ok {
		if strings.TrimSpace(cached.Remark) != "" {
			if label != "" {
				return fmt.Sprintf("%s / %s", label, strings.TrimSpace(cached.Remark))
			}
			return strings.TrimSpace(cached.Remark)
		}
		if label == "" && cached.ChannelID > 0 {
			if channel, ok := channelCache.Get(cached.ChannelID); ok {
				label = strings.TrimSpace(channel.Name)
			}
		}
	}
	if label == "" {
		label = "Channel Key"
	}
	return fmt.Sprintf("%s / Key %d", label, channelKeyID)
}

func opsBucketSince() int64 {
	return opsRetentionCutoff(time.Now())
}

func opsEntitySummaryFromBucket(row model.OpsMetricBucket) model.OpsEntitySummary {
	return opsEntitySummaryFromRaw(row.Scope, row.EntityKey, row.EntityLabel, row.SuccessCount, row.FailureCount, row.SkippedCount, row.WaitTime, row.InputToken, row.OutputToken, row.CacheReadToken, row.CacheWriteToken, row.CacheHitCount, row.CacheWriteCount, row.CacheSuccessCount)
}

func opsEntitySummaryFromRaw(scope, entityKey, entityLabel string, successCount, failureCount, skippedCount, waitTime, inputToken, outputToken, cacheReadToken, cacheWriteToken, cacheHitCount, cacheWriteCount, cacheSuccessCount int64) model.OpsEntitySummary {
	total := successCount + failureCount
	successRate := 0.0
	cacheHitRate := 0.0
	cacheRate := 0.0
	avgLatency := 0.0
	if total > 0 {
		successRate = float64(successCount) / float64(total)
		cacheHitRate = float64(cacheHitCount) / float64(total)
		cacheRate = float64(cacheSuccessCount) / float64(total)
		avgLatency = float64(waitTime) / float64(total)
	}
	return model.OpsEntitySummary{
		Scope:             scope,
		EntityKey:         entityKey,
		EntityLabel:       entityLabel,
		SuccessCount:      successCount,
		FailureCount:      failureCount,
		SkippedCount:      skippedCount,
		WaitTime:          waitTime,
		InputToken:        inputToken,
		OutputToken:       outputToken,
		CacheReadToken:    cacheReadToken,
		CacheWriteToken:   cacheWriteToken,
		CacheHitCount:     cacheHitCount,
		CacheWriteCount:   cacheWriteCount,
		CacheSuccessCount: cacheSuccessCount,
		SuccessRate:       successRate,
		CacheHitRate:      cacheHitRate,
		CacheRate:         cacheRate,
		AvgLatencyMS:      avgLatency,
	}
}

func OpsOverviewGet(ctx context.Context) (model.OpsOverview, error) {
	total, err := opsOverallSummary(ctx)
	if err != nil {
		return model.OpsOverview{}, err
	}
	topModels, err := OpsEntityList(ctx, model.OpsScopeModel, 8)
	if err != nil {
		return model.OpsOverview{}, err
	}
	topChannels, err := OpsEntityList(ctx, model.OpsScopeChannel, 8)
	if err != nil {
		return model.OpsOverview{}, err
	}
	topChannelKeys, err := OpsEntityList(ctx, model.OpsScopeChannelKey, 8)
	if err != nil {
		return model.OpsOverview{}, err
	}
	topAPIKeys, err := OpsEntityList(ctx, model.OpsScopeAPIKey, 8)
	if err != nil {
		return model.OpsOverview{}, err
	}
	topIPs, err := OpsEntityList(ctx, model.OpsScopeIP, 8)
	if err != nil {
		return model.OpsOverview{}, err
	}
	return model.OpsOverview{
		Window:         "12h",
		Total:          total,
		TopModels:      topModels,
		TopChannels:    topChannels,
		TopChannelKeys: topChannelKeys,
		TopAPIKeys:     topAPIKeys,
		TopIPs:         topIPs,
	}, nil
}

func opsOverallSummary(ctx context.Context) (model.OpsEntitySummary, error) {
	var row struct {
		SuccessCount      int64
		FailureCount      int64
		SkippedCount      int64
		WaitTime          int64
		InputToken        int64
		OutputToken       int64
		CacheReadToken    int64
		CacheWriteToken   int64
		CacheHitCount     int64
		CacheWriteCount   int64
		CacheSuccessCount int64
	}
	err := db.GetDB().WithContext(ctx).
		Model(&model.OpsMetricBucket{}).
		Select("COALESCE(SUM(success_count), 0) AS success_count, COALESCE(SUM(failure_count), 0) AS failure_count, COALESCE(SUM(skipped_count), 0) AS skipped_count, COALESCE(SUM(wait_time), 0) AS wait_time, COALESCE(SUM(input_token), 0) AS input_token, COALESCE(SUM(output_token), 0) AS output_token, COALESCE(SUM(cache_read_token), 0) AS cache_read_token, COALESCE(SUM(cache_write_token), 0) AS cache_write_token, COALESCE(SUM(cache_hit_count), 0) AS cache_hit_count, COALESCE(SUM(cache_write_count), 0) AS cache_write_count, COALESCE(SUM(cache_success_count), 0) AS cache_success_count").
		Where("scope = ? AND entity_key = ? AND bucket_start >= ?", model.OpsScopeOverall, model.OpsEntityOverall, opsBucketSince()).
		Scan(&row).Error
	if err != nil {
		return model.OpsEntitySummary{}, err
	}
	return opsEntitySummaryFromRaw(model.OpsScopeOverall, model.OpsEntityOverall, "Overall", row.SuccessCount, row.FailureCount, row.SkippedCount, row.WaitTime, row.InputToken, row.OutputToken, row.CacheReadToken, row.CacheWriteToken, row.CacheHitCount, row.CacheWriteCount, row.CacheSuccessCount), nil
}

func OpsEntityList(ctx context.Context, scope string, limit int) ([]model.OpsEntitySummary, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []struct {
		EntityKey         string
		EntityLabel       string
		SuccessCount      int64
		FailureCount      int64
		SkippedCount      int64
		WaitTime          int64
		InputToken        int64
		OutputToken       int64
		CacheReadToken    int64
		CacheWriteToken   int64
		CacheHitCount     int64
		CacheWriteCount   int64
		CacheSuccessCount int64
	}
	err := db.GetDB().WithContext(ctx).
		Model(&model.OpsMetricBucket{}).
		Select("entity_key, MAX(entity_label) AS entity_label, COALESCE(SUM(success_count), 0) AS success_count, COALESCE(SUM(failure_count), 0) AS failure_count, COALESCE(SUM(skipped_count), 0) AS skipped_count, COALESCE(SUM(wait_time), 0) AS wait_time, COALESCE(SUM(input_token), 0) AS input_token, COALESCE(SUM(output_token), 0) AS output_token, COALESCE(SUM(cache_read_token), 0) AS cache_read_token, COALESCE(SUM(cache_write_token), 0) AS cache_write_token, COALESCE(SUM(cache_hit_count), 0) AS cache_hit_count, COALESCE(SUM(cache_write_count), 0) AS cache_write_count, COALESCE(SUM(cache_success_count), 0) AS cache_success_count").
		Where("scope = ? AND bucket_start >= ?", scope, opsBucketSince()).
		Group("entity_key").
		Order("(COALESCE(SUM(success_count), 0) + COALESCE(SUM(failure_count), 0)) DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]model.OpsEntitySummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, opsEntitySummaryFromRaw(scope, row.EntityKey, row.EntityLabel, row.SuccessCount, row.FailureCount, row.SkippedCount, row.WaitTime, row.InputToken, row.OutputToken, row.CacheReadToken, row.CacheWriteToken, row.CacheHitCount, row.CacheWriteCount, row.CacheSuccessCount))
	}
	return items, nil
}

func OpsEntitySeries(ctx context.Context, scope, entityKey string) ([]model.OpsSeriesPoint, error) {
	entityKey = strings.TrimSpace(entityKey)
	if entityKey == "" {
		entityKey = model.OpsEntityOverall
	}
	start := opsBucketStart(time.Now().Add(-opsRetentionWindow))
	end := opsBucketStart(time.Now())

	var rows []model.OpsMetricBucket
	query := db.GetDB().WithContext(ctx).Where("scope = ? AND entity_key = ? AND bucket_start >= ? AND bucket_start <= ?", scope, entityKey, start, end)
	if err := query.Order("bucket_start asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	byBucket := make(map[int64]model.OpsMetricBucket, len(rows))
	for _, row := range rows {
		byBucket[row.BucketStart] = row
	}
	points := make([]model.OpsSeriesPoint, 0, int(end-start)/int(opsBucketSize/time.Second)+1)
	for bucket := start; bucket <= end; bucket += int64(opsBucketSize / time.Second) {
		row, ok := byBucket[bucket]
		if !ok {
			points = append(points, model.OpsSeriesPoint{
				BucketStart: bucket,
				Label:       time.Unix(bucket, 0).Format("15:04"),
			})
			continue
		}
		summary := opsEntitySummaryFromBucket(row)
		points = append(points, model.OpsSeriesPoint{
			BucketStart:       bucket,
			Label:             time.Unix(bucket, 0).Format("15:04"),
			SuccessCount:      summary.SuccessCount,
			FailureCount:      summary.FailureCount,
			SkippedCount:      summary.SkippedCount,
			WaitTime:          summary.WaitTime,
			InputToken:        summary.InputToken,
			OutputToken:       summary.OutputToken,
			CacheReadToken:    summary.CacheReadToken,
			CacheWriteToken:   summary.CacheWriteToken,
			CacheHitCount:     summary.CacheHitCount,
			CacheWriteCount:   summary.CacheWriteCount,
			CacheSuccessCount: summary.CacheSuccessCount,
			SuccessRate:       summary.SuccessRate,
			CacheHitRate:      summary.CacheHitRate,
			CacheRate:         summary.CacheRate,
			AvgLatencyMS:      summary.AvgLatencyMS,
		})
	}
	return points, nil
}

func OpsRecentDetails(ctx context.Context, scope, entityKey string, limit int) ([]model.OpsRecentDetail, error) {
	if limit <= 0 {
		limit = 20
	}
	start := int(time.Now().Add(-opsRetentionWindow).Unix())
	end := int(time.Now().Unix())
	logs, err := RelayLogExport(ctx, &start, &end, opsDetailScanLimit)
	if err != nil {
		return nil, err
	}
	items := make([]model.OpsRecentDetail, 0, limit)
	for _, item := range logs {
		if !matchOpsDetailScope(item, scope, entityKey) {
			continue
		}
		items = append(items, opsRecentDetailFromRelayLog(item))
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

func matchOpsDetailScope(item model.RelayLog, scope, entityKey string) bool {
	entityKey = strings.TrimSpace(entityKey)
	if scope == "" || scope == model.OpsScopeOverall || entityKey == "" || entityKey == model.OpsEntityOverall {
		return true
	}
	switch scope {
	case model.OpsScopeModel:
		return opsCanonicalModelKey(item.RequestModelName, item.ActualModelName) == entityKey
	case model.OpsScopeChannel:
		if entityKeyFromInt(item.ChannelId) == entityKey {
			return true
		}
		for _, attempt := range item.Attempts {
			if entityKeyFromInt(attempt.ChannelID) == entityKey {
				return true
			}
		}
	case model.OpsScopeChannelKey:
		for _, attempt := range item.Attempts {
			if entityKeyFromInt(attempt.ChannelKeyID) == entityKey {
				return true
			}
		}
	case model.OpsScopeAPIKey:
		return entityKeyFromInt(item.APIKeyID) == entityKey
	case model.OpsScopeIP:
		return strings.TrimSpace(item.ClientIP) == entityKey
	}
	return false
}

func opsRecentDetailFromRelayLog(item model.RelayLog) model.OpsRecentDetail {
	success := false
	channelKeyID := 0
	statusCode := 0
	for i := len(item.Attempts) - 1; i >= 0; i-- {
		attempt := item.Attempts[i]
		if channelKeyID == 0 && attempt.ChannelKeyID > 0 {
			channelKeyID = attempt.ChannelKeyID
		}
		if statusCode == 0 && attempt.StatusCode > 0 {
			statusCode = attempt.StatusCode
		}
		if attempt.Status == model.AttemptSuccess {
			success = true
			channelKeyID = attempt.ChannelKeyID
			statusCode = attempt.StatusCode
			break
		}
	}
	if !success && item.Error == "" && item.ChannelId != 0 {
		success = true
	}
	return model.OpsRecentDetail{
		ID:               item.ID,
		Time:             item.Time,
		ClientIP:         item.ClientIP,
		RequestModelName: item.RequestModelName,
		ActualModelName:  item.ActualModelName,
		APIKeyID:         item.APIKeyID,
		ChannelID:        item.ChannelId,
		ChannelName:      item.ChannelName,
		ChannelKeyID:     channelKeyID,
		InputTokens:      item.InputTokens,
		OutputTokens:     item.OutputTokens,
		CacheReadTokens:  item.CacheReadTokens,
		CacheWriteTokens: item.CacheWriteTokens,
		UseTime:          item.UseTime,
		Success:          success,
		StatusCode:       statusCode,
		Error:            item.Error,
		AttemptCount:     item.TotalAttempts,
	}
}

func OpsCleanup(ctx context.Context) error {
	return db.GetDB().WithContext(ctx).
		Where("bucket_start < ?", opsRetentionCutoff(time.Now())).
		Delete(&model.OpsMetricBucket{}).Error
}
