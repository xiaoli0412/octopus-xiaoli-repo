package task

import (
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/helper"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/op"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

func ChannelBaseUrlDelayTask() {
	log.Debugf("channel base url delay task started")
	startTime := time.Now()
	defer func() {
		log.Debugf("channel base url delay task finished, update time: %s", time.Since(startTime))
	}()
	ctx, cancel := taskContextWithTimeout(30 * time.Minute)
	defer cancel()
	channels, err := op.ChannelList(ctx)
	if err != nil {
		log.Errorf("failed to list channels: %v", err)
		return
	}
	for _, channel := range channels {
		helper.ChannelBaseUrlDelayUpdate(&channel, ctx)
	}
}
