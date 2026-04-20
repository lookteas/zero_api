package logic

import (
	"testing"
	"time"

	"api/internal/types"
	"api/model"
)

func TestBuildReviewPresentationKeepsNormalModeItemsSeparate(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -10)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.Add(-3 * time.Hour)},
		{Id: 2, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.Add(-2 * time.Hour)},
	}

	presentation := BuildReviewPresentation(items, now, lastActiveAt, 10)
	if presentation.RecoveryMode {
		t.Fatalf("expected normal mode presentation")
	}
	if len(presentation.VisibleItems) != 2 {
		t.Fatalf("expected 2 visible items in normal mode, got %d", len(presentation.VisibleItems))
	}
	if len(presentation.RecoveryGroups) != 0 {
		t.Fatalf("expected no recovery groups in normal mode, got %d", len(presentation.RecoveryGroups))
	}
}

func TestBuildReviewPresentationGroupsOverdueItemsInRecoveryMode(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -46)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -30)},
		{Id: 2, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.AddDate(0, 0, -26)},
		{Id: 3, DailyTaskId: 101, ReviewStage: "day30", Status: "pending", DueAt: now.AddDate(0, 0, -3)},
		{Id: 4, DailyTaskId: 202, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -8)},
	}

	presentation := BuildReviewPresentation(items, now, lastActiveAt, 10)
	if !presentation.RecoveryMode {
		t.Fatalf("expected recovery mode presentation")
	}
	if len(presentation.VisibleItems) != 0 {
		t.Fatalf("expected overdue items to be grouped in recovery mode, got %d visible items", len(presentation.VisibleItems))
	}
	if len(presentation.RecoveryGroups) != 2 {
		t.Fatalf("expected 2 recovery groups, got %d", len(presentation.RecoveryGroups))
	}
	if presentation.RecoveryGroups[0].DailyTaskId != 101 {
		t.Fatalf("expected first recovery group to belong to task 101, got %d", presentation.RecoveryGroups[0].DailyTaskId)
	}
	if presentation.RecoveryGroups[0].MergedStageLabel != "day3 / day7 / day30" {
		t.Fatalf("unexpected merged stage label: %s", presentation.RecoveryGroups[0].MergedStageLabel)
	}
	if len(presentation.RecoveryGroups[0].ReviewItemIds) != 3 {
		t.Fatalf("expected 3 merged review item ids, got %d", len(presentation.RecoveryGroups[0].ReviewItemIds))
	}
}

func TestBuildReviewPresentationHidesFutureStagesWhileRecoveryGroupExists(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -60)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -10)},
		{Id: 2, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.AddDate(0, 0, 2)},
	}

	presentation := BuildReviewPresentation(items, now, lastActiveAt, 10)
	if len(presentation.RecoveryGroups) != 1 {
		t.Fatalf("expected 1 recovery group, got %d", len(presentation.RecoveryGroups))
	}
	if len(presentation.RecoveryGroups[0].ReviewItemIds) != 1 || presentation.RecoveryGroups[0].ReviewItemIds[0] != 1 {
		t.Fatalf("expected only overdue item id 1 to be merged, got %#v", presentation.RecoveryGroups[0].ReviewItemIds)
	}
	if len(presentation.VisibleItems) != 0 {
		t.Fatalf("expected future stages to stay hidden while recovery exists, got %#v", presentation.VisibleItems)
	}
}

func TestReviewItemsResponseIncludesRecoveryGroups(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -50)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -5)},
		{Id: 2, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.AddDate(0, 0, -1)},
	}
	tasks := map[uint64]types.DailyTaskInfo{
		101: {Id: 101, TaskDate: "2026-04-01", TopicTitle: "Topic Alpha"},
	}

	data := BuildReviewItemListData(items, tasks, map[uint64]types.ReviewRecordInfo{}, now, lastActiveAt)
	if len(data.List) != 0 {
		t.Fatalf("expected no single review items in recovery mode, got %d", len(data.List))
	}
	if len(data.RecoveryGroups) != 1 {
		t.Fatalf("expected 1 recovery group, got %d", len(data.RecoveryGroups))
	}
	if data.RecoveryGroups[0].DailyTask.TopicTitle != "Topic Alpha" {
		t.Fatalf("expected task info to be attached to recovery group, got %#v", data.RecoveryGroups[0].DailyTask)
	}
}

func TestHomeReviewPresentationKeepsNearestPendingItemInNormalMode(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -3)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.Add(-2 * time.Hour)},
		{Id: 2, DailyTaskId: 202, ReviewStage: "day3", Status: "pending", DueAt: now.Add(-6 * time.Hour)},
	}
	tasks := map[uint64]types.DailyTaskInfo{
		101: {Id: 101, TopicTitle: "Topic A"},
		202: {Id: 202, TopicTitle: "Topic B"},
	}

	pending, recovery := BuildHomeReviewPresentation(items, tasks, now, lastActiveAt)
	if len(recovery) != 0 {
		t.Fatalf("expected no recovery groups in normal mode, got %d", len(recovery))
	}
	if len(pending) != 1 || pending[0].Id != 2 {
		t.Fatalf("expected nearest pending item id 2, got %#v", pending)
	}
}

func TestHomeReviewPresentationReturnsRecoveryGroupsInRecoveryMode(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -80)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -5)},
		{Id: 2, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.AddDate(0, 0, -1)},
		{Id: 3, DailyTaskId: 202, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -9)},
	}
	tasks := map[uint64]types.DailyTaskInfo{
		101: {Id: 101, TopicTitle: "Topic A"},
		202: {Id: 202, TopicTitle: "Topic B"},
	}

	pending, recovery := BuildHomeReviewPresentation(items, tasks, now, lastActiveAt)
	if len(pending) != 0 {
		t.Fatalf("expected no normal pending items in recovery mode, got %d", len(pending))
	}
	if len(recovery) != 1 {
		t.Fatalf("expected home to keep only one recovery group, got %d", len(recovery))
	}
	if recovery[0].DailyTaskId != 202 {
		t.Fatalf("expected earliest overdue recovery group to be task 202, got %d", recovery[0].DailyTaskId)
	}
}


func TestBuildReviewItemListDataKeepsOnlyNearestNormalReviewInListMode(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -2)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.Add(-8 * time.Hour)},
		{Id: 2, DailyTaskId: 202, ReviewStage: "day7", Status: "pending", DueAt: now.Add(-2 * time.Hour)},
	}
	tasks := map[uint64]types.DailyTaskInfo{
		101: {Id: 101, TopicTitle: "Topic A"},
		202: {Id: 202, TopicTitle: "Topic B"},
	}

	data := BuildReviewItemListData(items, tasks, map[uint64]types.ReviewRecordInfo{}, now, lastActiveAt)
	if len(data.List) != 1 || data.List[0].Id != 1 {
		t.Fatalf("expected only nearest due review item id 1, got %#v", data.List)
	}
	if data.PendingRemainingCount != 1 {
		t.Fatalf("expected 1 remaining pending review, got %d", data.PendingRemainingCount)
	}
}

func TestBuildReviewItemListDataKeepsOnlyEarliestRecoveryGroupInRecoveryMode(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.Local)
	lastActiveAt := now.AddDate(0, 0, -80)
	items := []model.ReviewItems{
		{Id: 1, DailyTaskId: 101, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -9)},
		{Id: 2, DailyTaskId: 101, ReviewStage: "day7", Status: "pending", DueAt: now.AddDate(0, 0, -5)},
		{Id: 3, DailyTaskId: 202, ReviewStage: "day3", Status: "pending", DueAt: now.AddDate(0, 0, -3)},
	}
	tasks := map[uint64]types.DailyTaskInfo{
		101: {Id: 101, TopicTitle: "Topic A"},
		202: {Id: 202, TopicTitle: "Topic B"},
	}

	data := BuildReviewItemListData(items, tasks, map[uint64]types.ReviewRecordInfo{}, now, lastActiveAt)
	if len(data.RecoveryGroups) != 1 || data.RecoveryGroups[0].DailyTaskId != 101 {
		t.Fatalf("expected only earliest recovery group for task 101, got %#v", data.RecoveryGroups)
	}
	if data.RecoveryRemainingCount != 1 {
		t.Fatalf("expected 1 remaining recovery group, got %d", data.RecoveryRemainingCount)
	}
}
