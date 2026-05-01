package op

import (
	"fmt"
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

var statsTokenBreakdownSink StatsTokenBreakdown
var channelKeyUpdateSink model.ChannelKey

func seedStatsTokenBreakdownBenchmark(channelCount, modelCount int) {
	resetOpTestState()

	for channelID := 1; channelID <= channelCount; channelID++ {
		channelCache.Set(channelID, model.Channel{
			ID:      channelID,
			Name:    fmt.Sprintf("channel-%03d", channelID),
			Enabled: true,
		})
		statsChannelCache.Set(channelID, model.StatsChannel{
			ChannelID: channelID,
			StatsMetrics: model.StatsMetrics{
				InputToken:     int64(channelID * 100),
				OutputToken:    int64(channelID * 80),
				RequestSuccess: int64(channelID),
			},
		})
	}

	for modelID := 1; modelID <= modelCount; modelID++ {
		name := fmt.Sprintf("bench-model-%04d", modelID)
		llmModelCache.Set(name, model.LLMInfo{
			Name:          name,
			CanonicalName: name,
			LLMPrice: model.LLMPrice{
				Input:      0.2,
				Output:     0.8,
				CacheRead:  0.05,
				CacheWrite: 0.1,
			},
			OfficialLLMPrice: model.OfficialLLMPrice{
				OfficialInput:  0.3,
				OfficialOutput: 1.0,
			},
		})

		statsModelCache.Set(modelID, model.StatsModel{
			ID:        modelID,
			Name:      name,
			ChannelID: (modelID % channelCount) + 1,
			StatsMetrics: model.StatsMetrics{
				InputToken:     int64(modelID * 13),
				OutputToken:    int64(modelID * 7),
				RequestSuccess: int64((modelID % 11) + 1),
			},
		})
	}
}

func BenchmarkStatsTokenBreakdownGet(b *testing.B) {
	for _, tc := range []struct {
		name         string
		channelCount int
		modelCount   int
	}{
		{name: "channels=32/models=256", channelCount: 32, modelCount: 256},
		{name: "channels=128/models=1024", channelCount: 128, modelCount: 1024},
	} {
		b.Run(tc.name, func(b *testing.B) {
			seedStatsTokenBreakdownBenchmark(tc.channelCount, tc.modelCount)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				statsTokenBreakdownSink = StatsTokenBreakdownGet()
			}
		})
	}
}

func BenchmarkChannelKeyUpdate(b *testing.B) {
	for _, keyCount := range []int{8, 32, 128} {
		b.Run(fmt.Sprintf("keys=%d", keyCount), func(b *testing.B) {
			resetOpTestState()

			keys := make([]model.ChannelKey, 0, keyCount)
			for i := 0; i < keyCount; i++ {
				keys = append(keys, model.ChannelKey{
					ID:         i + 1,
					ChannelID:  1,
					Enabled:    true,
					ChannelKey: fmt.Sprintf("bench-key-%03d", i+1),
					TotalCost:  float64(i),
				})
			}

			channelCache.Set(1, model.Channel{ID: 1, Name: "bench-channel", Enabled: true, Keys: keys})
			for _, key := range keys {
				channelKeyCache.Set(key.ID, key)
			}

			target := keys[keyCount/2]
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				target.TotalCost += 0.01
				target.StatusCode = 200 + (i & 1)
				target.LastUseTimeStamp = int64(i)
				if err := ChannelKeyUpdate(target); err != nil {
					b.Fatalf("ChannelKeyUpdate() error = %v", err)
				}
			}
			channelKeyUpdateSink = target
		})
	}
}
