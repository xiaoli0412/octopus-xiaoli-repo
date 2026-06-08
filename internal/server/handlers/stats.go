package handlers

import (
	"net/http"
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
