package logic

import (
	"database/sql"
	"testing"
	"time"

	"api/model"
)

func TestDailyTaskAccessAllowsEditingWithin24Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 16, 9, 0, 0, 0, time.Local)
	item := &model.DailyTasks{SubmittedAt: sql.NullTime{Time: time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local), Valid: true}}

	access := dailyTaskAccess(item, now)
	if !access.CanEditContent {
		t.Fatalf("expected task to remain editable within 24 hours")
	}
	if access.CanAppendReflection {
		t.Fatalf("expected reflection mode to stay off within 24 hours")
	}
}

func TestDailyTaskAccessLocksOriginalContentAfter24Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 16, 10, 1, 0, 0, time.Local)
	item := &model.DailyTasks{SubmittedAt: sql.NullTime{Time: time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local), Valid: true}}

	access := dailyTaskAccess(item, now)
	if access.CanEditContent {
		t.Fatalf("expected original content to be locked after 24 hours")
	}
	if !access.CanAppendReflection {
		t.Fatalf("expected reflection to remain allowed after edit window closes")
	}
}
