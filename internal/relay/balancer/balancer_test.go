package balancer

import (
	"testing"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
)

func TestRoundRobinCandidatesReturnStablePriorityOrder(t *testing.T) {
	items := []model.GroupItem{
		{ChannelID: 3, ModelName: "gpt-4o", Priority: 30},
		{ChannelID: 1, ModelName: "gpt-4o", Priority: 10},
		{ChannelID: 2, ModelName: "gpt-4o", Priority: 20},
	}

	got := (&RoundRobin{}).Candidates(items)
	want := []int{1, 2, 3}
	for i := range want {
		if got[i].ChannelID != want[i] {
			t.Fatalf("RoundRobin Candidates() order = [%d %d %d], want [1 2 3]", got[0].ChannelID, got[1].ChannelID, got[2].ChannelID)
		}
	}
}

func TestRoundRobinCandidatesUseStableTieBreakers(t *testing.T) {
	items := []model.GroupItem{
		{ID: 3, ChannelID: 9, ModelName: "z-model", Priority: 10, Weight: 1},
		{ID: 2, ChannelID: 7, ModelName: "a-model", Priority: 10, Weight: 1},
		{ID: 1, ChannelID: 7, ModelName: "a-model", Priority: 10, Weight: 2},
	}

	got := (&RoundRobin{}).Candidates(items)
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("RoundRobin Candidates() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i] {
			t.Fatalf("RoundRobin Candidates() order mismatch at %d: got ID %d want %d", i, got[i].ID, want[i])
		}
	}
}

func TestFailoverCandidatesReturnPriorityOrder(t *testing.T) {
	items := []model.GroupItem{
		{ChannelID: 30, ModelName: "gpt-4o", Priority: 3},
		{ChannelID: 10, ModelName: "gpt-4o", Priority: 1},
		{ChannelID: 20, ModelName: "gpt-4o", Priority: 2},
	}

	got := (&Failover{}).Candidates(items)
	want := []int{10, 20, 30}
	for i := range want {
		if got[i].ChannelID != want[i] {
			t.Fatalf("Failover Candidates() order = [%d %d %d], want [10 20 30]", got[0].ChannelID, got[1].ChannelID, got[2].ChannelID)
		}
	}
}

func TestRandomCandidatesPreserveMembershipWithoutPriorityBias(t *testing.T) {
	items := []model.GroupItem{
		{ID: 1, ChannelID: 30, ModelName: "gpt-4o", Priority: 3},
		{ID: 2, ChannelID: 10, ModelName: "gpt-4o", Priority: 1},
		{ID: 3, ChannelID: 20, ModelName: "gpt-4o", Priority: 2},
	}

	got := (&Random{}).Candidates(items)
	if len(got) != len(items) {
		t.Fatalf("Random Candidates() len = %d, want %d", len(got), len(items))
	}

	seen := make(map[int]struct{}, len(items))
	for _, item := range got {
		seen[item.ID] = struct{}{}
	}
	for _, item := range items {
		if _, ok := seen[item.ID]; !ok {
			t.Fatalf("Random Candidates() missing item ID %d in %#v", item.ID, got)
		}
	}
}

func TestWeightedCandidatesBuildStrictDeterministicSequence(t *testing.T) {
	items := []model.GroupItem{
		{ChannelID: 2, ModelName: "gpt-4o", Priority: 2, Weight: 3},
		{ChannelID: 1, ModelName: "gpt-4o", Priority: 1, Weight: 2},
		{ChannelID: 3, ModelName: "gpt-4o", Priority: 3, Weight: 1},
	}

	got := (&Weighted{}).Candidates(items)
	want := []int{1, 1, 2, 2, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("Weighted Candidates() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ChannelID != want[i] {
			t.Fatalf("Weighted Candidates() sequence mismatch at %d: got %d want %d", i, got[i].ChannelID, want[i])
		}
	}
}

func TestWeightedCandidatesUseStableTieBreakersBeforeExpansion(t *testing.T) {
	items := []model.GroupItem{
		{ID: 4, ChannelID: 20, ModelName: "gpt-4o", Priority: 1, Weight: 1},
		{ID: 2, ChannelID: 10, ModelName: "z-model", Priority: 1, Weight: 2},
		{ID: 1, ChannelID: 10, ModelName: "a-model", Priority: 1, Weight: 2},
		{ID: 3, ChannelID: 10, ModelName: "a-model", Priority: 1, Weight: 1},
	}

	got := (&Weighted{}).Candidates(items)
	want := []int{1, 1, 2, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("Weighted Candidates() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i] {
			t.Fatalf("Weighted Candidates() sequence mismatch at %d: got ID %d want %d", i, got[i].ID, want[i])
		}
	}
}

func TestWeightedCandidatesTreatNonPositiveWeightAsOne(t *testing.T) {
	items := []model.GroupItem{
		{ChannelID: 1, ModelName: "gpt-4o", Priority: 1, Weight: 0},
		{ChannelID: 2, ModelName: "gpt-4o", Priority: 2, Weight: -5},
	}

	got := (&Weighted{}).Candidates(items)
	want := []int{1, 2}
	if len(got) != len(want) {
		t.Fatalf("Weighted Candidates() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ChannelID != want[i] {
			t.Fatalf("Weighted Candidates() = [%d %d], want [1 2]", got[0].ChannelID, got[1].ChannelID)
		}
	}
}
