package logic

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"api/internal/svc"
	"api/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type recordingDailyTasksModel struct {
	model.DailyTasksModel
	seenTaskDate time.Time
	item         *model.DailyTasks
	err          error
	updatedItem  *model.DailyTasks
}

func (s *recordingDailyTasksModel) Insert(context.Context, *model.DailyTasks) (sql.Result, error) {
	panic("unexpected call")
}

func (s *recordingDailyTasksModel) FindOne(context.Context, uint64) (*model.DailyTasks, error) {
	panic("unexpected call")
}

func (s *recordingDailyTasksModel) FindOneByUserIdTaskDate(_ context.Context, _ uint64, taskDate time.Time) (*model.DailyTasks, error) {
	s.seenTaskDate = taskDate
	return s.item, s.err
}

func (s *recordingDailyTasksModel) Update(_ context.Context, item *model.DailyTasks) error {
	s.updatedItem = item
	s.item = item
	return nil
}

func (s *recordingDailyTasksModel) Delete(context.Context, uint64) error {
	panic("unexpected call")
}

func (s *recordingDailyTasksModel) withSession(sqlx.Session) model.DailyTasksModel {
	return s
}

func TestGetMyTodayTaskUsesDateOnlyLookup(t *testing.T) {
	t.Parallel()

	stub := &recordingDailyTasksModel{
		item: &model.DailyTasks{
			Id:           1,
			UserId:       1,
			TaskDate:     time.Date(2026, 4, 14, 0, 0, 0, 0, time.Local),
			TopicId:      1,
			TopicOrderNo: 1,
			TopicTitle:   "topic",
			TopicSummary: "summary",
			Status:       "draft",
			CreatedAt:    time.Date(2026, 4, 14, 10, 0, 0, 0, time.Local),
			UpdatedAt:    time.Date(2026, 4, 14, 10, 0, 0, 0, time.Local),
		},
	}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{DailyTasksModel: stub})
	_, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stub.seenTaskDate.Hour() != 0 || stub.seenTaskDate.Minute() != 0 || stub.seenTaskDate.Second() != 0 {
		t.Fatalf("expected date-only lookup, got %v", stub.seenTaskDate)
	}
}

func TestNormalizeDateClearsTimePortion(t *testing.T) {
	t.Parallel()

	input := time.Date(2026, 4, 14, 22, 45, 33, 123, time.Local)
	got := normalizeDate(input)
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("expected normalized date, got %v", got)
	}
	if got.Year() != 2026 || got.Month() != 4 || got.Day() != 14 {
		t.Fatalf("expected same date, got %v", got)
	}
}

type recordingTopicsLookupModel struct {
	model.TopicsModel
	seenDate time.Time
	item     *model.Topics
	err      error
}

func (m *recordingTopicsLookupModel) Insert(context.Context, *model.Topics) (sql.Result, error) {
	panic("unexpected call")
}
func (m *recordingTopicsLookupModel) FindOne(context.Context, uint64) (*model.Topics, error) {
	panic("unexpected call")
}
func (m *recordingTopicsLookupModel) FindOneByOrderNo(context.Context, int64) (*model.Topics, error) {
	panic("unexpected call")
}
func (m *recordingTopicsLookupModel) FindLatestActiveByScheduleDate(_ context.Context, scheduleDate time.Time) (*model.Topics, error) {
	m.seenDate = scheduleDate
	return m.item, m.err
}
func (m *recordingTopicsLookupModel) Update(context.Context, *model.Topics) error {
	panic("unexpected call")
}
func (m *recordingTopicsLookupModel) Delete(context.Context, uint64) error       { panic("unexpected call") }
func (m *recordingTopicsLookupModel) withSession(sqlx.Session) model.TopicsModel { return m }

func TestGetMyTodayTaskRefreshesEmptyDraftSnapshotWhenScheduledTopicChanged(t *testing.T) {
	t.Parallel()

	taskDate := normalizeDate(time.Now())
	description := "今天围绕这一点练，把情境、动作和判断标准都看清楚。"
	dailyTasks := &recordingDailyTasksModel{
		item: &model.DailyTasks{
			Id:           1,
			UserId:       1,
			TaskDate:     taskDate,
			TopicId:      1,
			TopicOrderNo: 1,
			TopicTitle:   "stale topic",
			TopicSummary: "stale summary",
			Status:       "draft",
			CreatedAt:    taskDate,
			UpdatedAt:    taskDate,
		},
	}
	topics := &recordingTopicsLookupModel{
		item: &model.Topics{
			Id:          19,
			Title:       "scheduled topic",
			Summary:     "scheduled summary",
			Description: sql.NullString{String: description, Valid: true},
			OrderNo:     19,
			Status:      1,
			CreatedAt:   taskDate,
			UpdatedAt:   taskDate,
		},
	}

	logic := NewGetMyTodayTaskLogic(context.Background(), &svc.ServiceContext{DailyTasksModel: dailyTasks, TopicsModel: topics})
	resp, err := logic.GetMyTodayTask()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dailyTasks.updatedItem == nil {
		t.Fatalf("expected task snapshot refresh")
	}
	if resp.Data.TopicTitle != "scheduled topic" || resp.Data.TopicOrderNo != 19 {
		t.Fatalf("expected refreshed scheduled topic, got %+v", resp.Data)
	}
	if resp.Data.TopicDescription != description {
		t.Fatalf("expected topic description %q, got %q", description, resp.Data.TopicDescription)
	}
}
