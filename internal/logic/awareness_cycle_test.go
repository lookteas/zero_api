package logic

import (
	"testing"
	"time"

	"api/model"
)

func TestResolveAwarenessCycleNormalDayMapsToExpectedAwarenessPoint(t *testing.T) {
	startDate := testAwarenessCycleDate(t, "2026-05-01")
	targetDate := testAwarenessCycleDate(t, "2026-05-02")
	points := testAwarenessPoints(3)

	result := resolveAwarenessCycleDay(startDate, targetDate, 7, points)

	if result.IsPreStart {
		t.Fatalf("expected normal day, got pre-start")
	}
	if result.IsRestDay {
		t.Fatalf("expected awareness day, got rest day")
	}
	if result.Awareness == nil {
		t.Fatalf("expected awareness point, got nil")
	}
	if result.Awareness.AwarenessId != 2 {
		t.Fatalf("expected awareness id 2, got %d", result.Awareness.AwarenessId)
	}
	if result.CycleIndex != 0 {
		t.Fatalf("expected cycle index 0, got %d", result.CycleIndex)
	}
	if result.RestDayIndex != 0 {
		t.Fatalf("expected rest day index 0, got %d", result.RestDayIndex)
	}
}

func TestResolveAwarenessCycleRestDayAfterLastPoint(t *testing.T) {
	startDate := testAwarenessCycleDate(t, "2026-05-01")
	targetDate := testAwarenessCycleDate(t, "2026-05-04")
	points := testAwarenessPoints(3)

	result := resolveAwarenessCycleDay(startDate, targetDate, 7, points)

	if !result.IsRestDay {
		t.Fatalf("expected rest day")
	}
	if result.Awareness != nil {
		t.Fatalf("expected no awareness point, got %+v", result.Awareness)
	}
	if result.RestDayIndex != 1 {
		t.Fatalf("expected first rest day, got %d", result.RestDayIndex)
	}
	if result.CycleIndex != 0 {
		t.Fatalf("expected cycle index 0, got %d", result.CycleIndex)
	}
}

func TestResolveAwarenessCycleRestartsAfterSevenRestDays(t *testing.T) {
	startDate := testAwarenessCycleDate(t, "2026-05-01")
	targetDate := testAwarenessCycleDate(t, "2026-05-11")
	points := testAwarenessPoints(3)

	result := resolveAwarenessCycleDay(startDate, targetDate, 7, points)

	if result.IsRestDay {
		t.Fatalf("expected awareness day after rest period")
	}
	if result.Awareness == nil {
		t.Fatalf("expected awareness point, got nil")
	}
	if result.Awareness.AwarenessId != 1 {
		t.Fatalf("expected awareness id 1, got %d", result.Awareness.AwarenessId)
	}
	if result.CycleIndex != 1 {
		t.Fatalf("expected cycle index 1, got %d", result.CycleIndex)
	}
}

func TestResolveAwarenessCyclePreStartDate(t *testing.T) {
	startDate := testAwarenessCycleDate(t, "2026-05-01")
	targetDate := testAwarenessCycleDate(t, "2026-04-30")
	points := testAwarenessPoints(3)

	result := resolveAwarenessCycleDay(startDate, targetDate, 7, points)

	if !result.IsPreStart {
		t.Fatalf("expected pre-start")
	}
	if result.Awareness != nil {
		t.Fatalf("expected no awareness point, got %+v", result.Awareness)
	}
	if result.IsRestDay {
		t.Fatalf("expected pre-start to not be a rest day")
	}
}

func TestParseAwarenessCycleStartFallsBackToDefaultDate(t *testing.T) {
	result := parseAwarenessCycleStart("not-a-date")

	if got, want := result.Format("2006-01-02"), "2026-05-01"; got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func testAwarenessCycleDate(t *testing.T, input string) time.Time {
	t.Helper()

	parsed, err := time.ParseInLocation("2006-01-02", input, time.Local)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	return parsed
}

func testAwarenessPoints(count int) []model.Awareness {
	points := make([]model.Awareness, 0, count)
	for i := 1; i <= count; i++ {
		points = append(points, model.Awareness{
			AwarenessId: uint64(i),
			PointTitle:  "test point",
		})
	}
	return points
}
