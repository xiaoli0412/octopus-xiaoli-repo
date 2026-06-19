package task

import (
	"context"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
	"golang.org/x/sync/errgroup"
)

const defaultUpstreamTaskConcurrency = 5

func UpstreamRefreshTask() {
	ctx, cancel := taskContextWithTimeout(30 * time.Minute)
	defer cancel()
	sites, err := op.UpstreamSiteList(ctx)
	if err != nil {
		log.Warnf("upstream refresh task failed to list sites: %v", err)
		return
	}
	now := time.Now()
	due := make([]model.UpstreamSite, 0, len(sites))
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
		due = append(due, site)
	}
	if len(due) == 0 {
		return
	}

	g, taskCtx := errgroup.WithContext(ctx)
	g.SetLimit(defaultUpstreamTaskConcurrency)
	for _, site := range due {
		site := site
		g.Go(func() error {
			siteCtx, siteCancel := context.WithTimeout(taskCtx, 2*time.Minute)
			defer siteCancel()
			if _, err := op.UpstreamSiteRefresh(siteCtx, model.UpstreamRefreshRequest{ID: site.ID, Manual: false}); err != nil {
				log.Warnf("upstream refresh failed (site=%d name=%q): %v", site.ID, site.Name, err)
				return nil
			}
			alert, err := op.CheckUpstreamBalance(siteCtx, site.ID)
			if err != nil {
				log.Warnf("upstream balance check failed (site=%d name=%q): %v", site.ID, site.Name, err)
				return nil
			}
			if alert.Alert {
				log.Warnf("upstream balance alert (site=%d name=%q): %s", site.ID, site.Name, alert.Message)
			} else {
				log.Debugf("upstream refresh completed (site=%d name=%q)", site.ID, site.Name)
			}
			return nil
		})
	}
	_ = g.Wait()
}

func UpstreamCheckinTask() {
	ctx, cancel := taskContextWithTimeout(30 * time.Minute)
	defer cancel()
	sites, err := op.UpstreamSiteList(ctx)
	if err != nil {
		log.Warnf("upstream checkin task failed to list sites: %v", err)
		return
	}
	now := time.Now()
	due := make([]model.UpstreamSite, 0, len(sites))
	for _, site := range sites {
		if !site.Enabled || !site.AutoCheckin {
			continue
		}
		interval := time.Duration(site.CheckinIntervalSecs) * time.Second
		if interval <= 0 {
			interval = 24 * time.Hour
		}
		if !site.LastCheckinAt.IsZero() && now.Sub(site.LastCheckinAt) < interval {
			continue
		}
		due = append(due, site)
	}
	if len(due) == 0 {
		return
	}

	g, taskCtx := errgroup.WithContext(ctx)
	g.SetLimit(defaultUpstreamTaskConcurrency)
	for _, site := range due {
		site := site
		g.Go(func() error {
			siteCtx, siteCancel := context.WithTimeout(taskCtx, 1*time.Minute)
			defer siteCancel()
			if _, err := op.UpstreamSiteCheckin(siteCtx, site.ID); err != nil {
				log.Warnf("upstream checkin failed (site=%d name=%q): %v", site.ID, site.Name, err)
				return nil
			}
			log.Debugf("upstream checkin completed (site=%d name=%q)", site.ID, site.Name)
			return nil
		})
	}
	_ = g.Wait()
}
