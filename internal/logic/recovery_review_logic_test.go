package logic

import (
	"testing"
	"time"

	"api/model"
)

func TestValidateReviewItemSubmissionRejectsFutureOrForeignItems(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	futureItem := &model.ReviewItems{Id: 1, UserId: 1, DailyTaskId: 11, ReviewStage: "day3", DueAt: now.Add(24 * time.Hour), Status: "pending"}
	foreignItem := &model.ReviewItems{Id: 2, UserId: 9, DailyTaskId: 11, ReviewStage: "day3", DueAt: now.Add(-24 * time.Hour), Status: "pending"}

	if err := validateReviewItemSubmission(futureItem, 1, now); err == nil {
		t.Fatalf("expected future review item submission to be rejected")
	}
	if err := validateReviewItemSubmission(foreignItem, 1, now); err == nil {
		t.Fatalf("expected foreign review item submission to be rejected")
	}
}

func TestValidateReviewItemSubmissionAllowsPendingOverdueOwnItem(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	item := &model.ReviewItems{Id: 1, UserId: 1, DailyTaskId: 11, ReviewStage: "day3", DueAt: now.Add(-24 * time.Hour), Status: "pending"}

	if err := validateReviewItemSubmission(item, 1, now); err != nil {
		t.Fatalf("expected overdue own pending item to pass validation, got %v", err)
	}
}

func TestValidateRecoveryReviewItemsAcceptsSameTaskOverdueItems(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	items := []*model.ReviewItems{
		{Id: 1, UserId: 1, DailyTaskId: 11, ReviewStage: "day3", DueAt: now.AddDate(0, 0, -10), Status: "pending"},
		{Id: 2, UserId: 1, DailyTaskId: 11, ReviewStage: "day7", DueAt: now.AddDate(0, 0, -6), Status: "pending"},
	}

	if err := validateRecoveryReviewItems(items, 1, now); err != nil {
		t.Fatalf("expected same-task overdue items to pass recovery validation, got %v", err)
	}
}

func TestValidateRecoveryReviewItemsRejectsMixedTaskOrFutureItems(t *testing.T) {
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.Local)
	mixedTaskItems := []*model.ReviewItems{
		{Id: 1, UserId: 1, DailyTaskId: 11, ReviewStage: "day3", DueAt: now.AddDate(0, 0, -10), Status: "pending"},
		{Id: 2, UserId: 1, DailyTaskId: 22, ReviewStage: "day7", DueAt: now.AddDate(0, 0, -6), Status: "pending"},
	}
	futureItems := []*model.ReviewItems{
		{Id: 1, UserId: 1, DailyTaskId: 11, ReviewStage: "day3", DueAt: now.AddDate(0, 0, -10), Status: "pending"},
		{Id: 3, UserId: 1, DailyTaskId: 11, ReviewStage: "day30", DueAt: now.AddDate(0, 0, 3), Status: "pending"},
	}

	if err := validateRecoveryReviewItems(mixedTaskItems, 1, now); err == nil {
		t.Fatalf("expected mixed-task recovery review to fail")
	}
	if err := validateRecoveryReviewItems(futureItems, 1, now); err == nil {
		t.Fatalf("expected future-stage recovery review to fail")
	}
}
