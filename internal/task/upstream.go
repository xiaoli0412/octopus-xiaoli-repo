package task

import (
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

func UpstreamRefreshTask() {
	ctx, cancel := taskContextWithTimeout(30 * time.Minute)
	defer cancel()
	sites, err := op.UpstreamSiteList(ctx)
	if err != nil {
		log.Warnf("upstream refresh task failed to list sites: %v", err)
		return
	}
	now := time.Now()
	for _, site := range sites {
		if !site.Enabled || !site.AutoRefresh {
			continue
		}
		interval := time.Duration(site.RefreshIntervalSecs) * time.Second
		if interval <= 0 {
			interval = 12 * time.Hour
		}
		if !site.LastRefreshAt.IsZero() && now.Sub(site.LastRefreshAt) < interval {
			continue
		}
		if _, err := op.UpstreamSiteRefresh(ctx, model.UpstreamRefreshRequest{ID: site.ID, Manual: false}); err != nil {
			log.Warnf("upstream refresh failed (site=%d): %v", site.ID, err)
		}
	}
}
