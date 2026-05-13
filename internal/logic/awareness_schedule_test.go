package logic

import (
	"database/sql"
	"testing"
	"time"

	"api/model"
)

func TestGenerateAwarenessSchedulePausesDoNotConsumeEffectiveDays(t *testing.T) {
	t.Parallel()

	cycle := &model.AwarenessCycles{
		CycleId:     1,
		CommunityId: 1,
		StartDate:   testAwarenessCycleDate(t, "2026-05-01"),
		RestDays:    1,
	}
	points := []model.Awareness{
		{AwarenessId: 101, PointTitle: "第一天", SortOrderGlobal: 1},
		{AwarenessId: 102, PointTitle: "第二天", SortOrderGlobal: 2},
	}
	pauses := []model.AwarenessCyclePauses{
		{
			PauseId:        9,
			CycleId:        1,
			CommunityId:    1,
			PauseStartDate: testAwarenessCycleDate(t, "2026-05-01"),
			PauseEndDate:   testAwarenessCycleDate(t, "2026-05-01"),
			Reason:         sql.NullString{String: "五一", Valid: true},
			Status:         1,
		},
	}

	items := generateAwarenessScheduleDays(cycle, points, pauses, testAwarenessCycleDate(t, "2026-05-01"), testAwarenessCycleDate(t, "2026-05-04"))
	if len(items) != 4 {
		t.Fatalf("expected 4 schedule days, got %d", len(items))
	}

	if items[0].DayType != scheduleDayPaused || items[0].AwarenessId.Valid || items[0].PauseReason.String != "五一" {
		t.Fatalf("expected 2026-05-01 paused without awareness, got %+v", items[0])
	}
	if items[1].DayType != scheduleDayNormal || items[1].AwarenessId.Int64 != 101 || items[1].EffectiveDayIndex.Int64 != 0 {
		t.Fatalf("expected 2026-05-02 first awareness day, got %+v", items[1])
	}
	if items[2].DayType != scheduleDayNormal || items[2].AwarenessId.Int64 != 102 || items[2].EffectiveDayIndex.Int64 != 1 {
		t.Fatalf("expected 2026-05-03 second awareness day, got %+v", items[2])
	}
	if items[3].DayType != scheduleDayRest || items[3].AwarenessId.Valid {
		t.Fatalf("expected 2026-05-04 rest day after two effective days, got %+v", items[3])
	}
}

func TestFindPauseForDateMatchesInclusiveRange(t *testing.T) {
	t.Parallel()

	pauses := []model.AwarenessCyclePauses{
		{
			PauseStartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
			PauseEndDate:   time.Date(2026, 5, 3, 0, 0, 0, 0, time.Local),
		},
	}

	if findPauseForDate(pauses, time.Date(2026, 5, 1, 12, 0, 0, 0, time.Local)) == nil {
		t.Fatalf("expected start date to match pause range")
	}
	if findPauseForDate(pauses, time.Date(2026, 5, 3, 12, 0, 0, 0, time.Local)) == nil {
		t.Fatalf("expected end date to match pause range")
	}
	if findPauseForDate(pauses, time.Date(2026, 5, 4, 0, 0, 0, 0, time.Local)) != nil {
		t.Fatalf("expected date after range not to match")
	}
}
