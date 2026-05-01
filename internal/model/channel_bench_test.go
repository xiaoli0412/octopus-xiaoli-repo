package model

import (
	"fmt"
	"testing"
)

var channelKeySelectionSink ChannelKey

func benchmarkChannelForSelection(keyCount int) *Channel {
	keys := make([]ChannelKey, 0, keyCount)
	for i := 0; i < keyCount; i++ {
		keys = append(keys, ChannelKey{
			ID:            i + 1,
			ChannelID:     1,
			Enabled:       true,
			ChannelKey:    fmt.Sprintf("bench-key-%03d", i+1),
			AllowedModels: "bench-model",
		})
	}

	return &Channel{
		ID:                1,
		Enabled:           true,
		Model:             "bench-model",
		KeyManagementMode: KeyManagementModeClassified,
		KeyRoutingPolicy:  KeyRoutingPolicyRoundRobin,
		Keys:              keys,
	}
}

func BenchmarkChannelGetChannelKeyForModelExcept(b *testing.B) {
	for _, tc := range []struct {
		name     string
		keyCount int
		excluded map[int]struct{}
	}{
		{name: "keys=8/no-excluded", keyCount: 8},
		{name: "keys=32/no-excluded", keyCount: 32},
		{name: "keys=128/no-excluded", keyCount: 128},
		{name: "keys=32/excluded-4", keyCount: 32, excluded: map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}}},
		{name: "keys=128/excluded-8", keyCount: 128, excluded: map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 7: {}, 8: {}}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			resetKeyRoundRobin()
			channel := benchmarkChannelForSelection(tc.keyCount)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				channelKeySelectionSink = channel.GetChannelKeyForModelExcept("bench-model", tc.excluded)
			}
		})
	}
}
