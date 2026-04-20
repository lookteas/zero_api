package logic

import (
	"testing"
	"time"
)

func TestBuildCycleSummaryUsesClampedRestDays(t *testing.T) {
	completedAt := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	cases := []struct {
		totalPoints int64
		wantRest    int64
	}{
		{totalPoints: 180, wantRest: 6},
		{totalPoints: 200, wantRest: 7},
		{totalPoints: 250, wantRest: 9},
		{totalPoints: 20, wantRest: 5},
		{totalPoints: 500, wantRest: 10},
	}

	for _, tc := range cases {
		summary := BuildCycleSummary(tc.totalPoints, tc.totalPoints, completedAt)
		if summary == nil {
			t.Fatalf("expected summary for total points %d", tc.totalPoints)
		}
		if summary.RestDays != tc.wantRest {
			t.Fatalf("expected %d rest days for total points %d, got %d", tc.wantRest, tc.totalPoints, summary.RestDays)
		}
	}
}

func TestBuildCycleSummaryOnlyAppearsAfterCycleCompletes(t *testing.T) {
	completedAt := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	if summary := BuildCycleSummary(200, 199, completedAt); summary != nil {
		t.Fatalf("expected no summary before cycle completion")
	}
}

func TestBuildCycleSummaryComputesNextCycleStartAfterRestWindow(t *testing.T) {
	completedAt := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	summary := BuildCycleSummary(200, 200, completedAt)
	if summary == nil {
		t.Fatalf("expected summary after cycle completion")
	}
	if summary.NextCycleStartDate != "2026-04-26" {
		t.Fatalf("expected next cycle start date 2026-04-26, got %s", summary.NextCycleStartDate)
	}
}
