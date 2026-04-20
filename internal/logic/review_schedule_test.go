package logic

import (
	"testing"
	"time"

	"api/model"
)

func TestBuildReviewStagePlansStartsAtDay3(t *testing.T) {
	base := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)

	plans := buildReviewStagePlans(base)
	if len(plans) != 3 {
		t.Fatalf("expected 3 review stages, got %d", len(plans))
	}

	wantNames := []string{"day3", "day7", "day30"}
	wantDates := []time.Time{
		base.AddDate(0, 0, 3),
		base.AddDate(0, 0, 7),
		base.AddDate(0, 0, 30),
	}

	for i := range plans {
		if plans[i].Name != wantNames[i] {
			t.Fatalf("expected stage %s at index %d, got %s", wantNames[i], i, plans[i].Name)
		}
		if !plans[i].DueAt.Equal(wantDates[i]) {
			t.Fatalf("expected due at %s, got %s", wantDates[i], plans[i].DueAt)
		}
	}
}

func TestVisibleDueReviewItemsReturnsOnlyNearestDuePendingItem(t *testing.T) {
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)
	items := []model.ReviewItems{
		{Id: 1, ReviewStage: "day7", Status: "pending", DueAt: now.Add(-2 * time.Hour)},
		{Id: 2, ReviewStage: "day3", Status: "pending", DueAt: now.Add(-6 * time.Hour)},
		{Id: 3, ReviewStage: "day30", Status: "pending", DueAt: now.Add(24 * time.Hour)},
		{Id: 4, ReviewStage: "day3", Status: "completed", DueAt: now.Add(-24 * time.Hour)},
	}

	visible := visibleDueReviewItems(items, now, 1)
	if len(visible) != 1 {
		t.Fatalf("expected exactly one visible review item, got %d", len(visible))
	}

	if visible[0].Id != 2 {
		t.Fatalf("expected nearest due pending item id 2, got %d", visible[0].Id)
	}
}

func TestVisibleDueReviewItemsIgnoresLegacyDay0(t *testing.T) {
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)
	items := []model.ReviewItems{
		{Id: 1, ReviewStage: "day0", Status: "pending", DueAt: now.Add(-24 * time.Hour)},
		{Id: 2, ReviewStage: "day3", Status: "pending", DueAt: now.Add(24 * time.Hour)},
	}

	visible := visibleDueReviewItems(items, now, 1)
	if len(visible) != 0 {
		t.Fatalf("expected legacy day0 item to be ignored, got %d items", len(visible))
	}
}

func TestVisibleDueReviewItemsReturnsEmptyBeforeDueDate(t *testing.T) {
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)
	items := []model.ReviewItems{
		{Id: 1, ReviewStage: "day3", Status: "pending", DueAt: now.Add(24 * time.Hour)},
		{Id: 2, ReviewStage: "day7", Status: "pending", DueAt: now.Add(72 * time.Hour)},
	}

	visible := visibleDueReviewItems(items, now, 1)
	if len(visible) != 0 {
		t.Fatalf("expected no visible review items before due date, got %d", len(visible))
	}
}
