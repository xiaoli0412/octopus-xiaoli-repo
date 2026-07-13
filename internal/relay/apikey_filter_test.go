package relay

import (
	"testing"

	dbmodel "github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestFilterGroupItemsByAllowedChannels(t *testing.T) {
	t.Parallel()

	items := []dbmodel.GroupItem{
		{ChannelID: 1, ModelName: "model-a"},
		{ChannelID: 2, ModelName: "model-a"},
		{ChannelID: 3, ModelName: "model-a"},
		{ChannelID: 4, ModelName: "model-a"},
	}

	cases := []struct {
		name    string
		items   []dbmodel.GroupItem
		allowed []int
		wantLen int
		wantIDs []int
	}{
		{
			name:    "nil allowed returns all items",
			items:   items,
			allowed: nil,
			wantLen: 4,
			wantIDs: []int{1, 2, 3, 4},
		},
		{
			name:    "empty allowed returns all items",
			items:   items,
			allowed: []int{},
			wantLen: 4,
			wantIDs: []int{1, 2, 3, 4},
		},
		{
			name:    "filter to channels 1 and 3",
			items:   items,
			allowed: []int{1, 3},
			wantLen: 2,
			wantIDs: []int{1, 3},
		},
		{
			name:    "single channel allowed",
			items:   items,
			allowed: []int{2},
			wantLen: 1,
			wantIDs: []int{2},
		},
		{
			name:    "no matching channels",
			items:   items,
			allowed: []int{99, 100},
			wantLen: 0,
			wantIDs: []int{},
		},
		{
			name:    "empty items with allowed",
			items:   []dbmodel.GroupItem{},
			allowed: []int{1, 2},
			wantLen: 0,
			wantIDs: []int{},
		},
		{
			name:    "nil items with allowed",
			items:   nil,
			allowed: []int{1, 2},
			wantLen: 0,
			wantIDs: []int{},
		},
		{
			name:    "all channels allowed",
			items:   items,
			allowed: []int{1, 2, 3, 4},
			wantLen: 4,
			wantIDs: []int{1, 2, 3, 4},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterGroupItemsByAllowedChannels(tc.items, tc.allowed)
			if len(got) != tc.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tc.wantLen)
			}
			gotIDs := make([]int, 0, len(got))
			for _, item := range got {
				gotIDs = append(gotIDs, item.ChannelID)
			}
			if !equalIntSlices(gotIDs, tc.wantIDs) {
				t.Fatalf("channel IDs = %v, want %v", gotIDs, tc.wantIDs)
			}
		})
	}
}

func TestFilterGroupItemsByAllowedChannelsDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	items := []dbmodel.GroupItem{
		{ChannelID: 1, ModelName: "model-a"},
		{ChannelID: 2, ModelName: "model-a"},
		{ChannelID: 3, ModelName: "model-a"},
	}
	original := make([]dbmodel.GroupItem, len(items))
	copy(original, items)

	_ = filterGroupItemsByAllowedChannels(items, []int{1, 3})

	for i := range items {
		if items[i].ChannelID != original[i].ChannelID {
			t.Fatalf("input slice was mutated at index %d: got ChannelID=%d, want %d", i, items[i].ChannelID, original[i].ChannelID)
		}
	}
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
