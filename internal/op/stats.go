package op

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/cache"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var statsDailyCache model.StatsDaily
var statsDailyCacheLock sync.RWMutex

var statsTotalCache model.StatsTotal
var statsTotalCacheLock sync.RWMutex

var statsHourlyCache [24]model.StatsHourly
var statsHourlyCacheLock sync.RWMutex

var statsChannelCache = cache.New[int, model.StatsChannel](16)
var statsChannelCacheNeedUpdate = make(map[int]struct{})
var statsChannelCacheNeedUpdateLock sync.Mutex

var statsModelCache = cache.New[int, model.StatsModel](16)
var statsModelCacheNeedUpdate = make(map[int]struct{})
var statsModelCacheNeedUpdateLock sync.Mutex

var statsAPIKeyCache = cache.New[int, model.StatsAPIKey](16)
var statsAPIKeyCacheNeedUpdate = make(map[int]struct{})
var statsAPIKeyCacheNeedUpdateLock sync.Mutex

func StatsSaveDBTask() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	log.Debugf("stats save db task started")
	startTime := time.Now()
	defer func() {
		log.Debugf("stats save db task finished, save time: %s", time.Since(startTime))
	}()
	if err := StatsSaveDB(ctx); err != nil {
		log.Errorf("stats save db error: %v", err)
		return
	}
}

func StatsSaveDB(ctx context.Context) error {
	statsTotalCacheLock.RLock()
	totalSnap := statsTotalCache
	statsTotalCacheLock.RUnlock()
	if totalSnap.ID == 0 {
		totalSnap.ID = 1
	}

	statsDailyCacheLock.RLock()
	dailySnap := statsDailyCache
	statsDailyCacheLock.RUnlock()

	statsHourlyCacheLock.RLock()
	hourlyAll := statsHourlyCache
	statsHourlyCacheLock.RUnlock()

	statsChannelCacheNeedUpdateLock.Lock()
	channelIDs := make([]int, 0, len(statsChannelCacheNeedUpdate))
	for id := range statsChannelCacheNeedUpdate {
		channelIDs = append(channelIDs, id)
	}
	statsChannelCacheNeedUpdate = make(map[int]struct{})
	statsChannelCacheNeedUpdateLock.Unlock()

	statsModelCacheNeedUpdateLock.Lock()
	modelIDs := make([]int, 0, len(statsModelCacheNeedUpdate))
	for id := range statsModelCacheNeedUpdate {
		modelIDs = append(modelIDs, id)
	}
	statsModelCacheNeedUpdate = make(map[int]struct{})
	statsModelCacheNeedUpdateLock.Unlock()

	statsAPIKeyCacheNeedUpdateLock.Lock()
	apiKeyIDs := make([]int, 0, len(statsAPIKeyCacheNeedUpdate))
	for id := range statsAPIKeyCacheNeedUpdate {
		apiKeyIDs = append(apiKeyIDs, id)
	}
	statsAPIKeyCacheNeedUpdate = make(map[int]struct{})
	statsAPIKeyCacheNeedUpdateLock.Unlock()

	return persistStatsSnapshots(ctx, totalSnap, dailySnap, hourlyAll, channelIDs, modelIDs, apiKeyIDs)
}

func persistStatsSnapshots(
	ctx context.Context,
	totalSnap model.StatsTotal,
	dailySnap model.StatsDaily,
	hourlyAll [24]model.StatsHourly,
	channelIDs []int,
	modelIDs []int,
	apiKeyIDs []int,
) error {
	dbConn := db.GetDB().WithContext(ctx)

	if result := dbConn.Save(&totalSnap); result.Error != nil {
		return result.Error
	}
	if result := dbConn.Save(&dailySnap); result.Error != nil {
		return result.Error
	}

	todayDate := time.Now().Format("20060102")
	hourlyStats := make([]model.StatsHourly, 0, 24)
	for hour := 0; hour < 24; hour++ {
		if hourlyAll[hour].Date == todayDate {
			hourlyStats = append(hourlyStats, hourlyAll[hour])
		}
	}
	if len(hourlyStats) > 0 {
		if result := dbConn.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "hour"}},
			UpdateAll: true,
		}).Create(&hourlyStats); result.Error != nil {
			return result.Error
		}
	}

	for _, id := range channelIDs {
		ch, ok := statsChannelCache.Get(id)
		if !ok {
			continue
		}
		if result := dbConn.Save(&ch); result.Error != nil {
			return result.Error
		}
	}

	for _, id := range modelIDs {
		m, ok := statsModelCache.Get(id)
		if !ok {
			continue
		}
		if result := dbConn.Save(&m); result.Error != nil {
			return result.Error
		}
	}

	for _, id := range apiKeyIDs {
		ak, ok := statsAPIKeyCache.Get(id)
		if !ok {
			continue
		}
		if result := dbConn.Save(&ak); result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func statsSaveDBWithDailyOverride(ctx context.Context, dailyOverride model.StatsDaily) error {
	statsTotalCacheLock.RLock()
	totalSnap := statsTotalCache
	statsTotalCacheLock.RUnlock()
	if totalSnap.ID == 0 {
		totalSnap.ID = 1
	}

	statsHourlyCacheLock.RLock()
	hourlyAll := statsHourlyCache
	statsHourlyCacheLock.RUnlock()

	statsChannelCacheNeedUpdateLock.Lock()
	channelIDs := make([]int, 0, len(statsChannelCacheNeedUpdate))
	for id := range statsChannelCacheNeedUpdate {
		channelIDs = append(channelIDs, id)
	}
	statsChannelCacheNeedUpdate = make(map[int]struct{})
	statsChannelCacheNeedUpdateLock.Unlock()

	statsModelCacheNeedUpdateLock.Lock()
	modelIDs := make([]int, 0, len(statsModelCacheNeedUpdate))
	for id := range statsModelCacheNeedUpdate {
		modelIDs = append(modelIDs, id)
	}
	statsModelCacheNeedUpdate = make(map[int]struct{})
	statsModelCacheNeedUpdateLock.Unlock()

	statsAPIKeyCacheNeedUpdateLock.Lock()
	apiKeyIDs := make([]int, 0, len(statsAPIKeyCacheNeedUpdate))
	for id := range statsAPIKeyCacheNeedUpdate {
		apiKeyIDs = append(apiKeyIDs, id)
	}
	statsAPIKeyCacheNeedUpdate = make(map[int]struct{})
	statsAPIKeyCacheNeedUpdateLock.Unlock()

	return persistStatsSnapshots(ctx, totalSnap, dailyOverride, hourlyAll, channelIDs, modelIDs, apiKeyIDs)
}

func StatsDailyUpdate(ctx context.Context, metrics model.StatsMetrics) error {
	today := time.Now().Format("20060102")

	statsDailyCacheLock.Lock()
	if statsDailyCache.Date == today {
		statsDailyCache.StatsMetrics.Add(metrics)
		statsDailyCacheLock.Unlock()
		return nil
	}

	prevDaily := statsDailyCache
	statsDailyCache = model.StatsDaily{Date: today}
	statsDailyCache.StatsMetrics.Add(metrics)
	statsDailyCacheLock.Unlock()

	return statsSaveDBWithDailyOverride(ctx, prevDaily)
}

func StatsTotalUpdate(metrics model.StatsMetrics) error {
	statsTotalCacheLock.Lock()
	defer statsTotalCacheLock.Unlock()
	if statsTotalCache.ID == 0 {
		statsTotalCache.ID = 1
	}
	statsTotalCache.StatsMetrics.Add(metrics)
	return nil
}

func StatsChannelUpdate(channelID int, metrics model.StatsMetrics) error {
	channelCache, ok := statsChannelCache.Get(channelID)
	if !ok {
		channelCache = model.StatsChannel{
			ChannelID: channelID,
		}
	}
	channelCache.StatsMetrics.Add(metrics)
	statsChannelCache.Set(channelID, channelCache)
	statsChannelCacheNeedUpdateLock.Lock()
	statsChannelCacheNeedUpdate[channelID] = struct{}{}
	statsChannelCacheNeedUpdateLock.Unlock()
	return nil
}

func StatsHourlyUpdate(metrics model.StatsMetrics) error {
	now := time.Now()
	nowHour := now.Hour()
	todayDate := time.Now().Format("20060102")

	statsHourlyCacheLock.Lock()
	defer statsHourlyCacheLock.Unlock()

	if statsHourlyCache[nowHour].Date != todayDate {
		statsHourlyCache[nowHour] = model.StatsHourly{
			Hour: nowHour,
			Date: todayDate,
		}
	}

	statsHourlyCache[nowHour].StatsMetrics.Add(metrics)
	return nil
}

func StatsModelUpdate(stats model.StatsModel) error {
	modelCache, ok := statsModelCache.Get(stats.ID)
	if !ok {
		modelCache = model.StatsModel{
			ID: stats.ID,
		}
	}
	modelCache.StatsMetrics.Add(stats.StatsMetrics)
	statsModelCache.Set(stats.ID, modelCache)
	statsModelCacheNeedUpdateLock.Lock()
	statsModelCacheNeedUpdate[stats.ID] = struct{}{}
	statsModelCacheNeedUpdateLock.Unlock()
	return nil
}

func StatsAPIKeyUpdate(apiKeyID int, metrics model.StatsMetrics) error {
	apiKeyCache, ok := statsAPIKeyCache.Get(apiKeyID)
	if !ok {
		apiKeyCache = model.StatsAPIKey{
			APIKeyID: apiKeyID,
		}
	}
	apiKeyCache.StatsMetrics.Add(metrics)
	statsAPIKeyCache.Set(apiKeyID, apiKeyCache)
	statsAPIKeyCacheNeedUpdateLock.Lock()
	statsAPIKeyCacheNeedUpdate[apiKeyID] = struct{}{}
	statsAPIKeyCacheNeedUpdateLock.Unlock()
	return nil
}

func StatsChannelDel(id int) error {
	if _, ok := statsChannelCache.Get(id); !ok {
		return nil
	}
	statsChannelCache.Del(id)
	statsChannelCacheNeedUpdateLock.Lock()
	delete(statsChannelCacheNeedUpdate, id)
	statsChannelCacheNeedUpdateLock.Unlock()
	return db.GetDB().Delete(&model.StatsChannel{}, id).Error
}

func StatsAPIKeyDel(id int) error {
	if _, ok := statsAPIKeyCache.Get(id); !ok {
		return nil
	}
	statsAPIKeyCache.Del(id)
	statsAPIKeyCacheNeedUpdateLock.Lock()
	delete(statsAPIKeyCacheNeedUpdate, id)
	statsAPIKeyCacheNeedUpdateLock.Unlock()
	return db.GetDB().Delete(&model.StatsAPIKey{}, id).Error
}

func StatsTotalGet() model.StatsTotal {
	statsTotalCacheLock.RLock()
	defer statsTotalCacheLock.RUnlock()
	return statsTotalCache
}

func StatsTodayGet() model.StatsDaily {
	statsDailyCacheLock.RLock()
	defer statsDailyCacheLock.RUnlock()
	return statsDailyCache
}

func StatsChannelGet(id int) model.StatsChannel {
	stats, ok := statsChannelCache.Get(id)
	if !ok {
		tmp := model.StatsChannel{
			ChannelID: id,
		}
		statsChannelCache.Set(id, tmp)
		statsChannelCacheNeedUpdateLock.Lock()
		statsChannelCacheNeedUpdate[id] = struct{}{}
		statsChannelCacheNeedUpdateLock.Unlock()
		return tmp
	}
	return stats
}

func StatsChannelSnapshot(id int) (model.StatsChannel, bool) {
	stats, ok := statsChannelCache.Get(id)
	return stats, ok
}

func StatsAPIKeyGet(id int) model.StatsAPIKey {
	stats, ok := statsAPIKeyCache.Get(id)
	if !ok {
		tmp := model.StatsAPIKey{
			APIKeyID: id,
		}
		statsAPIKeyCache.Set(id, tmp)
		statsAPIKeyCacheNeedUpdateLock.Lock()
		statsAPIKeyCacheNeedUpdate[id] = struct{}{}
		statsAPIKeyCacheNeedUpdateLock.Unlock()
		return tmp
	}
	return stats
}

func StatsAPIKeyList() []model.StatsAPIKey {
	return statsAPIKeyCache.Values()
}

// StatsChannelList 返回所有渠道统计的快照列表。
func StatsChannelList() []model.StatsChannel {
	return statsChannelCache.Values()
}

// StatsModelList 返回所有模型统计的快照列表。
func StatsModelList() []model.StatsModel {
	return statsModelCache.Values()
}

type StatsTokenBreakdownItem struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	InputToken  int64  `json:"input_token"`
	OutputToken int64  `json:"output_token"`
	TotalToken  int64  `json:"total_token"`
}

type StatsTokenBreakdown struct {
	TotalInputToken                int64                     `json:"total_input_token"`
	TotalOutputToken               int64                     `json:"total_output_token"`
	TotalToken                     int64                     `json:"total_token"`
	EstimatedOfficialInputCost     float64                   `json:"estimated_official_input_cost"`
	EstimatedOfficialOutputCost    float64                   `json:"estimated_official_output_cost"`
	EstimatedOfficialTotalCost     float64                   `json:"estimated_official_total_cost"`
	EstimatedGatewayInputCost      float64                   `json:"estimated_gateway_input_cost"`
	EstimatedGatewayOutputCost     float64                   `json:"estimated_gateway_output_cost"`
	EstimatedGatewayTotalCost      float64                   `json:"estimated_gateway_total_cost"`
	EstimatedPriceBasis            string                    `json:"estimated_price_basis"`
	EstimatedProbeInputCost        float64                   `json:"estimated_probe_input_cost"`
	EstimatedProbeOutputCost       float64                   `json:"estimated_probe_output_cost"`
	EstimatedProbeTotalCost        float64                   `json:"estimated_probe_total_cost"`
	RecentProbeCount               int64                     `json:"recent_probe_count"`
	RecentProbeSuccessCount        int64                     `json:"recent_probe_success_count"`
	RecentProbeFailedCount         int64                     `json:"recent_probe_failed_count"`
	RecentProbeLastAt              int64                     `json:"recent_probe_last_at"`
	RecentProbeLastStatus          string                    `json:"recent_probe_last_status"`
	RecentProbeLastChannel         string                    `json:"recent_probe_last_channel"`
	RecentProbeLastModel           string                    `json:"recent_probe_last_model"`
	RecentProbeLastMessage         string                    `json:"recent_probe_last_message"`
	ProbeSummaryBasis              string                    `json:"probe_summary_basis"`
	CircuitTrackedCount            int                       `json:"circuit_tracked_count"`
	CircuitOpenCount               int                       `json:"circuit_open_count"`
	CircuitHalfOpenCount           int                       `json:"circuit_half_open_count"`
	CircuitClosedCount             int                       `json:"circuit_closed_count"`
	CircuitMaxRemainingCooldownSec int                       `json:"circuit_max_remaining_cooldown_sec"`
	CircuitSummaryBasis            string                    `json:"circuit_summary_basis"`
	ByChannel                      []StatsTokenBreakdownItem `json:"by_channel"`
	ByModel                        []StatsTokenBreakdownItem `json:"by_model"`
	ByAPIKey                       []StatsTokenBreakdownItem `json:"by_api_key,omitempty"`
	ByChannelKey                   []StatsTokenBreakdownItem `json:"by_channel_key,omitempty"`
}

func StatsTokenBreakdownGetByWindow(window string) StatsTokenBreakdown {
	switch strings.TrimSpace(window) {
	case "", "1d":
		return StatsTokenBreakdownGet()
	case "12h", "3d", "7d", "30d":
		return statsTokenBreakdownFromRelayLogs(window)
	default:
		return StatsTokenBreakdownGet()
	}
}

func statsTokenBreakdownFromRelayLogs(window string) StatsTokenBreakdown {
	duration := 24 * time.Hour
	switch window {
	case "12h":
		duration = 12 * time.Hour
	case "3d":
		duration = 72 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	}
	end := int(time.Now().Unix())
	start := int(time.Now().Add(-duration).Unix())
	logs, err := RelayLogExport(context.Background(), &start, &end, 5000)
	if err != nil {
		return StatsTokenBreakdownGet()
	}
	result := StatsTokenBreakdown{EstimatedPriceBasis: "relay_logs_window"}
	byChannel := map[string]*StatsTokenBreakdownItem{}
	byModel := map[string]*StatsTokenBreakdownItem{}
	byAPIKey := map[string]*StatsTokenBreakdownItem{}
	byChannelKey := map[string]*StatsTokenBreakdownItem{}
	llmInfoByName, llmInfoByCanonical := statsLLMInfoMaps()
	channelNames := map[int]string{}
	for _, channel := range channelCache.Values() {
		channelNames[channel.ID] = channel.Name
	}
	apiKeyNames := map[int]string{}
	for _, key := range apiKeyCache.Values() {
		apiKeyNames[key.ID] = key.Name
	}
	for _, item := range logs {
		result.TotalInputToken += int64(item.InputTokens)
		result.TotalOutputToken += int64(item.OutputTokens)
		totalToken := int64(item.InputTokens + item.OutputTokens)
		result.TotalToken += totalToken
		actualModel := statsFirstNonEmpty(strings.TrimSpace(item.ActualModelName), strings.TrimSpace(item.RequestModelName))
		if gatewayPrice, ok := ResolveGatewayLLMPrice(actualModel, item.ChannelId); ok {
			inputCost, outputCost := estimateRelayLogCostWithPrice(item, gatewayPrice)
			result.EstimatedGatewayInputCost += inputCost
			result.EstimatedGatewayOutputCost += outputCost
		} else if info, ok := statsFindLLMInfo(actualModel, llmInfoByName, llmInfoByCanonical); ok {
			inputCost, outputCost := estimateRelayLogCostWithPrice(item, info.LLMPrice)
			result.EstimatedGatewayInputCost += inputCost
			result.EstimatedGatewayOutputCost += outputCost
		} else {
			result.EstimatedGatewayTotalCost += item.Cost
		}
		if info, ok := statsFindLLMInfo(actualModel, llmInfoByName, llmInfoByCanonical); ok {
			inputCost, outputCost := estimateRelayLogOfficialCost(item, info.OfficialLLMPrice)
			result.EstimatedOfficialInputCost += inputCost
			result.EstimatedOfficialOutputCost += outputCost
		}
		addBreakdownItem(byChannel, fmt.Sprintf("channel:%d", item.ChannelId), statsFirstNonEmpty(strings.TrimSpace(item.ChannelName), channelNames[item.ChannelId], fmt.Sprintf("Channel %d", item.ChannelId)), int64(item.InputTokens), int64(item.OutputTokens))
		addBreakdownItem(byModel, fmt.Sprintf("model:%s", strings.TrimSpace(item.ActualModelName)), statsFirstNonEmpty(strings.TrimSpace(item.ActualModelName), strings.TrimSpace(item.RequestModelName), "unknown-model"), int64(item.InputTokens), int64(item.OutputTokens))
		if item.APIKeyID > 0 {
			addBreakdownItem(byAPIKey, fmt.Sprintf("api_key:%d", item.APIKeyID), statsFirstNonEmpty(apiKeyNames[item.APIKeyID], fmt.Sprintf("API Key %d", item.APIKeyID)), int64(item.InputTokens), int64(item.OutputTokens))
		}
		for _, attempt := range item.Attempts {
			if attempt.ChannelKeyID <= 0 {
				continue
			}
			addBreakdownItem(byChannelKey, fmt.Sprintf("channel_key:%d", attempt.ChannelKeyID), fmt.Sprintf("%s / Key %d", statsFirstNonEmpty(attempt.ChannelName, item.ChannelName, fmt.Sprintf("Channel %d", attempt.ChannelID)), attempt.ChannelKeyID), int64(item.InputTokens), int64(item.OutputTokens))
			break
		}
	}
	result.ByChannel = flattenBreakdownItems(byChannel)
	result.ByModel = flattenBreakdownItems(byModel)
	result.ByAPIKey = flattenBreakdownItems(byAPIKey)
	result.ByChannelKey = flattenBreakdownItems(byChannelKey)
	result.EstimatedGatewayTotalCost += result.EstimatedGatewayInputCost + result.EstimatedGatewayOutputCost
	result.EstimatedOfficialTotalCost = result.EstimatedOfficialInputCost + result.EstimatedOfficialOutputCost
	return result
}

func addBreakdownItem(target map[string]*StatsTokenBreakdownItem, key, label string, inputToken, outputToken int64) {
	row, ok := target[key]
	if !ok {
		row = &StatsTokenBreakdownItem{Key: key, Label: label}
		target[key] = row
	}
	row.InputToken += inputToken
	row.OutputToken += outputToken
	row.TotalToken += inputToken + outputToken
}

func flattenBreakdownItems(source map[string]*StatsTokenBreakdownItem) []StatsTokenBreakdownItem {
	items := make([]StatsTokenBreakdownItem, 0, len(source))
	for _, item := range source {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].TotalToken > items[j].TotalToken })
	return items
}

func statsFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func statsLLMInfoMaps() (map[string]model.LLMInfo, map[string]model.LLMInfo) {
	llmInfos, _ := LLMList(context.Background())
	llmInfoByName := make(map[string]model.LLMInfo, len(llmInfos))
	llmInfoByCanonical := make(map[string]model.LLMInfo, len(llmInfos))
	for _, info := range llmInfos {
		normalizedName := strings.ToLower(strings.TrimSpace(info.Name))
		if normalizedName != "" {
			llmInfoByName[normalizedName] = info
		}
		canonicalName := strings.ToLower(strings.TrimSpace(info.CanonicalName))
		if canonicalName != "" {
			llmInfoByCanonical[canonicalName] = info
		}
	}
	return llmInfoByName, llmInfoByCanonical
}

func statsFindLLMInfo(modelName string, byName, byCanonical map[string]model.LLMInfo) (model.LLMInfo, bool) {
	normalizedName := strings.ToLower(strings.TrimSpace(modelName))
	if info, ok := byName[normalizedName]; ok {
		return info, true
	}
	canonicalName := llmname.CanonicalModelName(normalizedName)
	info, ok := byCanonical[strings.ToLower(strings.TrimSpace(canonicalName))]
	return info, ok
}

func estimateRelayLogCostWithPrice(item model.RelayLog, price model.LLMPrice) (float64, float64) {
	regularInput := item.InputTokens - item.CacheReadTokens
	if regularInput < 0 {
		regularInput = item.InputTokens
	}
	inputCost := (float64(regularInput)*price.Input + float64(item.CacheReadTokens)*price.CacheRead + float64(item.CacheWriteTokens)*price.CacheWrite) * 1e-6
	outputCost := float64(item.OutputTokens) * price.Output * 1e-6
	return inputCost, outputCost
}

func estimateRelayLogOfficialCost(item model.RelayLog, price model.OfficialLLMPrice) (float64, float64) {
	regularInput := item.InputTokens - item.CacheReadTokens
	if regularInput < 0 {
		regularInput = item.InputTokens
	}
	inputCost := (float64(regularInput)*price.OfficialInput + float64(item.CacheReadTokens)*price.OfficialCacheRead + float64(item.CacheWriteTokens)*price.OfficialCacheWrite) * 1e-6
	outputCost := float64(item.OutputTokens) * price.OfficialOutput * 1e-6
	return inputCost, outputCost
}

func StatsTokenBreakdownGet() StatsTokenBreakdown {
	channelStats := statsChannelCache.GetAll()
	modelStats := statsModelCache.GetAll()
	channels := channelCache.GetAll()
	llmInfoByName, llmInfoByCanonical := statsLLMInfoMaps()

	result := StatsTokenBreakdown{
		EstimatedPriceBasis: "model_tokens_only",
	}

	byChannel := make([]StatsTokenBreakdownItem, 0, len(channelStats))
	for channelID, stat := range channelStats {
		label := fmt.Sprintf("Channel %d", channelID)
		if ch, ok := channels[channelID]; ok && ch.Name != "" {
			label = ch.Name
		}
		total := stat.InputToken + stat.OutputToken
		result.TotalInputToken += stat.InputToken
		result.TotalOutputToken += stat.OutputToken
		result.TotalToken += total
		byChannel = append(byChannel, StatsTokenBreakdownItem{
			Key:         fmt.Sprintf("channel:%d", channelID),
			Label:       label,
			InputToken:  stat.InputToken,
			OutputToken: stat.OutputToken,
			TotalToken:  total,
		})
	}

	byModel := make([]StatsTokenBreakdownItem, 0, len(modelStats))
	for _, stat := range modelStats {
		total := stat.InputToken + stat.OutputToken
		normalizedName := strings.ToLower(strings.TrimSpace(stat.Name))
		info, ok := llmInfoByName[normalizedName]
		if !ok {
			canonicalName := llmname.CanonicalModelName(normalizedName)
			info, ok = llmInfoByCanonical[canonicalName]
		}
		if ok {
			result.EstimatedGatewayInputCost += float64(stat.InputToken) * info.Input * 1e-6
			result.EstimatedGatewayOutputCost += float64(stat.OutputToken) * info.Output * 1e-6
			result.EstimatedOfficialInputCost += float64(stat.InputToken) * info.OfficialInput * 1e-6
			result.EstimatedOfficialOutputCost += float64(stat.OutputToken) * info.OfficialOutput * 1e-6
		}
		byModel = append(byModel, StatsTokenBreakdownItem{
			Key:         fmt.Sprintf("model:%d", stat.ID),
			Label:       stat.Name,
			InputToken:  stat.InputToken,
			OutputToken: stat.OutputToken,
			TotalToken:  total,
		})
	}

	result.EstimatedGatewayTotalCost = result.EstimatedGatewayInputCost + result.EstimatedGatewayOutputCost
	result.EstimatedOfficialTotalCost = result.EstimatedOfficialInputCost + result.EstimatedOfficialOutputCost

	sort.Slice(byChannel, func(i, j int) bool { return byChannel[i].TotalToken > byChannel[j].TotalToken })
	sort.Slice(byModel, func(i, j int) bool { return byModel[i].TotalToken > byModel[j].TotalToken })

	result.ByChannel = byChannel
	result.ByModel = byModel
	return result
}

func StatsHourlyGet() []model.StatsHourly {
	now := time.Now()
	currentHour := now.Hour()
	todayDate := time.Now().Format("20060102")

	statsHourlyCacheLock.RLock()
	defer statsHourlyCacheLock.RUnlock()

	result := make([]model.StatsHourly, 0, currentHour+1)

	for hour := 0; hour <= currentHour; hour++ {
		if statsHourlyCache[hour].Date == todayDate {
			result = append(result, statsHourlyCache[hour])
		} else {
			result = append(result, model.StatsHourly{
				Hour: hour,
				Date: todayDate,
			})
		}
	}

	return result
}

func StatsGetDaily(ctx context.Context) ([]model.StatsDaily, error) {
	var statsDaily []model.StatsDaily
	result := db.GetDB().WithContext(ctx).Find(&statsDaily)
	if result.Error != nil {
		return nil, result.Error
	}
	return statsDaily, nil
}

func statsRefreshCache(ctx context.Context) error {
	dbConn := db.GetDB().WithContext(ctx)
	today := time.Now().Format("20060102")

	var loadedDaily model.StatsDaily
	result := dbConn.Last(&loadedDaily)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get daily stats: %v", result.Error)
	}
	if result.RowsAffected == 0 || loadedDaily.Date != today {
		loadedDaily = model.StatsDaily{Date: today}
	}

	var loadedTotal model.StatsTotal
	result = dbConn.First(&loadedTotal)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get total stats: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		loadedTotal = model.StatsTotal{ID: 1}
	} else if loadedTotal.ID == 0 {
		loadedTotal.ID = 1
	}

	var loadedChannels []model.StatsChannel
	result = dbConn.Find(&loadedChannels)
	if result.Error != nil {
		return fmt.Errorf("failed to get channels: %v", result.Error)
	}

	var loadedHourly []model.StatsHourly
	result = dbConn.Find(&loadedHourly)
	if result.Error != nil {
		return fmt.Errorf("failed to get hourly stats: %v", result.Error)
	}

	statsDailyCacheLock.Lock()
	statsDailyCache = loadedDaily
	statsDailyCacheLock.Unlock()

	statsTotalCacheLock.Lock()
	statsTotalCache = loadedTotal
	statsTotalCacheLock.Unlock()

	statsChannelCache.Clear()
	statsChannelCacheNeedUpdateLock.Lock()
	statsChannelCacheNeedUpdate = make(map[int]struct{})
	statsChannelCacheNeedUpdateLock.Unlock()
	for _, v := range loadedChannels {
		statsChannelCache.Set(v.ChannelID, v)
	}

	var loadedAPIKeys []model.StatsAPIKey
	result = dbConn.Find(&loadedAPIKeys)
	if result.Error != nil {
		return fmt.Errorf("failed to get api key stats: %v", result.Error)
	}

	statsAPIKeyCache.Clear()
	statsAPIKeyCacheNeedUpdateLock.Lock()
	statsAPIKeyCacheNeedUpdate = make(map[int]struct{})
	statsAPIKeyCacheNeedUpdateLock.Unlock()
	for _, v := range loadedAPIKeys {
		statsAPIKeyCache.Set(v.APIKeyID, v)
	}

	statsHourlyCacheLock.Lock()
	statsHourlyCache = [24]model.StatsHourly{}
	for _, v := range loadedHourly {
		if v.Hour >= 0 && v.Hour < 24 {
			statsHourlyCache[v.Hour] = v
		}
	}
	statsHourlyCacheLock.Unlock()

	return nil
}
