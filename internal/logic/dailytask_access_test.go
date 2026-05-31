package logic

import (
	"database/sql"
	"testing"
	"time"

	"api/model"
)

func TestDailyTaskAccessAllowsEditingWithin72Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 9, 0, 0, 0, time.Local)
	item := &model.DailyTasks{SubmittedAt: sql.NullTime{Time: time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local), Valid: true}}

	access := dailyTaskAccess(item, now)
	if !access.CanEditContent {
		t.Fatalf("expected task to remain editable within 72 hours")
	}
	if access.CanAppendReflection {
		t.Fatalf("expected reflection mode to stay off within 72 hours")
	}
}

func TestDailyTaskAccessLocksOriginalContentAfter72Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 10, 1, 0, 0, time.Local)
	item := &model.DailyTasks{SubmittedAt: sql.NullTime{Time: time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local), Valid: true}}

	access := dailyTaskAccess(item, now)
	if access.CanEditContent {
		t.Fatalf("expected original content to be locked after 72 hours")
	}
	if !access.CanAppendReflection {
		t.Fatalf("expected reflection to remain allowed after edit window closes")
	}
}
