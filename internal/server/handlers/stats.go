package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/relay/balancer"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/resp"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/task"
)

func init() {
	router.NewGroupRouter("/api/v1/stats").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/today", http.MethodGet).
				Handle(getStatsToday),
		).
		AddRoute(
			router.NewRoute("/daily", http.MethodGet).
				Handle(getStatsDaily),
		).
		AddRoute(
			router.NewRoute("/hourly", http.MethodGet).
				Handle(getStatsHourly),
		).
		AddRoute(
			router.NewRoute("/total", http.MethodGet).
				Handle(getStatsTotal),
		).
		AddRoute(
			router.NewRoute("/apikey", http.MethodGet).
				Handle(getStatsAPIKey),
		).
		AddRoute(
			router.NewRoute("/token-breakdown", http.MethodGet).
				Handle(getStatsTokenBreakdown),
		).
		AddRoute(
			router.NewRoute("/dynamic-routing-summary", http.MethodGet).
				Handle(getStatsDynamicRoutingSummary),
		).
		AddRoute(
			router.NewRoute("/export", http.MethodGet).
				Handle(getStatsExport),
		)
}

func getStatsToday(c *gin.Context) {
	resp.Success(c, op.StatsTodayGet())
}

func getStatsDaily(c *gin.Context) {
	statsDaily, err := op.StatsGetDaily(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to get daily stats")
		return
	}
	resp.Success(c, statsDaily)
}

func getStatsHourly(c *gin.Context) {
	resp.Success(c, op.StatsHourlyGet())
}

func getStatsTotal(c *gin.Context) {
	resp.Success(c, op.StatsTotalGet())
}

func getStatsAPIKey(c *gin.Context) {
	resp.Success(c, op.StatsAPIKeyList())
}

func getStatsTokenBreakdown(c *gin.Context) {
	window, _, err := parseOptionalNonEmptyTrimmedStringQuery(c, "window")
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	breakdown := op.StatsTokenBreakdownGetByWindow(window)
	probeSummary := op.ProbeSummaryGet(24 * time.Hour)
	breakdown.EstimatedProbeInputCost = probeSummary.EstimatedInputCost
	breakdown.EstimatedProbeOutputCost = probeSummary.EstimatedOutputCost
	breakdown.EstimatedProbeTotalCost = probeSummary.EstimatedTotalCost
	breakdown.RecentProbeCount = probeSummary.TotalCount
	breakdown.RecentProbeSuccessCount = probeSummary.SuccessCount
	breakdown.RecentProbeFailedCount = probeSummary.FailedCount
	breakdown.RecentProbeLastAt = probeSummary.LastAt
	breakdown.RecentProbeLastStatus = probeSummary.LastStatus
	breakdown.RecentProbeLastChannel = probeSummary.LastChannel
	breakdown.RecentProbeLastModel = probeSummary.LastModel
	breakdown.RecentProbeLastMessage = probeSummary.LastMessage
	breakdown.ProbeSummaryBasis = probeSummary.Basis

	circuitSummary := balancer.SnapshotSummary(time.Now())
	breakdown.CircuitTrackedCount = circuitSummary.TrackedCount
	breakdown.CircuitOpenCount = circuitSummary.OpenCount
	breakdown.CircuitHalfOpenCount = circuitSummary.HalfOpenCount
	breakdown.CircuitClosedCount = circuitSummary.ClosedCount
	breakdown.CircuitMaxRemainingCooldownSec = circuitSummary.MaxRemainingCooldownSec
	breakdown.CircuitSummaryBasis = circuitSummary.Basis

	resp.Success(c, breakdown)
}

func getStatsDynamicRoutingSummary(c *gin.Context) {
	resp.Success(c, task.GetDynamicRoutingSummaryScanSummary())
}

// statsExportDimension 支持导出的统计维度
type statsExportDimension string

const (
	statsExportDimensionChannel statsExportDimension = "channel"
	statsExportDimensionModel   statsExportDimension = "model"
	statsExportDimensionAPIKey  statsExportDimension = "apikey"
)

func parseStatsExportDimension(raw string) (statsExportDimension, error) {
	switch raw {
	case "channel":
		return statsExportDimensionChannel, nil
	case "model":
		return statsExportDimensionModel, nil
	case "apikey":
		return statsExportDimensionAPIKey, nil
	default:
		return "", fmt.Errorf("unsupported dimension: %s", raw)
	}
}

// statsExportRow CSV 导出的一行数据
type statsExportRow struct {
	ID           int
	Name         string
	RequestTotal int64
	Success      int64
	Failed       int64
	InputToken   int64
	OutputToken  int64
	Cost         float64
}

// getStatsExport 导出统计数据为 CSV。
// 查询参数：dimension（channel/model/apikey）、start（RFC3339）、end（RFC3339）。
// 注意：当前导出基于内存中的累计统计快照，start/end 参数已接受但暂不按时间过滤。
func getStatsExport(c *gin.Context) {
	dimension, err := parseStatsExportDimension(c.Query("dimension"))
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	// start/end 参数已接受，用于 API 前向兼容（当前导出累计值）
	if _, err := parseOptionalRFC3339TimeQuery(c, "start"); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := parseOptionalRFC3339TimeQuery(c, "end"); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	rows := collectStatsExportRows(c, dimension)
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	header := []string{"维度ID", "维度名称", "请求数", "成功数", "失败数", "输入Token", "输出Token", "成本"}
	if err := writer.Write(header); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to write csv header")
		return
	}

	for _, row := range rows {
		record := []string{
			strconv.Itoa(row.ID),
			row.Name,
			strconv.FormatInt(row.RequestTotal, 10),
			strconv.FormatInt(row.Success, 10),
			strconv.FormatInt(row.Failed, 10),
			strconv.FormatInt(row.InputToken, 10),
			strconv.FormatInt(row.OutputToken, 10),
			strconv.FormatFloat(row.Cost, 'f', 6, 64),
		}
		if err := writer.Write(record); err != nil {
			resp.Error(c, http.StatusInternalServerError, "failed to write csv row")
			return
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to flush csv")
		return
	}

	fileName := fmt.Sprintf("stats_%s_%s.csv", dimension, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

// collectStatsExportRows 根据维度收集统计数据行。
func collectStatsExportRows(c *gin.Context, dimension statsExportDimension) []statsExportRow {
	rows := make([]statsExportRow, 0)
	switch dimension {
	case statsExportDimensionChannel:
		channels, err := op.ChannelList(c.Request.Context())
		if err != nil {
			return rows
		}
		nameMap := make(map[int]string, len(channels))
		for _, ch := range channels {
			nameMap[ch.ID] = ch.Name
		}
		for _, stat := range op.StatsChannelList() {
			name := nameMap[stat.ChannelID]
			if name == "" {
				name = fmt.Sprintf("Channel %d", stat.ChannelID)
			}
			rows = append(rows, statsExportRow{
				ID:           stat.ChannelID,
				Name:         name,
				RequestTotal: stat.RequestSuccess + stat.RequestFailed,
				Success:      stat.RequestSuccess,
				Failed:       stat.RequestFailed,
				InputToken:   stat.InputToken,
				OutputToken:  stat.OutputToken,
				Cost:         stat.InputCost + stat.OutputCost,
			})
		}

	case statsExportDimensionModel:
		for _, stat := range op.StatsModelList() {
			rows = append(rows, statsExportRow{
				ID:           stat.ID,
				Name:         stat.Name,
				RequestTotal: stat.RequestSuccess + stat.RequestFailed,
				Success:      stat.RequestSuccess,
				Failed:       stat.RequestFailed,
				InputToken:   stat.InputToken,
				OutputToken:  stat.OutputToken,
				Cost:         stat.InputCost + stat.OutputCost,
			})
		}

	case statsExportDimensionAPIKey:
		apiKeys, err := op.APIKeyList(c.Request.Context())
		if err != nil {
			return rows
		}
		nameMap := make(map[int]string, len(apiKeys))
		for _, key := range apiKeys {
			nameMap[key.ID] = key.Name
		}
		for _, stat := range op.StatsAPIKeyList() {
			name := nameMap[stat.APIKeyID]
			if name == "" {
				name = fmt.Sprintf("APIKey %d", stat.APIKeyID)
			}
			rows = append(rows, statsExportRow{
				ID:           stat.APIKeyID,
				Name:         name,
				RequestTotal: stat.RequestSuccess + stat.RequestFailed,
				Success:      stat.RequestSuccess,
				Failed:       stat.RequestFailed,
				InputToken:   stat.InputToken,
				OutputToken:  stat.OutputToken,
				Cost:         stat.InputCost + stat.OutputCost,
			})
		}
	}
	return rows
}
